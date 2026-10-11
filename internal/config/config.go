package config

import (
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// ToxiproxyConfig хранит настройки Toxiproxy
type ToxiproxyConfig struct {
	Host      string `env:"TOXIPROXY_HOST" env-default:"localhost"`
	APIPort   int    `env:"TOXIPROXY_API_PORT" env-default:"8474"`
	ProxyPort int    `env:"TOXIPROXY_PROXY_PORT" env-default:"8666"`
}

// RetryConfig хранит настройки паттерна Retry
type RetryConfig struct {
	MaxAttempts  int           `env:"RETRY_MAX_ATTEMPTS"    env-default:"3"`
	InitialDelay time.Duration `env:"RETRY_INITIAL_DELAY"   env-default:"100ms"`
	MaxDelay     time.Duration `env:"RETRY_MAX_DELAY"       env-default:"5s"`
	Multiplier   float64       `env:"RETRY_MULTIPLIER"      env-default:"2.0"`
	Jitter       bool          `env:"RETRY_JITTER"          env-default:"true"`
}

// CircuitBreakerConfig хранит настройки паттерна Circuit Breaker
type CircuitBreakerConfig struct {
	Threshold       int           `env:"CB_THRESHOLD"          env-default:"3"`
	HalfOpenSuccess int           `env:"CB_HALF_OPEN_SUCCESS"  env-default:"2"`
	Timeout         time.Duration `env:"CB_TIMEOUT"            env-default:"5s"`
}

// TimeoutConfig хранит настройки паттерна Timeout
type TimeoutConfig struct {
	DefaultTimeout time.Duration `env:"TIMEOUT_DEFAULT" env-default:"1s"`
}

// BulkheadConfig хранит настройки паттерна Bulkhead
type BulkheadConfig struct {
	MaxConcurrent int64 `env:"BULKHEAD_MAX_CONCURRENT" env-default:"5"`
}

// RateLimiterConfig хранит настройки паттерна Rate Limiter
type RateLimiterConfig struct {
	RPS   float64 `env:"RATE_LIMIT_RPS"   env-default:"10"`
	Burst int     `env:"RATE_LIMIT_BURST" env-default:"20"`
}

// LogConfig хранит настройки логирования
type LogConfig struct {
	Level  string `env:"LOG_LEVEL" env-default:"info"`
	Format string `env:"LOG_FORMAT" env-default:"text"`
}

// ServerConfig хранит настройки демонстрационного HTTP-сервера
type ServerConfig struct {
	Host           string        `env:"SERVER_HOST" env-default:"0.0.0.0"`
	Port           int           `env:"SERVER_PORT" env-default:"8080"`
	DefaultSlowMs  int           `env:"SERVER_DEFAULT_SLOW_MS" env-default:"500"`
	DefaultErrorRt float64       `env:"SERVER_DEFAULT_ERROR_RATE" env-default:"0.3"`
	Middleware     bool          `env:"SERVER_MIDDLEWARE" env-default:"false"`
	RateLimitRPS   int           `env:"SERVER_RATE_LIMIT_RPS" env-default:"100"`
	RateLimitBurst int           `env:"SERVER_RATE_LIMIT_BURST" env-default:"200"`
	BulkheadMax    int64         `env:"SERVER_BULKHEAD_MAX" env-default:"50"`
	BulkheadWait   time.Duration `env:"SERVER_BULKHEAD_WAIT" env-default:"100ms"`
	RequestTimeout time.Duration `env:"SERVER_REQUEST_TIMEOUT" env-default:"5s"`
}

// RunnerConfig хранит настройки раннера сценариев
type RunnerConfig struct {
	ScenariosDir string `env:"RUNNER_SCENARIOS_DIR" env-default:"scenarios"`
	ResultsDir   string `env:"RUNNER_RESULTS_DIR" env-default:"results"`
	FixturesDir  string `env:"RUNNER_FIXTURES_DIR" env-default:"fixtures"`
}

// Config объединяет все настройки приложения
type Config struct {
	Toxiproxy      *ToxiproxyConfig
	Retry          *RetryConfig
	CircuitBreaker *CircuitBreakerConfig
	Timeout        *TimeoutConfig
	Bulkhead       *BulkheadConfig
	RateLimiter    *RateLimiterConfig
	Log            *LogConfig
	Server         *ServerConfig
	Runner         *RunnerConfig
}

// MustLoad загружает конфигурацию из .env файла
func MustLoad() *Config {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("Предупреждение: не удалось определить путь через runtime.Caller")
	} else {
		configDir := filepath.Dir(currentFile)
		internalDir := filepath.Dir(configDir)
		projectRoot := filepath.Dir(internalDir)

		envPath := filepath.Join(projectRoot, "configs", ".env")

		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Предупреждение: .env не загружен по пути %s: %v", envPath, err)
		}
	}

	var toxiproxyCfg ToxiproxyConfig
	var retryCfg RetryConfig
	var circuitBreakerCfg CircuitBreakerConfig
	var timeoutCfg TimeoutConfig
	var bulkheadCfg BulkheadConfig
	var rateLimiterCfg RateLimiterConfig
	var logCfg LogConfig
	var serverCfg ServerConfig
	var runnerCfg RunnerConfig

	if err := cleanenv.ReadEnv(&toxiproxyCfg); err != nil {
		log.Fatalln("Ошибка чтения настроек Toxiproxy из .env файла:", err)
	}
	if err := cleanenv.ReadEnv(&retryCfg); err != nil {
		log.Fatalln("Ошибка чтения настроек Retry из .env файла:", err)
	}
	if err := cleanenv.ReadEnv(&circuitBreakerCfg); err != nil {
		log.Fatalln("Ошибка чтения настроек Circuit Breaker из .env файла:", err)
	}
	if err := cleanenv.ReadEnv(&timeoutCfg); err != nil {
		log.Fatalln("Ошибка чтения настроек Timeout из .env файла:", err)
	}
	if err := cleanenv.ReadEnv(&bulkheadCfg); err != nil {
		log.Fatalln("Ошибка чтения настроек Bulkhead из .env файла:", err)
	}
	if err := cleanenv.ReadEnv(&rateLimiterCfg); err != nil {
		log.Fatalln("Ошибка чтения настроек Rate Limiter из .env файла:", err)
	}
	if err := cleanenv.ReadEnv(&logCfg); err != nil {
		log.Fatalln("Ошибка чтения настроек логирования из .env файла:", err)
	}
	if err := cleanenv.ReadEnv(&serverCfg); err != nil {
		log.Fatalln("Ошибка чтения настроек Server из .env файла:", err)
	}
	if err := cleanenv.ReadEnv(&runnerCfg); err != nil {
		log.Fatalln("Ошибка чтения настроек Runner из .env файла:", err)
	}

	return &Config{
		Toxiproxy:      &toxiproxyCfg,
		Retry:          &retryCfg,
		CircuitBreaker: &circuitBreakerCfg,
		Timeout:        &timeoutCfg,
		Bulkhead:       &bulkheadCfg,
		RateLimiter:    &rateLimiterCfg,
		Log:            &logCfg,
		Server:         &serverCfg,
		Runner:         &runnerCfg,
	}
}

// GetToxiproxyAddress возвращает полный адрес Toxiproxy API
func (c *Config) GetToxiproxyAddress() string {
	return c.Toxiproxy.Host + ":" + strconv.Itoa(c.Toxiproxy.APIPort)
}
