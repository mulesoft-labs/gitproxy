package mock

import (
	"github.com/gliderlabs/ssh"
	"github.com/mulesoft-labs/gitproxy/gitproxy"
	"github.com/mulesoft-labs/gitproxy/gitproxy/security"
)

type AuthServerProvider struct {}

func NewMockAuthServerProvider() *AuthServerProvider {
	return &AuthServerProvider{}
}

func (p *AuthServerProvider) Login(user string, password string) (string, error) {
	return user, nil
}

func (p *AuthServerProvider) IsAuthorized(token, repository string, op gitproxy.GitOperation) bool {

	return true
}

func (p *AuthServerProvider) FetchUserProfile(token string) (security.UserProfile, error) {

	return p, nil
}

func (p *AuthServerProvider) Id() string {
	return ""
}

func (p *AuthServerProvider) ValidatePublicKey(user string, key ssh.PublicKey) bool {
	return true
}