package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"swipe-mgz/internal/app"
	"swipe-mgz/internal/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(rootCtx, cfg, log)
	if err != nil {
		log.Error("init app", "error", err)
		os.Exit(1)
	}

	runErr := make(chan error, 1)
	go func() { runErr <- a.Run(rootCtx) }()

	select {
	case <-rootCtx.Done():
		log.Info("signal received")
	case err := <-runErr:
		if err != nil {
			log.Error("run", "error", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "error", err)
		os.Exit(1)
	}
	log.Info("server stopped")
}
