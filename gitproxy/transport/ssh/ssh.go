package ssh

import (
	"container/ring"
	sshserver "github.com/gliderlabs/ssh"
	"github.com/mulesoft-labs/gitproxy/gitproxy"
	"github.com/mulesoft-labs/gitproxy/gitproxy/config"
	"github.com/mulesoft-labs/gitproxy/gitproxy/security"
	sshclient "golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type sshAccount struct {
	user       string
	privateKey sshserver.Signer
}

type Transport struct {
	SshConfig config.SshConfig

	accounts *ring.Ring
	remoteHostKey sshserver.PublicKey
	provider security.Provider
}

func parsePrivateKey(privateKeyFile string) sshserver.Signer {
	key, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}
	signer, err := sshclient.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}
	return signer
}

func parsePublicKey(publicKeyFile string) sshserver.PublicKey {
	key, err := ioutil.ReadFile(publicKeyFile)
	if err != nil {
		log.Fatalf("unable to read public key: %v", err)
	}
	signer, err := sshclient.ParsePublicKey(key)
	if err != nil {
		log.Fatalf("unable to read public key: %v", err)
	}
	return signer
}

func (t *Transport) nextAccount() sshAccount {
	nextVal := t.accounts.Next().Value.(sshAccount)
	log.Printf("[DEBUG] SSH: Accessing account %v\n", nextVal)
	return nextVal
}


func NewSSHTransport(config config.SshConfig, provider security.Provider) (*Transport, error) {

	sshTransport := &Transport{
		SshConfig: config,
		remoteHostKey: parsePublicKey(config.RemoteHostKey),
		provider: provider,
	}

	sshTransport.accounts = ring.New(len(config.Accounts))
	accounts := sshTransport.accounts
	for i := 0; i < accounts.Len(); i++ {
		accounts = accounts.Next()
		accounts.Value = sshAccount{
			user:       config.Accounts[i].User,
			privateKey: parsePrivateKey(config.Accounts[i].PrivateKeyFile),
		}

	}

	return sshTransport, nil
}

func (t *Transport) newDownstreamClient() (*sshclient.Client, error) {
	account := t.nextAccount()
	config := &sshclient.ClientConfig{
		User: account.user,
		Auth: []sshclient.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			sshclient.PublicKeys(account.privateKey),
		},
		HostKeyCallback: sshclient.FixedHostKey(t.remoteHostKey),
		Timeout: time.Second * gitproxy.ConnectTimeout,
	}
	return sshclient.Dial("tcp", t.SshConfig.RemoteAddr, config)
}

func (t *Transport) exchange(downStream *sshclient.Session, upstream *sshserver.Session, cmd string) int {
	var exitCode int
	log.Printf("[DEBUG] SSH: Command to send: %s", cmd)

	reader,err := downStream.StdoutPipe()
	if err != nil {
		log.Panicf("unable to create stdout: %v", err)
	}
	writer,err := downStream.StdinPipe()
	if err != nil {
		log.Panicf("unable to create stdin: %v", err)
	}
	err = downStream.Start(cmd)
	if err != nil {
		log.Panicf("unable to run command: %v", err)
	}

	eof := make(chan bool, 2)

	copyConn := func(dst io.Writer, src io.Reader) {
		n, err := io.Copy(dst, src)
		log.Printf("[DEBUG] Copying %d bytes", n)
		if err != nil {
			log.Printf("[DEBUG] SSH: unable to copy: %s", err)
		}
		eof <- true
	}

	go copyConn(*upstream, reader)
	go copyConn(writer, *upstream)

	<- eof
	<- eof
	err = downStream.Wait()

	if err == nil {
		exitCode = 0
	} else if exitError,ok := err.(*sshclient.ExitError); ok {
		exitCode = exitError.ExitStatus()
	} else if _, ok := err.(*sshclient.ExitMissingError); ok {
		exitCode = -1
	}

	return exitCode
}

func (t *Transport) Serve() {

	hostKeyFile := sshserver.HostKeyFile(t.SshConfig.HostKeyFile)
	publicKeyOption := sshserver.PublicKeyAuth(func(ctx sshserver.Context, key sshserver.PublicKey) bool {
		return t.provider.ValidatePublicKey(ctx.User(), key)
	})

	handler := func(upstream sshserver.Session) {
		var err error
		exitCode := 1

		defer func() {
			if err := recover(); err != nil{
				log.Printf("[ERROR] %v", err)
				upstream.Exit(1)
			}
		}()

		if isGitShell(upstream.Command()) {

			service := upstream.Command()[0]
			repo := upstream.Command()[1]
			user := upstream.User()
			log.Printf("[DEBUG] Serving %s for %s@%s", service, user, repo)
			if t.provider.IsAuthorized(user, repo, gitproxy.GetOperation(service)) {

				downstreamClient, err := t.newDownstreamClient()
				if err != nil {
					log.Panicf("unable to connect: %v", err)
				}
				defer downstreamClient.Close()

				downstream, err := downstreamClient.NewSession()
				if err != nil {
					log.Panicf("unable to create downstream: %v", err)
				}
				defer downstream.Close()

				exitCode = t.exchange(downstream, &upstream, strings.Join(upstream.Command(), " "))

			}
		}

		err = upstream.Exit(exitCode)
		if err != nil {
			log.Panicf("unable to send exit code: %v", err)
		}

	}

	go func() {
		log.Printf("[INFO] SSH: Serving on %s", t.SshConfig.Addr)
		err := sshserver.ListenAndServe(t.SshConfig.Addr, handler, publicKeyOption, hostKeyFile)
		if err != nil {
			log.Panic(err)
		}
	}()
}

func isGitShell(cmd []string) bool {
	return len(cmd) == 2 && isGitCommand(cmd[0])
}

func isGitCommand(cmd string) bool {
	return cmd == gitproxy.GitReceivePack || cmd == gitproxy.GitUploadPack
}

