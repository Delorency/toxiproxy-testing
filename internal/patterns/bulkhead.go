package patterns

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"

	"golang.org/x/sync/semaphore"

	"pr4/internal/config"
)

// ErrBulkheadFull возвращается, когда все слоты заняты (fail-fast)
var ErrBulkheadFull = errors.New("bulkhead is full")

// ErrBulkheadPanic оборачивает панику внутри защищённой функции
var ErrBulkheadPanic = errors.New("panic in bulkhead function")

// Bulkhead ограничивает количество параллельных выполнений.
// Реализует паттерн Bulkhead (Release It!): при переполнении — fail-fast,
// не блокируется, чтобы не превратиться в очередь и не усугубить лавину.
type Bulkhead struct {
	sem      *semaphore.Weighted
	maxSize  int64
	acquired int64
}

// NewBulkhead создает новый изолятор ресурсов
func NewBulkhead(cfg *config.BulkheadConfig) *Bulkhead {
	maxConcurrent := cfg.MaxConcurrent
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	return &Bulkhead{
		sem:     semaphore.NewWeighted(maxConcurrent),
		maxSize: maxConcurrent,
	}
}

// Execute — fail-fast версия: если нет свободного слота, сразу возвращает ErrBulkheadFull.
// Не блокируется, поэтому не нуждается в ctx для ожидания.
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
	if !b.sem.TryAcquire(1) {
		return ErrBulkheadFull
	}
	return b.run(fn)
}

// ExecuteWithWait — ждёт свободный слот до отмены ctx.
// Это уже не чистый Bulkhead, а worker pool — используйте осознанно.
func (b *Bulkhead) ExecuteWithWait(ctx context.Context, fn func() error) error {
	if err := b.sem.Acquire(ctx, 1); err != nil {
		return err // ctx.Err()
	}
	return b.run(fn)
}

// run — общая часть: слот уже захвачен, нужно его освободить
func (b *Bulkhead) run(fn func() error) error {
	defer b.sem.Release(1)

	atomic.AddInt64(&b.acquired, 1)
	defer atomic.AddInt64(&b.acquired, -1)

	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%w: %v\n%s", ErrBulkheadPanic, r, debug.Stack())
			}
		}()
		err = fn()
	}()

	return err
}

// Available возвращает количество свободных слотов (приблизительно, снимок)
func (b *Bulkhead) Available() int64 {
	val := b.maxSize - atomic.LoadInt64(&b.acquired)
	if val < 0 {
		return 0
	}
	return val
}

// InUse возвращает количество занятых слотов (приблизительно, снимок)
func (b *Bulkhead) InUse() int64 {
	return atomic.LoadInt64(&b.acquired)
}
