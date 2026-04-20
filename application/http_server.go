package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	httpPkg "github.com/cerhlhgr/golang-lib/http"
)

func runHTTPServer(ctx context.Context, a *Application) error {
	cfg := a.httpConfig
	cfg.setDefaults()
	if err := cfg.Validate(); err != nil {
		return err
	}

	h := a.HTTP
	if h == nil {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
		h = mux
		a.HTTP = h
	}

	if cfg.EnableHealthEndpoints {
		h = a.withHealthHandlers(h)
	}

	httpServer, err := httpPkg.New(cfg.Server)
	if err != nil {
		return fmt.Errorf("create HTTP server: %w", err)
	}

	httpServer.SetHandler(h)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			a.logger.Printf("http: serve error: %v", err)
			return
		}
		a.logger.Printf("http: server stopped")
	}()

	if err := httpServer.WaitsForStarted(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	a.logger.Printf("http: server started on %s", cfg.Server.Addr)

	a.AddCloserContext("http", func(ctx context.Context) error {
		shutdownTimeout := cfg.Server.ShutdownTimeout
		if shutdownTimeout <= 0 {
			shutdownTimeout = 10 * time.Second
		}

		shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	})

	a.HTTPServer = httpServer
	return nil
}

func (a *Application) withHealthHandlers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case a.httpConfig.HealthPath:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		case a.httpConfig.ReadinessPath:
			if err := a.runChecks(r.Context(), a.readinessChecks); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
			return
		case a.httpConfig.LivenessPath:
			if err := a.runChecks(r.Context(), a.livenessChecks); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("alive"))
			return
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func (a *Application) runChecks(ctx context.Context, checks map[string]HealthCheck) error {
	a.checkMu.RLock()
	copyChecks := make(map[string]HealthCheck, len(checks))
	for name, check := range checks {
		copyChecks[name] = check
	}
	a.checkMu.RUnlock()

	for name, check := range copyChecks {
		checkCtx, cancel := context.WithTimeout(ctx, a.httpConfig.CheckTimeout)
		err := check(checkCtx)
		cancel()
		if err != nil {
			return fmt.Errorf("health check %s failed: %w", name, err)
		}
	}

	return nil
}
