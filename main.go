package main

import (
	"github.com/hashicorp/logutils"
	"github.com/mulesoft-labs/gitproxy/gitproxy/config"
	"github.com/mulesoft-labs/gitproxy/gitproxy/security"
	"github.com/mulesoft-labs/gitproxy/gitproxy/security/authserver"
	"github.com/mulesoft-labs/gitproxy/gitproxy/security/mock"
	"github.com/mulesoft-labs/gitproxy/gitproxy/transport/http"
	"github.com/mulesoft-labs/gitproxy/gitproxy/transport/ssh"
	"log"
	"os"
	"os/signal"
)

func main() {

	configuration := config.NewConfig()

	initLog(configuration.Log)

	var provider security.Provider
	if configuration.AuthServer.Mock {
		log.Println("[DEBUG] !!!!!! Running with MOCK Authentication Server !!!!!!!!")
		provider = mock.NewMockAuthServerProvider()
	} else {
		provider = authserver.NewAuthenticationServerProvider(configuration.AuthServer)
	}

	if configuration.HttpConfig.Enabled {
		httpTransport, err := http.NewHttpTransport(configuration.HttpConfig, provider)
		if err != nil {
			log.Panic(err.Error())
		}
		httpTransport.Serve()
	}

	if configuration.SshConfig.Enabled {
		sshTransport, err := ssh.NewSSHTransport(configuration.SshConfig, provider)
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

func initLog(config config.Log) {
	filter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(config.MinLevel),
		Writer: os.Stdout,
	}
	log.SetOutput(filter)
}