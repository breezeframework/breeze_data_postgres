package pg

import (
	"context"
	"github.com/simpleGorm/pg/internal/pg_api"
	"github.com/simpleGorm/pg/internal/transaction"
)

type DbClient interface {
	Close() error
	//pg() pg_api.PgDbClient
	RunTransaction(ctx context.Context, txOptions transaction.TxOptions, f pg_api.TransactionalFlow) error
}

func NewDBClient(ctx context.Context, dsn string) (DbClient, error) {
	return pg_api.NewPgDBClient(ctx, dsn)
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
