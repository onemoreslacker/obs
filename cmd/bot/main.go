package main

import (
	"log/slog"

	botservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot_service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error(err.Error())
	}

	service, err := botservice.New(cfg)
	if err != nil {
		slog.Error(err.Error())
	}

	if err := service.Run(); err != nil {
		slog.Error(err.Error())
	}
}
