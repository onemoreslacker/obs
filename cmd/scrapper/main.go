package main

import (
	"log/slog"

	scrapperservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper_service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		slog.Error(
			"config init error",
			slog.String("msg", err.Error()),
		)
	}

	service, err := scrapperservice.New(cfg)
	if err != nil {
		slog.Error(
			"scrapper init error",
			slog.String("msg", err.Error()),
		)
	}

	if err := service.Run(); err != nil {
		slog.Error(
			"scrapper service error",
			slog.String("msg", err.Error()),
		)
	}
}
