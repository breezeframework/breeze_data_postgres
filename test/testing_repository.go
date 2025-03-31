package testrepository

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

// Mandatory declare ID field
type MyObj struct {
	ID     int64
	Field1 int64
	Field2 string
}

type MyObjRepository struct {
	pg.Repository[MyObj]
}

var increaseField1Builder = sq.Update("MyObjTable").PlaceholderFormat(sq.Dollar).
	Set("field1", sq.Expr("field1 + 1")).Suffix("RETURNING id, field1, field2")

func NewMyObjRepository(db pg.DbClient) MyObjRepository {
	return MyObjRepository{
		pg.NewPostgreRepository[MyObj](
			db,
			sq.Insert("MyObjTable").PlaceholderFormat(sq.Dollar).Columns("field1", "field2"),
			sq.Select("id", "field1", "field2").PlaceholderFormat(sq.Dollar).From("MyObjTable"),
			sq.Update("MyObjTable").PlaceholderFormat(sq.Dollar),
			sq.Delete("MyObjTable").PlaceholderFormat(sq.Dollar),
			myObjConverter),
	}
}

func myObjConverter(row pgx.Row) MyObj {
	var myObj MyObj
	if err := row.Scan(&myObj.ID, &myObj.Field1, &myObj.Field2); err != nil {
		panic(err)
	}
	return myObj
}

func (repo *MyObjRepository) GetOneByField2(ctx context.Context, field2 string) MyObj {
	list := repo.GetBy(ctx, sq.Eq{"field2": field2})
	if list != nil && len(list) > 0 {
		return (list)[0]
	}
	var ret MyObj
	return ret
}

func (repo *MyObjRepository) IncreaseField1(ctx context.Context, id int64) int64 {
	updated := repo.UpdateReturning(ctx, increaseField1Builder.Where(sq.Eq{"id": id}))
	return updated.Field1
}
