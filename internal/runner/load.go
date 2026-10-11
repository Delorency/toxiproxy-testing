package runner

import (
	"context"
	"sync"
	"time"

	"pr4/internal/client"
)

// RunLoad гонит нагрузку с заданным RPS в течение duration
// Возвращает срез метрик, собранных за этот период
func RunLoad(
	ctx context.Context,
	c *client.Client,
	rps int,
	duration time.Duration,
	endpoint string,
	metricsCh <-chan client.RequestMetrics,
) []client.RequestMetrics {
	if rps <= 0 {
		rps = 1
	}
	interval := time.Second / time.Duration(rps)

	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	// генератор запросов
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				go func() {
					_ = c.Get(ctx, endpoint)
				}()
			}
		}
	}()

	// сбор метрик, пока идёт нагрузка
	collected := make([]client.RequestMetrics, 0, rps*int(duration.Seconds()))
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-metricsCh:
				collected = append(collected, m)
			}
		}
	}()

	wg.Wait()
	// дать секунду на то, чтобы долетели последние метрики
	time.Sleep(500 * time.Millisecond)
	cancel()
	<-done

	return collected
}
