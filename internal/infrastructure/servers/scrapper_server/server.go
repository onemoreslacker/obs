package scrapperserver

import (
	"net"
	"net/http"
	"time"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

// New instantiates a new http.Server entity.
func New(cfg *config.Config, api *scrapperapi.API) *http.Server {
	mux := http.NewServeMux()

	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Serving.ScrapperHost, cfg.Serving.ScrapperPort),
		Handler:           scrapperapi.HandlerFromMux(api, mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
}
