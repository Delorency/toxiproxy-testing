package patterns

import (
	"context"

	"pr4/internal/config"
)

// TimeoutError возвращается при превышении времени ожидания
var TimeoutError = context.DeadlineExceeded

// WithTimeout выполняет функцию с ограничением по времени.
// Если функция не завершилась за отведённое время, возвращается ошибка.
func WithTimeout(ctx context.Context, cfg *config.TimeoutConfig, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, cfg.DefaultTimeout)
	defer cancel()
	return fn(ctx)
}
