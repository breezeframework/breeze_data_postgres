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
	anchor        T
	DB            PgDbClient
	InsertBuilder sq.InsertBuilder
	SelectBuilder sq.SelectBuilder
	UpdateBuilder sq.UpdateBuilder
	DeleteBuilder sq.DeleteBuilder
	Converter     func(row pgx.Row) any // type is any to allow generalization
	Relations     []Relation[any]       // the relation type is any because it really any entity
	AddRelated    func(*T, any)
	AddRelation   func(Relation[any])
}

func WrapRepository[R any](repo Repository[R]) Repository[any] {
	return Repository[any]{
		anchor:        repo.anchor,
		DB:            repo.DB,
		InsertBuilder: repo.InsertBuilder,
		SelectBuilder: repo.SelectBuilder,
		UpdateBuilder: repo.UpdateBuilder,
		DeleteBuilder: repo.DeleteBuilder,
		Converter: func(row pgx.Row) any {
			return repo.Converter(row) // Уже возвращает any, можно передавать напрямую
		},
		Relations: repo.Relations, // Уже []IRelation[any], копирование не нужно
		AddRelated: func(target *any, related any) {
			if tgt, ok := (*target).(R); ok {
				repo.AddRelated(&tgt, related)
				*target = any(tgt) // Обновляем значение
			}
		},
		AddRelation: repo.AddRelation, // Можно передать напрямую, так как уже `Relation[any]`
	}
}

type Identifiable interface {
	GetID() int64
}

type Related interface {
	GetParentID() int64
	PushToParent(parent any)
}

type Relation[R any] struct {
	ForeignKey     string
	Repo           Repository[R]
	ParentIdGetter func(R) int64
}

func WrapRelation[R any](r Relation[R]) Relation[any] {
	return Relation[any]{
		ForeignKey: r.ForeignKey,
		Repo:       WrapRepository(r.Repo), // Приведение репозитория к `any`
		ParentIdGetter: func(t any) int64 {
			if val, ok := t.(R); ok {
				return r.ParentIdGetter(val)
			}
			panic("cannot cast foreign key")
		},
	}
}

func (r Relation[R]) GetParentId(child R) int64 {
	return r.ParentIdGetter(child)
}

func (r Relation[R]) getRepo() Repository[R] {
	return r.Repo
}

func (r Relation[R]) GetForeignKey() string {
	return r.ForeignKey
}

func NewRepository[T any](
	anchor T,
	db DbClient,
	insertBuilder sq.InsertBuilder,
	selectBuilder sq.SelectBuilder,
	updateBuilder sq.UpdateBuilder,
	deleteBuilder sq.DeleteBuilder,
	converter func(row pgx.Row) any) Repository[T] {
	return Repository[T]{
		anchor:        anchor,
		DB:            db.Pg(),
		InsertBuilder: insertBuilder, SelectBuilder: selectBuilder, UpdateBuilder: updateBuilder, DeleteBuilder: deleteBuilder,
		Converter: converter,
	}
}

func (repo *Repository[T]) loadRelations(ctx context.Context, parentEntities []*T) {
	if len(repo.Relations) == 0 {
		return
	}
	var parentIds []int64
	for _, entity := range parentEntities {
		var ident Identifiable
		if v, ok := any(*entity).(Identifiable); ok { // Используем any для приведения к интерфейсу
			ident = v
			parentIds = append(parentIds, ident.GetID())
		}
	}

	parentMap := make(map[int64]*T)
	for i := 0; i < len(parentEntities); i++ {
		parentMap[parentIds[i]] = parentEntities[i]
	}

	for _, rel := range repo.Relations {
		whereClause := sq.Eq{rel.GetForeignKey(): parentIds}
		relatedObjects := rel.getRepo().GetBy(ctx, whereClause)

		for _, related := range relatedObjects {
			var parentId int64
			if rel, ok := any(related).(Related); ok {
				parentId = rel.GetParentID()
				if parent, ok := parentMap[parentId]; ok {
					rel.PushToParent(parent)
				}
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
		objs = append(objs, obj)
		repo.loadRelations(ctx, objs)
	}
	return *obj
}

func (repo Repository[T]) getById(ctx context.Context, builder sq.SelectBuilder) *T {
	row := repo.DB.API().QueryRowContextSelect(ctx, builder)
	obj := repo.Converter(row)
	return obj.(*T)
}

func (repo Repository[T]) convertToObjects(rows pgx.Rows) []T {
	var objs []T
	for rows.Next() {
		obj := repo.Converter(rows)
		if t, ok := obj.(*T); ok {
			objs = append(objs, *t)
		} else {
			objs = append(objs, obj.(T))
		}
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
