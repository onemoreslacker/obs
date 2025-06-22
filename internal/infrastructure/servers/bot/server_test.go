package botserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	botserver "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/servers/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/didip/tollbooth/v8/limiter"
	"github.com/es-debug/backend-academy-2024-go-template/config"
)

func TestRateLimitIntegration(t *testing.T) {
	cfg := &config.Config{
		Serving: config.Serving{
			BotHost: "127.0.0.1",
			BotPort: "8080",
		},
		TimeoutPolicy: config.Timeouts{
			ServerRead:  5 * time.Second,
			ServerWrite: 5 * time.Second,
			ServerIdle:  2 * time.Second,
		},
		RateLimiter: config.RateLimiter{
			TTL:          time.Hour,
			MaxPerSecond: 5.0,
			Burst:        10,
			Methods:      []string{"GET", "POST", "PUT", "DELETE"},
		},
	}

	api := mocks.NewMockServerInterface(t)
	api.On("PostUpdates", mock.Anything, mock.Anything).
		Return()

	lmt := limiter.New(&limiter.ExpirableOptions{DefaultExpirationTTL: cfg.RateLimiter.TTL})
	lmt.SetMax(cfg.RateLimiter.MaxPerSecond)
	lmt.SetBurst(cfg.RateLimiter.Burst)
	lmt.SetMethods(cfg.RateLimiter.Methods)
	lmt.SetIPLookup(
		limiter.IPLookup{
			Name:           "RemoteAddr",
			IndexFromRight: 0,
		},
	)

	srv := botserver.New(cfg, api, lmt)
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	client := &http.Client{}
	path := ts.URL + "/updates"

	for range cfg.RateLimiter.Burst {
		req, err := http.NewRequest("POST", path, http.NoBody)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	req, err := http.NewRequest("POST", path, http.NoBody)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
