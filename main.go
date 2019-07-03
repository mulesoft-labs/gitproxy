package main

import (
	"github.com/mulesoft-labs/gitproxy/gitproxy/security/authserver"
	"github.com/mulesoft-labs/gitproxy/gitproxy/transport/http"
	"github.com/mulesoft-labs/gitproxy/gitproxy/transport/ssh"
	"log"
	"os"
	"os/signal"
)

func main() {

	baseUrl := getEnv("AUTHSERVER_URL", "https://devx.anypoint.mulesoft.com/accounts")
	authenticationServerProvider := authserver.NewAuthenticationServerProvider(baseUrl)

	config := http.Config{
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

	httpTransport, err := http.NewHttpTransport(config, authenticationServerProvider)
	if err != nil {
		log.Panic(err.Error())
	}
	httpTransport.Serve()


	sshConfig := ssh.Config{
		Addr: ":2222",
		RemoteAddr: "github.com:22",
		RemoteHostKey: "github.key",
		HostKeyFile: "key.pem",
		Accounts: []ssh.Account{
			{
				User: "patricio78",
				PrivateKeyFile: "key.pem",
			},
		},
	}

	sshTransport, err := ssh.NewSSHTransport(sshConfig, authenticationServerProvider)
	if err != nil {
		log.Panic(err.Error())
	}
	sshTransport.Serve()

	waitForCtrlC()
}

func waitForCtrlC() {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

