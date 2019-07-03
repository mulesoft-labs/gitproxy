package authserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mulesoft-labs/gitproxy/gitproxy"
	"github.com/mulesoft-labs/gitproxy/gitproxy/security"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	Namespace string = "vcs"
)


type loginResponse struct {
	AccessToken string `json:"access_token"`
}

type resource string

type resourceRequest struct {
	Namespace string     `json:"namespace"`
	Action string        `json:"action"`
	Resources []resource `json:"resources"`
}

type isAuthorizedResponse struct {
	Resources []resource
}

type UserProfile struct {
	User string `json:"user_id"`
}

type AuthenticationServerProvider struct {
	client http.Client
	baseUrl string
}

func (u *UserProfile) Id() string {
	return u.User
}

func NewAuthenticationServerProvider(baseUrl string) *AuthenticationServerProvider {
	return &AuthenticationServerProvider{
		client: http.Client{
			Timeout: time.Second * 2,
		},
		baseUrl: baseUrl + "%s",
	}
}

func (p *AuthenticationServerProvider) Login(user string, password string) (string, error) {

	body := make(map[string]string)
	body["username"] = user
	body["password"] = password

	loginUrl := p.buildUrl("/login")

	loginResponse := &loginResponse{}
	err := p.post(nil, loginUrl, body, loginResponse)
	if err != nil {
		return "", err
	}

	return (*loginResponse).AccessToken, nil
}

func (p *AuthenticationServerProvider) IsAuthorized(token, repository string, op gitproxy.GitOperation) bool {

	body := resourceRequest{
		Namespace: Namespace,
		Action: gitproxy.GetAction(op),
		Resources: []resource{
			resource(repository),
		},
	}

	authorizeUrl := p.buildUrl("/api/authorize")

	isAuthorizedResponse := &isAuthorizedResponse{}
	err := p.post(&token, authorizeUrl, body, isAuthorizedResponse)
	if err != nil {
		return false
	}

	return len(isAuthorizedResponse.Resources) > 0
}

func (p *AuthenticationServerProvider) FetchUserProfile(token string) (security.UserProfile, error) {

	userProfileUrl := p.buildUrl("/api/access_token/me")

	userProfile := &UserProfile{}
	err := p.post(&token, userProfileUrl, nil, userProfile)
	if err != nil {
		return nil, err
	}

	return userProfile, nil
}

func (p *AuthenticationServerProvider) FetchUserPK(user string) ([][]byte, error) {
	return nil, nil
}

func (p *AuthenticationServerProvider) post(token *string, reqUrl string, body interface{}, response interface{}) error {

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != nil {
		req.Header.Set("Authorization", "Bearer " + *token)
	}

	res, err := p.client.Do(req)
	if err != nil {
		return err
	}

	bodyResp, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyResp, response)
	if err != nil {
		return err
	}

	return nil
}

func (p *AuthenticationServerProvider) buildUrl(suffix string) string {
	return fmt.Sprintf(p.baseUrl, suffix)
}
