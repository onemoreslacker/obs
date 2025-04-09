package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	scrapperservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper_service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/storage"
)

func main() {
	configFileName := flag.String("config", "config/config.yaml", "path to config file")

	flag.Parse()

	cfg, err := config.Load(*configFileName)
	if err != nil {
		slog.Error(
			"Scrapper: config was not loaded",
			slog.String("msg", err.Error()),
		)
		os.Exit(1)
	}

	var (
		pool *pgxpool.Pool
	)

	if cfg.Database.AccessType != "in-memory" {
		pool, err = storage.NewPool(cfg)
		if err != nil {
			slog.Error(
				"Scrapper: pool was not created",
				slog.String("msg", err.Error()),
			)
			os.Exit(1)
		}

		defer pool.Close()
	}

	repository, err := storage.New(cfg, pool)
	if err != nil {
		slog.Error(
			"Scrapper: repository was not initialized",
			slog.String("msg", err.Error()),
		)
		os.Exit(1) //nolint:gocritic // to fix.
	}

	service, err := scrapperservice.New(cfg, repository)
	if err != nil {
		slog.Error(
			"Scrapper: initialization error",
			slog.String("msg", err.Error()),
		)
		os.Exit(1)
	}

	if err := service.Run(); err != nil {
		slog.Error(
			"Scrapper: service is down",
			slog.String("msg", err.Error()),
		)
		os.Exit(1)
	}
}
