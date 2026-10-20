package runner

import (
	"context"
	"sync"
	"time"

	"pr4/internal/client"
)

func RunLoad(
	ctx context.Context,
	c *client.Client,
	rps int,
	duration time.Duration,
	endpoint string,
	metricsCh chan client.RequestMetrics,
) []client.RequestMetrics {
	if rps <= 0 {
		rps = 1
	}
	interval := time.Second / time.Duration(rps)

	loadCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	var wg sync.WaitGroup
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-loadCtx.Done():
			break loop
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				reqCtx, reqCancel := context.WithTimeout(ctx, duration)
				defer reqCancel()
				_ = c.Get(reqCtx, endpoint)
			}()
		}
	}

	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	// забираем всё, что накопилось, не блокируясь
	collected := make([]client.RequestMetrics, 0, rps*int(duration.Seconds()))
	for {
		select {
		case m := <-metricsCh:
			collected = append(collected, m)
		default:
			return collected
		}
	}
}
