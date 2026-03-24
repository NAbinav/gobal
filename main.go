package main

import (
	"flag"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

type RoundRobin struct {
	backends []*url.URL
	counter  uint64
}

func New(items []string) *RoundRobin {
	var urls []*url.URL
	for _, item := range items {
		u, _ := url.Parse(item)
		urls = append(urls, u)
	}
	return &RoundRobin{backends: urls}
}

func (r *RoundRobin) Balance() *url.URL {
	n := atomic.AddUint64(&r.counter, 1)
	return r.backends[(n-1)%uint64(len(r.backends))]
}

var transport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     90 * time.Second,
}

func main() {
	var backends []string
	flag.Func("b", "backend", func(s string) error {
		backends = append(backends, s)
		return nil
	})
	out := flag.String("out", ":8080", "listen address")
	flag.Parse()

	if len(backends) == 0 {
		panic("at least one backend is required")
	}

	rr := New(backends)

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			target := rr.Balance()
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path + req.URL.Path
			if target.RawQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				req.Header.Set("User-Agent", "")
			}
		},
		Transport: transport,
	}

	server := &http.Server{
		Addr:    *out,
		Handler: proxy,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
