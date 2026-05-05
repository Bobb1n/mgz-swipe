package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"swipe-mgz/internal/config"
	"swipe-mgz/internal/events"
	"swipe-mgz/internal/repository"
	grpctransport "swipe-mgz/internal/transport/grpc"
	httptransport "swipe-mgz/internal/transport/http"
	"swipe-mgz/internal/usecase"
	swipev1 "swipe-mgz/pkg/api/swipe/v1"
	pgconn "swipe-mgz/pkg/postgres"
	rdconn "swipe-mgz/pkg/redis"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type App struct {
	cfg       *config.Config
	log       *slog.Logger
	pool      *pgxpool.Pool
	rdb       *redis.Client
	publisher *events.Publisher
	echo      *echo.Echo
	httpSrv   *http.Server
	grpcSrv   *grpc.Server
	grpcLis   net.Listener
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
	pool, err := pgconn.NewPoolFromURL(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	rdb, err := rdconn.New(ctx, rdconn.Config{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("redis: %w", err)
	}

	publisher := events.NewPublisher(cfg.KafkaBrokers)

	swipeRepo := repository.NewSwipeRepo(pool)
	matchRepo := repository.NewMatchRepo(pool)
	locationRepo := repository.NewLocationRepo(rdb)

	uc := usecase.NewSwipeUseCase(swipeRepo, matchRepo, locationRepo, publisher, cfg.GeoRadiusKm)

	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.Use(echomw.CORS())
	httptransport.NewHandler(uc).RegisterRoutes(e)

	httpSrv := &http.Server{
		Addr:              net.JoinHostPort("", cfg.ServerPort),
		Handler:           e,
		ReadHeaderTimeout: 5 * time.Second,
	}

	grpcSrv := grpc.NewServer()
	swipev1.RegisterSwipeServiceServer(grpcSrv, grpctransport.NewServer(uc))

	grpcLis, err := net.Listen("tcp", net.JoinHostPort("", cfg.GRPCPort))
	if err != nil {
		pool.Close()
		_ = rdb.Close()
		_ = publisher.Close()
		return nil, fmt.Errorf("grpc listen: %w", err)
	}

	return &App{
		cfg:       cfg,
		log:       log,
		pool:      pool,
		rdb:       rdb,
		publisher: publisher,
		echo:      e,
		httpSrv:   httpSrv,
		grpcSrv:   grpcSrv,
		grpcLis:   grpcLis,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		a.log.Info("http server started", "port", a.cfg.ServerPort)
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http: %w", err)
		}
	}()

	go func() {
		a.log.Info("grpc server started", "port", a.cfg.GRPCPort)
		if err := a.grpcSrv.Serve(a.grpcLis); err != nil {
			errCh <- fmt.Errorf("grpc: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	a.log.Info("shutting down")

	httpErr := a.httpSrv.Shutdown(ctx)

	stopped := make(chan struct{})
	go func() {
		a.grpcSrv.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-ctx.Done():
		a.grpcSrv.Stop()
	}

	_ = a.publisher.Close()
	_ = a.rdb.Close()
	a.pool.Close()
	return httpErr
}
