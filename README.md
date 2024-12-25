# Golang module for CRUD database operations management

## Example

### 1. Create specific repository
``` go

package repository

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/breezeframework/breeze_data/breeze_data"
	"github.com/breezeframework/breeze_data/breeze_data/pg"
	"github.com/jackc/pgx/v5"
)

// Mandatory declare ID field
type MyObj struct {
	ID     int64
	Field1 int64
	Field2 string
}

type MyRepository interface {
	IncreaseField1(ctx context.Context, id int64) int64
	GetByField2(ctx context.Context, url string) MyObj
	breeze_data.CrudRepository[MyObj]
}

type myRepository struct {
	breeze_data.CrudRepository[MyObj]
}

func (repo *myRepository) GetByField2(ctx context.Context, field2 string) MyObj {
	list := repo.GetBy(ctx, sq.Eq{"field2": field2})
	if list != nil && len(*list) > 0 {
		return (*list)[0]
	}
	var ret MyObj
	return ret
}

type myBuilders struct {
	InsertBuilder         sq.InsertBuilder
	SelectBuilder         sq.SelectBuilder
	UpdateBuilder         sq.UpdateBuilder
	DeleteBuilder         sq.DeleteBuilder
	IncreaseField1Builder sq.UpdateBuilder
}

var MyBuilders = myBuilders{
	InsertBuilder: sq.Insert("MyObjTable").PlaceholderFormat(sq.Dollar).Columns("field1", "field2"),
	SelectBuilder: sq.Select("field1", "field2").PlaceholderFormat(sq.Dollar).From("MyObjTable"),
	UpdateBuilder: sq.Update("MyObjTable").PlaceholderFormat(sq.Dollar),
	DeleteBuilder: sq.Delete("MyObjTable").PlaceholderFormat(sq.Dollar),
	IncreaseField1Builder: sq.Update("MyObjTable").PlaceholderFormat(sq.Dollar).
		Set("field1", sq.Expr("field1 + 1")).Suffix("RETURNING id, field1, field2"),
}

func NewMyRepository(db breeze_data.DbClient) MyRepository {
	return &myRepository{
		breeze_data.NewPostgreSQLCRUDRepository[MyObj](
			db,
			MyBuilders.InsertBuilder,
			MyBuilders.SelectBuilder,
			MyBuilders.UpdateBuilder,
			MyBuilders.DeleteBuilder,
			MyObjConverter),
	}
}

func (repo *myRepository) IncreaseField1(ctx context.Context, id int64) int64 {
	ret := repo.UpdateReturning(ctx, MyBuilders.IncreaseField1Builder.
		Where(sq.Eq{"id": id}), MyObjConverter)
	return ret.Field1
}

func MyObjConverter(row pgx.Row) MyObj {
	var myObj MyObj
	if err := row.Scan(&myObj.ID, &myObj.Field1, &myObj.Field2); err != nil {
		panic(err)
	}
	return myObj
}

```
### 2. Usage
``` go
    DSN := "connection string"
    ctx := context.Background()
    dbClient, err := pg.NewPgDBClient(ctx, DSN)
    myRepository := NewMyRepository(dbClient)
    field1_value := 10
    field2_value := "field2_value"
    id := myRepository.Create(ctx, field1_value, field2_value)
    ....
    myRepository.GetById(ctx,id)
    ....
    myRepository.GetByField2(ctx, field2_value)
    ....
    myRepository.GetAll(ctx)
    ....
    newField1Value := myRepository.IncreaseField1(ctx, id)
    print(newField1Value) // 11
    
    // Transaction
    err := dbClient.RunTransaction(ctx, transaction.TxOptions{IsoLevel: transaction.ReadCommitted},
		func(ctx context.Context) error {
			id1 := myRepository.Create(ctx, ...)
			id2 := myRepository.Create(ctx, ...)
			if(...some error condition...){
			    // Rollback
			    return errors.Wrap(err, "Error condition achieved")
			}
			// Commit
			return nil
		}
	)
    
```
