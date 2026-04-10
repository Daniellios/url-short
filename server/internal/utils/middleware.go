package utils

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimit struct {
	Count       int
	WindowStart time.Time
}

var (
	rateLimitMu      sync.Mutex
	rateLimitByIP    = make(map[string]*rateLimit)
	rateLimitWindow  = time.Minute
	rateLimitMaxHits = 10
)

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	return r.RemoteAddr
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		ip := clientIP(r)

		rateLimitMu.Lock()
		entry, exists := rateLimitByIP[ip]
		if !exists || now.Sub(entry.WindowStart) >= rateLimitWindow {
			entry = &rateLimit{
				Count:       0,
				WindowStart: now,
			}
			rateLimitByIP[ip] = entry
		}

		entry.Count++
		exceeded := entry.Count > rateLimitMaxHits
		rateLimitMu.Unlock()

		if exceeded {
			RespondWithError(w, 429, "Too many requests")
			return
		}

		next.ServeHTTP(w, r)
	})
}
