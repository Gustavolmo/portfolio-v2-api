package email

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	limit    rate.Limit
	burst    int
}

func NewRateLimiter(limit rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		limit:    limit,
		burst:    burst,
	}
}

func (r *RateLimiter) Allow(ip string) bool {
	r.mu.Lock()

	limiter, exists := r.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(r.limit, r.burst)
		r.visitors[ip] = limiter
	}

	r.mu.Unlock()

	return limiter.Allow()
}

func ClientIP(r *http.Request) (string, error) {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		first, _, _ := strings.Cut(forwardedFor, ",")
		if address, err := netip.ParseAddr(strings.TrimSpace(first)); err == nil {
			return address.String(), nil
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	return host, nil
}
