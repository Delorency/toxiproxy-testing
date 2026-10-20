package client

import (
	"sort"
	"time"
)

// RequestMetrics — метрики одного запроса
type RequestMetrics struct {
	Timestamp time.Time
	Endpoint  string
	OK        bool
	Latency   time.Duration
	Attempts  int    // сколько попыток Retry сделал
	CBState   string // состояние CB на момент запроса: closed / open / half-open
	Err       string // текст ошибки, если была
}

// Snapshot - агрегированные метрики за период
type Snapshot struct {
	Total   int
	Success int
	Failed  int
	Retries int

	// распределение запросов по состоянию CB
	CBStateClosed   int
	CBStateOpen     int
	CBStateHalfOpen int

	P50Ms float64
	P95Ms float64
	P99Ms float64
	AvgMs float64
}

// NewSnapshot считает агрегаты по срезу метрик
func NewSnapshot(ms []RequestMetrics) Snapshot {
	s := Snapshot{Total: len(ms)}
	if s.Total == 0 {
		return s
	}

	latencies := make([]float64, 0, len(ms))
	for _, m := range ms {
		if m.OK {
			s.Success++
		} else {
			s.Failed++
		}
		if m.Attempts > 1 {
			s.Retries += m.Attempts - 1
		}
		switch m.CBState {
		case "closed":
			s.CBStateClosed++
		case "open":
			s.CBStateOpen++
		case "half-open":
			s.CBStateHalfOpen++
		}
		latencies = append(latencies, float64(m.Latency.Milliseconds()))
	}

	sort.Float64s(latencies)
	s.P50Ms = percentileSorted(latencies, 0.50)
	s.P95Ms = percentileSorted(latencies, 0.95)
	s.P99Ms = percentileSorted(latencies, 0.99)
	s.AvgMs = avg(latencies)

	return s
}

// percentileSorted принимает уже отсортированный срез
func percentileSorted(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(q * float64(len(sorted)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func avg(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var sum float64
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}
