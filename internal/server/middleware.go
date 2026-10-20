package server

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

type Middleware func(http.Handler) http.Handler

// Лимитер общий на весь сервер. Per-client вариант - отдельная задача.
func RateLimitMiddleware(rps int, burst int) Middleware {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Ограничивает число одновременных запросов на сервере.
// Ждёт освобождения слота не дольше waitTimeout, иначе отдаёт 503.
func BulkheadMiddleware(maxConcurrent int64, waitTimeout time.Duration) Middleware {
	sem := semaphore.NewWeighted(maxConcurrent)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), waitTimeout)
			defer cancel()

			if err := sem.Acquire(ctx, 1); err != nil {
				http.Error(w, "bulkhead full", http.StatusServiceUnavailable)
				return
			}
			defer sem.Release(1)

			next.ServeHTTP(w, r)
		})
	}
}

// Ограничивает время обработки запроса. Пробрасывает ctx с дедлайном вниз,
// если хендлер не успел записать ответ - отдаёт 504.
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)

			if sw.wroteHeader {
				return
			}

			if ctx.Err() == context.DeadlineExceeded {
				http.Error(sw, "server timeout", http.StatusGatewayTimeout)
				return
			}

			sw.WriteHeader(http.StatusOK)
		})
	}
}

// Запоминает, был ли записан заголовок ответа.
type statusWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (sw *statusWriter) WriteHeader(code int) {
	if sw.wroteHeader {
		return
	}
	sw.wroteHeader = true
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if !sw.wroteHeader {
		sw.WriteHeader(http.StatusOK)
	}
	return sw.ResponseWriter.Write(b)
}

// Счётчики отказов по middleware.
type Counters struct {
	RateLimited  int64
	BulkheadFull int64
	Timeouts     int64
}

func NewCounters() *Counters { return &Counters{} }

// RateLimitMiddleware со счётчиком отказов.
func RateLimitMiddlewareWithCounter(rps int, burst int, c *Counters) Middleware {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				atomic.AddInt64(&c.RateLimited, 1)
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BulkheadMiddleware со счётчиком отказов.
func BulkheadMiddlewareWithCounter(maxConcurrent int64, waitTimeout time.Duration, c *Counters) Middleware {
	sem := semaphore.NewWeighted(maxConcurrent)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), waitTimeout)
			defer cancel()

			if err := sem.Acquire(ctx, 1); err != nil {
				atomic.AddInt64(&c.BulkheadFull, 1)
				http.Error(w, "bulkhead full", http.StatusServiceUnavailable)
				return
			}
			defer sem.Release(1)

			next.ServeHTTP(w, r)
		})
	}
}
