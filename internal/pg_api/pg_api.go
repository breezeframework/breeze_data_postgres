package pg_api

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simpleGorm/pg/internal/prettier"
	"github.com/simpleGorm/pg/pkg"
	"log"
)

type key string

const (
	TxKey key = "tx"
)

type PG struct {
	API *pgxpool.Pool
}

func (pg PG) ExecDelete(ctx context.Context, builder sq.DeleteBuilder) int64 {
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
		tag, err = pg.API.Exec(ctx, query, args...)
	}
	if err != nil {
		log.Printf("err: %+v", err)
		log.Panic(err)
	}
	return tag.RowsAffected()
}

func (pg PG) ExecUpdate(ctx context.Context, builder sq.UpdateBuilder) int64 {
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	pkg.Logger().Info("[ExecUpdate] query: %s", query)
	pkg.Logger().Info("[ExecUpdate] args: %+v", args)
	pkg.Logger().Info("[ExecUpdate] err: %+v", err)
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	var tag pgconn.CommandTag
	if ok {
		tag, err = tx.Exec(ctx, query, args...)
	} else {
		tag, err = pg.API.Exec(ctx, query, args...)
	}
	if err != nil {
		log.Panic(err)
	}
	return tag.RowsAffected()
}

func (pg PG) QueryContextSelect(ctx context.Context, builder sq.SelectBuilder, where map[string]interface{}) pgx.Rows {
	if where != nil {
		builder.Where(where)
	}
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	pkg.Logger().Info("Generated SQL query:", query)
	pkg.Logger().Info("Arguments:", args)
	pkg.Logger().Info("ctx:", ctx)
	pkg.Logger().Info("ctx.Value(TxKey):", ctx.Value(TxKey))
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	var rows pgx.Rows
	if ok {
		rows, err = tx.Query(ctx, query, args...)
	} else {
		rows, err = pg.API.Query(ctx, query, args...)
	}
	if err != nil {
		panic(err)
	}
	return rows
}

func (pg PG) QueryRowContextSelect(ctx context.Context, builder sq.SelectBuilder) pgx.Row {
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	pkg.Logger().Info("Generated SQL query:", query)
	pkg.Logger().Info("Arguments:", args)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, query, args...)
	}

	return pg.API.QueryRow(ctx, query, args...)
}

func (pg PG) QueryRowContextInsert(ctx context.Context, builder sq.InsertBuilder) pgx.Row {

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	pkg.Logger().Info("Generated SQL query:", query)
	pkg.Logger().Info("Arguments:", args)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, query, args...)
	}

	return pg.API.QueryRow(ctx, query, args...)
}

func (pg PG) UpdateReturning(ctx context.Context, builder sq.UpdateBuilder) pgx.Row {
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		panic(err)
	}

	pkg.Logger().Info("Generated SQL query:", query)
	pkg.Logger().Info("Arguments:", args)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, query, args...)
	}

	return pg.API.QueryRow(ctx, query, args...)
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

func (pg PG) ExecContext(ctx context.Context, q Query, args ...interface{}) (pgconn.CommandTag, error) {
	logQuery(q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, q.QueryRaw, args...)
	}

	return pg.API.Exec(ctx, q.QueryRaw, args...)
}

func (pg PG) QueryContext(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error) {
	logQuery(q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, q.QueryRaw, args...)
	}

	return pg.API.Query(ctx, q.QueryRaw, args...)
}

func (pg PG) QueryRowContext(ctx context.Context, q Query, args ...interface{}) pgx.Row {
	logQuery(q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, q.QueryRaw, args...)
	}

	return pg.API.QueryRow(ctx, q.QueryRaw, args...)
}

func (pg PG) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return pg.API.BeginTx(ctx, txOptions)
}

func (pg PG) Ping(ctx context.Context) error {
	return pg.API.Ping(ctx)
}

func (pg PG) Close() {
	pg.API.Close()
}

func MakeContextTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}

func logQuery(q Query, args ...interface{}) {
	prettyQuery := prettier.Pretty(q.QueryRaw, prettier.PlaceholderDollar, args...)
	pkg.Logger().Info("sql: %s", q.Name)
	pkg.Logger().Info("query: %s", prettyQuery)

}
