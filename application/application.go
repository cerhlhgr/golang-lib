package application

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/cerhlhgr/golang-lib/db"
	httpPkg "github.com/cerhlhgr/golang-lib/http"
	"github.com/cerhlhgr/golang-lib/redis"
	"github.com/minio/minio-go/v7"
)

type Application struct {
	HTTP       http.Handler
	HTTPServer *httpPkg.Server

	dependencies entities
	services     entities

	Database *db.Instance
	Redis    *redis.Instance
	S3       *minio.Client

	closers []func() error
}

func newApplication() *Application {
	return &Application{
		closers:      make([]func() error, 0, 4),
		services:     make(entities),
		dependencies: make(entities),
	}
}

func (a *Application) RegisterHTTPServe(handler http.Handler) {
	a.HTTP = handler
}

func WithHTTPServer(opts ...HTTPServerOption) Option {
	return func(a *Application) error {
		for _, opt := range opts {
			opt(a)
		}

		return addServiceToApp(httpServer, a)
	}
}

func WithInfraDependencies(deps ...Dependency) Option {
	return func(a *Application) error {
		for _, d := range deps {
			if added := a.dependencies.addDependency(d); !added {
				return fmt.Errorf("%s: %w", d.name, ErrDependencyAlreadyEnabled)
			}
		}

		return nil
	}
}

func (a *Application) apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(a); err != nil {
			return err
		}
	}

	return nil
}

func (a *Application) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for _, dep := range a.dependencies {
		if dep.initFn != nil {
			if err := dep.initFn(ctx, a); err != nil {
				return err
			}
		}
	}

	for _, svc := range a.services {
		if svc.runFn == nil {
			continue
		}
		if err := svc.runFn(ctx, a); err != nil {
			return err
		}
	}

	<-ctx.Done()
	return a.Close()
}

func (a *Application) AddCloser(f func() error) { a.closers = append(a.closers, f) }

func (a *Application) Close() error {
	var retErr error
	for i := len(a.closers) - 1; i >= 0; i-- {
		if err := a.closers[i](); err != nil && retErr == nil {
			retErr = err
		}
	}
	return retErr
}
