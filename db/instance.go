package db

import "github.com/jackc/pgx/v5/pgxpool"

type Instance struct {
	Postgres *pgxpool.Pool
}

func NewInstance(postgres *pgxpool.Pool) *Instance {
	return &Instance{Postgres: postgres}
}

func (i *Instance) Close() {
	if i == nil || i.Postgres == nil {
		return
	}

	i.Postgres.Close()
}
