package application

import (
	"context"

	"github.com/cerhlhgr/golang-lib/db"
)

func initDB(ctx context.Context, a *Application) error {
	database, err := db.NewPostgres(ctx)
	if err != nil {
		return err
	}

	a.Database = database
	a.AddCloser(func() error {
		database.Close()
		return nil
	})

	return nil
}
