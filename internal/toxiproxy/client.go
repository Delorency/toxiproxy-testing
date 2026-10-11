package toxiproxy

import (
	"fmt"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"

	"pr4/internal/config"
)

// ToxiClient управляет подключением к Toxiproxy
type ToxiClient struct {
	client *toxiproxy.Client
	proxy  *toxiproxy.Proxy
	name   string
	cfg    *config.ToxiproxyConfig
}

// NewToxiClient создаёт клиент Toxiproxy и прокси с указанным upstream
// upstream это адрес реального сервера, например "server:8080" или "localhost:8080"
func NewToxiClient(cfg *config.ToxiproxyConfig, proxyName string, upstream string) (*ToxiClient, error) {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.APIPort)
	client := toxiproxy.NewClient(address)

	// удаляем существующий прокси с таким именем, если есть
	existing, _ := client.Proxy(proxyName)
	if existing != nil {
		_ = existing.Delete()
	}

	proxyAddress := fmt.Sprintf("0.0.0.0:%d", cfg.ProxyPort)
	proxy, err := client.CreateProxy(proxyName, proxyAddress, upstream)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy: %w", err)
	}

	return &ToxiClient{
		client: client,
		proxy:  proxy,
		name:   proxyName,
		cfg:    cfg,
	}, nil
}

// GetProxyAddress возвращает адрес, по которому клиент подключается к прокси
func (tc *ToxiClient) GetProxyAddress() string {
	return tc.proxy.Listen
}

// AddLatency добавляет токсин задержки
func (tc *ToxiClient) AddLatency(latencyMs int64, jitterMs int64) error {
	_, err := tc.proxy.AddToxic(
		"latency_toxic",
		"latency",
		"downstream",
		1.0,
		toxiproxy.Attributes{
			"latency": latencyMs,
			"jitter":  jitterMs,
		},
	)
	return err
}

// AddTimeout добавляет токсин тайм-аута
func (tc *ToxiClient) AddTimeout(timeoutMs int) error {
	_, err := tc.proxy.AddToxic(
		"timeout_toxic",
		"timeout",
		"downstream",
		1.0,
		toxiproxy.Attributes{
			"timeout": timeoutMs,
		},
	)
	return err
}

// AddBandwidthLimit добавляет токсин ограничения пропускной способности
func (tc *ToxiClient) AddBandwidthLimit(rateKBps int64) error {
	_, err := tc.proxy.AddToxic(
		"bandwidth_toxic",
		"bandwidth",
		"downstream",
		1.0,
		toxiproxy.Attributes{
			"rate": rateKBps,
		},
	)
	return err
}

// AddResetPeer добавляет токсин мгновенного сброса соединения (RST)
func (tc *ToxiClient) AddResetPeer(timeoutMs int) error {
	_, err := tc.proxy.AddToxic(
		"reset_peer_toxic",
		"reset_peer",
		"downstream",
		1.0,
		toxiproxy.Attributes{
			"timeout": timeoutMs,
		},
	)
	return err
}

// AddSlowClose добавляет токсин медленного закрытия соединения
func (tc *ToxiClient) AddSlowClose(delayMs int) error {
	_, err := tc.proxy.AddToxic(
		"slow_close_toxic",
		"slow_close",
		"downstream",
		1.0,
		toxiproxy.Attributes{
			"delay": delayMs,
		},
	)
	return err
}

// RemoveToxic удаляет токсин по имени
func (tc *ToxiClient) RemoveToxic(name string) error {
	return tc.proxy.RemoveToxic(name)
}

// Reset снимает все токсины с прокси
func (tc *ToxiClient) Reset() error {
	// RemoveToxic возвращает ошибку, если токсина нет.
	// Перебираем известные имена и игнорируем ошибки.
	names := []string{
		"latency_toxic",
		"timeout_toxic",
		"bandwidth_toxic",
		"reset_peer_toxic",
		"slow_close_toxic",
	}
	for _, n := range names {
		_ = tc.proxy.RemoveToxic(n)
	}
	return nil
}

// Close удаляет прокси
func (tc *ToxiClient) Close() error {
	return tc.proxy.Delete()
}

// ResetState сбрасывает состояние всех прокси и токсинов во всём Toxiproxy
func (tc *ToxiClient) ResetState() error {
	return tc.client.ResetState()
}
