package testrepository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/simpleGorm/pg"
	"github.com/simpleGorm/pg/internal/closer"
	"github.com/simpleGorm/pg/internal/transaction"
	"github.com/simpleGorm/pg/test"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"path/filepath"
)

func TestMyRepositoryIntegration(t *testing.T) {

	ctx := context.Background()
	dbName := "users"
	dbUser := "postgres"
	dbPassword := "postgres"

	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
		postgres.WithConfigFile(filepath.Join("testdata", "my-postgres.conf")),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)

	logs, _ := postgresContainer.Logs(ctx)
	io.Copy(os.Stdout, logs)

	require.NoError(t, err)
	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	conf, err := postgresContainer.Inspect(ctx)
	require.NoError(t, err)

	t.Logf("Postgres container id: %s", conf.ID)

	var prettyJSON bytes.Buffer
	jsonBytes, err := json.Marshal(*conf.Config)
	require.NoError(t, err)

	err = json.Indent(&prettyJSON, jsonBytes, "", "  ")
	require.NoError(t, err)
	//t.Logf("Postgres Config: %s", prettyJSON.String())

	jsonBytes, err = json.Marshal(*conf.NetworkSettings)
	require.NoError(t, err)
	err = json.Indent(&prettyJSON, jsonBytes, "", "  ")
	require.NoError(t, err)
	//t.Logf("Postgres NetworkSettings: %+v", prettyJSON.String())

	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)
	t.Logf("Postgres host: %s", host)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)
	t.Logf("Postgres port: %s", port.Port())

	DSN := fmt.Sprintf("host=%s port=%s dbname=%s user=%s  password=%s sslmode=disable", "127.0.0.1", port.Port(), dbName, dbUser, dbPassword)
	t.Logf("Postgres DSN: %s", DSN)

	dbClient, err := pg.NewDBClient(ctx, DSN)

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

	myRepository := testrepository.NewMyObjRepository(dbClient)
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
