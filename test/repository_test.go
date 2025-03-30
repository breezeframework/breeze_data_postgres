package repository_test

import (
	"context"
	"github.com/jackc/pgx/v5"
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

	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
		postgres.WithConfigFile(filepath.Join("testdata", "my-postgres.conf")),
		postgres.WithDatabase("users"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
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

	postgresURI, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	t.Logf("Postgres URI: %s", postgresURI)

	conn, err := pgx.Connect(ctx, postgresURI)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	type MyObj struct {
		ID     int64
		Field1 int64
		Field2 string
	}

	var myObj MyObj

	query := "SELECT FROM MyObjTable (ID, Field1, Field2)"
	row := conn.QueryRow(ctx, query)
	err = row.Scan(&myObj.ID, &myObj.Field1, &myObj.Field2)
	require.NoError(t, err)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

}
