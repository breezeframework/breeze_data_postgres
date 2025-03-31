package pg_api

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simpleGorm/pg/internal/prettier"
	"log"
)

type key string

const (
	TxKey key = "tx"
)

type pg struct {
	api *pgxpool.Pool
}

func (pg pg) ExecDelete(ctx context.Context, builder sq.DeleteBuilder) int64 {
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	log.Printf("[ExecDelete] query: %s", query)
	log.Printf("[ExecDelete] args: %+v", args)
	log.Printf("[ExecDelete] err: %+v", err)
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	var tag pgconn.CommandTag
	if ok {
		tag, err = tx.Exec(ctx, query, args...)
	} else {
		tag, err = pg.api.Exec(ctx, query, args...)
	}
	if err != nil {
		log.Printf("err: %+v", err)
		log.Panic(err)
	}
	return tag.RowsAffected()
}

func (pg pg) ExecUpdate(ctx context.Context, builder sq.UpdateBuilder) int64 {
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	log.Printf("[ExecUpdate] query: %s", query)
	log.Printf("[ExecUpdate] args: %+v", args)
	log.Printf("[ExecUpdate] err: %+v", err)
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	var tag pgconn.CommandTag
	if ok {
		tag, err = tx.Exec(ctx, query, args...)
	} else {
		tag, err = pg.api.Exec(ctx, query, args...)
	}
	if err != nil {
		log.Printf("err: %+v", err)
		log.Panic(err)
	}
	return tag.RowsAffected()
}

func (pg pg) QueryContextSelect(ctx context.Context, builder sq.SelectBuilder, where map[string]interface{}) pgx.Rows {
	if where != nil {
		builder.Where(where)
	}
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated SQL query:", query)
	fmt.Println("Arguments:", args)
	fmt.Println("ctx:", ctx)
	fmt.Println("ctx.Value(TxKey):", ctx.Value(TxKey))
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	var rows pgx.Rows
	if ok {
		rows, err = tx.Query(ctx, query, args...)
	} else {
		rows, err = pg.api.Query(ctx, query, args...)
	}
	if err != nil {
		panic(err)
	}
	return rows
}

func (pg pg) QueryRowContextSelect(ctx context.Context, builder sq.SelectBuilder) pgx.Row {
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated SQL query:", query)
	fmt.Println("Arguments:", args)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, query, args...)
	}

	return pg.api.QueryRow(ctx, query, args...)
}

func (pg pg) QueryRowContextInsert(ctx context.Context, builder sq.InsertBuilder) pgx.Row {

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated SQL query:", query)
	fmt.Println("Arguments:", args)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, query, args...)
	}

	return pg.api.QueryRow(ctx, query, args...)
}

func (pg pg) UpdateReturning(ctx context.Context, builder sq.UpdateBuilder) pgx.Row {
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated SQL query:", query)
	fmt.Println("Arguments:", args)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, query, args...)
	}

	return pg.api.QueryRow(ctx, query, args...)
}
func NewDB(dbc *pgxpool.Pool) pg {
	return pg{
		api: dbc,
	}
}

/*func (pkg *pkg) ScanOneContext(ctx context.Context, dest interface{}, q db.Query, args ...interface{}) error {
	logQuery(ctx, q, args...)

	row, err := p.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanOne(dest, row)
}

func (pkg *pkg) ScanAllContext(ctx context.Context, dest interface{}, q db.Query, args ...interface{}) error {
	logQuery(ctx, q, args...)

	rows, err := p.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanAll(dest, rows)
}*/

func (pg pg) ExecContext(ctx context.Context, q Query, args ...interface{}) (pgconn.CommandTag, error) {
	logQuery(ctx, q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, q.QueryRaw, args...)
	}

	return pg.api.Exec(ctx, q.QueryRaw, args...)
}

func (pg pg) QueryContext(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error) {
	logQuery(ctx, q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, q.QueryRaw, args...)
	}

	return pg.api.Query(ctx, q.QueryRaw, args...)
}

func (pg pg) QueryRowContext(ctx context.Context, q Query, args ...interface{}) pgx.Row {
	logQuery(ctx, q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, q.QueryRaw, args...)
	}

	return pg.api.QueryRow(ctx, q.QueryRaw, args...)
}

func (pg pg) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return pg.api.BeginTx(ctx, txOptions)
}

func (pg pg) Ping(ctx context.Context) error {
	return pg.api.Ping(ctx)
}

func (pg pg) Close() {
	pg.api.Close()
}

func MakeContextTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}

func logQuery(ctx context.Context, q Query, args ...interface{}) {
	prettyQuery := prettier.Pretty(q.QueryRaw, prettier.PlaceholderDollar, args...)
	log.Println(
		ctx,
		fmt.Sprintf("sql: %s", q.Name),
		fmt.Sprintf("query: %s", prettyQuery),
	)
}
