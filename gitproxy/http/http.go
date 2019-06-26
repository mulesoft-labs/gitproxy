package http

import (
	"container/ring"
	"github.com/mulesoft-labs/git-proxy/gitproxy"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	HttpsScheme string = "https"
	GitServiceName string = "service"
)

type Account struct {
	User string
	Password string
}

type HttpConfig struct {
	Addr string
	CertFile string
	KeyFile string
	RemoteAddr string
	Accounts []Account
}

type HttpTransport struct {
	HttpConfig HttpConfig

	origin *url.URL

	accounts *ring.Ring

}

func NewHttpTransport(config *HttpConfig) (*HttpTransport, error) {

	httpTransport := &HttpTransport{
		HttpConfig: *config,
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

func (t *HttpTransport) nextAccount() Account {
	nextVal := t.accounts.Next().Value.(int)
	log.Printf("[DEBUG] HTTP: Accessing account #%d\n", nextVal)
	return t.HttpConfig.Accounts[nextVal]
}

func (t *HttpTransport) Serve()  {

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
		service := r.URL.Query().Get(GitServiceName)
		repo := r.URL.Path
		user, password, _ := r.BasicAuth()
		log.Printf("[DEBUG] SSH: Serving %s for %s/%s", service, user, repo)
		token, err := gitproxy.Login(user, password)

		if err == nil && gitproxy.IsAuthorized(token, repo, gitproxy.GetOperation(service)) {
			proxy.ServeHTTP(w, r)
		} else {
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
