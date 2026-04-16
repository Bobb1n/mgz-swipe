package main

import (
	"context"
	"log/slog"
	"os"

	"swipe-mgz/internal/config"
	"swipe-mgz/internal/events"
	"swipe-mgz/internal/repository"
	httphandler "swipe-mgz/internal/transport/http"
	"swipe-mgz/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("postgres connect", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		slog.Error("postgres ping", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer func() { _ = rdb.Close() }()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("redis ping", "error", err)
		os.Exit(1)
	}

	publisher := events.NewPublisher(cfg.KafkaBrokers)
	defer func() { _ = publisher.Close() }()

	swipeRepo := repository.NewSwipeRepo(db)
	matchRepo := repository.NewMatchRepo(db)
	locationRepo := repository.NewLocationRepo(rdb)

	uc := usecase.NewSwipeUseCase(swipeRepo, matchRepo, locationRepo, publisher, cfg.GeoRadiusKm)

	e := echo.New()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORS())

	httphandler.NewHandler(uc).RegisterRoutes(e)

	slog.Info("swipe-mgz starting", "port", cfg.ServerPort)
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		slog.Error("server", "error", err)
		os.Exit(1)
	}
}
