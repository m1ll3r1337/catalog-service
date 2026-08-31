package sproduct

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/m1ll3r1337/catalog-service/internal/app/entity"
	"github.com/m1ll3r1337/catalog-service/internal/app/repository"
	"github.com/m1ll3r1337/catalog-service/internal/app/service"
)

type srv struct {
	repoProduct  repository.Product
	repoCategory repository.Category
}

func NewService(repoProduct repository.Product, repoCategory repository.Category) service.Product {
	return &srv{repoProduct: repoProduct, repoCategory: repoCategory}
}

func (s *srv) Create(ctx context.Context, req entity.RequestProductCreate) (entity.Product, error) {
	existing, err := s.repoProduct.List(ctx, &req.Name, nil, nil, nil)
	if err != nil {
		return entity.Product{}, err
	}
	if len(existing) > 0 {
		return entity.Product{}, entity.ErrAlreadyExists
	}

	categories, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{req.CategoryGUID})
	if err != nil {
		return entity.Product{}, err
	}
	if len(categories) == 0 {
		return entity.Product{}, entity.ErrNotFound
	}

	now := time.Now()
	product := entity.Product{
		GUID:         uuid.Must(uuid.NewV4()),
		Name:         req.Name,
		Description:  req.Description,
		CategoryGUID: req.CategoryGUID,
		Price:        req.Price,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repoProduct.Create(ctx, product); err != nil {
		return entity.Product{}, err
	}

	return product, nil
}

func (s *srv) GetByGUIDs(ctx context.Context, guids []uuid.UUID) ([]entity.Product, error) {
	products, err := s.repoProduct.GetByGUIDs(ctx, guids)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *srv) Update(ctx context.Context, guid uuid.UUID, req entity.RequestProductUpdate) (entity.Product, error) {
	products, err := s.repoProduct.GetByGUIDs(ctx, []uuid.UUID{guid})
	if err != nil {
		return entity.Product{}, err
	}
	if len(products) == 0 {
		return entity.Product{}, entity.ErrNotFound
	}

	product := products[0]

	if req.Name != nil {
		existing, err := s.repoProduct.List(ctx, req.Name, nil, nil, nil)
		if err != nil {
			return entity.Product{}, err
		}
		for _, p := range existing {
			if p.GUID != guid {
				return entity.Product{}, entity.ErrAlreadyExists
			}
		}
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.CategoryGUID != nil {
		categories, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{*req.CategoryGUID})
		if err != nil {
			return entity.Product{}, err
		}
		if len(categories) == 0 {
			return entity.Product{}, entity.ErrNotFound
		}
		product.CategoryGUID = *req.CategoryGUID
	}
	if req.Price != nil {
		product.Price = *req.Price
	}

	product.UpdatedAt = time.Now()

	if err := s.repoProduct.Update(ctx, product); err != nil {
		return entity.Product{}, err
	}

	return product, nil
}

func (s *srv) Delete(ctx context.Context, guid uuid.UUID) error {
	products, err := s.repoProduct.GetByGUIDs(ctx, []uuid.UUID{guid})
	if err != nil {
		return err
	}
	if len(products) == 0 {
		return entity.ErrNotFound
	}

	if err := s.repoProduct.Delete(ctx, guid); err != nil {
		return err
	}

	return nil
}

func (s *srv) List(ctx context.Context, req entity.RequestProductList) ([]entity.Product, error) {
	return s.repoProduct.List(ctx, nil, req.CategoryGUID, req.MinPrice, req.MaxPrice)
}
