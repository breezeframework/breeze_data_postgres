package repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

type Child1Entity struct {
	ID   int64
	TYPE string
}

type Child1EntityRepository struct {
	pg.Repository[Child1Entity]
}

const CHILD1_TABLE = "TEST_CHILD1_TABLE "

func NewChild1EntityRepository(db pg.DbClient) Child1EntityRepository {
	repo, err := pg.NewRepository[Child1Entity](
		db,
		"",
		sq.Insert(CHILD1_TABLE).PlaceholderFormat(sq.Dollar).Columns("TYPE", "PARENT_ID"),
		sq.SelectBuilder{},
		sq.UpdateBuilder{},
		sq.DeleteBuilder{},
		child1EntityConverter)
	if err != nil {
		panic(err)
	}
	return Child1EntityRepository{repo}
}

func child1EntityConverter(row pgx.Row) Child1Entity {
	var obj Child1Entity
	if err := row.Scan(&obj.ID, &obj.TYPE); err != nil {
		panic(err)
	}
	return obj
}
