package patterns

import (
	"context"
	"time"

	"pr4/internal/config"
)

// Retry выполняет функцию с повторными попытками при ошибках
// Использует экспоненциальную задержку между попытками
func Retry(ctx context.Context, cfg *config.RetryConfig, fn func() error) error {
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		if attempt == cfg.MaxAttempts {
			return err
		}

		// ждем перед следующей попыткой
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// увеличиваем задержку экспоненциально
			delay = time.Duration(float64(delay) * cfg.Multiplier)
			delay = min(delay, cfg.MaxDelay)
		}
	}

	return nil
}
