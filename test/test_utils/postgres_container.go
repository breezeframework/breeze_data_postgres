package test_utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"path/filepath"
	"testing"
	"time"
)

func StartPostgresContainer(ctx context.Context, t *testing.T) (string, error) {
	dbName := "users"
	dbUser := "postgres"
	dbPassword := "postgres"
	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithInitScripts(filepath.Join("../test_env", "init-db.sh")),
		postgres.WithConfigFile(filepath.Join("../test_env", "postgres.conf")),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	//logs, _ := postgresContainer.Logs(ctx)
	//_, err = io.Copy(os.Stdout, logs)

	require.NoError(t, err)

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
	return DSN, err
}
