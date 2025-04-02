package one_to_many

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/simpleGorm/pg"
	"github.com/simpleGorm/pg/internal/closer"
	"github.com/simpleGorm/pg/test/one_to_many/test_repository"
	"github.com/simpleGorm/pg/test/test_utils"
	"github.com/stretchr/testify/require"
	"reflect"
	"runtime/debug"
	"testing"
)

func TestOneToManyEntityRepositoryIT(t *testing.T) {

	ctx := context.Background()
	DSN, err := test_utils.StartPostgresContainer(ctx, t)
	dbClient, err := pg.NewDBClient(ctx, DSN)
	require.NoError(t, err)
	// Define dbclient gracefull shutdown
	closer.Add(dbClient.Close)
	defer func() {
		if r := recover(); r != nil {
			// handle panic errors
			closer.CloseAll()
			closer.Wait()
			t.Logf("panic: %v", r)
			t.Fatalf("panic: %v", string(debug.Stack()))
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

	var expected test_repository.ParentEntity
	err = json.Unmarshal([]byte(EXPECTED), &expected)
	require.NoError(t, err)
	if !reflect.DeepEqual(expected, parentEntity) {
		t.Errorf("Expected: %v\nGot: %v", expected, parentEntity)
	}
	jsonData, err := json.MarshalIndent(parentEntity, "", "  ")
	require.NoError(t, err)
	fmt.Println(string(jsonData))
}

var EXPECTED = `{
		"ID": 1,
		"Name": "PARENT",
		"Children1": [
			{"ID": 1, "TYPE": "TYPE1", "PARENT_ID": 1},
			{"ID": 2, "TYPE": "TYPE2", "PARENT_ID": 1}
		],
		"Children2": [
			{"ID": 1, "SIZE": 0.5, "PARENT_ID": 1},
			{"ID": 2, "SIZE": 0.7, "PARENT_ID": 1}
		]
	}`
