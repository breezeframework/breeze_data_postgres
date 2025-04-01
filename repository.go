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

type Repository struct {
	db            pg_api.PgDbClient
	insertBuilder sq.InsertBuilder
	selectBuilder sq.SelectBuilder
	updateBuilder sq.UpdateBuilder
	deleteBuilder sq.DeleteBuilder
	converter     func(row pgx.Row) any
	relations     []IRelation
	idGetter      func(any) int64
}

type Relation[T any, R any] struct {
	ForeignKey     string
	Repo           Repository
	ParentSetter   func(any, any)
	ParentIdGetter func(R) int64
}

func (r Relation[T, R]) getRepo() Repository {
	return r.Repo
}

func (r Relation[T, R]) SetParent(parent any, related any) {
	r.ParentSetter(parent, related)
}

type IRelation interface {
	getRepo() Repository
	GetForeignKey() string
	SetParent(parent any, related any)
	GetParentId(parent any) int64
}

func (r Relation[T, R]) GetForeignKey() string {
	return r.ForeignKey
}

func (r Relation[T, R]) GetParentId(child any) int64 {
	return r.ParentIdGetter(child.(R))
}

func NewRepository(
	db DbClient,
	insertBuilder sq.InsertBuilder,
	selectBuilder sq.SelectBuilder,
	updateBuilder sq.UpdateBuilder,
	deleteBuilder sq.DeleteBuilder,
	converter func(row pgx.Row) any,
	relations []IRelation,
	idGetter func(any) int64) Repository {
	return Repository{
		db:            db.(pg_api.PgDbClient),
		insertBuilder: insertBuilder, selectBuilder: selectBuilder, updateBuilder: updateBuilder, deleteBuilder: deleteBuilder,
		converter: converter,
		relations: relations,
		idGetter:  idGetter}
}

func (repo *Repository) loadRelations(ctx context.Context, parentEntities []any) {
	if len(repo.relations) == 0 {
		return
	}
	var parentIds []int64
	for _, entity := range parentEntities {
		parentIds = append(parentIds, repo.idGetter(entity))
	}

	parentMap := make(map[int64]any)
	for i := 0; i < len(parentEntities); i++ {
		parentMap[parentIds[i]] = parentEntities[i]
	}

	for _, rel := range repo.relations {
		whereClause := sq.Eq{rel.GetForeignKey(): parentIds}
		relatedObjects := rel.getRepo().GetBy(ctx, whereClause)

		for _, related := range relatedObjects {
			if parent, ok := parentMap[rel.GetParentId(related)]; ok {
				rel.SetParent(parent, related)
			}
		}
	}
}

func (repo Repository) Create(ctx context.Context, values ...interface{}) int64 {
	builder := repo.insertBuilder.Suffix(RETURNING_ID).Values(values...)
	var id int64
	err := repo.db.API().QueryRowContextInsert(ctx, builder).Scan(&id)
	if err != nil {
		panic(err)
	}
	return id
}

func (repo Repository) GetById(ctx context.Context, id int64) any {
	builder := repo.selectBuilder.Where(sq.Eq{idColumn: id})
	obj := repo.getById(ctx, builder)
	if len(repo.relations) > 0 {
		var objs []any
		objs = append(objs, &obj)
		repo.loadRelations(ctx, objs)
	}
	return obj
}

func (repo Repository) getById(ctx context.Context, builder sq.SelectBuilder) any {
	row := repo.db.API().QueryRowContextSelect(ctx, builder)
	obj := repo.converter(row)
	return obj
}

func (repo Repository) convertToObjects(rows pgx.Rows) []any {
	var objs []any
	for rows.Next() {
		obj := repo.converter(rows)
		objs = append(objs, obj)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return objs
}

func (repo Repository) GetAll(ctx context.Context) []any {
	rows := repo.db.API().QueryContextSelect(ctx, repo.selectBuilder, nil)
	objs := repo.convertToObjects(rows)
	if len(repo.relations) > 0 {
		var objPtrs []any
		for i := range objs {
			objPtrs = append(objPtrs, &objs[i])
		}
		repo.loadRelations(ctx, objPtrs)
	}
	return objs
}

func (repo Repository) GetBy(ctx context.Context, where sq.Sqlizer) []any {
	builder := repo.selectBuilder.Where(where)
	rows := repo.db.API().QueryContextSelect(ctx, builder, nil)
	objs := repo.convertToObjects(rows)
	return objs
}

func (repo Repository) Delete(ctx context.Context, id int64) int64 {
	builder := repo.deleteBuilder.Where(sq.Eq{idColumn: id})
	return repo.db.API().ExecDelete(ctx, builder)
}

func update(ctx context.Context, api DbApi, updateBuilder sq.UpdateBuilder, fields map[string]interface{}) int64 {
	for column, value := range fields {
		updateBuilder = updateBuilder.Set(column, value)
	}
	return api.ExecUpdate(ctx, updateBuilder)
}

func (repo Repository) Update(ctx context.Context, fields map[string]interface{}, id int64) int64 {
	builder := repo.updateBuilder.Where(sq.Eq{idColumn: id})
	return update(ctx, repo.db.API(), builder, fields)
}

func (repo Repository) UpdateCollection(ctx context.Context, fields map[string]interface{}, where sq.Sqlizer) int64 {
	builder := repo.updateBuilder.Where(where)
	return update(ctx, repo.db.API(), builder, fields)
}

func (repo Repository) UpdateReturning(ctx context.Context, builder sq.UpdateBuilder) any {
	row := repo.db.API().UpdateReturning(ctx, builder)
	return repo.converter(row)
}

func (repo Repository) UpdateReturningWithExtendedConverter(ctx context.Context, builder sq.UpdateBuilder, entityConverter func(row pgx.Row) any) any {
	row := repo.db.API().UpdateReturning(ctx, builder)
	return entityConverter(row)
}
