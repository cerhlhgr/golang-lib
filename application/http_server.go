package application

import (
	"context"
	"fmt"
	"log"
	"net/http"

	httpPkg "github.com/cerhlhgr/golang-lib/http"
)

const defaultServerName = "HTTP"

func runHTTPServer(ctx context.Context, a *Application) error {
	h := a.HTTP
	if h == nil {
		// provide a default mux so developer doesn't need to wire it in main
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
		h = mux
		a.HTTP = h
	}

	// create server on :8080 by default
	httpServer := httpPkg.NewServer(":8080", defaultServerName)
	httpServer.SetHandler(h)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Printf("http: serve error: %v", err)
			return
		}
		log.Printf("http: server stopped")
	}()

	if err := httpServer.WaitsForStarted(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	log.Printf("http: server started on %s", ":8080")

	a.AddCloser(func() error { return httpServer.Shutdown(context.Background()) })

	a.HTTPServer = httpServer
	return nil
}
