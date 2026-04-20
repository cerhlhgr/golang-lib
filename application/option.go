package application

import (
	"context"
	"errors"
	"fmt"
	"log"
)

var (
	ErrDependencyAlreadyEnabled = errors.New("dependency is already enabled")
	ErrServiceAlreadyRegistered = errors.New("service is already registered")
)

var (
	httpServer = NewService("http_server", NoopActionFn, runHTTPServer)
)

var (
	Postgres = Dependency{name: "postgres", initFn: initDB, runFn: NoopActionFn}
	Redis    = Dependency{name: "redis", initFn: initRedis, runFn: NoopActionFn}
	S3       = Dependency{name: "s3", initFn: initS3, runFn: NoopActionFn}
)

type (
	HTTPServerOption  func(a *Application)
	Option            func(application *Application) error
	ActionFunc        func(context.Context, *Application) error
	LifecycleObserver interface {
		OnComponentStart(name string)
		OnComponentStop(name string, err error)
	}

	Service    entity
	Dependency entity

	entity struct {
		initFn ActionFunc
		runFn  ActionFunc
		name   string
	}

	entities struct {
		items []entity
		index map[string]struct{}
	}
)

func WithLogger(logger *log.Logger) Option {
	return func(a *Application) error {
		if logger != nil {
			a.logger = logger
		}
		return nil
	}
}

func WithObserver(observer LifecycleObserver) Option {
	return func(a *Application) error {
		a.observer = observer
		return nil
	}
}

func WithMetrics(observer LifecycleObserver) Option {
	return WithObserver(observer)
}

func WithTracer(observer LifecycleObserver) Option {
	return WithObserver(observer)
}

func NewService(name string, init, run ActionFunc) Service {
	return Service{
		name:   name,
		initFn: init,
		runFn:  run,
	}
}

func NoopActionFn(context.Context, *Application) error {
	return nil
}

func newEntities() entities {
	return entities{
		items: make([]entity, 0, 4),
		index: make(map[string]struct{}, 4),
	}
}

func (e *entities) addService(s Service) bool {
	return e.add(entity(s))
}

func (e *entities) addDependency(d Dependency) bool {
	return e.add(entity(d))
}

func (e *entities) add(ent entity) bool {
	if e.contains(ent.name) {
		return false
	}

	e.items = append(e.items, ent)
	e.index[ent.name] = struct{}{}

	return true
}

func (e *entities) contains(name string) bool {
	_, ok := e.index[name]

	return ok
}

func (e *entities) list() []entity {
	return e.items
}

func addServiceToApp(s Service, a *Application) error {
	if ok := a.services.addService(s); !ok {
		return fmt.Errorf("%s: %w", s.name, ErrServiceAlreadyRegistered)
	}

	return nil
}
