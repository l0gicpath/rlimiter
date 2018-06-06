package proxy

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"rlimiter/buckets"
	"rlimiter/conf"
	"strconv"
	"time"
)

// Proxy is an HTTP handler than sets up the reverse proxy
type Proxy struct {
	target      *url.URL
	reverseProx *httputil.ReverseProxy
}

func (lp *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-R-Limit-Limit", strconv.Itoa(int(conf.Cfg.RPM)))

	key := r.RequestURI
	if key == "" {
		key = "/"
	}

	waitUntil := buckets.PathBuckets.GetOrCreate(r.RequestURI).Take(1)
	if waitUntil > 0 {
		w.Header().Set("X-R-Limit-Wait", waitUntil.String())
		http.Error(w, "Too many connections, try again later", http.StatusTooManyRequests)
	} else {
		lp.reverseProx.ServeHTTP(w, r)
	}
}

// newProxy accepts a target to proxy to and returns a single host reverse proxy
// instance that redirects to the target
func newProxy(target string) *Proxy {
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

// RateLimitingServer holds two pieces, the HTTP server component that accepts HTTP requests
type RateLimitingServer struct {
	HTTP *http.Server
}

func New(target, addr string) *RateLimitingServer {
	prox := newProxy(target)
	serv := &http.Server{
		Handler: prox,
		Addr:    addr,
	}

	return &RateLimitingServer{serv}
}
