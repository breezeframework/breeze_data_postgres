# Golang module for CRUD database operations management

The library is simplifying write database access layer in Go programs 
by supplying ORM approach - defining Entities and Repositories with minimal required code 
and utilize them during your regular development without needs to write a lot of boilerplate code.

Unlike other ORM libraries and frameworks this library doesn't use reflection and have other performance relevant code optimisation therefore has no performance penalty comparing to all-by-hand written code 

# Examples
 
### Since the library is actively developing, the examples written below might be out of date, but explains the common idea, therefore for actual code please reference to links presented at top of each example.
  

### 1. Create specific entity/repository
Up to date working examples locates [here](internal/test/plain/plain_entity_repository.go) and [here](internal/test/one_to_many/test_repository)
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
Up to date working examples locates [here](internal/test/plain/plain_entity_test.go) and [here](internal/test/one_to_many/one_to_many_entity_test.go)
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
