package main

import (
	"fmt"
	rl "github.com/juju/ratelimit"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"rlimiter/conf"
	"time"
)

type Buckets map[string]*rl.Bucket

var buckets Buckets = make(Buckets)

func (b Buckets) GetOrCreate(key string) (bucket *rl.Bucket) {
	bucket, ok := b[key]
	if !ok {
		bucket = rl.NewBucketWithQuantum(60*time.Second, conf.Cfg.RPM, conf.Cfg.RPM)
		b[key] = bucket
	}

	return
}

// Proxy is an HTTP handler than sets up the reverse proxy
type Proxy struct {
	target      *url.URL
	reverseProx *httputil.ReverseProxy
}

// ServeHTTP will log the request, create a rate limited transport
// and hand the flow over to the reverse proxy
func (lp *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.RequestURI
	if key == "" {
		key = "/"
	}
	waitUntil := buckets.GetOrCreate(r.RequestURI).Take(1)
	if waitUntil > 0 {
		http.Error(w, fmt.Sprintf("Too many connections, try again in %v", waitUntil), http.StatusTooManyRequests)
	} else {
		lp.reverseProx.ServeHTTP(w, r)
	}
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

	// Custom dialer and custom Transport to fix this random context cancelled I've been
	// receiving. Seems I'm not the only one, initially I tried implementing this using
	// DialContext but it still failed. Seems like a scope issue with context being
	// cancelled too early? Or this:
	// https://groups.google.com/forum/#!msg/golang-nuts/oiBBZfUb2hM/9S_JB6g2EAAJ
	// Either way, it's quite annoying, so I'm keeping that for now.
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
		Dial:                dialer.Dial,
	}

	return proxy
}

func main() {
	err := conf.CommandLineArgs()
	if err != nil {
		fmt.Println(err)
		conf.DisplayUsage()
		os.Exit(1)
	}

	proxy := NewProxy(conf.Cfg.Target)
	server := &http.Server{
		Handler: proxy,
		Addr:    conf.Cfg.Addr(),
	}

	log.Printf("Starting rlimiter... binding to: %s", conf.Cfg.Addr())
	log.Printf("Limiting access to %s by %v reqs/m", conf.Cfg.Target, conf.Cfg.RPM)
	log.Fatal(server.ListenAndServe())
}
