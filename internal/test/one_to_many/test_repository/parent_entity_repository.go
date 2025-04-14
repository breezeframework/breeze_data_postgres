package test_repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/simpleGorm/pg"
	"github.com/simpleGorm/pg/pkg/repository"
)

const TABLE_NAME = "TEST_PARENT_ENTITY_TABLE "

var (
	ParentEntity_id   = "id"
	ParentEntity_name = "name"
)

var ParentEntity_Fields = []string{
	ParentEntity_id,
	ParentEntity_name,
}

type ParentEntity struct {
	ID        int64  `json:"ID"` // ID field is mandatory
	Name      string `json:"Name"`
	Children1 []any  `json:"Children1"`
	Children2 []any  `json:"Children2"`
}

type ParentEntityRepository struct {
	repository.Repository[ParentEntity]
}

func (parent ParentEntity) GetID() int64 {
	return parent.ID
}

func (parent *ParentEntity) AddRelatedEntity(related any) {
	parent.Children1 = append(parent.Children1, related)
}

func NewParentEntityRepository(db pg.DbClient) ParentEntityRepository {
	repo := repository.NewRepository(
		ParentEntity{},
		db,
		sq.Insert(TABLE_NAME).PlaceholderFormat(sq.Dollar).Columns(ParentEntity_name),
		sq.Select(ParentEntity_Fields...).PlaceholderFormat(sq.Dollar).From(TABLE_NAME),
		sq.UpdateBuilder{},
		sq.DeleteBuilder{},
		parentEntityConverter)
	child1Rel := repository.WrapRelation(OneToManyChild1EntityRelation(db))
	child2Rel := repository.WrapRelation(OneToManyChild2EntityRelation(db))
	repo.Relations = append(repo.Relations, child1Rel, child2Rel)
	return ParentEntityRepository{repo}
}

func parentEntityConverter(rows pgx.Row) any {
	var parent ParentEntity
	if err := rows.Scan(&parent.ID, &parent.Name); err != nil {
		panic(err)
	}
	return &parent
}
