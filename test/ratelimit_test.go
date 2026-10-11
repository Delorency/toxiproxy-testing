package test

// import (
// 	"context"
// 	"fmt"
// 	"net"
// 	"net/http"
// 	"sync"
// 	"testing"
// 	"time"

// 	toxiproxyLib "github.com/Shopify/toxiproxy/v2/client"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	"pr4/internal/config"
// 	"pr4/internal/patterns"
// 	"pr4/internal/toxiproxy"
// )

// // TestRateLimiter_Burst проверяет что лимитер пропускает burst запросов
// func TestRateLimiter_Burst(t *testing.T) {
// 	cfg := config.MustLoad()

// 	// создаем ограничитель частоты
// 	limiter := patterns.NewRateLimiter(cfg.RateLimiter)

// 	// проверяем что burst запросов проходят мгновенно
// 	start := time.Now()
// 	for i := 0; i < cfg.RateLimiter.Burst; i++ {
// 		assert.True(t, limiter.Allow())
// 	}
// 	burstTime := time.Since(start)
// 	assert.Less(t, burstTime, 100*time.Millisecond)

// 	// следующий запрос должен быть отклонен
// 	assert.False(t, limiter.Allow())
// }

// // TestRateLimiter_WithRealLoad проверяет лимитер под реальной нагрузкой через Toxiproxy
// func TestRateLimiter_WithRealLoad(t *testing.T) {
// 	cfg := config.MustLoad()

// 	// сбрасываем состояние Toxiproxy
// 	resetClient := toxiproxyLib.NewClient(cfg.GetToxiproxyAddress())
// 	resetClient.ResetState()

// 	// получаем порт из конфига
// 	serverPort := cfg.RateLimiter.TestPort

// 	// запускаем тестовый сервер
// 	server := startTestServer(t, serverPort)
// 	defer server.Close()

// 	// создаем уникальное имя прокси
// 	proxyName := fmt.Sprintf("ratelimit-test-%d", time.Now().UnixNano())

// 	// upstream указывает на наш сервер
// 	upstream := fmt.Sprintf("localhost:%d", serverPort)

// 	// создаем клиент Toxiproxy
// 	toxi, err := toxiproxy.NewToxiClient(cfg.Toxiproxy, proxyName, upstream)
// 	require.NoError(t, err)
// 	defer toxi.Close()

// 	// добавляем небольшую задержку чтобы запросы не были мгновенными
// 	err = toxi.AddLatency(100, 0)
// 	require.NoError(t, err)

// 	// создаем HTTP клиент с прохождением через прокси
// 	httpClient := &http.Client{
// 		Transport: &http.Transport{
// 			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
// 				return net.Dial("tcp", toxi.GetProxyAddress())
// 			},
// 		},
// 	}

// 	// создаем RateLimiter
// 	limiter := patterns.NewRateLimiter(cfg.RateLimiter)

// 	var wg sync.WaitGroup
// 	start := time.Now()
// 	successCount := 0
// 	var mu sync.Mutex

// 	// отправляем больше запросов чем лимит (3 * RPS)
// 	totalRequests := cfg.RateLimiter.RPS * 3

// 	for i := 0; i < totalRequests; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()

// 			// ждем разрешения от лимитера
// 			if err := limiter.Wait(context.Background()); err != nil {
// 				return
// 			}

// 			// реальный HTTP запрос через прокси
// 			resp, err := httpClient.Get("http://" + toxi.GetProxyAddress() + "/test")
// 			if err == nil {
// 				mu.Lock()
// 				successCount++
// 				mu.Unlock()
// 				resp.Body.Close()
// 			}
// 		}()
// 	}

// 	wg.Wait()
// 	elapsed := time.Since(start)

// 	// проверка: totalRequests / RPS = минимальное время
// 	minExpected := time.Duration(totalRequests/cfg.RateLimiter.RPS) * time.Second
// 	assert.GreaterOrEqual(t, elapsed, minExpected,
// 		"Лимитер пропустил запросы слишком быстро: %v (минимум ожидалось: %v)", elapsed, minExpected)

// 	t.Logf("Total time: %v (RPS: %d, requests: %d, success: %d)",
// 		elapsed, cfg.RateLimiter.RPS, totalRequests, successCount)
// }
