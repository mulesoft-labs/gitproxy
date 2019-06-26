package gitproxy

import (
	"github.com/hashicorp/logutils"
	"log"
	"os"
	"os/signal"
	"time"
)

func init() {
	filter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("DEBUG"),
		Writer: os.Stderr,
	}
	log.SetOutput(filter)
}

type GitOperation int

const (
	GitRead           GitOperation  = 0
	GitWrite          GitOperation  = 1
	ConnectTimeout time.Duration = 2
	IdleConnectionTimeout time.Duration = 2

	GitUploadPack string = "git-upload-pack"
	GitReceivePack string = "git-receive-pack"
)

type Transport interface {
	Serve()
}

func GetOperation(service string) GitOperation {
	if service == "git-upload-pack" {
		return GitRead
	} else {
		return GitWrite
	}
}

func WaitForCtrlC() {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
}
