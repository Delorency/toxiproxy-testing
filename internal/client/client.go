package client

import (
	"context"
	"net/http"
	"pr4/internal/patterns"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	handler    Handler
	metricsCh  chan<- RequestMetrics
	cb         *patterns.CircuitBreaker
}

func (c *Client) SetCB(cb *patterns.CircuitBreaker) {
	c.cb = cb
}

func New(
	baseURL string,
	metricsCh chan<- RequestMetrics,
	mw ...Middleware,
) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		baseURL:   baseURL,
		metricsCh: metricsCh,
	}

	// собираем пирог с конца: первый mw — самый внешний
	var h Handler = c.HTTPHandler()
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	c.handler = h

	return c
}

// Get выполняет запрос через собранный пирог.
func (c *Client) Get(ctx context.Context, path string) error {
	start := time.Now()
	err := c.handler(ctx, path)
	latency := time.Since(start)

	cbState := ""
	if c.cb != nil {
		cbState = patterns.StateToString(c.cb.State())
	}

	select {
	case c.metricsCh <- RequestMetrics{
		Timestamp: start,
		Endpoint:  path,
		OK:        err == nil,
		Latency:   latency,
		CBState:   cbState,
		Err:       errString(err),
	}:
	default:
	}
	return err
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
