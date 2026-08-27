package pcategory

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/m1ll3r1337/catalog-service/internal/app/entity"
	"github.com/m1ll3r1337/catalog-service/internal/app/repository"
	rcpostgres "github.com/m1ll3r1337/catalog-service/internal/app/repository/conn/postgres"
	"github.com/uptrace/bun"
)

type (
	repoPg struct {
		*_DB
	}

	_DB = rcpostgres.Client
)

func NewRepoFromPostgres(client *rcpostgres.Client) repository.Category {
	return &repoPg{_DB: client}
}

func (r *repoPg) Create(ctx context.Context, category entity.Category) error {
	_, err := r.NewInsert().Model(&category).Exec(ctx)
	return err
}

func (r *repoPg) GetByGUIDs(ctx context.Context, guids []uuid.UUID) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.NewSelect().Model(&categories).Where("guid IN (?)", bun.List(guids)).Scan(ctx)
	return categories, err
}

func (r *repoPg) Update(ctx context.Context, category entity.Category) error {
	res, err := r.NewUpdate().Model(&category).WherePK().ExcludeColumn("id", "created_at").Exec(ctx)
	return rcpostgres.UpdateErr(res, err)
}

func (r *repoPg) Delete(ctx context.Context, guid uuid.UUID) error {
	_, err := r.NewDelete().Model((*entity.Category)(nil)).Where("guid = ?", guid).Exec(ctx)
	return rcpostgres.DeleteErr(err)
}

func (r *repoPg) List(ctx context.Context, name *string) ([]entity.Category, error) {
	var categories []entity.Category
	q := r.NewSelect().Model(&categories)
	if name != nil {
		q.Where("name = ?", *name)
	}
	err := q.Scan(ctx)
	return categories, err
}
