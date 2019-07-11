package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AuthServer AuthenticationServerConfig
	HttpConfig HttpConfig
	SshConfig SshConfig
	Log Log
}

type Log struct {
	MinLevel string
}

type AuthenticationServerConfig struct {
	BaseUrl string
	Mock bool
}

type HttpAccount struct {
	User string
	Password string
}

type HttpConfig struct {
	Addr string
	CertFile string
	KeyFile string
	RemoteAddr string
	Accounts []HttpAccount
	Enabled bool
}

type SshAccount struct {
	User string
	PrivateKeyFile string

}

type SshConfig struct {
	Addr string
	HostKeyFile string
	RemoteAddr string
	RemoteHostKey string
	Accounts []SshAccount
	Enabled bool
}

func NewConfig() Config {
	return Config{
		Log: Log {
			MinLevel: getEnv("LOG_MINLEVEL", "WARN"),
		},
		AuthServer: AuthenticationServerConfig{
			BaseUrl: getEnv("AUTHSERVER_URL", ""),
			Mock: boolVal(getEnv("AUTHSERVER_MOCK", "false")),
		},
		HttpConfig:buildHttpConfig(),
		SshConfig:buildSshConfig(),
	}
}
func buildHttpConfig() HttpConfig {

	enabled := boolVal(getEnv("HTTPCONFIG_ENABLED", "false"))

	if enabled {
		return HttpConfig{
			Addr:getEnv("HTTPCONFIG_ADDR", ":443"),
			KeyFile:getEnv("HTTPCONFIG_KEYFILE", ""),
			CertFile:getEnv("HTTPCONFIG_CERTFILE", ""),
			RemoteAddr:getEnv("HTTPCONFIG_REMOTEADDR", ""),
			Accounts: buildHttpAccounts(getEnv("HTTPCONFIG_ACCOUNTS", "")),
			Enabled: true,
		}
	} else {
		return HttpConfig{
			Enabled: false,
		}
	}
}

func buildSshConfig() SshConfig {

	enabled := boolVal(getEnv("SSHCONFIG_ENABLED", "true"))

	if enabled {
		return SshConfig{
			Addr:          getEnv("SSHCONFIG_ADDR", ":22"),
			HostKeyFile:   getEnv("SSHCONFIG_HOSTKEYFILE", ""),
			RemoteAddr:    getEnv("SSHCONFIG_REMOTEADDR", ""),
			RemoteHostKey: getEnv("SSHCONFIG_REMOTEHOSTKEY", ""),
			Accounts:      buildSshAccounts(getEnv("SSHCONFIG_ACCOUNTS", "")),
			Enabled:       true,
		}
	} else {
		return SshConfig{
			Enabled: false,
		}
	}
}

func boolVal(val string) bool {
	b, err := strconv.ParseBool(val)
	if err != nil {
		log.Panic(err.Error())
	}
	return b
}

func buildHttpAccounts(accounts string) []HttpAccount {
	split := strings.Split(accounts, ",")
	httpAccounts := make([]HttpAccount, len(split))
	for i,val := range split {
		account := strings.Split(val, ":")
		if len(account) != 2 {
			log.Panicf("Invalid http account")
		}
		httpAccounts[i] = HttpAccount{
			User: account[0],
			Password: account[1],
		}
	}
	return httpAccounts
}

func buildSshAccounts(accounts string) []SshAccount {
	split := strings.Split(accounts, ",")
	httpAccounts := make([]SshAccount, len(split))
	for i,val := range split {
		account := strings.Split(val, ":")
		if len(account) != 2 {
			log.Panicf("Invalid ssh account")
		}
		httpAccounts[i] = SshAccount{
			User: account[0],
			PrivateKeyFile: account[1],
		}
	}
	return httpAccounts
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		if len(fallback) > 0 {
			return fallback
		} else {
			log.Panicf("Missing configuration %s", key)
		}
	}
	return value
}
