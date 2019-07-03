package security

import (
	"github.com/mulesoft-labs/gitproxy/gitproxy"
)

type UserProfile interface {
	Id() string
}

type Provider interface {
	Login(user, password string) (string, error)

	IsAuthorized(token, repository string, op gitproxy.GitOperation) bool

	FetchUserProfile(token string) (UserProfile, error)

	FetchUserPK(user string) ([][]byte, error)
}