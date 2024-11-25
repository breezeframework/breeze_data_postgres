package breeze_data

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/breezeframework/breeze_data/breeze_data/transaction"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DbClient interface {
	API() DbApi
	Close() error
	RunTransaction(ctx context.Context, txOptions transaction.TxOptions, f TransactionalFlow) error
}

type DbApi interface {
	SQLExecutor
	Transactor
	Pinger

	Close()
}

type TransactionalFlow func(ctx context.Context) error

type TxManager interface {
	Transaction(ctx context.Context, opts transaction.TxOptions, f TransactionalFlow) error
}

type Query struct {
	Name     string
	QueryRaw string
}

type Transactor interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type SQLExecutor interface {
	//NamedQueryExecutor
	QueryExecutor
}

type NamedQueryExecutor interface {
	ScanOneContext(ctx context.Context, dest interface{}, q Query, args ...interface{}) error
	ScanAllContext(ctx context.Context, dest interface{}, q Query, args ...interface{}) error
}

type QueryExecutor interface {
	UpdateReturning(ctx context.Context, builder squirrel.UpdateBuilder) pgx.Row
	ExecUpdate(ctx context.Context, builder squirrel.UpdateBuilder) pgconn.CommandTag
	QueryContextSelect(ctx context.Context, builder squirrel.SelectBuilder, where map[string]interface{}) pgx.Rows
	QueryRowContextSelect(ctx context.Context, builder squirrel.SelectBuilder) pgx.Row
	QueryRowContextInsert(ctx context.Context, builder squirrel.InsertBuilder) pgx.Row
}

type Pinger interface {
	Ping(ctx context.Context) error
}
