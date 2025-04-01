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

type Repository[T any] struct {
	db            pg_api.PgDbClient
	tableAlias    string
	insertBuilder sq.InsertBuilder
	selectBuilder sq.SelectBuilder
	updateBuilder sq.UpdateBuilder
	deleteBuilder sq.DeleteBuilder
	converter     func(row pgx.Row) T
	relations     []relation[T]
	getID         func(T) int64
}

type relation[T any] struct {
	foreignKey  string
	repo        Repository[any]
	setFunc     func(*T, any)
	getParentID func(T) int64
}

func (repo Repository[T]) AddRelation(
	relatedRepo Repository[any],
	foreignKey string,
	setFunc func(*T, []any),
) {
	repo.relations = append(repo.relations, relation[T]{
		foreignKey: foreignKey,
		repo:       relatedRepo,
		setFunc: func(parent *T, related any) {
			setFunc(parent, related.([]any)) // Приводим `any` обратно в `[]R`
		},
	})
}

func NewRepository[T any](
	db DbClient,
	tableAlias string,
	insertBuilder sq.InsertBuilder,
	selectBuilder sq.SelectBuilder,
	updateBuilder sq.UpdateBuilder,
	deleteBuilder sq.DeleteBuilder,
	converter func(row pgx.Row) T) (Repository[T], error) {
	return Repository[T]{
		db:            db.(pg_api.PgDbClient),
		tableAlias:    tableAlias,
		insertBuilder: insertBuilder, selectBuilder: selectBuilder, updateBuilder: updateBuilder, deleteBuilder: deleteBuilder,
		converter: converter,
	}, nil
}

func (repo *Repository[T]) loadRelations(ctx context.Context, parentEntities []T) {
	if len(repo.relations) == 0 {
		return
	}
	var parentIds []int64
	for _, entity := range parentEntities {
		parentIds = append(parentIds, repo.getID(entity))
	}

	parentMap := make(map[int64]*T)
	for i := 0; i < len(parentEntities); i++ {
		parentMap[parentIds[i]] = &parentEntities[i]
	}

	for _, rel := range repo.relations {
		whereClause := sq.Eq{rel.foreignKey: parentIds}
		relatedObjects := rel.repo.GetBy(ctx, whereClause)

		for _, related := range relatedObjects {
			if parent, ok := parentMap[rel.getParentID(related)]; ok {
				rel.setFunc(parent, []any{related})
			}
		}
	}
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
	var idClouse string
	if repo.tableAlias == "" {
		idClouse = idColumn
	} else {
		idClouse = repo.tableAlias + "." + idColumn
	}
	builder := repo.selectBuilder.Where(sq.Eq{idClouse: id})
	obj := repo.getById(ctx, builder)
	if len(repo.relations) > 0 {
		repo.loadRelations(ctx, []T{obj})
	}
	return obj
}

func (repo Repository[T]) getById(ctx context.Context, builder sq.SelectBuilder) T {
	row := repo.db.API().QueryRowContextSelect(ctx, builder)
	obj := repo.converter(row)
	return obj
}

func (repo Repository[T]) convertToObjects(rows pgx.Rows) []T {
	var objs []T
	for rows.Next() {
		obj := repo.converter(rows)
		objs = append(objs, obj)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return objs
}

func (repo Repository[T]) GetAll(ctx context.Context) []T {
	rows := repo.db.API().QueryContextSelect(ctx, repo.selectBuilder, nil)
	objs := repo.convertToObjects(rows)
	if len(repo.relations) > 0 {
		repo.loadRelations(ctx, objs)
	}
	return objs
}

func (repo Repository[T]) GetBy(ctx context.Context, where sq.Sqlizer) []T {
	builder := repo.selectBuilder.Where(where)
	rows := repo.db.API().QueryContextSelect(ctx, builder, nil)
	objs := repo.convertToObjects(rows)
	return objs
}

func (repo Repository[T]) Delete(ctx context.Context, id int64) int64 {
	builder := repo.deleteBuilder.Where(sq.Eq{idColumn: id})
	return repo.db.API().ExecDelete(ctx, builder)
}

func update(ctx context.Context, api DbApi, updateBuilder sq.UpdateBuilder, fields map[string]interface{}) int64 {
	for column, value := range fields {
		updateBuilder = updateBuilder.Set(column, value)
	}
	return api.ExecUpdate(ctx, updateBuilder)
}

func (repo Repository[T]) Update(ctx context.Context, fields map[string]interface{}, id int64) int64 {
	builder := repo.updateBuilder.Where(sq.Eq{idColumn: id})
	return update(ctx, repo.db.API(), builder, fields)
}

func (repo Repository[T]) UpdateCollection(ctx context.Context, fields map[string]interface{}, where sq.Sqlizer) int64 {
	builder := repo.updateBuilder.Where(where)
	return update(ctx, repo.db.API(), builder, fields)
}

func (repo Repository[T]) UpdateReturning(ctx context.Context, builder sq.UpdateBuilder) T {
	row := repo.db.API().UpdateReturning(ctx, builder)
	return repo.converter(row)
}

func (repo Repository[T]) UpdateReturningWithExtendedConverter(ctx context.Context, builder sq.UpdateBuilder, entityConverter func(row pgx.Row) T) T {
	row := repo.db.API().UpdateReturning(ctx, builder)
	return entityConverter(row)
}
