package pg

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/simpleGorm/pg/internal/pg_api"
	"github.com/simpleGorm/pg/pkg/transaction"
)

type PgDbClient struct {
	masterDBC          pg_api.PG
	transactionManager *pg_api.PgTransactionManager
}

func NewPgDBClient(ctx context.Context, dsn string) (PgDbClient, error) {
	dbc, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return PgDbClient{}, errors.Errorf("failed to connect to db: %v", err)
	}

	return PgDbClient{
		masterDBC:          pg_api.PG{dbc},
		transactionManager: pg_api.NewPgTransactionManager(dbc),
	}, nil
}

func (c PgDbClient) RunTransaction(ctx context.Context, txOpts transaction.TxOptions, f pg_api.TransactionalFlow) error {
	return c.transactionManager.Transaction(ctx, txOpts, f)
}

func (c PgDbClient) API() pg_api.PG {
	return c.masterDBC
}

func (c PgDbClient) Close() error {
	if c.masterDBC.API != nil {
		c.masterDBC.Close()
	}

	return nil
}
