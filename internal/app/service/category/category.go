package scategory

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/m1ll3r1337/catalog-service/internal/app/entity"
	"github.com/m1ll3r1337/catalog-service/internal/app/repository"
	"github.com/m1ll3r1337/catalog-service/internal/app/service"
)

type srv struct {
	repoCategory repository.Category
	repoProduct  repository.Product
}

func NewService(repoCategory repository.Category, repoProduct repository.Product) service.Category {
	return &srv{
		repoCategory: repoCategory,
		repoProduct:  repoProduct,
	}
}

func (s *srv) Create(ctx context.Context, req entity.RequestCategoryCreate) (entity.Category, error) {
	var category entity.Category

	err := s.repoCategory.InsideTx(ctx, func(ctx context.Context) error {
		existing, err := s.repoCategory.List(ctx, &req.Name)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return entity.ErrAlreadyExists
		}

		now := time.Now()
		category = entity.Category{
			GUID:      uuid.Must(uuid.NewV4()),
			Name:      req.Name,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := s.repoCategory.Create(ctx, category); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return entity.Category{}, err
	}

	return category, nil
}

func (s *srv) GetByGUIDs(ctx context.Context, guids []uuid.UUID) ([]entity.Category, error) {
	categories, err := s.repoCategory.GetByGUIDs(ctx, guids)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *srv) Update(ctx context.Context, guid uuid.UUID, req entity.RequestCategoryUpdate) (entity.Category, error) {
	var category entity.Category

	err := s.repoCategory.InsideTx(ctx, func(ctx context.Context) error {
		categories, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{guid})
		if err != nil {
			return err
		}
		if len(categories) == 0 {
			return entity.ErrNotFound
		}

		list, err := s.repoCategory.List(ctx, &req.Name)
		if err != nil {
			return err
		}
		if len(list) > 0 && list[0].GUID != guid {
			return entity.ErrAlreadyExists
		}

		categories[0].Name = req.Name
		categories[0].UpdatedAt = time.Now()
		category = categories[0]

		if err := s.repoCategory.Update(ctx, category); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return entity.Category{}, err
	}

	return category, nil
}

func (s *srv) Delete(ctx context.Context, guid uuid.UUID) error {
	err := s.repoCategory.InsideTx(ctx, func(ctx context.Context) error {
		cats, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{guid})
		if err != nil {
			return err
		}
		if len(cats) == 0 {
			return entity.ErrNotFound
		}

		products, err := s.repoProduct.List(ctx, nil, &guid, nil, nil)
		if err != nil {
			return err
		}
		if len(products) > 0 {
			return entity.ErrCategoryHasProducts
		}

		return s.repoCategory.Delete(ctx, guid)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *srv) List(ctx context.Context) ([]entity.Category, error) {
	categories, err := s.repoCategory.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	return categories, nil
}
