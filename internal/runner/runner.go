package runner

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pr4/internal/client"
	"pr4/internal/config"
	"pr4/internal/toxiproxy"
)

// Runner — прогон сценария
type Runner struct {
	cfg *config.Config
}

// New создаёт раннер
func New(cfg *config.Config) *Runner {
	return &Runner{cfg: cfg}
}

// Run выполняет сценарий и сохраняет результат в CSV
func (r *Runner) Run(scenarioPath string) (string, error) {
	sc, err := LoadScenario(scenarioPath)
	if err != nil {
		return "", err
	}

	log.Printf("scenario loaded: %s, steps=%d, server=%s", sc.Name, len(sc.Steps), sc.ServerURL)

	// общий контекст на весь сценарий
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// upstream для Toxiproxy — без схемы
	upstream := trimScheme(sc.ServerURL)
	log.Printf("upstream for toxiproxy: %s", upstream)

	// создаём прокси
	toxi, err := toxiproxy.NewToxiClient(r.cfg.Toxiproxy, sc.ProxyName, upstream)
	if err != nil {
		return "", fmt.Errorf("create toxiproxy: %w", err)
	}
	defer toxi.Close()
	defer func() { _ = toxi.Reset() }()

	// адрес прокси, на который будет ходить клиент
	clientBaseURL, err := r.proxyClientURL(toxi.GetProxyAddress())
	if err != nil {
		return "", fmt.Errorf("parse proxy address: %w", err)
	}
	log.Printf("client will connect to: %s", clientBaseURL)

	// канал метрик — общий на весь сценарий
	metricsCh := make(chan client.RequestMetrics, 200000)

	// создаём клиент
	c := client.New(
		clientBaseURL,
		r.cfg.Bulkhead,
		r.cfg.RateLimiter,
		r.cfg.CircuitBreaker,
		r.cfg.Retry,
		r.cfg.Timeout,
		metricsCh,
	)

	// прогрев — метрики после него сбрасываем, чтобы не попали в первый шаг
	if sc.WarmupS > 0 {
		log.Printf("warmup %ds...", sc.WarmupS)
		_ = RunLoad(ctx, c, 20, time.Duration(sc.WarmupS)*time.Second, "/fast", metricsCh)
		drainMetrics(metricsCh)
	}

	// основной прогон
	results := make([]StepResult, 0, len(sc.Steps))
	for i, step := range sc.Steps {
		log.Printf("step %d/%d: duration=%ds rps=%d toxic=%v",
			i+1, len(sc.Steps), step.DurationS, step.RPS, step.Toxic)

		// включаем токсин перед шагом
		toxicApplied := false
		if step.Toxic != nil {
			if err := applyToxic(toxi, step.Toxic); err != nil {
				log.Printf("warning: failed to apply toxic: %v", err)
			} else {
				toxicApplied = true
			}
		}

		// гарантированно снимаем токсин после шага
		func() {
			if toxicApplied {
				defer func() { _ = toxi.Reset() }()
			}

			metrics := RunLoad(ctx, c, step.RPS, step.Duration(), step.Endpoint, metricsCh)
			snap := client.NewSnapshot(metrics)

			results = append(results, StepResult{
				Step:      i + 1,
				DurationS: step.DurationS,
				RPS:       step.RPS,
				Toxic:     toxicName(step.Toxic),
				Snapshot:  snap,
			})

			log.Printf("  total=%d success=%d failed=%d retries=%d p50=%.0fms p95=%.0fms p99=%.0fms cb: closed=%d open=%d half-open=%d",
				snap.Total, snap.Success, snap.Failed, snap.Retries,
				snap.P50Ms, snap.P95Ms, snap.P99Ms,
				snap.CBStateClosed, snap.CBStateOpen, snap.CBStateHalfOpen)
		}()

		// чистим канал от «хвостов» предыдущего шага
		drainMetrics(metricsCh)
	}

	// сохраняем CSV
	if err := os.MkdirAll(r.cfg.Runner.ResultsDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir results: %w", err)
	}
	outPath := filepath.Join(r.cfg.Runner.ResultsDir, sc.Name+".csv")
	if err := writeCSV(outPath, results); err != nil {
		return "", fmt.Errorf("write csv: %w", err)
	}

	return outPath, nil
}

// StepResult — результат шага
type StepResult struct {
	Step      int
	DurationS int
	RPS       int
	Toxic     string
	Snapshot  client.Snapshot
}

// writeCSV сохраняет результаты в CSV
func writeCSV(path string, results []StepResult) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{
		"step", "duration_s", "rps", "toxic",
		"total", "success", "failed", "retries",
		"p50_ms", "p95_ms", "p99_ms", "avg_ms",
		"cb_closed", "cb_open", "cb_half_open",
	}); err != nil {
		return err
	}

	for _, r := range results {
		row := []string{
			strconv.Itoa(r.Step),
			strconv.Itoa(r.DurationS),
			strconv.Itoa(r.RPS),
			r.Toxic,
			strconv.Itoa(r.Snapshot.Total),
			strconv.Itoa(r.Snapshot.Success),
			strconv.Itoa(r.Snapshot.Failed),
			strconv.Itoa(r.Snapshot.Retries),
			fmt.Sprintf("%.1f", r.Snapshot.P50Ms),
			fmt.Sprintf("%.1f", r.Snapshot.P95Ms),
			fmt.Sprintf("%.1f", r.Snapshot.P99Ms),
			fmt.Sprintf("%.1f", r.Snapshot.AvgMs),
			strconv.Itoa(r.Snapshot.CBStateClosed),
			strconv.Itoa(r.Snapshot.CBStateOpen),
			strconv.Itoa(r.Snapshot.CBStateHalfOpen),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// applyToxic включает указанный toxic
func applyToxic(toxi *toxiproxy.ToxiClient, t *Toxic) error {
	switch t.Type {
	case "latency":
		return toxi.AddLatency(t.LatencyMs, t.JitterMs)
	case "timeout":
		return toxi.AddTimeout(t.TimeoutMs)
	case "bandwidth":
		return toxi.AddBandwidthLimit(t.RateKBps)
	case "reset_peer":
		return toxi.AddResetPeer(t.TimeoutMs)
	case "slow_close":
		return toxi.AddSlowClose(t.DelayMs)
	default:
		return fmt.Errorf("unknown toxic type: %s", t.Type)
	}
}

// toxicName возвращает строковое имя токсина для отчёта
func toxicName(t *Toxic) string {
	if t == nil {
		return "-"
	}
	return t.Type
}

// trimScheme убирает http:// или https:// из URL
func trimScheme(u string) string {
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "https://")
	return u
}

// proxyClientURL строит адрес прокси, к которому подключается клиент.
// Хост берётся из конфига Toxiproxy — это позволяет работать и локально,
// и внутри docker-сети (localhost → toxiproxy).
func (r *Runner) proxyClientURL(listen string) (string, error) {
	_, port, err := net.SplitHostPort(listen)
	if err != nil {
		return "", err
	}
	host := r.cfg.Toxiproxy.Host
	if host == "" || host == "localhost" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port), nil
}

// drainMetrics опустошает канал метрик, не блокируясь.
// Нужен, чтобы метрики от предыдущего шага не попали в следующий.
func drainMetrics(ch chan client.RequestMetrics) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
