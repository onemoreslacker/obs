package cbtransport_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
	cbt "github.com/es-debug/backend-academy-2024-go-template/pkg/cbtransport"
	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreakerTransport_RoundTrip_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	transport := createTestTransport(t, []int{500, 502, 503})

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCircuitBreakerTransport_RoundTrip_RetryableError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		retryableCodes []int
		maxAttempts    uint
		shouldRetry    bool
	}{
		{
			name:           "500 is retryable",
			statusCode:     500,
			retryableCodes: []int{500, 502, 503},
			maxAttempts:    3,
			shouldRetry:    true,
		},
		{
			name:           "502 is retryable",
			statusCode:     502,
			retryableCodes: []int{500, 502, 503},
			maxAttempts:    3,
			shouldRetry:    true,
		},
		{
			name:           "503 is retryable",
			statusCode:     503,
			retryableCodes: []int{500, 502, 503},
			maxAttempts:    3,
			shouldRetry:    true,
		},
		{
			name:           "404 is not retryable",
			statusCode:     404,
			retryableCodes: []int{500, 502, 503},
			maxAttempts:    3,
			shouldRetry:    false,
		},
		{
			name:           "400 is not retryable",
			statusCode:     400,
			retryableCodes: []int{500, 502, 503},
			maxAttempts:    3,
			shouldRetry:    false,
		},
		{
			name:           "401 is not retryable",
			statusCode:     401,
			retryableCodes: []int{500, 502, 503},
			maxAttempts:    3,
			shouldRetry:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attemptCount int64

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt64(&attemptCount, 1)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(fmt.Sprintf("error %d", tt.statusCode)))
			}))
			defer server.Close()

			transport := createTestTransportWithConfig(t, &cbt.Config{
				MaxAttempts:    tt.maxAttempts,
				RetryDelay:     10 * time.Millisecond,
				RetryableCodes: tt.retryableCodes,
			})

			req, err := http.NewRequest("GET", server.URL, nil)
			require.NoError(t, err)

			resp, err := transport.RoundTrip(req)

			if tt.shouldRetry {
				assert.Equal(t, int64(tt.maxAttempts), atomic.LoadInt64(&attemptCount))
				assert.Error(t, err)
				assert.Contains(t, err.Error(), fmt.Sprintf("HTTP %d (retryable)", tt.statusCode))
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, int64(1), atomic.LoadInt64(&attemptCount))
				assert.Error(t, err)
				assert.True(t, errors.Is(err, clients.ErrNonRetryableRequest))
				assert.Nil(t, resp)
			}
		})
	}
}

func TestCircuitBreakerTransport_RoundTrip_RetryThenSuccess(t *testing.T) {
	var attemptCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt64(&attemptCount, 1)
		if count <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("temporary failure"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	transport := createTestTransportWithConfig(t, &cbt.Config{
		MaxAttempts:    3,
		RetryDelay:     10 * time.Millisecond,
		RetryableCodes: []int{500},
	})

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)

	assert.Equal(t, int64(3), atomic.LoadInt64(&attemptCount))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCircuitBreakerTransport_RoundTrip_NetworkError(t *testing.T) {
	mockTransport := &mockRoundTripper{
		responses: []mockResponse{
			{err: errors.New("network error 1")},
			{err: errors.New("network error 2")},
			{err: errors.New("network error 3")},
		},
	}

	cb := createTestCircuitBreaker()
	transport := cbt.New(mockTransport, cb, &cbt.Config{
		MaxAttempts:    3,
		RetryDelay:     10 * time.Millisecond,
		RetryableCodes: []int{500},
	})

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)

	assert.Equal(t, 3, mockTransport.callCount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
	assert.Nil(t, resp)
}

func TestCircuitBreakerTransport_RoundTrip_ContextCancellation(t *testing.T) {
	var attemptCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&attemptCount, 1)
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	transport := createTestTransportWithConfig(t, &cbt.Config{
		MaxAttempts:    5,
		RetryDelay:     50 * time.Millisecond,
		RetryableCodes: []int{500},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)

	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "context")
	assert.Nil(t, resp)

	attempts := atomic.LoadInt64(&attemptCount)
	assert.Less(t, attempts, int64(5))
}

func TestCircuitBreakerTransport_CircuitBreakerOpen(t *testing.T) {
	cb := gobreaker.NewCircuitBreaker[*http.Response](gobreaker.Settings{
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     200 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.TotalFailures >= 2
		},
	})

	mockTransport := &mockRoundTripper{
		responses: []mockResponse{
			{err: errors.New("network error")},
			{err: errors.New("network error")},
			{err: errors.New("network error")},
		},
	}

	transport := cbt.New(mockTransport, cb, &cbt.Config{
		MaxAttempts:    3,
		RetryDelay:     10 * time.Millisecond,
		RetryableCodes: []int{500},
	})

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	_, err1 := transport.RoundTrip(req)
	assert.Error(t, err1)

	_, err2 := transport.RoundTrip(req)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "circuit breaker is open")
}

func createTestTransport(t *testing.T, retryableCodes []int) *cbt.CircuitBreakerTransport {
	return createTestTransportWithConfig(t, &cbt.Config{
		MaxAttempts:    3,
		RetryDelay:     10 * time.Millisecond,
		RetryableCodes: retryableCodes,
	})
}

func createTestTransportWithConfig(t *testing.T, cfg *cbt.Config) *cbt.CircuitBreakerTransport {
	cb := createTestCircuitBreaker()
	return cbt.New(http.DefaultTransport, cb, cfg)
}

func createTestCircuitBreaker() *gobreaker.CircuitBreaker[*http.Response] {
	return gobreaker.NewCircuitBreaker[*http.Response](gobreaker.Settings{
		MaxRequests: 3,
		Interval:    time.Minute,
		Timeout:     time.Minute,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.TotalFailures >= 10 // High threshold for most tests
		},
	})
}

type mockResponse struct {
	resp *http.Response
	err  error
}

type mockRoundTripper struct {
	responses []mockResponse
	callCount int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.callCount >= len(m.responses) {
		return nil, errors.New("no more mock responses")
	}

	response := m.responses[m.callCount]
	m.callCount++

	return response.resp, response.err
}
