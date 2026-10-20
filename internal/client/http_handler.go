package client

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) HTTPHandler() Handler {
	return func(ctx context.Context, path string) error {
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
}
