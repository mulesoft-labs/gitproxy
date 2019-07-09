package gitproxy

import (
	"time"
)

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

func GetAction(op GitOperation) string {
	if op == GitRead {
		return "GET"
	} else {
		return "POST"
	}
}

