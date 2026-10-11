package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Config — параметры HTTP-сервера
type Config struct {
	Host           string
	Port           int
	DefaultSlowMs  int
	DefaultErrorRt float64
}

// Server — HTTP-сервер с эндпоинтами разной тяжести
type Server struct {
	cfg     Config
	httpSrv *http.Server
	mux     *http.ServeMux
}

// New создаёт сервер
func New(cfg Config) *Server {
	s := &Server{
		cfg: cfg,
		mux: http.NewServeMux(),
	}
	s.routes()
	return s
}

// routes регистрирует эндпоинты
func (s *Server) routes() {
	s.mux.HandleFunc("/fast", s.handleFast)
	s.mux.HandleFunc("/slow", s.handleSlow)
	s.mux.HandleFunc("/error", s.handleError)
	s.mux.HandleFunc("/random", s.handleRandom)
	s.mux.HandleFunc("/heavy", s.handleHeavy)
	s.mux.HandleFunc("/health", s.handleHealth)
}

// Handler возвращает http.Handler с middleware (в порядке применения)
func (s *Server) Handler(mw ...Middleware) http.Handler {
	var h http.Handler = s.mux
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// Start запускает сервер и блокируется до ошибки или остановки
func (s *Server) Start(mw ...Middleware) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	s.httpSrv = &http.Server{
		Addr:              addr,
		Handler:           s.Handler(mw...),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s.httpSrv.ListenAndServe()
}

// Shutdown корректно останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}

// --- handlers ---

// handleFast отвечает мгновенно
func (s *Server) handleFast(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "endpoint": "fast"})
}

// handleSlow отвечает с задержкой (?delay=500ms) или дефолтной
func (s *Server) handleSlow(w http.ResponseWriter, r *http.Request) {
	delay := time.Duration(s.cfg.DefaultSlowMs) * time.Millisecond
	if d := r.URL.Query().Get("delay"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			delay = parsed
		}
	}

	select {
	case <-time.After(delay):
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "endpoint": "slow"})
	case <-r.Context().Done():
		return
	}
}

// handleError всегда возвращает 500
func (s *Server) handleError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

// handleRandom отвечает 500 с заданной вероятностью (?error_rate=0.3)
func (s *Server) handleRandom(w http.ResponseWriter, r *http.Request) {
	rate := s.cfg.DefaultErrorRt
	if v := r.URL.Query().Get("error_rate"); v != "" {
		if _, err := fmt.Sscanf(v, "%f", &rate); err != nil {
			rate = s.cfg.DefaultErrorRt
		}
	}

	if randFloat() < rate {
		http.Error(w, "random failure", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "endpoint": "random"})
}

// handleHeavy имитирует нагрузку (?ms=200)
func (s *Server) handleHeavy(w http.ResponseWriter, r *http.Request) {
	ms := 200
	if v := r.URL.Query().Get("ms"); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &ms); err != nil {
			ms = 200
		}
	}

	select {
	case <-time.After(time.Duration(ms) * time.Millisecond):
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "endpoint": "heavy"})
	case <-r.Context().Done():
		return
	}
}

// handleHealth — для health-check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
