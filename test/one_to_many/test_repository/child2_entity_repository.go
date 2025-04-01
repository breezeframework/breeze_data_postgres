package test_repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

const CHILD2_TABLE = "TEST_CHILD2_TABLE "

type Child2Entity struct {
	ID        int64
	SIZE      string
	PARENT_ID int64
}

const (
	CHILD2ENTITY_ID        = "ID"
	CHILD2ENTITY_SIZE      = "SIZE"
	CHILD2ENTITY_PARENT_ID = "PARENT_ID"
)

var Child2Entity_Fields = []string{
	CHILD2ENTITY_ID,
	CHILD2ENTITY_SIZE,
	CHILD2ENTITY_PARENT_ID,
}

type Child2EntityRepository struct {
	pg.Repository
}

func NewChild2EntityRepository(db pg.DbClient) Child2EntityRepository {
	repo := pg.NewRepository(
		db,
		sq.Insert(CHILD2_TABLE).PlaceholderFormat(sq.Dollar).Columns(CHILD2ENTITY_SIZE, CHILD2ENTITY_PARENT_ID),
		sq.Select(Child2Entity_Fields...).From(CHILD2_TABLE),
		sq.UpdateBuilder{},
		sq.DeleteBuilder{},
		child2EntityConverter,
		nil,
		func(entity any) int64 { return entity.(*Child2Entity).ID })
	return Child2EntityRepository{repo}
}

func child2EntityConverter(row pgx.Row) any {
	var obj Child2Entity
	if err := row.Scan(&obj.ID, &obj.SIZE, &obj.PARENT_ID); err != nil {
		panic(err)
	}
	return obj
}

func OneToManyChild2Entity(db pg.DbClient) pg.Relation[ParentEntity, Child2Entity] {
	return pg.Relation[ParentEntity, Child2Entity]{
		ForeignKey: "ID",
		Repo:       NewChild2EntityRepository(db).Repository,
		ParentSetter: func(parent any, related any) {
			p := (*parent.(*interface{})).(*ParentEntity)
			if p.Children2 == nil {
				p.Children2 = []any{}
			}
			p.Children2 = append(p.Children2, related)
		},
		ParentIdGetter: func(child Child2Entity) int64 {
			return child.PARENT_ID
		},
	}
}
