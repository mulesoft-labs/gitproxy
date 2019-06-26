package gitproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	Namespace string = "vcs"
)

var baseUrl string

var client = http.Client{
	Timeout: time.Second * 2,
}

func init() {
	baseUrl = getEnv("AUTHSERVER_URL", "https://devx.anypoint.mulesoft.com/accounts") + "%s"
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type Resource string

type ResourceRequest struct {
	Namespace string `json:"namespace"`
	Action string `json:"action"`
	Resources []Resource `json:"resources"`
}

type IsAuthorizedResponse struct {
	Resources []Resource
}

type UserProfile struct {
	User string `json:"user_id"`
}

func Login(user string, password string) (string, error) {

	body := make(map[string]string)
	body["username"] = user
	body["password"] = password

	loginUrl := buildUrl("/login")

	loginResponse := &LoginResponse{}
	err := post(nil, loginUrl, body, loginResponse)
	if err != nil {
		return "", err
	}

	return (*loginResponse).AccessToken, nil
}

func IsAuthorized(token, repository string, op GitOperation) bool {

	body := ResourceRequest{
		Namespace: Namespace,
		Action: getAction(op),
		Resources: []Resource {
			Resource(repository),
		},
	}

	authorizeUrl := buildUrl("/api/authorize")

	isAuthorizedResponse := &IsAuthorizedResponse{}
	err := post(&token, authorizeUrl, body, isAuthorizedResponse)
	if err != nil {
		return false
	}

	return len(isAuthorizedResponse.Resources) > 0
}

func FetchUserProfile(token string) (*UserProfile, error) {

	userProfileUrl := buildUrl("/api/access_token/me")

	userProfile := &UserProfile{}
	err := post(&token, userProfileUrl, nil, userProfile)
	if err != nil {
		return nil, err
	}

	return userProfile, nil
}

func post(token *string, reqUrl string, body interface{}, response interface{}) error {

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

	res, err := client.Do(req)
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


func FetchUserPK(user string) ([]byte, error) {
	pkDir := getEnv("PK_DIR", "/tmp")
	return ioutil.ReadFile(fmt.Sprintf("%s/%s.pub", pkDir, user))
}

func getAction(op GitOperation) string {
	if op == GitRead {
		return "GET"
	} else {
		return "POST"
	}
}


func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func buildUrl(suffix string) string {
	return fmt.Sprintf(baseUrl, suffix)
}