package repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

// Mandatory declare ID field
type ParentEntity struct {
	ID        int64
	Name      string
	Children1 []Child1Entity
	Children2 []Child2Entity
}

type ParentEntityRepository struct {
	pg.Repository[ParentEntity]
}

const TABLE_NAME = "TEST_PARENT_ENTITY_TABLE "

func NewParentEntityRepository(db pg.DbClient) ParentEntityRepository {
	repo, err := pg.NewRepository[ParentEntity](
		db,
		"p",
		sq.Insert(TABLE_NAME).PlaceholderFormat(sq.Dollar).Columns("NAME"),
		sq.Select("p.id", "p.name", "c1.id", "c1.type", "c2.id", "c2.size").PlaceholderFormat(sq.Dollar).
			From(TABLE_NAME+" p").
			LeftJoin(CHILD1_TABLE+"c1 "+"ON p.id = c1.PARENT_ID").
			LeftJoin(CHILD2_TABLE+"c2 "+"ON p.id = c2.PARENT_ID"),
		sq.UpdateBuilder{},
		sq.DeleteBuilder{},
		parentEntityConverter)
	if err != nil {
		panic(err)
	}
	return ParentEntityRepository{repo}
}

func parentEntityConverter(rows pgx.Rows) ParentEntity {
	var result ParentEntity
	var children1 []Child1Entity
	var children2 []Child2Entity
	firstIteration := true

	for rows.Next() {
		var main ParentEntity
		var child1 Child1Entity
		var child2 Child2Entity

		if err := rows.Scan(&main.ID, &main.Name, &child1.ID, &child1.TYPE, &child2.ID, &child2.SIZE); err != nil {
			panic(err)
		}

		if firstIteration {
			result = main
			firstIteration = false
		}

		children1 = append(children1, child1)
		children2 = append(children2, child2)
	}

	result.Children1 = children1
	result.Children2 = children2

	return result
}
