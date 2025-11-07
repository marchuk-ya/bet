package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type RateLimiter struct {
	requestsPerWindow int
	windowDuration    time.Duration
	clients           map[string]*clientLimiter
	mu                sync.RWMutex
	cleanupInterval   time.Duration
	lastCleanup       time.Time
	logger            *zap.Logger
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

type clientLimiter struct {
	count       int
	windowStart time.Time
}

type RateLimitConfig struct {
	RequestsPerMinute int
	Logger            *zap.Logger
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	if config.RequestsPerMinute <= 0 {
		config.RequestsPerMinute = 60
	}

	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		requestsPerWindow: config.RequestsPerMinute,
		windowDuration:    time.Minute,
		clients:           make(map[string]*clientLimiter),
		cleanupInterval:   5 * time.Minute,
		lastCleanup:       time.Now(),
		logger:            config.Logger,
		ctx:               ctx,
		cancel:            cancel,
	}

	rl.wg.Add(1)
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	defer rl.wg.Done()

	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.ctx.Done():
			rl.mu.Lock()
			now := time.Now()
			for ip, client := range rl.clients {
				if now.Sub(client.windowStart) > rl.windowDuration {
					delete(rl.clients, ip)
				}
			}
			rl.lastCleanup = now
			rl.mu.Unlock()
			rl.logger.Info("rate limiter cleanup goroutine stopped")
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, client := range rl.clients {
				if now.Sub(client.windowStart) > rl.windowDuration {
					delete(rl.clients, ip)
				}
			}
			rl.lastCleanup = now
			rl.mu.Unlock()
		}
	}
}

func (rl *RateLimiter) Shutdown(ctx context.Context) error {
	rl.logger.Info("shutting down rate limiter...")
	rl.cancel()

	done := make(chan struct{})
	go func() {
		rl.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		rl.logger.Info("rate limiter shutdown complete")
		return nil
	case <-ctx.Done():
		rl.logger.Warn("rate limiter shutdown timeout")
		return ctx.Err()
	}
}

func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[ip]

	if !exists {
		rl.clients[ip] = &clientLimiter{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if now.Sub(client.windowStart) >= rl.windowDuration {
		client.count = 1
		client.windowStart = now
		return true
	}

	if client.count >= rl.requestsPerWindow {
		return false
	}

	client.count++
	return true
}

func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			path := r.URL.Path
			if path == "/health" || path == "/ready" || path == "/live" {
				next.ServeHTTP(w, r)
				return
			}

			clientIP := getClientIP(r)

			if !limiter.Allow(clientIP) {
				limiter.logger.Warn("rate limit exceeded",
					zap.String("ip", clientIP),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)

				response := map[string]string{
					"error": "rate limit exceeded",
				}

				json.NewEncoder(w).Encode(response)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.requestsPerWindow))
			next.ServeHTTP(w, r)
		})
	}
}
