package test_repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

const CHILD2_TABLE = "TEST_CHILD2_TABLE "

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

type Child2Entity struct {
	ID        int64   `json:"ID"` // ID field is mandatory
	SIZE      float64 `json:"size"`
	PARENT_ID int64   `json:"PARENT_ID"`
}

func (child *Child2Entity) GetParentID() int64 {
	return child.PARENT_ID
}

func (child *Child2Entity) PushToParent(parent any) {
	par := parent.(*ParentEntity)
	par.Children2 = append(par.Children2, child)
}

type Child2EntityRepository struct {
	pg.Repository[Child2Entity]
}

func NewChild2EntityRepository(db pg.DbClient) Child2EntityRepository {
	repo := pg.NewRepository[Child2Entity](
		Child2Entity{},
		db,
		sq.Insert(CHILD2_TABLE).PlaceholderFormat(sq.Dollar).Columns(CHILD2ENTITY_SIZE, CHILD2ENTITY_PARENT_ID),
		sq.Select(Child2Entity_Fields...).From(CHILD2_TABLE),
		sq.UpdateBuilder{},
		sq.DeleteBuilder{},
		sq.InsertBuilder{},
		child2EntityConverter)
	return Child2EntityRepository{repo}
}

func child2EntityConverter(row pgx.Row) *Child2Entity {
	var obj Child2Entity
	if err := row.Scan(&obj.ID, &obj.SIZE, &obj.PARENT_ID); err != nil {
		panic(err)
	}
	return &obj
}

func OneToManyChild2EntityRelation(db pg.DbClient) pg.Relation[Child2Entity] {
	return pg.Relation[Child2Entity]{
		ForeignKey: CHILD2ENTITY_PARENT_ID,
		Repo:       NewChild2EntityRepository(db).Repository,
		ParentIdGetter: func(child Child2Entity) int64 {
			return child.PARENT_ID
		},
	}
}
