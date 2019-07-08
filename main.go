package main

import (
	"github.com/mulesoft-labs/gitproxy/gitproxy/config"
	"github.com/mulesoft-labs/gitproxy/gitproxy/security/authserver"
	"github.com/mulesoft-labs/gitproxy/gitproxy/transport/http"
	"github.com/mulesoft-labs/gitproxy/gitproxy/transport/ssh"
	"log"
	"os"
	"os/signal"
)

func main() {

	configuration := config.NewConfig()

	authenticationServerProvider := authserver.NewAuthenticationServerProvider(configuration.AuthServer)

	if configuration.HttpConfig.Enabled {
		httpTransport, err := http.NewHttpTransport(configuration.HttpConfig, authenticationServerProvider)
		if err != nil {
			log.Panic(err.Error())
		}
		httpTransport.Serve()
	}

	if configuration.SshConfig.Enabled {
		sshTransport, err := ssh.NewSSHTransport(configuration.SshConfig, authenticationServerProvider)
		if err != nil {
			log.Panic(err.Error())
		}
		sshTransport.Serve()
	}

	waitForCtrlC()
}

func waitForCtrlC() {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
}


