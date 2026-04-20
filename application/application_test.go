package application

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestCloseOrderIsLIFO(t *testing.T) {
	t.Parallel()

	app := newApplication()
	order := make([]string, 0, 2)

	app.AddCloserContext("first", func(context.Context) error {
		order = append(order, "first")
		return nil
	})
	app.AddCloserContext("second", func(context.Context) error {
		order = append(order, "second")
		return nil
	})

	if err := app.CloseContext(context.Background()); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	want := []string{"second", "first"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("unexpected close order: got %v, want %v", order, want)
	}
}

func TestStartAndShutdownHTTP(t *testing.T) {
	t.Parallel()

	httpCfg := DefaultHTTPConfig()
	httpCfg.Server.Addr = "127.0.0.1:0"
	httpCfg.Server.ShutdownTimeout = 2 * time.Second

	app, err := New(context.Background(),
		WithHTTP(httpCfg),
		WithHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})),
	)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	if err := app.Start(context.Background()); err != nil {
		t.Fatalf("start app: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := app.CloseContext(ctx); err != nil {
		t.Fatalf("close app: %v", err)
	}
}
