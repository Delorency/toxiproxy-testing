package test

// import (
// 	"context"
// 	"fmt"
// 	"net"
// 	"net/http"
// 	"testing"
// 	"time"

// 	toxiproxyLib "github.com/Shopify/toxiproxy/v2/client"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	"pr4/internal/config"
// 	"pr4/internal/patterns"
// 	"pr4/internal/toxiproxy"
// )

// // TestTimeout_WithLatency проверяет что тайм-аут срабатывает раньше чем задержка
// func TestTimeout_WithLatency(t *testing.T) {
// 	cfg := config.MustLoad()

// 	// сбрасываем состояние Toxiproxy
// 	resetClient := toxiproxyLib.NewClient(cfg.GetToxiproxyAddress())
// 	resetClient.ResetState()

// 	// получаем порт из конфига
// 	serverPort := cfg.Timeout.TestPort

// 	// запускаем тестовый сервер
// 	server := startTestServer(t, serverPort)
// 	defer server.Close()

// 	// создаем уникальное имя прокси
// 	proxyName := fmt.Sprintf("timeout-test-%d", time.Now().UnixNano())

// 	// upstream указывает на наш сервер
// 	upstream := fmt.Sprintf("localhost:%d", serverPort)

// 	// создаем клиент Toxiproxy
// 	toxi, err := toxiproxy.NewToxiClient(cfg.Toxiproxy, proxyName, upstream)
// 	require.NoError(t, err)
// 	defer toxi.Close()

// 	// добавляем токсин задержки
// 	latencyMs := cfg.Timeout.LatencyMs
// 	err = toxi.AddLatency(latencyMs, 0)
// 	require.NoError(t, err)

// 	// создаем HTTP клиент с прохождением через прокси
// 	httpClient := &http.Client{
// 		Timeout: 10 * time.Second,
// 		Transport: &http.Transport{
// 			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
// 				return net.Dial("tcp", toxi.GetProxyAddress())
// 			},
// 		},
// 	}

// 	start := time.Now()

// 	// выполняем запрос с таймаутом из конфига (по умолчанию 1 секунда)
// 	err = patterns.WithTimeout(context.Background(), cfg.Timeout, func(ctx context.Context) error {
// 		req, err := http.NewRequestWithContext(ctx, "GET", "http://"+toxi.GetProxyAddress()+"/test", nil)
// 		if err != nil {
// 			return err
// 		}

// 		resp, err := httpClient.Do(req)
// 		if err != nil {
// 			return err
// 		}
// 		defer resp.Body.Close()
// 		return nil
// 	})

// 	elapsed := time.Since(start)

// 	// проверяем что тайм-аут сработал и время меньше чем задержка
// 	timeoutDuration := cfg.Timeout.DefaultTimeout
// 	assert.ErrorIs(t, err, patterns.TimeoutError, "Должна быть ошибка тайм-аута")
// 	assert.Less(t, elapsed, time.Duration(latencyMs)*time.Millisecond,
// 		"Запрос должен упасть быстрее чем за %d мс", latencyMs)
// 	assert.GreaterOrEqual(t, elapsed, timeoutDuration,
// 		"Тайм-аут должен сработать не раньше %v", timeoutDuration)
// }
