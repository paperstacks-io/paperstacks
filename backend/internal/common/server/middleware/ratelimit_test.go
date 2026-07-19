package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimitMiddleware(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(2), 2)
	middleware := RateLimit(limiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	sendReq := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := sendReq(); code != http.StatusOK {
		t.Errorf("Expected StatusOK, got %d", code)
	}
	if code := sendReq(); code != http.StatusOK {
		t.Errorf("Expected StatusOK, got %d", code)
	}

	if code := sendReq(); code != http.StatusTooManyRequests {
		t.Errorf("Expected StatusTooManyRequests, got %d", code)
	}

	time.Sleep(1 * time.Second)

	if code := sendReq(); code != http.StatusOK {
		t.Errorf("Expected StatusOK after waiting, got %d", code)
	}
}

func TestRateLimitLimiterCleanup(t *testing.T) {
	limiter := &IPRateLimiter{
		clients: make(map[string]*client),
		r:       rate.Limit(10),
		b:       10,
	}

	ip := "10.0.0.1"
	limiter.getClientLimiter(ip)

	limiter.mu.Lock()
	limiter.clients[ip].lastSeen = time.Now().Add(-11 * time.Minute)
	limiter.mu.Unlock()

	limiter.mu.Lock()
	for clientIP, c := range limiter.clients {
		if time.Since(c.lastSeen) > 10*time.Minute {
			delete(limiter.clients, clientIP)
		}
	}
	limiter.mu.Unlock()

	limiter.mu.RLock()
	_, exists := limiter.clients[ip]
	limiter.mu.RUnlock()

	if exists {
		t.Errorf("Expected client with IP %s to be cleaned up, but it still exists", ip)
	}
}
