package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"pr4/internal/config"
	"pr4/internal/patterns"
)

// Bulkhead → RateLimiter → CircuitBreaker → Retry → Timeout → HTTP
type Client struct {
	httpClient *http.Client
	baseURL    string

	bh *patterns.Bulkhead
	rl *patterns.RateLimiter
	cb *patterns.CircuitBreaker

	retryCf *config.RetryConfig
	toCf    *config.TimeoutConfig

	metricsCh chan<- RequestMetrics
}

// New создаёт клиент.
// baseURL — адрес сервера (например, адрес Toxiproxy).
func New(
	baseURL string,
	bhCfg *config.BulkheadConfig,
	rlCfg *config.RateLimiterConfig,
	cbCfg *config.CircuitBreakerConfig,
	retryCfg *config.RetryConfig,
	toCfg *config.TimeoutConfig,
	metricsCh chan<- RequestMetrics,
) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		baseURL:   baseURL,
		bh:        patterns.NewBulkhead(bhCfg),
		rl:        patterns.NewRateLimiter(rlCfg),
		cb:        patterns.NewCircuitBreaker(cbCfg),
		retryCf:   retryCfg,
		toCf:      toCfg,
		metricsCh: metricsCh,
	}
}

// Get выполняет GET-запрос через полный конвейер паттернов.
func (c *Client) Get(ctx context.Context, path string) error {
	start := time.Now()
	attempts := 0

	err := c.bh.Execute(ctx, func() error {
		// RateLimiter: fail-fast при переполнении
		if rlErr := c.rl.TryAcquire(); rlErr != nil {
			return rlErr
		}

		// CircuitBreaker видит итог всех ретраев
		return c.cb.Call(func() error {
			return patterns.Retry(ctx, c.retryCf, func() error {
				// Timeout — на каждую попытку отдельно
				return patterns.WithTimeout(ctx, c.toCf, func(tctx context.Context) error {
					attempts++
					return c.doRequest(tctx, path)
				})
			})
		})
	})

	latency := time.Since(start)

	// CBState — состояние ПОСЛЕ запроса, видит переходы closed → open
	select {
	case c.metricsCh <- RequestMetrics{
		Timestamp: start,
		Endpoint:  path,
		OK:        err == nil,
		Latency:   latency,
		Attempts:  attempts,
		CBState:   c.cbStateString(),
		Err:       errString(err),
	}:
	default:
	}

	return err
}

// doRequest — один HTTP-запрос.
func (c *Client) doRequest(ctx context.Context, path string) error {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("client error: %d", resp.StatusCode)
	}
	return nil
}

// CBStateString возвращает текущее состояние CB.
func (c *Client) cbStateString() string {
	return patterns.StateToString(c.cb.State())
}

// CBState — состояние CB (экспорт для тестов).
func (c *Client) CBState() patterns.State {
	return c.cb.State()
}

// BulkheadInUse — текущая загрузка bulkhead (для тестов).
func (c *Client) BulkheadInUse() int64 {
	return c.bh.InUse()
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
