package botserver

import (
	"net"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/bot"
)

func New(cfg config.Serving, api *botapi.API) *http.Server {
	return &http.Server{
		Addr:    net.JoinHostPort(cfg.BotHost, cfg.BotPort),
		Handler: botapi.HandlerFromMux(api, http.NewServeMux()),
	}
}
