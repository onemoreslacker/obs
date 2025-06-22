package botserver

import (
	"net"
	"net/http"

	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
	"github.com/es-debug/backend-academy-2024-go-template/config"
	botapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/bot"
)

func New(
	cfg *config.Config,
	api botapi.ServerInterface,
	lmt *limiter.Limiter,
) *http.Server {
	handler := botapi.HandlerFromMux(api, http.NewServeMux())
	rateLimitedHandler := tollbooth.LimitHandler(lmt, handler)

	return &http.Server{
		Addr:         net.JoinHostPort(cfg.Serving.BotHost, cfg.Serving.BotPort),
		Handler:      rateLimitedHandler,
		ReadTimeout:  cfg.TimeoutPolicy.ServerWrite,
		WriteTimeout: cfg.TimeoutPolicy.ServerRead,
		IdleTimeout:  cfg.TimeoutPolicy.ServerIdle,
	}
}
