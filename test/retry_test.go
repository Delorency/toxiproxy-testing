package test

// import (
// 	"context"
// 	"fmt"
// 	"net"
// 	"net/http"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/require"

// 	toxiproxyLib "github.com/Shopify/toxiproxy/v2/client"

// 	"pr4/internal/config"
// 	"pr4/internal/patterns"
// 	"pr4/internal/toxiproxy"
// )

// // TestRetry_WithTimeout проверяет что Retry делает повторные попытки при тайм-ауте
// func TestRetry_WithTimeout(t *testing.T) {
// 	cfg := config.MustLoad()

// 	// сбрасываем состояние Toxiproxy
// 	resetClient := toxiproxyLib.NewClient(cfg.GetToxiproxyAddress())
// 	resetClient.ResetState()

// 	serverPort := cfg.Retry.TestPort
// 	server := startTestServer(t, serverPort)
// 	defer server.Close()

// 	// создаем уникальное имя прокси
// 	proxyName := fmt.Sprintf("retry-test-%d", time.Now().UnixNano())

// 	upstream := fmt.Sprintf("localhost:%d", serverPort)

// 	// создаем клиент Toxiproxy
// 	toxi, err := toxiproxy.NewToxiClient(cfg.Toxiproxy, proxyName, upstream)
// 	require.NoError(t, err)
// 	defer toxi.Close()

// 	// добавляем тайм-аут 10 миллисекунд
// 	err = toxi.AddTimeout(10)
// 	require.NoError(t, err)

// 	// создаем HTTP клиент с прохождением через прокси
// 	client := &http.Client{
// 		Timeout: 5 * time.Second,
// 		Transport: &http.Transport{
// 			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
// 				return net.Dial("tcp", toxi.GetProxyAddress())
// 			},
// 		},
// 	}

// 	attemptCount := 0

// 	// выполняем запрос с повторами
// 	err = patterns.Retry(context.Background(), cfg.Retry, func() error {
// 		attemptCount++
// 		resp, err := client.Get("http://" + toxi.GetProxyAddress() + "/test")
// 		if err != nil {
// 			return err
// 		}
// 		defer resp.Body.Close()
// 		return nil
// 	})

// 	t.Logf("Attempts: %d, err: %v", attemptCount, err)
// }
