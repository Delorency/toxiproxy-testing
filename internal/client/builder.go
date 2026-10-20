package client

import (
	"fmt"

	"pr4/internal/config"
	"pr4/internal/patterns"
)

type MiddlewareDeps struct {
	Bulkhead       *patterns.Bulkhead
	RateLimiter    *patterns.RateLimiter
	CircuitBreaker *patterns.CircuitBreaker
	RetryCfg       *config.RetryConfig
	TimeoutCfg     *config.TimeoutConfig
}

func BuildMiddlewares(names []string, deps *MiddlewareDeps) ([]Middleware, error) {
	mw := make([]Middleware, 0, len(names))

	for _, name := range names {
		switch name {
		case "bulkhead":
			if deps.Bulkhead == nil {
				return nil, fmt.Errorf("bulkhead: deps.Bulkhead is nil")
			}
			mw = append(mw, BulkheadMiddleware(deps.Bulkhead))

		case "ratelimit":
			if deps.RateLimiter == nil {
				return nil, fmt.Errorf("ratelimit: deps.RateLimiter is nil")
			}
			mw = append(mw, RateLimitMiddleware(deps.RateLimiter))

		case "circuitbreaker":
			if deps.CircuitBreaker == nil {
				return nil, fmt.Errorf("circuitbreaker: deps.CircuitBreaker is nil")
			}
			mw = append(mw, CircuitBreakerMiddleware(deps.CircuitBreaker))

		case "retry":
			if deps.RetryCfg == nil {
				return nil, fmt.Errorf("retry: deps.RetryCfg is nil")
			}
			mw = append(mw, RetryMiddleware(deps.RetryCfg))

		case "timeout":
			if deps.TimeoutCfg == nil {
				return nil, fmt.Errorf("timeout: deps.TimeoutCfg is nil")
			}
			mw = append(mw, TimeoutMiddleware(deps.TimeoutCfg))

		default:
			return nil, fmt.Errorf("unknown middleware: %q", name)
		}
	}

	return mw, nil
}
