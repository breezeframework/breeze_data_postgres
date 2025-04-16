package oneToMany_entity_repository_test

import (
	"context"
	"encoding/json"
	"github.com/Masterminds/squirrel"
	"github.com/simpleGorm/pg"
	"github.com/simpleGorm/pg/internal/closer"
	"github.com/simpleGorm/pg/internal/logger"
	"github.com/simpleGorm/pg/internal/test/one_to_many/test_repository"
	"github.com/simpleGorm/pg/internal/test/test_utils"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"runtime/debug"
	"testing"
)

func TestOneToMany(t *testing.T) {
	logger.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	})))
	ctx := context.Background()
	DSN, err := test_utils.StartPostgresContainer(ctx, t)
	dbClient, err := pg.NewDBClient(ctx, DSN)
	require.NoError(t, err)
	dbClient.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})))
	closer.Add(dbClient.Close)
	// Define dbclient gracefull shutdown
	defer func() {
		if r := recover(); r != nil {
			// handle panic errors
			t.Logf("panic: %v", r)
			closer.CloseAll()
			closer.Wait()
			t.Fatalf("panic: %v", string(debug.Stack()))
		} else {
			closer.CloseAll()
		}
		// Close db connection
	}()

	parentRepository := test_repository.NewParentEntityRepository(dbClient)
	child1Repository := test_repository.NewChild1EntityRepository(dbClient)
	child2Repository := test_repository.NewChild2EntityRepository(dbClient)

	parentId := parentRepository.Create(ctx, "PARENT")
	child1Repository.Create(ctx, "TYPE1", parentId)
	child1Repository.Create(ctx, "TYPE2", parentId)
	child2Repository.Create(ctx, 0.5, parentId)
	child2Repository.Create(ctx, 0.7, parentId)

	parentEntity := parentRepository.GetById(ctx, parentId)
	actual := marshallActual(t, err, parentEntity)
	if EXPECTED_ONE != actual {
		t.Errorf("\nExpected:\n%v\nGot:\n%v", EXPECTED_ONE, actual)
	}

	parentEntities := parentRepository.GetAll(ctx)
	actual = marshallActual(t, err, parentEntities)
	if EXPECTED_MANY != actual {
		t.Errorf("\nExpected:\n%v\nGot:\n%v", EXPECTED_MANY, actual)
	}

	parentEntities = parentRepository.GetBy(ctx, squirrel.NotEq{"id": 0})
	actual = marshallActual(t, err, parentEntities)
	if EXPECTED_MANY != actual {
		t.Errorf("\nExpected:\n%v\nGot:\n%v", EXPECTED_MANY, actual)
	}

}

func marshallActual(t *testing.T, err error, obj any) string {
	marshalled, err := json.Marshal(&obj)
	require.NoError(t, err)
	actual := string(marshalled)
	logger.Logger().Info(actual)
	return actual
}

var EXPECTED_ONE = `{"ID":1,"Name":"PARENT","Children1":[{"ID":1,"type":"TYPE1","PARENT_ID":1},{"ID":2,"type":"TYPE2","PARENT_ID":1}],"Children2":[{"ID":1,"size":0.5,"PARENT_ID":1},{"ID":2,"size":0.7,"PARENT_ID":1}]}`

var EXPECTED_MANY = "[" + EXPECTED_ONE + "]"
