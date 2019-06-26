package main

import (
	"github.com/mulesoft-labs/git-proxy/gitproxy"
	"github.com/mulesoft-labs/git-proxy/gitproxy/http"
	"github.com/mulesoft-labs/git-proxy/gitproxy/ssh"
	"log"
)

func main() {

	config := &http.HttpConfig{
		Addr: ":443",
		RemoteAddr: "https://github.com",
		CertFile: "cert.pem",
		KeyFile: "key.pem",
		Accounts: []http.Account{
			{
				User:     "patricio78",
				Password: "",
			},
		},
	}

	httpTransport, err := http.NewHttpTransport(config)
	if err != nil {
		log.Panic(err.Error())
	}
	httpTransport.Serve()


	sshConfig := &ssh.SshConfig{
		Addr: ":2222",
		RemoteAddr: "github.com:22",
		RemoteHostKey: "github.key",
		HostKeyFile: "/etc/ssh/git_proxy",
		Accounts: []ssh.Account {
			{
				User: "patricio78",
				PrivateKeyFile: "/etc/ssh/git_proxy",
			},
		},
	}

	sshTransport, err := ssh.NewSSHTransport(sshConfig)
	if err != nil {
		log.Panic(err.Error())
	}
	sshTransport.Serve()

	gitproxy.WaitForCtrlC()
}
