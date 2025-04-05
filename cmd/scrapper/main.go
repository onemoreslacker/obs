package main

import (
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"log/slog"
	"os"

	scrapperservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper_service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error(
			"config init error",
			slog.String("msg", err.Error()),
			slog.String("service", "cfg"),
		)
		os.Exit(1)
	}

	service, err := scrapperservice.New(cfg)
	if err != nil {
		slog.Error(
			"scrapper init error",
			slog.String("msg", err.Error()),
			slog.String("service", "cfg"),
		)
		os.Exit(1)
	}

	if err := service.Run(); err != nil {
		slog.Error(
			"scrapper service error",
			slog.String("msg", err.Error()),
			slog.String("service", "cfg"),
		)
		os.Exit(1)
	}
}
