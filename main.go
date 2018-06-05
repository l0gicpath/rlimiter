package main

import (
	"flag"
	"fmt"
	rl "github.com/juju/ratelimit"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

type Buckets map[string]*rl.Bucket

var buckets Buckets = make(Buckets)

func (b Buckets) GetOrCreate(key string) (bucket *rl.Bucket) {
	bucket, ok := b[key]
	if !ok {
		bucket = rl.NewBucketWithQuantum(60*time.Second, config.ratePerMinute, config.ratePerMinute)
		b[key] = bucket
	}

	return
}

// Config carries command line configurations
type Config struct {
	port          string
	target        string
	ip            string
	ratePerMinute int64
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

const (
	ipUsageHelp            = "IP address for rlimiter to bind to"
	portUsageHelp          = "Port number for rlimiter to listen on"
	targetUsageHelp        = "Address of server we are protecting"
	ratePerMinuteUsageHelp = "Rate of connections allowed per minute"
)

var config *Config
var defaultConfig *Config = &Config{
	port:          "2400",
	target:        "http://0.0.0.0:80",
	ip:            "127.0.0.1",
	ratePerMinute: 100,
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s -target=TARGET [other options]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	config = &Config{}

	flag.StringVar(&config.port, "port", defaultConfig.port, portUsageHelp)
	flag.StringVar(&config.target, "target", defaultConfig.target, targetUsageHelp)
	flag.StringVar(&config.ip, "ip", defaultConfig.ip, ipUsageHelp)
	flag.Int64Var(&config.ratePerMinute, "rpm", defaultConfig.ratePerMinute, ratePerMinuteUsageHelp)

	flag.Usage = usage

	flag.Parse()

	proxy := NewProxy(config.target)
	server := &http.Server{
		Handler: proxy,
		Addr:    config.ip + ":" + config.port,
	}

	log.Printf("Starting http rate limiter on port %s:%s", config.ip, config.port)
	log.Printf("Limiting access to %s by %v reqs/m", config.target, config.ratePerMinute)
	log.Fatal(server.ListenAndServe())
}
