/*
Proxy package contains the implementation of a catch-all HTTP handler that is rate limited.
The handler implementation acts as a gatekeeper to the underlaying protected service which
is in concept represented by the SingleHostReverseProxy instance.

If the limit is not exceeded then the request is passed to the reverse proxy but if the
limit is exceeded the request is dropped with a 429 HTTP status code.

Headers are set for each limit-exceeding request:
	- X-R-Limit-Wait		This gives the calling client an indicator to how long it should wait before a re-try

Headers that are set for each response:
	- X-R-Limit-Limit		This informs the client of how big is the current limit
*/
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

// Proxy is an HTTP handler than sets up the reverse proxy and catches-all HTTP traffic
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

// RateLimitingServer holds the HTTP server, the abstraction is in place because later
// it will encapsulate the logic for making the service a daemon and for gracefull
// shutdown.
type RateLimitingServer struct {
	HTTP *http.Server
}

// Create a new RateLimitingService accepting the target to be protected and the
// address on which the HTTP server would bind to.
func New(target, addr string) *RateLimitingServer {
	prox := newProxy(target)
	serv := &http.Server{
		Handler: prox,
		Addr:    addr,
	}

	return &RateLimitingServer{serv}
}
