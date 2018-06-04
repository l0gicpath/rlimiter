package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	defaultPort     = "2400"
	portUsageHelp   = "rate limiter reverse proxy port"
	defaultTarget   = "http://0.0.0.0:80"
	targetUsageHelp = "server to limit access to"
)

// Proxy is an HTTP handler than sets up the reverse proxy
type Proxy struct {
	target      *url.URL
	reverseProx *httputil.ReverseProxy
}

// ServeHTTP will log the request, create a rate limited transport
// and hand the flow over to the reverse proxy
func (lp *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-R-Limit-Proxy", "Rlimit-0.1.0")
	lp.reverseProx.ServeHTTP(w, r)
}

// NewProxy accepts a target string, that should be a valid URL that
// represents the http server/load-balancer/api-server we are setting
// a reverse proxy for.
//
// It will return a *Proxy instance with the parsed target and a
// SingleHostReverseProxy.
func NewProxy(target string) *Proxy {
	url, err := url.Parse(target)
	if err != nil {
		log.Panic(err)
	}

	proxy := &Proxy{
		target:      url,
		reverseProx: httputil.NewSingleHostReverseProxy(url),
	}

	dialer := net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	proxy.reverseProx.Transport = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        240,
		IdleConnTimeout:     120 * time.Second,
		TLSHandshakeTimeout: 30 * time.Second,
		DialContext:         dialer.DialContext,
	}

	return proxy
}

func main() {
	port := flag.String("port", defaultPort, portUsageHelp)
	target := flag.String("target", defaultTarget, targetUsageHelp)
	flag.Parse()

	proxy := NewProxy(*target)
	server := &http.Server{
		Handler: proxy,
		Addr:    "0.0.0.0:" + *port,
	}

	log.Printf("Starting http rate limiter on port 0.0.0.0:%s", *port)
	log.Printf("Limiting access to %s", *target)
	log.Fatal(server.ListenAndServe())
}
