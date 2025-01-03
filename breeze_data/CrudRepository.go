package breeze_data

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type CrudRepository[T any] interface {
	Create(ctx context.Context, values ...interface{}) int64
	GetById(ctx context.Context, id int64) T
	GetAll(ctx context.Context) []T
	GetBy(ctx context.Context, where sq.Sqlizer) []T
	Delete(ctx context.Context, id int64) int64
	Update(ctx context.Context, fields map[string]interface{}, where string, args ...interface{}) int64
	UpdateReturning(ctx context.Context, builder sq.UpdateBuilder, entityConverter func(row pgx.Row) T) T
}
