package pg

import (
	"context"
	"github.com/breezeframework/breeze_data/breeze_data"
	"github.com/breezeframework/breeze_data/breeze_data/transaction"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pkg/errors"
)

type pgDbClient struct {
	masterDBC          breeze_data.DbApi
	transactionManager *PgTransactionManager
}

func (c *pgDbClient) RunTransaction(ctx context.Context, txOpts transaction.TxOptions, f breeze_data.TransactionalFlow) error {
	return c.transactionManager.Transaction(ctx, txOpts, f)
}

func NewPgDBClient(ctx context.Context, dsn string) (breeze_data.DbClient, error) {
	dbc, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.Errorf("failed to connect to db: %v", err)
	}

	return &pgDbClient{
		masterDBC:          &pg{api: dbc},
		transactionManager: NewPgTransactionManager(dbc),
	}, nil
}

func (c *pgDbClient) API() breeze_data.DbApi {
	return c.masterDBC
}

func (c *pgDbClient) Close() error {
	if c.masterDBC != nil {
		c.masterDBC.Close()
	}

	return nil
}
