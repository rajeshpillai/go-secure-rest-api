package main

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"

	"golang.org/x/time/rate"
)

// limiter gives every client IP its own token bucket.
type limiter struct {
	mu      sync.Mutex
	clients map[string]*rate.Limiter
	every   rate.Limit // tokens added per second
	burst   int        // bucket size
}

func newLimiter(every rate.Limit, burst int) *limiter {
	return &limiter{clients: map[string]*rate.Limiter{}, every: every, burst: burst}
}

func (l *limiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.clients[ip]
	if !ok {
		b = rate.NewLimiter(l.every, l.burst)
		l.clients[ip] = b
	}
	return b.Allow()
}

func (l *limiter) wrap(next http.Handler) http.Handler {
	wait := strconv.Itoa(int(math.Ceil(1 / float64(l.every)))) // seconds per token
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !l.allow(ip) {
			w.Header().Set("Retry-After", wait)
			writeError(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}
