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
	"strconv"
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

// We add limit information to X-R-Limit-Limit header. Get an existing bucket
// or create one based on the request path, then we take a token out of it.
// If we have enough tokens, we'll pass this over to the reverse proxy otherwise
// we'll give back a 429 http status and set the X-R-Limit-Wait header to the
// time the client has to wait before making a new request.
func (lp *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-R-Limit-Limit", strconv.Itoa(int(conf.Cfg.RPM)))

	key := r.RequestURI
	if key == "" {
		key = "/"
	}

	waitUntil := buckets.GetOrCreate(r.RequestURI).Take(1)
	if waitUntil > 0 {
		w.Header().Set("X-R-Limit-Wait", waitUntil.String())
		http.Error(w, "Too many connections, try again later", http.StatusTooManyRequests)
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
