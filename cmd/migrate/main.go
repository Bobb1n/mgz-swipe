package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"swipe-mgz/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	direction := flag.String("dir", "up", "migration direction: up | down")
	steps := flag.Int("steps", 0, "number of steps for up/down (0 = all)")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	migrationsDir := os.Getenv("MIGRATIONS_PATH")
	if migrationsDir == "" {
		migrationsDir = "internal/migrations"
	}
	abs, err := filepath.Abs(migrationsDir)
	if err != nil {
		log.Error("migrations path", "error", err)
		os.Exit(1)
	}
	srcURL := "file://" + filepath.ToSlash(abs)

	m, err := migrate.New(srcURL, cfg.DatabaseURL)
	if err != nil {
		log.Error("migrate new", "error", err, "source", srcURL)
		os.Exit(1)
	}
	defer func() { _, _ = m.Close() }()

	if err := run(m, *direction, *steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("migrate", "dir", *direction, "error", err)
		os.Exit(1)
	}
	log.Info("migrations applied", "dir", *direction, "steps", *steps, "source", srcURL)
}

func run(m *migrate.Migrate, dir string, steps int) error {
	switch dir {
	case "up":
		if steps > 0 {
			return m.Steps(steps)
		}
		return m.Up()
	case "down":
		if steps > 0 {
			return m.Steps(-steps)
		}
		return m.Down()
	default:
		return fmt.Errorf("dir must be 'up' or 'down'")
	}
}
