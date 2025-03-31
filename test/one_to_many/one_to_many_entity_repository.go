package oneToMany_entity_repository

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

type Child1Entity struct {
	ID   int64
	TYPE string
}

type Child2Entity struct {
	ID   int64
	SIZE float64
}

type ParentEntityRepository struct {
	pg.Repository[ParentEntity]
}

type Child1EntityRepository struct {
	pg.Repository[Child1Entity]
}

type Child2EntityRepository struct {
	pg.Repository[Child2Entity]
}

const TABLE_NAME = "TEST_PARENT_ENTITY_TABLE "

const CHILD1_TABLE = "TEST_CHILD1_TABLE "

const CHILD2_TABLE = "TEST_CHILD2_TABLE "

func NewParentEntityRepository(db pg.DbClient) ParentEntityRepository {
	repo, err := pg.NewPostgreOneToManyEntityRepository[ParentEntity](
		db,
		"p",
		sq.Insert(TABLE_NAME).PlaceholderFormat(sq.Dollar).Columns("NAME"),
		sq.Select("p.id", "p.name", "c1.id", "c1.type", "c2.id", "c2.size").PlaceholderFormat(sq.Dollar).
			From(TABLE_NAME+" p").
			LeftJoin(CHILD1_TABLE+"c1 "+"ON p.id = c1.PARENT_ID").
			LeftJoin(CHILD2_TABLE+"c2 "+"ON p.id = c2.PARENT_ID"),
		sq.UpdateBuilder{},
		sq.DeleteBuilder{},
		testParentEntityConverter)
	if err != nil {
		panic(err)
	}
	return ParentEntityRepository{repo}
}

func NewChild1EntityRepository(db pg.DbClient) Child1EntityRepository {
	repo, err := pg.NewPostgrePlainEntityRepository[Child1Entity](
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

func child1EntityConverter(row pgx.Row) Child1Entity {
	var obj Child1Entity
	if err := row.Scan(&obj.ID, &obj.TYPE); err != nil {
		panic(err)
	}
	return obj
}

func child2EntityConverter(row pgx.Row) Child2Entity {
	var obj Child2Entity
	if err := row.Scan(&obj.ID, &obj.SIZE); err != nil {
		panic(err)
	}
	return obj
}

func testParentEntityConverter(rows pgx.Rows) ParentEntity {
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
