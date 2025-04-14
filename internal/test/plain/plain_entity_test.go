package plain_test

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/simpleGorm/pg/internal/test/plain"
	"github.com/simpleGorm/pg/internal/test/test_utils"
	"github.com/simpleGorm/pg/internal/transaction"
	"github.com/simpleGorm/pg/pkg"
	"github.com/simpleGorm/pg/pkg/closer"
	"github.com/simpleGorm/pg/pkg/logger"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"runtime/debug"
	"testing"
)

func TestPlain(t *testing.T) {

	logger.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError})))

	ctx := context.Background()
	DSN, err := test_utils.StartPostgresContainer(ctx, t)

	dbClient, err := pkg.NewDBClient(ctx, DSN)
	require.NoError(t, err)
	closer.Add(dbClient.Close)
	// Define dbclient gracefull shutdown
	defer func() {
		if r := recover(); r != nil {
			// handle panic errors
			logger.Logger().Error("panic: %v", r)
			t.Fatalf("panic: %v", string(debug.Stack()))
			//closer.CloseAll()
			//closer.Wait()
		} else {
			closer.CloseAll()
		}
		// Close db connection
	}()

	myRepository := plain.NewTestPlainEntityRepository(dbClient)
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
			logger.Logger().Info("", slog.Int64("id1", id1))
			logger.Logger().Info("", slog.Int64("id2", id2))
			if id1 != id2 {
				// Rollback
				return errors.Wrap(errors.New("Test error"), "Error condition achieved")
			}
			// Commit
			return nil
		},
	)
	require.Error(t, err, "Transaction is not rolled back")
}
