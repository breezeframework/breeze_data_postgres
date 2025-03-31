package plain_entity_test

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/simpleGorm/pg"
	"github.com/simpleGorm/pg/internal/closer"
	"github.com/simpleGorm/pg/internal/transaction"
	"github.com/simpleGorm/pg/test/plain_entity"
	"github.com/simpleGorm/pg/test/test_utils"
	"testing"
)

func TestPlainEntityRepositoryIT(t *testing.T) {

	ctx := context.Background()
	dbName := "users"
	dbUser := "postgres"
	dbPassword := "postgres"

	DSN, err := test_utils.StartPostgresContainer(ctx, dbName, dbUser, dbPassword, t)

	dbClient, err := pg.NewDBClient(ctx, DSN)

	if err != nil {
		t.Fatalf("failed to connect to Postgres container: %v", err)
	}
	// Define dbclient gracefull shutdown
	closer.Add(dbClient.Close)
	defer func() {
		if r := recover(); r != nil {
			// handle panic errors
			closer.CloseAll()
			closer.Wait()
			t.Fatalf("panic: %v", r)
		}
		// Close db connection
	}()

	myRepository := plain_entity.NewTestPlainEntityRepository(dbClient)
	field1_value := 10
	field2_value := "field2_value"
	id := myRepository.Create(ctx, field1_value, field2_value)
	myRepository.GetById(ctx, id)

	myRepository.GetOneByField2(ctx, field2_value)

	myRepository.GetAll(ctx)

	newField1Value := myRepository.IncreaseField1(ctx, id)
	print(newField1Value) // 11

	fields := map[string]interface{}{"field2": "updated_field2"}

	updatedRowsCount := myRepository.Update(ctx, fields, id)
	if updatedRowsCount != 1 {
		t.Fatalf("MyObject not updated")
	}

	where := squirrel.NotEq{"field2": "---"}
	updatedRowsCount = myRepository.UpdateCollection(ctx, fields, where)
	if updatedRowsCount == 0 {
		t.Fatalf("No objects was updated")
	}

	// Transaction
	err = dbClient.RunTransaction(ctx, transaction.TxOptions{IsoLevel: transaction.ReadCommitted},
		func(ctx context.Context) error {
			id1 := myRepository.Create(ctx, 2, "field2_value_2")
			id2 := myRepository.Create(ctx, 3, "field2_value_3")
			t.Logf("id1: %d", id1)
			t.Logf("id2: %d", id2)
			if id1 != id2 {
				// Rollback
				return errors.Wrap(errors.New("Test error"), "Error condition achieved")
			}
			// Commit
			return nil
		},
	)
	if err == nil {
		t.Fatalf("Transaction is not rolled back")
	}

}
