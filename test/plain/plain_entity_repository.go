package plain

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

// Mandatory declare ID field
type TestPlainEntity struct {
	ID     int64
	Field1 int64
	Field2 string
}

type TestPlainEntityRepository struct {
	pg.Repository[TestPlainEntity]
}

const TABLE_NAME = "TEST_PLAIN_ENTITY_TABLE"

var increaseField1Builder = sq.Update(TABLE_NAME).PlaceholderFormat(sq.Dollar).
	Set("field1", sq.Expr("field1 + 1")).Suffix("RETURNING id, field1, field2")

func NewTestPlainEntityRepository(db pg.DbClient) TestPlainEntityRepository {
	repo, err := pg.NewRepository[TestPlainEntity](
		db,
		"",
		sq.Insert(TABLE_NAME).PlaceholderFormat(sq.Dollar).Columns("field1", "field2"),
		sq.Select("id", "field1", "field2").PlaceholderFormat(sq.Dollar).From(TABLE_NAME),
		sq.Update(TABLE_NAME).PlaceholderFormat(sq.Dollar),
		sq.Delete(TABLE_NAME).PlaceholderFormat(sq.Dollar),
		testPlainEntityConverter)
	if err != nil {
		panic(err)
	}
	return TestPlainEntityRepository{repo}
}

func testPlainEntityConverter(row pgx.Row) TestPlainEntity {
	var myObj TestPlainEntity
	if err := row.Scan(&myObj.ID, &myObj.Field1, &myObj.Field2); err != nil {
		panic(err)
	}
	return myObj
}

func (repo *TestPlainEntityRepository) GetOneByField2(ctx context.Context, field2 string) TestPlainEntity {
	list := repo.GetBy(ctx, sq.Eq{"field2": field2})
	if list != nil && len(list) > 0 {
		return (list)[0]
	}
	var ret TestPlainEntity
	return ret
}

func (repo *TestPlainEntityRepository) IncreaseField1(ctx context.Context, id int64) int64 {
	updated := repo.UpdateReturning(ctx, increaseField1Builder.Where(sq.Eq{"id": id}))
	return updated.Field1
}
