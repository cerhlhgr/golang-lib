package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	httpPkg "github.com/cerhlhgr/golang-lib/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	HTTP        http.Handler
	HTTPServer  *httpPkg.Server
	PGConn      *pgxpool.Pool
	RedisClient *redis.Client
	S3          *minio.Client

	dependencies entities
	services     entities

	logger *log.Logger

	observer LifecycleObserver

	httpConfig     HTTPConfig
	postgresConfig *PostgresConfig
	redisConfig    *RedisConfig
	s3Config       *S3Config

	checkMu         sync.RWMutex
	readinessChecks map[string]HealthCheck
	livenessChecks  map[string]HealthCheck

	mu           sync.Mutex
	bootstrapped bool
	started      bool
	closed       bool
	closers      []namedCloser
}

type HealthCheck func(ctx context.Context) error

type namedCloser struct {
	name string
	fn   func(context.Context) error
}

func newApplication() *Application {
	return &Application{
		closers:         make([]namedCloser, 0, 4),
		services:        newEntities(),
		dependencies:    newEntities(),
		logger:          log.Default(),
		httpConfig:      DefaultHTTPConfig(),
		readinessChecks: map[string]HealthCheck{},
		livenessChecks:  map[string]HealthCheck{},
	}
}

func (a *Application) RegisterHTTPServe(handler http.Handler) {
	a.HTTP = handler
}

func (a *Application) RegisterReadinessCheck(name string, check HealthCheck) {
	a.checkMu.Lock()
	defer a.checkMu.Unlock()
	a.readinessChecks[name] = check
}

func (a *Application) RegisterLivenessCheck(name string, check HealthCheck) {
	a.checkMu.Lock()
	defer a.checkMu.Unlock()
	a.livenessChecks[name] = check
}

func WithHTTPServer(opts ...HTTPServerOption) Option {
	return func(a *Application) error {
		if err := WithHTTP(a.httpConfig)(a); err != nil {
			return err
		}

		for _, opt := range opts {
			opt(a)
		}

		return nil
	}
}

func WithInfraDependencies(deps ...Dependency) Option {
	return func(a *Application) error {
		for _, d := range deps {
			if added := a.dependencies.addDependency(d); !added {
				return fmt.Errorf("%s: %w", d.name, ErrDependencyAlreadyEnabled)
			}

			switch d.name {
			case Postgres.name:
				cfg := PostgresConfigFromEnv()
				a.postgresConfig = &cfg
			case Redis.name:
				cfg := RedisConfigFromEnv()
				a.redisConfig = &cfg
			case S3.name:
				cfg := S3ConfigFromEnv()
				a.s3Config = &cfg
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

func New(ctx context.Context, opts ...Option) (*Application, error) {
	app := newApplication()
	if err := app.apply(opts...); err != nil {
		return nil, err
	}

	if err := app.bootstrap(ctx); err != nil {
		return nil, err
	}

	return app, nil
}

func MustNew(ctx context.Context, opts ...Option) *Application {
	app, err := New(ctx, opts...)
	if err != nil {
		panic(err)
	}

	return app
}

func (a *Application) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.started {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	if err := a.bootstrap(ctx); err != nil {
		return err
	}

	for _, svc := range a.services.list() {
		if svc.runFn == nil {
			continue
		}

		a.notifyStart(svc.name)

		if err := svc.runFn(ctx, a); err != nil {
			a.notifyStop(svc.name, err)
			return fmt.Errorf("start service %s: %w", svc.name, err)
		}

		a.notifyStop(svc.name, nil)
	}

	a.mu.Lock()
	a.started = true
	a.mu.Unlock()

	return nil
}

func (a *Application) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := a.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.httpConfig.Server.ShutdownTimeout)
	defer cancel()

	return a.CloseContext(shutdownCtx)
}

func (a *Application) AddCloser(f func() error) {
	if f == nil {
		return
	}
	a.AddCloserContext("custom", func(context.Context) error { return f() })
}

func (a *Application) AddCloserContext(name string, f func(context.Context) error) {
	a.closers = append(a.closers, namedCloser{name: name, fn: f})
}

func (a *Application) bootstrap(ctx context.Context) error {
	a.mu.Lock()
	if a.bootstrapped {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	for _, dep := range a.dependencies.list() {
		if dep.initFn == nil {
			continue
		}

		a.notifyStart(dep.name)

		if err := dep.initFn(ctx, a); err != nil {
			a.notifyStop(dep.name, err)
			return fmt.Errorf("init dependency %s: %w", dep.name, err)
		}

		a.notifyStop(dep.name, nil)
	}

	a.mu.Lock()
	a.bootstrapped = true
	a.mu.Unlock()

	return nil
}

func (a *Application) Close() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.httpConfig.Server.ShutdownTimeout)
	defer cancel()

	return a.CloseContext(shutdownCtx)
}

func (a *Application) CloseContext(ctx context.Context) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.closed = true
	closers := make([]namedCloser, len(a.closers))
	copy(closers, a.closers)
	a.mu.Unlock()

	var errs []error

	for i := len(closers) - 1; i >= 0; i-- {
		closer := closers[i]
		if err := closer.fn(ctx); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", closer.name, err))
		}
	}

	return errors.Join(errs...)
}

func (a *Application) notifyStart(name string) {
	if a.observer != nil {
		a.observer.OnComponentStart(name)
	}
}

func (a *Application) notifyStop(name string, err error) {
	if a.observer != nil {
		a.observer.OnComponentStop(name, err)
	}
}
