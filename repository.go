package pg

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg/internal/pg_api"
)

const (
	idColumn     = "id"
	RETURNING_ID = "RETURNING id"
)

/*type PostgreRepository[T any] interface {
	Create(ctx context.Context, values ...interface{}) int64
	GetById(ctx context.Context, id int64) T
	GetAll(ctx context.Context) []T
	GetBy(ctx context.Context, where sq.Sqlizer) []T
	Delete(ctx context.Context, id int64) int64
	UpdateCollection(ctx context.Context, fields map[string]interface{}, where sq.Sqlizer) int64
	Update(ctx context.Context, fields map[string]interface{}, id int64) int64
	UpdateReturning(ctx context.Context, builder sq.UpdateBuilder, entityConverter func(row pgx.Row) T) T
}*/

type Repository[T any] struct {
	db              pg_api.PgDbClient
	insertBuilder   sq.InsertBuilder
	selectBuilder   sq.SelectBuilder
	updateBuilder   sq.UpdateBuilder
	deleteBuilder   sq.DeleteBuilder
	entityConverter func(row pgx.Row) T
}

func NewPostgreRepository[T any](
	db DbClient,
	insertBuilder sq.InsertBuilder,
	selectBuilder sq.SelectBuilder,
	updateBuilder sq.UpdateBuilder,
	deleteBuilder sq.DeleteBuilder,
	entityConverter func(row pgx.Row) T) Repository[T] {
	return Repository[T]{
		db:            db.(pg_api.PgDbClient),
		insertBuilder: insertBuilder, selectBuilder: selectBuilder, updateBuilder: updateBuilder, deleteBuilder: deleteBuilder,
		entityConverter: entityConverter}
}

func (repo Repository[T]) Create(ctx context.Context, values ...interface{}) int64 {
	builder := repo.insertBuilder.Suffix(RETURNING_ID).Values(values...)
	var id int64
	err := repo.db.API().QueryRowContextInsert(ctx, builder).Scan(&id)
	if err != nil {
		panic(err)
	}
	return id
}

func (repo Repository[T]) GetById(ctx context.Context, id int64) T {
	builder := repo.selectBuilder.Where(sq.Eq{idColumn: id})
	row := repo.db.API().QueryRowContextSelect(ctx, builder)
	return repo.entityConverter(row)
}

func (repo Repository[T]) ConvertToObjects(rows pgx.Rows) []T {
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

func (repo Repository[T]) GetAll(ctx context.Context) []T {
	rows := repo.db.API().QueryContextSelect(ctx, repo.selectBuilder, nil)
	objs := repo.ConvertToObjects(rows)
	return objs
}

func (repo Repository[T]) GetBy(ctx context.Context, where sq.Sqlizer) []T {
	builder := repo.selectBuilder.Where(where)
	rows := repo.db.API().QueryContextSelect(ctx, builder, nil)
	objs := repo.ConvertToObjects(rows)
	return objs
}

func (repo Repository[T]) Delete(ctx context.Context, id int64) int64 {
	builder := repo.deleteBuilder.Where(sq.Eq{idColumn: id})
	return repo.db.API().ExecDelete(ctx, builder)
}

func updateInternal(ctx context.Context, api DbApi, updateBuilder sq.UpdateBuilder, fields map[string]interface{}) int64 {
	for column, value := range fields {
		updateBuilder = updateBuilder.Set(column, value)
	}
	return api.ExecUpdate(ctx, updateBuilder)
}

func (repo Repository[T]) Update(ctx context.Context, fields map[string]interface{}, id int64) int64 {
	builder := repo.updateBuilder.Where(sq.Eq{idColumn: id})
	return updateInternal(ctx, repo.db.API(), builder, fields)
}

func (repo Repository[T]) UpdateCollection(ctx context.Context, fields map[string]interface{}, where sq.Sqlizer) int64 {
	builder := repo.updateBuilder.Where(where)
	return updateInternal(ctx, repo.db.API(), builder, fields)
}

func (repo Repository[T]) UpdateReturning(ctx context.Context, builder sq.UpdateBuilder) T {
	row := repo.db.API().UpdateReturning(ctx, builder)
	return repo.entityConverter(row)
}

func (repo Repository[T]) UpdateReturningWithExtendedConverter(ctx context.Context, builder sq.UpdateBuilder, entityConverter func(row pgx.Row) T) T {
	row := repo.db.API().UpdateReturning(ctx, builder)
	return entityConverter(row)
}
