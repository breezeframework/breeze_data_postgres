package one_to_many

import (
	"fmt"
	"github.com/simpleGorm/pg"
	"github.com/simpleGorm/pg/test/one_to_many/test_repository"
	"testing"
)

type IRelation[T any, R any] interface {
	GetRepo() pg.Repository[R]
	GetForeignKey() string
	//AddToParent(parent T, related R)
	GetParentId(related R) int64
}

type RelationImpl[T any, R any] struct {
	repo       pg.Repository[R]
	foreignKey string
}

func (r *RelationImpl[T, R]) GetRepo() pg.Repository[R] {
	return r.repo
}

func (r *RelationImpl[T, R]) GetForeignKey() string {
	return r.foreignKey
}

func (r *RelationImpl[T, R]) AddToParent(parent T, child R) {
	fmt.Println("AddToParent called")
}

func (r *RelationImpl[T, R]) GetParentId(child R) int64 {
	fmt.Println("GetParentId called")
	return 0
}

func NewRelationImpl[T any, R any](repo pg.Repository[R], foreignKey string) IRelation[T, R] {
	return &RelationImpl[T, R]{
		repo:       repo,
		foreignKey: foreignKey,
	}
}

func TestCasting(t *testing.T) {
	repo := pg.Repository[test_repository.ParentEntity]{}
	arr := NewRelationImpl[test_repository.ParentEntity, test_repository.ParentEntity](repo, "")
	cast := any(arr).(pg.IRelation[test_repository.ParentEntity, any])
	t.Logf("cast %+v", cast)
}
