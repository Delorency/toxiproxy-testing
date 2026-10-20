package patterns

import (
	"errors"
	"sync"
	"time"

	"pr4/internal/config"
)

// ErrCircuitOpen возвращается, когда выключатель разомкнут
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State - состояние автоматического выключателя
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker защищает сервис от каскадных отказов
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failures         int
	successes        int
	threshold        int
	halfOpenSuccess  int
	halfOpenInFlight int
	timeout          time.Duration
	lastFailureTime  time.Time
}

// NewCircuitBreaker создаёт новый автоматический выключатель
func NewCircuitBreaker(cfg *config.CircuitBreakerConfig) *CircuitBreaker {
	halfOpenSuccess := cfg.HalfOpenSuccess
	if halfOpenSuccess < 1 {
		halfOpenSuccess = 1
	}

	threshold := cfg.Threshold
	if threshold < 1 {
		threshold = 1
	}

	return &CircuitBreaker{
		state:           StateClosed,
		threshold:       threshold,
		halfOpenSuccess: halfOpenSuccess,
		timeout:         cfg.Timeout,
	}
}

// Call выполняет функцию под защитой выключателя.
// Гарантирует, что результат будет записан даже при панике в fn,
// чтобы halfOpenInFlight не залип в состоянии HalfOpen.
func (cb *CircuitBreaker) Call(fn func() error) (err error) {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	defer func() {
		if r := recover(); r != nil {
			cb.recordResult(false)
			panic(r)
		}
		cb.recordResult(err == nil)
	}()

	err = fn()
	return err
}

// State возвращает текущее состояние выключателя,
// при этом обновляя переход Open -> HalfOpen по истечении таймаута.
// Важно для тестов и метрик, чтобы чтение состояния видело
// актуальную картину, а не устаревшую.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.maybeHalfOpenLocked()
	return cb.state
}

// StateToString возвращает строковое представление состояния
func StateToString(s State) string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// maybeHalfOpenLocked переводит Open -> HalfOpen, если истёк таймаут.
// Должен вызываться под удержанием cb.mu.
func (cb *CircuitBreaker) maybeHalfOpenLocked() {
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.state = StateHalfOpen
		cb.successes = 0
		cb.halfOpenInFlight = 0
	}
}

// allowRequest проверяет, разрешён ли запрос
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.maybeHalfOpenLocked()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		return false

	case StateHalfOpen:
		// в HalfOpen пропускаем только один пробный запрос одновременно
		if cb.halfOpenInFlight > 0 {
			return false
		}
		cb.halfOpenInFlight = 1
		return true

	default:
		return false
	}
}

// recordResult обновляет состояние выключателя на основе результата
func (cb *CircuitBreaker) recordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		switch cb.state {
		case StateClosed:
			cb.failures = 0

		case StateHalfOpen:
			cb.halfOpenInFlight = 0
			cb.successes++
			if cb.successes >= cb.halfOpenSuccess {
				cb.state = StateClosed
				cb.failures = 0
				cb.successes = 0
			}
		}
		return
	}

	// обработка ошибки
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.threshold {
			cb.state = StateOpen
		}

	case StateHalfOpen:
		cb.halfOpenInFlight = 0
		cb.successes = 0
		cb.state = StateOpen
	}
}
