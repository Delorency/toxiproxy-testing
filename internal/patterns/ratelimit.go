package patterns

import (
	"context"
	"errors"

	"golang.org/x/time/rate"

	"pr4/internal/config"
)

// ErrRateLimited возвращается, когда токенов нет
var ErrRateLimited = errors.New("rate limited")

// RateLimiter ограничивает частоту запросов (token bucket)
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter создаёт новый лимитер
func NewRateLimiter(cfg *config.RateLimiterConfig) *RateLimiter {
	limit := rate.Limit(cfg.RPS)
	if cfg.RPS <= 0 {
		limit = rate.Inf
	}
	return &RateLimiter{
		limiter: rate.NewLimiter(limit, cfg.Burst),
	}
}

// TryAcquire - неблокирующая попытка взять токен
func (rl *RateLimiter) TryAcquire() error {
	if !rl.limiter.Allow() {
		return ErrRateLimited
	}
	return nil
}

// Wait блокируется до получения токена или отмены ctx
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

// Reserve позволяет узнать, сколько ждать, и решить самому
func (rl *RateLimiter) Reserve() *rate.Reservation {
	return rl.limiter.Reserve()
}

func (rl *RateLimiter) Limit() float64 { return float64(rl.limiter.Limit()) }
func (rl *RateLimiter) Burst() int     { return rl.limiter.Burst() }
