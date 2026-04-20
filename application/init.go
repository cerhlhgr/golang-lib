package application

import "context"

func InitApplication(opts ...Option) (*Application, error) {
	app := newApplication()
	if err := app.apply(opts...); err != nil {
		return nil, err
	}

	return app, nil
}

func InitAndBootstrap(ctx context.Context, opts ...Option) (*Application, error) {
	return New(ctx, opts...)
}
