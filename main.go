package main

import (
	"flag"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type RoundRobin struct {
	items []string
	index int
	mu    sync.Mutex
}

func New(items []string) *RoundRobin {
	return &RoundRobin{
		items: items,
		index: 0,
	}
}

func (r *RoundRobin) Balance() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.items) == 0 {
		return ""
	}
	r.index = (r.index + 1) % len(r.items)
	return r.items[r.index]
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
		target, _ := url.Parse(rr.Balance())
		httputil.NewSingleHostReverseProxy(target).ServeHTTP(w, r)
	}))

}
