package pproduct

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

func NewRepoFromPostgres(client *rcpostgres.Client) repository.Product {
	return &repoPg{_DB: client}
}

func (r *repoPg) Create(ctx context.Context, product entity.Product) error {
	_, err := r.NewInsert().Model(&product).Exec(ctx)
	return err
}

func (r *repoPg) GetByGUIDs(ctx context.Context, guids []uuid.UUID) ([]entity.Product, error) {
	var products []entity.Product
	err := r.NewSelect().Model(&products).Where("guid IN (?)", bun.List(guids)).Scan(ctx)
	return products, err
}

func (r *repoPg) Update(ctx context.Context, product entity.Product) error {
	res, err := r.NewUpdate().Model(&product).WherePK().ExcludeColumn("id", "created_at").Exec(ctx)
	return rcpostgres.UpdateErr(res, err)
}

func (r *repoPg) Delete(ctx context.Context, guid uuid.UUID) error {
	_, err := r.NewDelete().Model((*entity.Product)(nil)).Where("guid = ?", guid).Exec(ctx)
	return rcpostgres.DeleteErr(err)
}

func (r *repoPg) List(ctx context.Context, name *string, categoryGUID *uuid.UUID) ([]entity.Product, error) {
	var products []entity.Product
	q := r.NewSelect().Model(&products)
	if name != nil {
		q.Where("name = ?", *name)
	}
	if categoryGUID != nil {
		q.Where("category_guid = ?", *categoryGUID)
	}
	err := q.Scan(ctx)
	return products, err
}
