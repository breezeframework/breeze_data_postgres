package pg_api

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/simpleGorm/pg/internal/transaction"
)

type PgDbClient struct {
	masterDBC          PG
	transactionManager *PgTransactionManager
}

func NewPgDBClient(ctx context.Context, dsn string) (PgDbClient, error) {
	dbc, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return PgDbClient{}, errors.Errorf("failed to connect to db: %v", err)
	}

	return PgDbClient{
		masterDBC:          PG{dbc},
		transactionManager: NewPgTransactionManager(dbc),
	}, nil
}

func (c PgDbClient) RunTransaction(ctx context.Context, txOpts transaction.TxOptions, f TransactionalFlow) error {
	return c.transactionManager.Transaction(ctx, txOpts, f)
}

func (c PgDbClient) API() PG {
	return c.masterDBC
}

func (c PgDbClient) Close() error {
	if c.masterDBC.API != nil {
		c.masterDBC.Close()
	}

	return nil
}
