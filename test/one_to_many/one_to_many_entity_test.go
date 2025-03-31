package oneToMany_entity_repository_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/simpleGorm/pg"
	"github.com/simpleGorm/pg/internal/closer"
	"github.com/simpleGorm/pg/test/one_to_many"
	"github.com/simpleGorm/pg/test/test_utils"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestOneToManyEntityRepositoryIT(t *testing.T) {

	ctx := context.Background()
	dbName := "users"
	dbUser := "postgres"
	dbPassword := "postgres"
	DSN, err := test_utils.StartPostgresContainer(ctx, dbName, dbUser, dbPassword, t)
	dbClient, err := pg.NewDBClient(ctx, DSN)
	require.NoError(t, err)
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

	parentRepository := oneToMany_entity_repository.NewParentEntityRepository(dbClient)
	child1Repository := oneToMany_entity_repository.NewChild1EntityRepository(dbClient)
	child2Repository := oneToMany_entity_repository.NewChild2EntityRepository(dbClient)

	parentId := parentRepository.Create(ctx, "PARENT")
	child1Repository.Create(ctx, "TYPE1", parentId)
	child1Repository.Create(ctx, "TYPE2", parentId)
	child2Repository.Create(ctx, 0.5, parentId)
	child2Repository.Create(ctx, 0.7, parentId)
	parentEntity := parentRepository.GetById(ctx, parentId)

	var expected oneToMany_entity_repository.ParentEntity
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
			{"ID": 1, "TYPE": "TYPE1"},
			{"ID": 1, "TYPE": "TYPE1"},
			{"ID": 2, "TYPE": "TYPE2"},
			{"ID": 2, "TYPE": "TYPE2"}
		],
		"Children2": [
			{"ID": 1, "SIZE": 0.5},
			{"ID": 2, "SIZE": 0.7},
			{"ID": 1, "SIZE": 0.5},
			{"ID": 2, "SIZE": 0.7}
		]
	}`
