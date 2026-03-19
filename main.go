package main

import (
	"flag"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type RoundRobin struct {
	proxies []*httputil.ReverseProxy
	index   int
	mu      sync.Mutex
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
	r.mu.Lock()
	proxy := r.proxies[r.index]
	r.index = (r.index + 1) % len(r.proxies)
	r.mu.Unlock()
	return proxy
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

	http.ListenAndServe(*out, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr.Balance().ServeHTTP(w, r)
	}))

}
