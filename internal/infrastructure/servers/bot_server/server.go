package botserver

import (
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"net"
	"net/http"
	"time"

	bot "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/bot_api"
)

// New instantiates a new http.Server entity.
func New(cfg *config.Config, api *bot.API) *http.Server {
	mux := http.NewServeMux()

	// TODO: identify an appropriate ReadHeaderTimeout.
	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Serving.Host, cfg.Serving.BotPort),
		Handler:           bot.HandlerFromMux(api, mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
}
