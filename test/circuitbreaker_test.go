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

// // TestCircuitBreaker_OpensAfterFailures проверяет что выключатель открывается после N ошибок
// func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
// 	cfg := config.MustLoad()

// 	// сбрасываем состояние Toxiproxy
// 	resetClient := toxiproxyLib.NewClient(cfg.GetToxiproxyAddress())
// 	resetClient.ResetState()

// 	// получаем порт из конфига
// 	serverPort := cfg.CircuitBreaker.TestPort

// 	server := startTestServer(t, serverPort)
// 	defer server.Close()

// 	// создаем уникальное имя прокси
// 	proxyName := fmt.Sprintf("cb-test-%d", time.Now().UnixNano())

// 	// upstream указывает на наш сервер
// 	upstream := fmt.Sprintf("localhost:%d", serverPort)

// 	// создаем клиент Toxiproxy
// 	toxi, err := toxiproxy.NewToxiClient(cfg.Toxiproxy, proxyName, upstream)
// 	require.NoError(t, err)
// 	defer toxi.Close()

// 	// добавляем тайм-аут 100 миллисекунд (запросы будут падать)
// 	err = toxi.AddTimeout(100)
// 	require.NoError(t, err)

// 	// создаем HTTP клиент с прохождением через прокси
// 	httpClient := &http.Client{
// 		Timeout: 5 * time.Second,
// 		Transport: &http.Transport{
// 			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
// 				return net.Dial("tcp", toxi.GetProxyAddress())
// 			},
// 		},
// 	}

// 	// создаем автоматический выключатель
// 	cb := patterns.NewCircuitBreaker(cfg.CircuitBreaker)

// 	// делаем Threshold запроса которые должны упасть из-за тайм-аута
// 	for i := 1; i <= cfg.CircuitBreaker.Threshold; i++ {
// 		err := cb.Call(func() error {
// 			resp, err := httpClient.Get("http://" + toxi.GetProxyAddress() + "/test")
// 			if err != nil {
// 				return err
// 			}
// 			defer resp.Body.Close()
// 			return nil
// 		})
// 		assert.Error(t, err, "Запрос %d должен упасть", i)
// 	}

// 	// четвертый запрос должен упасть мгновенно с ошибкой открытого выключателя
// 	start := time.Now()
// 	err = cb.Call(func() error {
// 		// даже если функция вернет nil, выключатель уже открыт
// 		return nil
// 	})
// 	elapsed := time.Since(start)

// 	assert.Equal(t, patterns.ErrCircuitOpen, err, "Должна быть ошибка открытого выключателя")
// 	assert.Less(t, elapsed, 50*time.Millisecond, "Ошибка должна быть мгновенной")
// }
