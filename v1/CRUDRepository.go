package v1

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	idColumn     = "id"
	RETURNING_ID = "RETURNING id"
)

type CRUDRepository[T any] struct {
	db              DbClient
	insertBuilder   sq.InsertBuilder
	selectBuilder   sq.SelectBuilder
	updateBuilder   sq.UpdateBuilder
	deleteBuilder   sq.DeleteBuilder
	entityConverter func(row pgx.Row) T
}

func NewCRUDRepository[T any](
	db DbClient,
	insertBuilder sq.InsertBuilder,
	selectBuilder sq.SelectBuilder,
	updateBuilder sq.UpdateBuilder,
	deleteBuilder sq.DeleteBuilder,
	entityConverter func(row pgx.Row) T) CRUDRepository[T] {
	return CRUDRepository[T]{
		db:            db,
		insertBuilder: insertBuilder, selectBuilder: selectBuilder, updateBuilder: updateBuilder, deleteBuilder: deleteBuilder,
		entityConverter: entityConverter}
}

func (repo CRUDRepository[T]) Create(ctx context.Context, values ...interface{}) int64 {
	builder := repo.insertBuilder.Suffix(RETURNING_ID).Values(values...)
	var id int64
	err := repo.db.API().QueryRowContextInsert(ctx, builder).Scan(&id)
	if err != nil {
		panic(err)
	}
	return id
}

func (repo CRUDRepository[T]) GetById(ctx context.Context, id int64) T {
	builder := repo.selectBuilder.Where(sq.Eq{idColumn: id})
	row := repo.db.API().QueryRowContextSelect(ctx, builder)
	return repo.entityConverter(row)
}

func (repo CRUDRepository[T]) ConvertToObjects(rows pgx.Rows) []T {
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

func (repo *CRUDRepository[T]) GetAll(ctx context.Context) []T {
	rows := repo.db.API().QueryContextSelect(ctx, repo.selectBuilder, nil)
	objs := repo.ConvertToObjects(rows)
	return objs
}

func (repo *CRUDRepository[T]) GetBy(ctx context.Context, where sq.Sqlizer) []T {
	builder := repo.selectBuilder.Where(where)
	rows := repo.db.API().QueryContextSelect(ctx, builder, nil)
	objs := repo.ConvertToObjects(rows)
	return objs
}

func (repo *CRUDRepository[T]) Delete(ctx context.Context, id int64) int64 {
	builder := repo.deleteBuilder.Where(sq.Eq{idColumn: id})
	return repo.db.API().ExecDelete(ctx, builder)
}

func updateInternal(ctx context.Context, api DbApi, updateBuilder sq.UpdateBuilder, fields map[string]interface{}) int64 {
	for column, value := range fields {
		updateBuilder = updateBuilder.Set(column, value)
	}
	return api.ExecUpdate(ctx, updateBuilder)
}

func (repo *CRUDRepository[T]) Update(ctx context.Context, fields map[string]interface{}, id int64) int64 {
	builder := repo.updateBuilder.Where(sq.Eq{idColumn: id})
	return updateInternal(ctx, repo.db.API(), builder, fields)
}

func (repo *CRUDRepository[T]) UpdateCollection(ctx context.Context, fields map[string]interface{}, where sq.Sqlizer) int64 {
	builder := repo.updateBuilder.Where(where)
	return updateInternal(ctx, repo.db.API(), builder, fields)
}

func (repo *CRUDRepository[T]) UpdateReturning(ctx context.Context, builder sq.UpdateBuilder, entityConverter func(row pgx.Row) T) T {
	row := repo.db.API().UpdateReturning(ctx, builder)
	return entityConverter(row)
}
