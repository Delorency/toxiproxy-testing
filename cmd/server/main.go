package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pr4/internal/config"
	"pr4/internal/server"
)

func main() {
	cfg := config.MustLoad()

	srv := server.New(server.Config{
		Host:           cfg.Server.Host,
		Port:           cfg.Server.Port,
		DefaultSlowMs:  cfg.Server.DefaultSlowMs,
		DefaultErrorRt: cfg.Server.DefaultErrorRt,
	})

	// SERVER_MIDDLEWARE=true/false
	var mw []server.Middleware
	if cfg.Server.Middleware {
		mw = append(mw,
			server.RateLimitMiddleware(cfg.Server.RateLimitRPS, cfg.Server.RateLimitBurst),
			server.BulkheadMiddleware(cfg.Server.BulkheadMax, cfg.Server.BulkheadWait),
			server.TimeoutMiddleware(cfg.Server.RequestTimeout),
		)
	}

	go func() {
		log.Printf("server listening on :%d (middleware=%v)", cfg.Server.Port, cfg.Server.Middleware)
		if err := srv.Start(mw...); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("server stopped")
}
