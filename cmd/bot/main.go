package main

import (
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"log/slog"
	"os"

	botservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot_service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error(
			"config was not loaded",
			slog.String("msg", err.Error()),
			slog.String("service", "cfg"),
		)
		os.Exit(1)
	}

	service, err := botservice.New(cfg)
	if err != nil {
		slog.Error(
			"service was not initialized",
			slog.String("msg", err.Error()),
			slog.String("service", "scrapper"),
		)
		os.Exit(1)
	}

	if err := service.Run(); err != nil {
		slog.Error(
			"service job was failed",
			slog.String("msg", err.Error()),
			slog.String("service", "scrapper"),
		)
		os.Exit(1)
	}
}
