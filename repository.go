package pg

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	idColumn     = "id"
	RETURNING_ID = "RETURNING id"
)

type IRepository interface {
	Create(ctx context.Context, values ...interface{}) int64
	GetById(ctx context.Context, id int64) any
	GetAll(ctx context.Context) []any
	GetBy(ctx context.Context, where sq.Sqlizer) []any
	Delete(ctx context.Context, id int64) int64
	UpdateCollection(ctx context.Context, fields map[string]interface{}, where sq.Sqlizer) int64
	Update(ctx context.Context, fields map[string]interface{}, id int64) int64
	UpdateReturning(ctx context.Context, builder sq.UpdateBuilder, entityConverter func(row pgx.Row) any) any
}

type Repository[T any] struct {
	DB            PgDbClient
	InsertBuilder sq.InsertBuilder
	SelectBuilder sq.SelectBuilder
	UpdateBuilder sq.UpdateBuilder
	DeleteBuilder sq.DeleteBuilder
	Converter     func(row pgx.Row) any
	Relations     []IRelation[T, any]
	IdGetter      func(any) int64
}

type Relation[T any, R any] struct {
	ForeignKey     string
	Repo           Repository[R]
	ParentSetter   func(any, any)
	ParentIdGetter func(R) int64
}

type IRelation[T any, R any] interface {
	getRepo() Repository[R]
	GetForeignKey() string
	SetParent(parent T, related R)
	GetParentId(parent any) int64
}

func (r Relation[T, any]) SetParent(parent T, related any) {
	r.ParentSetter(parent, related)
}

func (r Relation[T, R]) GetParentId(child any) int64 {
	return r.ParentIdGetter(child.(R))
}

func (r Relation[T, R]) getRepo() Repository[R] {
	return r.Repo
}

func convertRelations[T any](relations []IRelation[any, any]) []IRelation[T, any] {
	var result []IRelation[T, any]
	for _, rel := range relations {
		if typedRel, ok := rel.(IRelation[T, any]); ok {
			result = append(result, typedRel)
		}
	}
	return result
}

func ConvertRepo[T any](anyRepo Repository[any]) Repository[T] {
	convertedRepo := Repository[T]{
		DB:            anyRepo.DB,
		InsertBuilder: anyRepo.InsertBuilder,
		SelectBuilder: anyRepo.SelectBuilder,
		UpdateBuilder: anyRepo.UpdateBuilder,
		DeleteBuilder: anyRepo.DeleteBuilder,
		Converter:     anyRepo.Converter,
		Relations:     convertRelations[T](anyRepo.Relations),
		IdGetter:      anyRepo.IdGetter,
	}
	return convertedRepo
}

func (r Relation[T, R]) GetForeignKey() string {
	return r.ForeignKey
}

func wrapFunc[T any](f func(T) int64) func(any) int64 {
	return func(x any) int64 {
		v, ok := x.(T) // Проверяем, что x действительно типа T
		if !ok {
			panic("Неверный тип аргумента")
		}
		return f(v)
	}
}
func NewRepository[T any](
	db DbClient,
	insertBuilder sq.InsertBuilder,
	selectBuilder sq.SelectBuilder,
	updateBuilder sq.UpdateBuilder,
	deleteBuilder sq.DeleteBuilder,
	converter func(row pgx.Row) any,
	relations []IRelation[T, any],
	idGetter func(T) int64) Repository[T] {
	return Repository[T]{
		DB:            db.Pg(),
		InsertBuilder: insertBuilder, SelectBuilder: selectBuilder, UpdateBuilder: updateBuilder, DeleteBuilder: deleteBuilder,
		Converter: converter,
		Relations: relations,
		IdGetter:  wrapFunc(idGetter)}
}

func (repo *Repository[T]) loadRelations(ctx context.Context, parentEntities []*T) {
	if len(repo.Relations) == 0 {
		return
	}
	var parentIds []int64
	for _, entity := range parentEntities {
		parentIds = append(parentIds, repo.IdGetter(entity))
	}

	parentMap := make(map[int64]any)
	for i := 0; i < len(parentEntities); i++ {
		parentMap[parentIds[i]] = parentEntities[i]
	}

	for _, rel := range repo.Relations {
		whereClause := sq.Eq{rel.GetForeignKey(): parentIds}
		relatedObjects := rel.getRepo().GetBy(ctx, whereClause)

		for _, related := range relatedObjects {
			if parent, ok := parentMap[rel.GetParentId(related.(T))]; ok {
				rel.SetParent(parent.(T), related.(any))
			}
		}
	}
}

func (repo Repository[T]) Create(ctx context.Context, values ...interface{}) int64 {
	builder := repo.InsertBuilder.Suffix(RETURNING_ID).Values(values...)
	var id int64
	err := repo.DB.API().QueryRowContextInsert(ctx, builder).Scan(&id)
	if err != nil {
		panic(err)
	}
	return id
}

func (repo Repository[T]) GetById(ctx context.Context, id int64) T {
	builder := repo.SelectBuilder.Where(sq.Eq{idColumn: id})
	obj := repo.getById(ctx, builder)
	if len(repo.Relations) > 0 {
		var objs []*T
		objs = append(objs, &obj)
		repo.loadRelations(ctx, objs)
	}
	return obj
}

func (repo Repository[T]) getById(ctx context.Context, builder sq.SelectBuilder) T {
	row := repo.DB.API().QueryRowContextSelect(ctx, builder)
	obj := repo.Converter(row)
	return obj.(T)
}

func (repo Repository[T]) convertToObjects(rows pgx.Rows) []T {
	var objs []T
	for rows.Next() {
		obj := repo.Converter(rows)
		objs = append(objs, obj.(T))
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return objs
}

func (repo Repository[T]) GetAll(ctx context.Context) []T {
	rows := repo.DB.API().QueryContextSelect(ctx, repo.SelectBuilder, nil)
	objs := repo.convertToObjects(rows)
	if len(repo.Relations) > 0 {
		var objPtrs []*T
		for i := range objs {
			objPtrs = append(objPtrs, &objs[i])
		}
		repo.loadRelations(ctx, objPtrs)
	}
	return objs
}

func (repo Repository[T]) GetBy(ctx context.Context, where sq.Sqlizer) []T {
	builder := repo.SelectBuilder.Where(where)
	rows := repo.DB.API().QueryContextSelect(ctx, builder, nil)
	objs := repo.convertToObjects(rows)
	return objs
}

func (repo Repository[T]) Delete(ctx context.Context, id int64) int64 {
	builder := repo.DeleteBuilder.Where(sq.Eq{idColumn: id})
	return repo.DB.API().ExecDelete(ctx, builder)
}

func update(ctx context.Context, api DbApi, updateBuilder sq.UpdateBuilder, fields map[string]interface{}) int64 {
	for column, value := range fields {
		updateBuilder = updateBuilder.Set(column, value)
	}
	return api.ExecUpdate(ctx, updateBuilder)
}

func (repo Repository[T]) Update(ctx context.Context, fields map[string]interface{}, id int64) int64 {
	builder := repo.UpdateBuilder.Where(sq.Eq{idColumn: id})
	return update(ctx, repo.DB.API(), builder, fields)
}

func (repo Repository[T]) UpdateCollection(ctx context.Context, fields map[string]interface{}, where sq.Sqlizer) int64 {
	builder := repo.UpdateBuilder.Where(where)
	return update(ctx, repo.DB.API(), builder, fields)
}

func (repo Repository[T]) UpdateReturning(ctx context.Context, builder sq.UpdateBuilder) any {
	row := repo.DB.API().UpdateReturning(ctx, builder)
	return repo.Converter(row)
}

func (repo Repository[T]) UpdateReturningWithExtendedConverter(ctx context.Context, builder sq.UpdateBuilder, entityConverter func(row pgx.Row) any) any {
	row := repo.DB.API().UpdateReturning(ctx, builder)
	return entityConverter(row)
}
