package plain

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

const TABLE_NAME = "TEST_PLAIN_ENTITY_TABLE"

var (
	Entity_id     = "id"
	Entity_field1 = "field1"
	Entity_field2 = "field2"
)

var Entity_Fields = []string{
	Entity_id,
	Entity_field1,
	Entity_field2,
}

type TestPlainEntity struct {
	ID     int64 // ID field is mandatory
	Field1 int64
	Field2 string
}

func (p TestPlainEntity) GetID() int64 {
	return p.ID
}

type TestPlainEntityRepository struct {
	pg.Repository[TestPlainEntity]
}

var increaseField1Builder = sq.Update(TABLE_NAME).PlaceholderFormat(sq.Dollar).
	Set(Entity_field1, sq.Expr(Entity_field1+"+ 1")).Suffix("RETURNING " + Entity_id + ", " + Entity_field1 + "" + ", " + Entity_field2)

func NewTestPlainEntityRepository(db pg.DbClient) TestPlainEntityRepository {
	repo := pg.NewRepository(
		TestPlainEntity{},
		db,
		sq.Insert(TABLE_NAME).PlaceholderFormat(sq.Dollar).Columns(Entity_field1, Entity_field2),
		sq.Select(Entity_Fields...).PlaceholderFormat(sq.Dollar).From(TABLE_NAME),
		sq.Update(TABLE_NAME).PlaceholderFormat(sq.Dollar),
		sq.Delete(TABLE_NAME).PlaceholderFormat(sq.Dollar),
		testPlainEntityConverter)

	return TestPlainEntityRepository{repo}
}

func testPlainEntityConverter(row pgx.Row) any {
	var entity TestPlainEntity
	if err := row.Scan(&entity.ID, &entity.Field1, &entity.Field2); err != nil {
		panic(err)
	}
	return &entity
}

func (repo *TestPlainEntityRepository) GetOneByField2(ctx context.Context, field2 string) any {
	list := repo.GetBy(ctx, sq.Eq{Entity_field2: field2})
	if list != nil && len(list) > 0 {
		return (list)[0]
	}
	var ret TestPlainEntity
	return ret
}

func (repo *TestPlainEntityRepository) IncreaseField1(ctx context.Context, id int64) int64 {
	updated := repo.UpdateReturning(ctx, increaseField1Builder.Where(sq.Eq{Entity_id: id}))
	return updated.(*TestPlainEntity).Field1
}
