package repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

type Child2Entity struct {
	ID   int64
	SIZE float64
}

type Child2EntityRepository struct {
	pg.Repository[Child2Entity]
}

const CHILD2_TABLE = "TEST_CHILD2_TABLE "

func NewChild2EntityRepository(db pg.DbClient) Child2EntityRepository {
	repo, err := pg.NewPostgrePlainEntityRepository[Child2Entity](
		db,
		"",
		sq.Insert(CHILD2_TABLE).PlaceholderFormat(sq.Dollar).Columns("SIZE", "PARENT_ID"),
		sq.SelectBuilder{},
		sq.UpdateBuilder{},
		sq.DeleteBuilder{},
		child2EntityConverter)
	if err != nil {
		panic(err)
	}
	return Child2EntityRepository{repo}
}

func child2EntityConverter(row pgx.Row) Child2Entity {
	var obj Child2Entity
	if err := row.Scan(&obj.ID, &obj.SIZE); err != nil {
		panic(err)
	}
	return obj
}
