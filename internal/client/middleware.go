package client

import (
	"context"

	"pr4/internal/config"
	"pr4/internal/patterns"
)

func BulkheadMiddleware(bh *patterns.Bulkhead) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, path string) error {
			return bh.Execute(ctx, func() error {
				return next(ctx, path)
			})
		}
	}
}

func RateLimitMiddleware(rl *patterns.RateLimiter) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, path string) error {
			if err := rl.TryAcquire(); err != nil {
				return err
			}
			return next(ctx, path)
		}
	}
}

func CircuitBreakerMiddleware(cb *patterns.CircuitBreaker) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, path string) error {
			return cb.Call(func() error {
				return next(ctx, path)
			})
		}
	}
}

func RetryMiddleware(cfg *config.RetryConfig) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, path string) error {
			return patterns.Retry(ctx, cfg, func() error {
				return next(ctx, path)
			})
		}
	}
}

func TimeoutMiddleware(cfg *config.TimeoutConfig) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, path string) error {
			return patterns.WithTimeout(ctx, cfg, func(tctx context.Context) error {
				return next(tctx, path)
			})
		}
	}
}
