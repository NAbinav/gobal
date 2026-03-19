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
	proxies []*httputil.ReverseProxy
	counter uint64
}

func New(items []string) *RoundRobin {
	proxies := make([]*httputil.ReverseProxy, len(items))
	for i, backend := range items {
		target, _ := url.Parse(backend)
		proxies[i] = httputil.NewSingleHostReverseProxy(target)
	}
	return &RoundRobin{proxies: proxies}
}

func (r *RoundRobin) Balance() *httputil.ReverseProxy {
	n := atomic.AddUint64(&r.counter, 1)
	return r.proxies[n%uint64(len(r.proxies))]
}

func main() {
	var backends []string
	flag.Func("b", "backend", func(s string) error {
		backends = append(backends, s)
		return nil
	})
	out := flag.String("out", ":8080", "listen address")
	flag.Parse()

	rr := New(backends)
	server := &http.Server{
		Addr: *out,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rr.Balance().ServeHTTP(w, r)
		}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	server.ListenAndServe()

}
