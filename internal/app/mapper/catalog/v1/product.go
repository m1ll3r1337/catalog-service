package mcatv1

import (
	"github.com/m1ll3r1337/catalog-service/internal/app/entity"
	catalogv1 "github.com/m1ll3r1337/catalog-service/internal/pkg/grpc/gen/catalog/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ProductToProto(p entity.Product) *catalogv1.Product {
	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}

	return &catalogv1.Product{
		Guid:         p.GUID.String(),
		Name:         p.Name,
		Description:  desc,
		Price:        p.Price,
		CategoryGuid: p.CategoryGUID.String(),
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
	}
}
