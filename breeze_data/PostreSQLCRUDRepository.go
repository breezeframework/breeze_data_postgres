package breeze_data

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"log"
)

const (
	idColumn     = "id"
	RETURNING_ID = "RETURNING id"
)

type PostgreSQLCRUDRepository[T any] struct {
	db              DbClient
	insertBuilder   sq.InsertBuilder
	selectBuilder   sq.SelectBuilder
	updateBuilder   sq.UpdateBuilder
	deleteBuilder   sq.DeleteBuilder
	entityConverter func(row pgx.Row) T
}

func NewPostgreSQLCRUDRepository[T any](
	db DbClient,
	insertBuilder sq.InsertBuilder,
	selectBuilder sq.SelectBuilder,
	updateBuilder sq.UpdateBuilder,
	deleteBuilder sq.DeleteBuilder,
	entityConverter func(row pgx.Row) T) CrudRepository[T] {
	return &PostgreSQLCRUDRepository[T]{
		db:            db,
		insertBuilder: insertBuilder, selectBuilder: selectBuilder, updateBuilder: updateBuilder, deleteBuilder: deleteBuilder,
		entityConverter: entityConverter}
}

func (repo *PostgreSQLCRUDRepository[T]) Create(ctx context.Context, values ...interface{}) int64 {
	builder := repo.insertBuilder.Suffix(RETURNING_ID).Values(values...)
	var id int64
	err := repo.db.API().QueryRowContextInsert(ctx, &builder).Scan(&id)
	if err != nil {
		panic(err)
	}
	return id
}

func (repo *PostgreSQLCRUDRepository[T]) GetById(ctx context.Context, id int64) T {
	builder := repo.selectBuilder.Where(sq.Eq{idColumn: id})
	row := repo.db.API().QueryRowContextSelect(ctx, &builder)
	return repo.entityConverter(row)
}

func (repo *PostgreSQLCRUDRepository[T]) ConvertToObjects(rows pgx.Rows) []T {
	var objs []T
	for rows.Next() {
		obj := repo.entityConverter(rows)
		objs = append(objs, obj)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return objs
}

func (repo *PostgreSQLCRUDRepository[T]) GetAll(ctx context.Context) []T {

	defer func() {
		if r := recover(); r != nil {
			err := errors.New(fmt.Sprintf("%v", r))
			log.Print(err)
		}
	}()
	rows := repo.db.API().QueryContextSelect(ctx, &repo.selectBuilder, nil)
	objs := repo.ConvertToObjects(rows)
	return objs
}

func (repo *PostgreSQLCRUDRepository[T]) GetBy(ctx context.Context, where sq.Sqlizer) []T {
	defer func() {
		if r := recover(); r != nil {
			errors.New(fmt.Sprintf("%v", r))
		}
	}()
	builder := repo.selectBuilder.Where(where)
	rows := repo.db.API().QueryContextSelect(ctx, &builder, nil)
	objs := repo.ConvertToObjects(rows)
	return objs
}

func (repo *PostgreSQLCRUDRepository[T]) Delete(ctx context.Context, id int64) {
	panic("implement me")
}

func (repo *PostgreSQLCRUDRepository[T]) Update(ctx context.Context, builder sq.UpdateBuilder) pgconn.CommandTag {
	return repo.db.API().ExecUpdate(ctx, &builder)
}

func (repo *PostgreSQLCRUDRepository[T]) UpdateReturning(ctx context.Context, builder sq.UpdateBuilder, entityConverter func(row pgx.Row) T) T {
	row := repo.db.API().UpdateReturning(ctx, &builder)
	return entityConverter(row)
}
