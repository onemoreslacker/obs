package cbtransport

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
	"github.com/sony/gobreaker/v2"
)

type Config struct {
	MaxAttempts         uint
	RetryDelay          time.Duration
	RetryableCodes      []int
	AdditionalRetryOpts []retry.Option
}

type CircuitBreakerTransport struct {
	transport      http.RoundTripper
	cb             *gobreaker.CircuitBreaker[*http.Response]
	retryableCodes []int
	opts           []retry.Option
}

func New(
	baseTransport http.RoundTripper,
	cb *gobreaker.CircuitBreaker[*http.Response],
	cfg *Config,
) *CircuitBreakerTransport {
	return &CircuitBreakerTransport{
		transport:      baseTransport,
		cb:             cb,
		retryableCodes: cfg.RetryableCodes,
		opts: append([]retry.Option{
			retry.Attempts(cfg.MaxAttempts),
			retry.Delay(cfg.RetryDelay),
			retry.LastErrorOnly(true),
		}, cfg.AdditionalRetryOpts...),
	}
}

func (c *CircuitBreakerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	return retry.DoWithData[*http.Response](
		func() (*http.Response, error) {
			result, err := c.cb.Execute(func() (*http.Response, error) {
				return c.transport.RoundTrip(req)
			})
			if err != nil {
				return nil, err
			}

			if result.StatusCode < 400 {
				return result, nil
			}

			if c.shouldRetry(result.StatusCode) {
				return nil, fmt.Errorf("HTTP %d (retryable)", result.StatusCode)
			}

			return nil, clients.ErrNonRetryableRequest
		},
		append(c.opts, retry.Context(ctx), retry.RetryIf(func(err error) bool {
			return !errors.Is(err, clients.ErrNonRetryableRequest)
		}))...,
	)
}

func (c *CircuitBreakerTransport) shouldRetry(statusCode int) bool {
	return slices.Contains(c.retryableCodes, statusCode)
}
