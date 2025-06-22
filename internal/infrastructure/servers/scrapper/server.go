package scrapperserver

import (
	"net"
	"net/http"

	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
	"github.com/es-debug/backend-academy-2024-go-template/config"
	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
)

func New(
	cfg *config.Config,
	api *scrapperapi.API,
	lmt *limiter.Limiter,
) *http.Server {
	handler := scrapperapi.HandlerFromMux(api, http.NewServeMux())
	rateLimitedHandler := tollbooth.LimitHandler(lmt, handler)

	return &http.Server{
		Addr:         net.JoinHostPort(cfg.Serving.ScrapperHost, cfg.Serving.ScrapperPort),
		Handler:      rateLimitedHandler,
		ReadTimeout:  cfg.TimeoutPolicy.ServerWrite,
		WriteTimeout: cfg.TimeoutPolicy.ServerRead,
		IdleTimeout:  cfg.TimeoutPolicy.ServerIdle,
	}
}
