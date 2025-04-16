package pg

import (
	"context"
	"github.com/simpleGorm/pg/internal/logger"
	"github.com/simpleGorm/pg/internal/pg_api"
	"log/slog"
)

type DbClient struct {
	DbApi
}

func NewDBClient(ctx context.Context, dsn string) (DbClient, error) {
	client, err := pg_api.NewPgDBClient(ctx, dsn)
	if err != nil {
		return DbClient{}, err
	}
	return DbClient{client}, nil
}

type DbApi interface {
	pg_api.SQLExecutor
	pg_api.Transactor
	Pinger
	Closer
}

type Closer interface {
	Close() error
}

type Pinger interface {
	Ping(ctx context.Context) error
}

func (db DbClient) SetLogger(l *slog.Logger) {
	logger.SetLogger(l)
}

func (db DbClient) Logger() *slog.Logger {
	return logger.Logger()
}
