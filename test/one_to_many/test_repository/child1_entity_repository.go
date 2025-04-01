package test_repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

const CHILD1_TABLE = "TEST_CHILD1_TABLE "

type Child1Entity struct {
	ID        int64  `json:"ID"`
	TYPE      string `json:"type"`
	PARENT_ID int64  `json:"PARENT_ID"`
}

const (
	CHILD1ENTITY_ID        = "ID"
	CHILD1ENTITY_TYPE      = "TYPE"
	CHILD1ENTITY_PARENT_ID = "PARENT_ID"
)

var Child1Entity_Fields = []string{
	CHILD1ENTITY_ID,
	CHILD1ENTITY_TYPE,
	CHILD1ENTITY_PARENT_ID,
}

type Child1EntityRepository struct {
	pg.Repository
}

func NewChild1EntityRepository(db pg.DbClient) Child1EntityRepository {
	repo := pg.NewRepository(
		db,
		sq.Insert(CHILD1_TABLE).PlaceholderFormat(sq.Dollar).Columns(CHILD1ENTITY_TYPE, CHILD1ENTITY_PARENT_ID),
		sq.Select(Child1Entity_Fields...).From(CHILD1_TABLE),
		sq.UpdateBuilder{},
		sq.DeleteBuilder{},
		child1EntityConverter,
		nil,
		func(entity any) int64 { return entity.(*Child1Entity).ID })
	return Child1EntityRepository{repo}
}

func child1EntityConverter(row pgx.Row) any {
	var obj Child1Entity
	if err := row.Scan(&obj.ID, &obj.TYPE, &obj.PARENT_ID); err != nil {
		panic(err)
	}
	return obj
}

func OneToManyChild1Entity(db pg.DbClient) pg.Relation[ParentEntity, Child1Entity] {
	return pg.Relation[ParentEntity, Child1Entity]{
		ForeignKey: CHILD1ENTITY_PARENT_ID,
		Repo:       NewChild1EntityRepository(db).Repository,
		ParentSetter: func(parent any, related any) {
			p := (*parent.(*interface{})).(*ParentEntity)
			if p.Children1 == nil {
				p.Children1 = []any{}
			}
			p.Children1 = append(p.Children1, related)
		},
		ParentIdGetter: func(child Child1Entity) int64 {
			return child.PARENT_ID
		},
	}
}
