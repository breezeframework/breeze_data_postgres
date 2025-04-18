package test_repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

const CHILD1_TABLE = "TEST_CHILD1_TABLE "

type Child1Entity struct {
	ID        int64  `json:"ID"` // ID field is mandatory
	TYPE      string `json:"type"`
	PARENT_ID int64  `json:"PARENT_ID"`
}

func (child *Child1Entity) GetParentID() int64 {
	return child.PARENT_ID
}

func (child *Child1Entity) PushToParent(parent any) {
	par := parent.(*ParentEntity)
	par.Children1 = append(par.Children1, child)
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
	pg.Repository[Child1Entity]
}

func NewChild1EntityRepository(db pg.DbClient) Child1EntityRepository {
	repo := pg.NewRepository[Child1Entity](
		Child1Entity{},
		db,
		sq.Insert(CHILD1_TABLE).PlaceholderFormat(sq.Dollar).Columns(CHILD1ENTITY_TYPE, CHILD1ENTITY_PARENT_ID),
		sq.Select(Child1Entity_Fields...).From(CHILD1_TABLE),
		sq.UpdateBuilder{},
		sq.DeleteBuilder{},
		sq.InsertBuilder{},
		child1EntityConverter)
	return Child1EntityRepository{repo}
}

func child1EntityConverter(row pgx.Row) *Child1Entity {
	var obj Child1Entity
	if err := row.Scan(&obj.ID, &obj.TYPE, &obj.PARENT_ID); err != nil {
		panic(err)
	}
	return &obj
}

func OneToManyChild1EntityRelation(db pg.DbClient) pg.Relation[Child1Entity] {
	return pg.Relation[Child1Entity]{
		ForeignKey: CHILD1ENTITY_PARENT_ID,
		Repo:       NewChild1EntityRepository(db).Repository,
		ParentIdGetter: func(child Child1Entity) int64 {
			return child.PARENT_ID
		},
	}
}
