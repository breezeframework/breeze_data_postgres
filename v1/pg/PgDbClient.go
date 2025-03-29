package pg

import (
	"context"
	lib "github.com/breezeframework/breeze_data_postgres/v1"
	"github.com/breezeframework/breeze_data_postgres/v1/transaction"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pkg/errors"
)

type pgDbClient struct {
	masterDBC          lib.DbApi
	transactionManager *PgTransactionManager
}

func (c *pgDbClient) RunTransaction(ctx context.Context, txOpts transaction.TxOptions, f lib.TransactionalFlow) error {
	return c.transactionManager.Transaction(ctx, txOpts, f)
}

func NewPgDBClient(ctx context.Context, dsn string) (lib.DbClient, error) {
	dbc, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.Errorf("failed to connect to db: %v", err)
	}

	return &pgDbClient{
		masterDBC:          &pg{api: dbc},
		transactionManager: NewPgTransactionManager(dbc),
	}, nil
}

func (c *pgDbClient) API() lib.DbApi {
	return c.masterDBC
}

func (c *pgDbClient) Close() error {
	if c.masterDBC != nil {
		c.masterDBC.Close()
	}

	return nil
}
