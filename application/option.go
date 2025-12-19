package application

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrDependencyAlreadyEnabled = errors.New("зависимость не может быть инициализирована")
	ErrServiceAlreadyRegistered = errors.New("сервис не может быть запущен")
)

var (
	httpServer = NewService("http_server", NoopActionFn, runHTTPServer)
)

var (
	Postgres = Dependency{name: "postgres", initFn: initDB, runFn: NoopActionFn}
)

type (
	HTTPServerOption func(a *Application)
	Option           func(application *Application) error
	ActionFunc       func(context.Context, *Application) error

	Service    entity
	Dependency entity

	entity struct {
		initFn ActionFunc
		runFn  ActionFunc
		name   string
	}

	entities map[string]entity
)

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

func (e entities) addService(s Service) bool {
	return e.add(entity(s))
}

func (e entities) addDependency(d Dependency) bool {
	return e.add(entity(d))
}

func (e entities) add(ent entity) bool {
	if e.contains(ent.name) {
		return false
	}

	e[ent.name] = ent

	return true
}

func (e entities) contains(name string) bool {
	_, ok := e[name]

	return ok
}

func addServiceToApp(s Service, a *Application) error {
	if ok := a.services.addService(s); !ok {
		return fmt.Errorf("%s: %w", s.name, ErrServiceAlreadyRegistered)
	}

	return nil
}
