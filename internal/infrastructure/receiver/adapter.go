package receiver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type SyncAdapter struct {
	srv Server
}

func NewSyncUpdateReceiver(srv Server) *SyncAdapter {
	return &SyncAdapter{
		srv: srv,
	}
}

func (s *SyncAdapter) Start(_ context.Context) error {
	if err := s.srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (s *SyncAdapter) Stop(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("update receiver: failed to shutdown server: %w", err)
	}

	return nil
}
