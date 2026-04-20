package http

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestServerStartAndShutdown(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Addr = "127.0.0.1:0"

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	srv.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	go func() { _ = srv.ListenAndServe() }()

	if err := srv.WaitsForStarted(); err != nil {
		t.Fatalf("wait for started: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Addr = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty addr")
	}
}
