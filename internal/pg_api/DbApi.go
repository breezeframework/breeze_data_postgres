package pg_api

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg/internal/transaction"
)

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
	ExecDelete(ctx context.Context, builder squirrel.DeleteBuilder) int64
	ExecUpdate(ctx context.Context, builder squirrel.UpdateBuilder) int64
	QueryContextSelect(ctx context.Context, builder squirrel.SelectBuilder, where map[string]interface{}) pgx.Rows
	QueryRowContextSelect(ctx context.Context, builder squirrel.SelectBuilder) pgx.Row
	QueryRowContextInsert(ctx context.Context, builder squirrel.InsertBuilder) pgx.Row
}

type Pinger interface {
	Ping(ctx context.Context) error
}
