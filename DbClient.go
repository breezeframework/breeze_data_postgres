package pg

import (
	"context"
	"github.com/simpleGorm/pg/internal/pg_api"
)

type DbClient interface {
	Close() error
	API() pg_api.PgDbClient
}

func NewDBClient(ctx context.Context, dsn string) (DbClient, error) {
	client, err := pg_api.NewPgDBClient(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &dbClientWrapper{pgClient: client}, nil
}

type dbClientWrapper struct {
	pgClient pg_api.PgDbClient
}

func (d *dbClientWrapper) Close() error {
	return d.pgClient.Close()
}

func (d *dbClientWrapper) API() pg_api.PgDbClient {
	return d.pgClient
}

type DbApi interface {
	pg_api.SQLExecutor
	pg_api.Transactor
	Pinger

	Close()
}

type Pinger interface {
	Ping(ctx context.Context) error
}
