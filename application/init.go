package application

import "context"

func InitApplication(opts ...Option) (*Application, error) {
	app := newApplication()
	if err := app.apply(opts...); err != nil {
		return nil, err
	}
	if err := app.initDependencies(context.Background()); err != nil {
		return nil, err
	}

	return app, nil
}
