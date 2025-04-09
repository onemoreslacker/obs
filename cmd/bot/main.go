package main

import (
	"flag"
	"log/slog"
	"os"

	botservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot_service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

func main() {
	configFileName := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configFileName)
	if err != nil {
		slog.Error(
			"Bot: config was not loaded",
			slog.String("msg", err.Error()),
		)
		os.Exit(1)
	}

	service, err := botservice.New(cfg)
	if err != nil {
		slog.Error(
			"Bot: initialization error",
			slog.String("msg", err.Error()),
		)
		os.Exit(1)
	}

	if err := service.Run(); err != nil {
		slog.Error(
			"Bot: service is down",
			slog.String("msg", err.Error()),
		)
		os.Exit(1)
	}
}
