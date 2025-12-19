package application

func InitApplication(opts ...Option) (*Application, error) {
	app := newApplication()
	if err := app.apply(opts...); err != nil {
		return nil, err
	}

	return app, nil
}
