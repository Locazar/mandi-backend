package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/config"
)

type rateEntry struct {
	count       int
	windowStart time.Time
}

type bruteEntry struct {
	count        int
	windowStart  time.Time
	blockedUntil time.Time
}

type securityState struct {
	mu           sync.Mutex
	rateEntries  map[string]*rateEntry
	bruteEntries map[string]*bruteEntry
}

func newSecurityState() *securityState {
	return &securityState{
		rateEntries:  make(map[string]*rateEntry),
		bruteEntries: make(map[string]*bruteEntry),
	}
}

func SecurityHeadersMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("X-Frame-Options", "DENY")
		ctx.Header("X-XSS-Protection", "1; mode=block")
		ctx.Header("Referrer-Policy", "same-origin")
		ctx.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		ctx.Header("Content-Security-Policy", "default-src 'self'")

		if ctx.Request.TLS != nil && cfg.Security.HSTSMaxAge > 0 {
			hstsValue := fmt.Sprintf("max-age=%d", cfg.Security.HSTSMaxAge)
			if cfg.Security.HSTSIncludeSubDomains {
				hstsValue += "; includeSubDomains"
			}
			if cfg.Security.HSTSPreload {
				hstsValue += "; preload"
			}
			ctx.Header("Strict-Transport-Security", hstsValue)
		}

		ctx.Next()
	}
}

func SecurityMiddleware(cfg config.Config) gin.HandlerFunc {
	state := newSecurityState()
	return func(ctx *gin.Context) {
		clientIP := getClientIP(ctx)
		now := time.Now()

		if cfg.Security.BruteForceProtectionEnabled {
			if state.isBlocked(clientIP, now) {
				ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"success": false,
					"message": "Too many authentication attempts. Try again later.",
				})
				return
			}
		}

		if cfg.Security.RateLimitingEnabled {
			if exceeded := state.registerRate(clientIP, cfg.Security.RateLimitRequests, time.Duration(cfg.Security.RateLimitWindowSeconds)*time.Second, now); exceeded {
				ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"success": false,
					"message": "Rate limit exceeded. Slow down your requests.",
				})
				return
			}
		}

		ctx.Next()

		if cfg.Security.BruteForceProtectionEnabled {
			status := ctx.Writer.Status()
			if status == http.StatusUnauthorized || status == http.StatusForbidden {
				state.registerFailure(clientIP, cfg.Security.BruteForceMaxAttempts, time.Duration(cfg.Security.BruteForceWindowSeconds)*time.Second, time.Duration(cfg.Security.BruteForceBlockDurationSeconds)*time.Second, now)
			}
		}
	}
}

func getClientIP(ctx *gin.Context) string {
	clientIP := ctx.ClientIP()
	if clientIP == "" {
		clientIP = "unknown"
	}
	return clientIP
}

func (s *securityState) isBlocked(clientIP string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.bruteEntries[clientIP]
	if !exists {
		return false
	}
	if now.After(entry.blockedUntil) {
		entry.count = 0
		entry.windowStart = now
		entry.blockedUntil = time.Time{}
		return false
	}
	return true
}

func (s *securityState) registerRate(clientIP string, limit int, window time.Duration, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.rateEntries[clientIP]
	if !exists || now.After(entry.windowStart.Add(window)) {
		s.rateEntries[clientIP] = &rateEntry{count: 1, windowStart: now}
		return false
	}
	entry.count++
	return entry.count > limit
}

func (s *securityState) registerFailure(clientIP string, maxAttempts int, window, blockDuration time.Duration, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.bruteEntries[clientIP]
	if !exists || now.After(entry.windowStart.Add(window)) {
		s.bruteEntries[clientIP] = &bruteEntry{count: 1, windowStart: now}
		return
	}
	entry.count++
	if entry.count >= maxAttempts {
		entry.blockedUntil = now.Add(blockDuration)
		entry.count = 0
		entry.windowStart = now
	}
}
