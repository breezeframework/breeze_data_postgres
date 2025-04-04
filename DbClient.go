package pg

import (
	"context"
	"github.com/simpleGorm/pg/internal/pg_api"
	"github.com/simpleGorm/pg/internal/transaction"
)

type DbClient interface {
	Close() error
	Pg() PgDbClient
	RunTransaction(ctx context.Context, txOptions transaction.TxOptions, f pg_api.TransactionalFlow) error
}

func NewDBClient(ctx context.Context, dsn string) (DbClient, error) {
	client, err := NewPgDBClient(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &dbClientWrapper{pgClient: client}, nil
}

type dbClientWrapper struct {
	pgClient PgDbClient
}

func (d *dbClientWrapper) Close() error {
	return d.pgClient.Close()
}

func (d *dbClientWrapper) Pg() PgDbClient {
	return d.pgClient
}

func (d *dbClientWrapper) RunTransaction(ctx context.Context, txOptions transaction.TxOptions, f pg_api.TransactionalFlow) error {
	return d.pgClient.RunTransaction(ctx, txOptions, f)
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
