package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*client
	r       rate.Limit
	b       int
}

// NewIPRateLimiter creates a rate limiter tracker with rate r and burst size b.
// It also starts a background goroutine to clean up inactive clients.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients: make(map[string]*client),
		r:       r,
		b:       b,
	}

	go limiter.cleanupClients()

	return limiter
}

func (i *IPRateLimiter) getClientLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	c, exists := i.clients[ip]
	if !exists {
		c = &client{
			limiter:  rate.NewLimiter(i.r, i.b),
			lastSeen: time.Now(),
		}
		i.clients[ip] = c
	} else {
		c.lastSeen = time.Now()
	}

	return c.limiter
}

// cleanupClients runs periodically in the background and removes clients that haven't
// made a request in the last 10 minutes to prevent memory leaks.
func (i *IPRateLimiter) cleanupClients() {
	ticker := time.NewTicker(3 * time.Minute)
	for range ticker.C {
		i.mu.Lock()
		for ip, c := range i.clients {
			if time.Since(c.lastSeen) > 10*time.Minute {
				delete(i.clients, ip)
			}
		}
		i.mu.Unlock()
	}
}

// RateLimit returns a middleware that limits requests using the provided IPRateLimiter.
func RateLimit(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			lim := limiter.getClientLimiter(ip)
			if !lim.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
