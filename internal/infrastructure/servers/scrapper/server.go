package scrapperserver

import (
	"net"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
)

func New(cfg config.Serving, api *scrapperapi.API) *http.Server {
	return &http.Server{
		Addr:    net.JoinHostPort(cfg.ScrapperHost, cfg.ScrapperPort),
		Handler: scrapperapi.HandlerFromMux(api, http.NewServeMux()),
	}
}
