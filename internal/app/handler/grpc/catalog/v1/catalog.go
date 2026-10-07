package ghcatalogv1

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/m1ll3r1337/catalog-service/internal/app/entity"
	"github.com/m1ll3r1337/catalog-service/internal/app/mapper"
	mcatv1 "github.com/m1ll3r1337/catalog-service/internal/app/mapper/catalog/v1"
	"github.com/m1ll3r1337/catalog-service/internal/app/service"
	catalogv1 "github.com/m1ll3r1337/catalog-service/internal/pkg/grpc/gen/catalog/v1"
)

type handler struct {
	catalogv1.UnimplementedCatalogServiceServer
	srv service.Product
}

func NewHandler(srv service.Product) catalogv1.CatalogServiceServer {
	return &handler{srv: srv}
}

func (h *handler) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	raw := req.GetGuid()
	guid, err := uuid.FromString(raw)
	if err != nil {
		return nil, mapper.ErrorToGRPC(entity.ErrIncorrectParameters)
	}

	products, err := h.srv.GetByGUIDs(ctx, []uuid.UUID{guid})
	if err != nil {
		return nil, mapper.ErrorToGRPC(err)
	}

	if len(products) == 0 {
		return nil, mapper.ErrorToGRPC(entity.ErrNotFound)
	}

	return &catalogv1.GetProductResponse{
		Product: mcatv1.ProductToProto(products[0]),
	}, nil
}

func (h *handler) GetProducts(ctx context.Context, req *catalogv1.GetProductsRequest) (*catalogv1.GetProductsResponse, error) {
	guidsRaw := req.GetGuids()
	if len(guidsRaw) == 0 {
		return &catalogv1.GetProductsResponse{}, nil
	}

	guids := make([]uuid.UUID, 0, len(guidsRaw))
	for _, g := range guidsRaw {
		guid, err := uuid.FromString(g)
		if err != nil {
			return nil, mapper.ErrorToGRPC(entity.ErrIncorrectParameters)
		}

		guids = append(guids, guid)
	}

	products, err := h.srv.GetByGUIDs(ctx, guids)
	if err != nil {
		return nil, mapper.ErrorToGRPC(err)
	}

	found := make(map[uuid.UUID]struct{}, len(products))
	results := make([]*catalogv1.Product, 0, len(products))
	for _, p := range products {
		found[p.GUID] = struct{}{}

		results = append(results, mcatv1.ProductToProto(p))
	}

	missing := make([]string, 0, len(guids))
	for _, g := range guids {
		if _, ok := found[g]; !ok {
			missing = append(missing, g.String())
		}
	}

	return &catalogv1.GetProductsResponse{
		Products:     results,
		MissingGuids: missing,
	}, nil
}
