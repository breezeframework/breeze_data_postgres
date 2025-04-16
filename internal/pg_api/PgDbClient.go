package pg_api

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/simpleGorm/pg/pkg/transaction"
)

type PgDbClient struct {
	masterDBC          PG
	transactionManager *PgTransactionManager
}

func (c PgDbClient) UpdateReturning(ctx context.Context, builder sq.UpdateBuilder) pgx.Row {
	return c.masterDBC.UpdateReturning(ctx, builder)
}

func (c PgDbClient) ExecDelete(ctx context.Context, builder sq.DeleteBuilder) int64 {
	return c.masterDBC.ExecDelete(ctx, builder)
}

func (c PgDbClient) ExecUpdate(ctx context.Context, builder sq.UpdateBuilder) int64 {
	return c.masterDBC.ExecUpdate(ctx, builder)
}

func (c PgDbClient) QueryContextSelect(ctx context.Context, builder sq.SelectBuilder, where map[string]interface{}) pgx.Rows {
	return c.masterDBC.QueryContextSelect(ctx, builder, where)
}

func (c PgDbClient) QueryRowContextSelect(ctx context.Context, builder sq.SelectBuilder) pgx.Row {
	return c.masterDBC.QueryRowContextSelect(ctx, builder)
}

func (c PgDbClient) QueryRowContextInsert(ctx context.Context, builder sq.InsertBuilder) pgx.Row {
	return c.masterDBC.QueryRowContextInsert(ctx, builder)
}

func (c PgDbClient) Ping(ctx context.Context) error {
	return c.masterDBC.Ping(ctx)
}

func (c PgDbClient) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return c.masterDBC.BeginTx(ctx, txOptions)
}

func (c PgDbClient) Close() error {
	if c.masterDBC.API != nil {
		return c.masterDBC.Close()
	}
	return nil
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
