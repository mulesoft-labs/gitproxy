package http

import (
	"container/ring"
	"github.com/mulesoft-labs/gitproxy/gitproxy"
	"github.com/mulesoft-labs/gitproxy/gitproxy/config"
	"github.com/mulesoft-labs/gitproxy/gitproxy/security"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	HttpsScheme string = "https"
	GitServiceName string = "service"
)

type Transport struct {
	HttpConfig config.HttpConfig

	origin *url.URL

	accounts *ring.Ring
	provider security.Provider

}

func NewHttpTransport(config config.HttpConfig, provider security.Provider) (*Transport, error) {

	httpTransport := &Transport{
		HttpConfig: config,
		provider: provider,
	}

	var err error

	origin, err := url.Parse(httpTransport.HttpConfig.RemoteAddr)
	if err != nil {
		return nil, err
	}

	httpTransport.origin = origin

	httpTransport.accounts = ring.New(len(config.Accounts))
	accounts := httpTransport.accounts
	for i := 0; i < accounts.Len(); i++ {
		accounts = accounts.Next()
		accounts.Value = i
	}

	return httpTransport, nil
}

func (t *Transport) nextAccount() config.HttpAccount {
	nextVal := t.accounts.Next().Value.(int)
	log.Printf("[DEBUG] HTTP: Accessing account #%d\n", nextVal)
	return t.HttpConfig.Accounts[nextVal]
}

func (t *Transport) Serve()  {

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", t.origin.Host)
		req.URL.Host = t.origin.Host
		req.URL.Scheme = HttpsScheme
		req.Host = t.origin.Host

		account := t.nextAccount()
		req.SetBasicAuth(account.User, account.Password)
	}
	var DefaultTransport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSHandshakeTimeout: gitproxy.ConnectTimeout * time.Second,
		IdleConnTimeout: gitproxy.IdleConnectionTimeout * time.Second,
	}
	proxy := &httputil.ReverseProxy{
		Director: director,
		Transport: DefaultTransport,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var token string
		var err error = nil
		service := r.URL.Query().Get(GitServiceName)
		repo := r.URL.Path
		log.Printf("[DEBUG] SSH: Serving %s/%s", service, repo)
		user, password, ok := r.BasicAuth()
		if ok {
			token, err = t.provider.Login(user, password)
		} else {
			authorizationHeader := r.Header.Get("Authorization")
			if len(authorizationHeader) > 0 {
				token = strings.Split(authorizationHeader, "Bearer")[1]
			}
		}

		if err == nil && len(token) > 0 && t.provider.IsAuthorized(token, repo, gitproxy.GetOperation(service)) {
			proxy.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Bearer")
			w.Header().Add("WWW-Authenticate", "Basic")
			w.WriteHeader(401)
		}
	})

	log.Printf("[INFO] HTTP: Serving TLS on %s", t.HttpConfig.Addr)
	go func() {
		err := http.ListenAndServeTLS(t.HttpConfig.Addr, t.HttpConfig.CertFile, t.HttpConfig.KeyFile, nil)
		if err != nil {
			log.Panic(err)
		}
	}()
}
