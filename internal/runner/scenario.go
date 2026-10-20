package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Scenario — JSON-сценарий эксперимента
type Scenario struct {
	Name        string    `json:"name"`
	ServerURL   string    `json:"server_url"`
	ProxyName   string    `json:"proxy_name"`
	WarmupS     int       `json:"warmup_s"`
	Middlewares *[]string `json:"middlewares,omitempty"`
	Steps       []Step    `json:"steps"`
}

// Step — шаг сценария
type Step struct {
	DurationS int    `json:"duration_s"`
	RPS       int    `json:"rps"`
	Toxic     *Toxic `json:"toxic,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"` // по умолчанию /fast
}

// Toxic — описание включаемого токсина
type Toxic struct {
	Type      string `json:"type"` // latency | timeout | bandwidth | reset_peer | slow_close
	LatencyMs int64  `json:"latency_ms,omitempty"`
	JitterMs  int64  `json:"jitter_ms,omitempty"`
	TimeoutMs int    `json:"timeout_ms,omitempty"`
	RateKBps  int64  `json:"rate_kbps,omitempty"`
	DelayMs   int    `json:"delay_ms,omitempty"`
}

// LoadScenario читает JSON-сценарий из файла
func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario: %w", err)
	}

	var s Scenario
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse scenario: %w", err)
	}

	if s.ProxyName == "" {
		s.ProxyName = "client_to_server"
	}
	if s.WarmupS == 0 {
		s.WarmupS = 3
	}
	for i := range s.Steps {
		if s.Steps[i].Endpoint == "" {
			s.Steps[i].Endpoint = "/fast"
		}
		if s.Steps[i].DurationS == 0 {
			s.Steps[i].DurationS = 10
		}
		if s.Steps[i].RPS == 0 {
			s.Steps[i].RPS = 50
		}
	}

	if s.Middlewares == nil {
		defaultMW := []string{"bulkhead", "ratelimit", "circuitbreaker", "retry", "timeout"}
		s.Middlewares = &defaultMW
	}

	return &s, nil
}

// Duration возвращает длительность шага
func (s Step) Duration() time.Duration {
	return time.Duration(s.DurationS) * time.Second
}
