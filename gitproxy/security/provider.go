package security

import (
	"github.com/gliderlabs/ssh"
	"github.com/mulesoft-labs/gitproxy/gitproxy"
)

type UserProfile interface {
	Id() string
}

type Provider interface {
	Login(user, password string) (string, error)

	IsAuthorized(token, repository string, op gitproxy.GitOperation) bool

	FetchUserProfile(token string) (UserProfile, error)

	ValidatePublicKey(user string, key ssh.PublicKey) bool
}