package test

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"net"
// 	"net/http"
// 	"sync/atomic"
// 	"testing"
// 	"time"

// 	toxiproxyLib "github.com/Shopify/toxiproxy/v2/client"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	"pr4/internal/config"
// 	"pr4/internal/patterns"
// 	"pr4/internal/toxiproxy"
// )

// // TestBulkhead_WithBandwidthLimit проверяет что Bulkhead ограничивает параллельные запросы
// func TestBulkhead_WithBandwidthLimit(t *testing.T) {
// 	cfg := config.MustLoad()

// 	// сбрасываем состояние Toxiproxy
// 	resetClient := toxiproxyLib.NewClient(cfg.GetToxiproxyAddress())
// 	resetClient.ResetState()

// 	// получаем порт из конфига
// 	serverPort := cfg.Bulkhead.TestPort

// 	// запускаем тестовый сервер
// 	server := startTestServer(t, serverPort)
// 	defer server.Close()

// 	// создаем уникальное имя прокси
// 	proxyName := fmt.Sprintf("bulkhead-test-%d", time.Now().UnixNano())

// 	// upstream указывает на наш сервер
// 	upstream := fmt.Sprintf("localhost:%d", serverPort)

// 	// создаем клиент Toxiproxy
// 	toxi, err := toxiproxy.NewToxiClient(cfg.Toxiproxy, proxyName, upstream)
// 	require.NoError(t, err)
// 	defer toxi.Close()

// 	// добавляем ограничение скорости из конфига
// 	bandwidthKBps := cfg.Bulkhead.BandwidthKBps
// 	err = toxi.AddBandwidthLimit(bandwidthKBps)
// 	require.NoError(t, err)

// 	// создаем HTTP клиент с прохождением через прокси
// 	httpClient := &http.Client{
// 		Transport: &http.Transport{
// 			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
// 				return net.Dial("tcp", toxi.GetProxyAddress())
// 			},
// 		},
// 	}

// 	// создаем Bulkhead с лимитом из конфига
// 	bulkhead := patterns.NewBulkhead(cfg.Bulkhead)

// 	var active int32
// 	var maxActive int32

// 	// запускаем горутины (количество из конфига)
// 	concurrentRequests := cfg.Bulkhead.ConcurrentRequests
// 	for range concurrentRequests {
// 		go func() {
// 			err := bulkhead.Execute(context.Background(), func() error {
// 				// увеличиваем счетчик активных запросов
// 				current := atomic.AddInt32(&active, 1)
// 				defer atomic.AddInt32(&active, -1)

// 				// обновляем максимальное значение
// 				for current > atomic.LoadInt32(&maxActive) {
// 					atomic.CompareAndSwapInt32(&maxActive, atomic.LoadInt32(&maxActive), current)
// 				}

// 				resp, err := httpClient.Get("http://" + toxi.GetProxyAddress() + "/test")
// 				if err != nil {
// 					return err
// 				}
// 				defer resp.Body.Close()

// 				// читаем тело чтобы bandwidth limit сработал
// 				_, err = io.Copy(io.Discard, resp.Body)
// 				return err
// 			})
// 			assert.NoError(t, err)
// 		}()
// 	}

// 	// даем время на выполнение (из конфига)
// 	testDuration := time.Duration(cfg.Bulkhead.TestDurationSec) * time.Second
// 	time.Sleep(testDuration)

// 	// проверяем что максимум параллельных запросов не превышает лимит
// 	maxActiveValue := atomic.LoadInt32(&maxActive)
// 	assert.LessOrEqual(t, maxActiveValue, int32(cfg.Bulkhead.MaxConcurrent),
// 		"Максимум параллельных запросов %d превышает лимит %d", maxActiveValue, cfg.Bulkhead.MaxConcurrent)

// 	t.Logf("Max concurrent requests: %d (limit: %d)", maxActiveValue, cfg.Bulkhead.MaxConcurrent)
// }
