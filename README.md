# Golang module for CRUD database operations management

## Example

### 1. Create specific repository
Working example locates [here](test/testing_repository.go)
``` go

package repository

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
)

// Mandatory declare ID field
type TestObj struct {
	ID     int64
	Field1 int64
	Field2 string
}

type TestObjRepository struct {
	pg.Repository[TestObj]
}

const TABLE_NAME = "TestObjTable"

var increaseField1Builder = sq.Update(TABLE_NAME).PlaceholderFormat(sq.Dollar).
	Set("field1", sq.Expr("field1 + 1")).Suffix("RETURNING id, field1, field2")

func NewMyObjRepository(db pg.DbClient) TestObjRepository {
	return TestObjRepository{
		pg.NewPostgreRepository[TestObj](
			db,
			sq.Insert(TABLE_NAME).PlaceholderFormat(sq.Dollar).Columns("field1", "field2"),
			sq.Select("id", "field1", "field2").PlaceholderFormat(sq.Dollar).From(TABLE_NAME),
			sq.Update(TABLE_NAME).PlaceholderFormat(sq.Dollar),
			sq.Delete(TABLE_NAME).PlaceholderFormat(sq.Dollar),
			myObjConverter),
	}
}

func myObjConverter(row pgx.Row) TestObj {
	var myObj TestObj
	if err := row.Scan(&myObj.ID, &myObj.Field1, &myObj.Field2); err != nil {
		panic(err)
	}
	return myObj
}

func (repo *TestObjRepository) GetOneByField2(ctx context.Context, field2 string) TestObj {
	list := repo.GetBy(ctx, sq.Eq{"field2": field2})
	if list != nil && len(list) > 0 {
		return (list)[0]
	}
	var ret TestObj
	return ret
}

func (repo *TestObjRepository) IncreaseField1(ctx context.Context, id int64) int64 {
	updated := repo.UpdateReturning(ctx, increaseField1Builder.Where(sq.Eq{"id": id}))
	return updated.Field1
}

```
### 2. Usage
Working test locates [here](test/repository_test.go)
``` go
    DSN := "connection string"
    ctx := context.Background()
    dbClient, err := pg.NewPgDBClient(ctx, DSN)
    
    // Define dbclient gracefull shutdown
    closer.Add(dbClient.Close)
    defer func() {
        if r := recover(); r != nil {
            // handle panic errors
        }
        // Close db connection
        closer.CloseAll()
        closer.Wait()
    }()
	
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
    .....
    
   fields := map[string]interface{}{"field2": "updated_field2"}

   updated := myRepository.Update(ctx, fields, id)
   if updated != 1 {
      return errors.New("MyObject not updated")
   }
   
   where := squirrel.Eq{"field2": field2}
   updated = myRepository.UpdateCollection(ctx, fields, where)
   if updated == 0 {
      return errors.New("No objects was updated")
   }
    
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
   if err != nil {
     ...  // handle the error
   }
```
