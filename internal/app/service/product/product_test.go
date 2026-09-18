package sproduct

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/m1ll3r1337/catalog-service/internal/app/entity"
	"github.com/m1ll3r1337/catalog-service/internal/app/repository/mocks"
	"github.com/m1ll3r1337/catalog-service/internal/pkg/testutil"
)

type createProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *createProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestCreateProductSuite(t *testing.T) {
	suite.Run(t, new(createProductSuite))
}

func (s *createProductSuite) TestCreate() {
	type args struct {
		req entity.RequestProductCreate
	}
	type want struct {
		err error
	}

	categoryGUID := uuid.Must(uuid.NewV4())

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Test Product",
					Description:  testutil.PtrString("A test product"),
					Price:        1000,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: nil},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.req.CategoryGUID}).
					Return([]entity.Category{{GUID: categoryGUID}}, nil).
					Once()

				s.productRepo.EXPECT().
					Create(s.ctx, mock.MatchedBy(func(p entity.Product) bool {
						return p.Name == args.req.Name &&
							p.Description == args.req.Description &&
							p.Price == args.req.Price &&
							p.CategoryGUID == args.req.CategoryGUID
					})).
					Return(nil).
					Once()
			},
		},
		{
			name: "already exists",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Existing Product",
					Price:        500,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: entity.ErrAlreadyExists},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{{Name: "Existing Product"}}, nil).
					Once()
			},
		},
		{
			name: "category not found",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "New Product",
					Price:        1000,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.req.CategoryGUID}).
					Return([]entity.Category{}, nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			result, err := s.srv.Create(s.ctx, tc.args.req)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(result.GUID)
			} else {
				s.NoError(err)
				s.NotEmpty(result.GUID)
				s.Equal(tc.args.req.Name, result.Name)
				s.Equal(tc.args.req.Description, result.Description)
				s.Equal(tc.args.req.Price, result.Price)
				s.Equal(tc.args.req.CategoryGUID, result.CategoryGUID)
			}
		})
	}
}

type getByGUIDsProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *getByGUIDsProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestGetByGUIDsProductSuite(t *testing.T) {
	suite.Run(t, new(getByGUIDsProductSuite))
}

func (s *getByGUIDsProductSuite) TestGetByGUIDs() {
	type args struct {
		guids []uuid.UUID
	}
	type want struct {
		err      error
		products []entity.Product
	}

	guids := []uuid.UUID{
		uuid.Must(uuid.NewV4()),
	}

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "single product found",
			args: args{
				guids: guids,
			},
			want: want{
				err:      nil,
				products: []entity.Product{{GUID: guids[0]}},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, args.guids).
					Return([]entity.Product{{GUID: guids[0]}}, nil).
					Once()
			},
		},
		{
			name: "not found returns empty slice",
			args: args{
				guids: guids,
			},
			want: want{
				err:      nil,
				products: []entity.Product{},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, args.guids).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			result, err := s.srv.GetByGUIDs(s.ctx, tc.args.guids)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(result)
			} else {
				s.NoError(err)
				s.Equal(tc.want.products, result)
			}
		})
	}
}

type deleteProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *deleteProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestDeleteProductSuite(t *testing.T) {
	suite.Run(t, new(deleteProductSuite))
}

func (s *deleteProductSuite) TestDelete() {
	type args struct {
		guid uuid.UUID
	}
	type want struct {
		err error
	}

	guid := uuid.Must(uuid.NewV4())

	var errDelete error = errors.New("delete error")

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				guid: guid,
			},
			want: want{err: nil},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{guid}).
					Return([]entity.Product{{GUID: guid}}, nil).
					Once()

				s.productRepo.EXPECT().
					Delete(s.ctx, args.guid).
					Return(nil).
					Once()
			},
		},
		{
			name: "not found",
			args: args{guid: guid},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{guid}).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
		{
			name: "delete error",
			args: args{guid: guid},
			want: want{err: errDelete},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{guid}).
					Return([]entity.Product{{GUID: guid}}, nil).
					Once()

				s.productRepo.EXPECT().
					Delete(s.ctx, args.guid).
					Return(errDelete).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			err := s.srv.Delete(s.ctx, tc.args.guid)
			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
			} else {
				s.NoError(err)
			}
		})
	}
}

type listProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *listProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestListProductSuite(t *testing.T) {
	suite.Run(t, new(listProductSuite))
}

func (s *listProductSuite) TestList() {
	type args struct {
		req entity.RequestProductList
	}
	type want struct {
		err   error
		count int
		name  string
		price int64
	}

	categoryGUID := uuid.Must(uuid.NewV4())

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				req: entity.RequestProductList{
					CategoryGUID: testutil.PtrUUID(categoryGUID),
					MinPrice:     testutil.PtrInt64(100),
					MaxPrice:     testutil.PtrInt64(1000),
				},
			},
			want: want{
				err:   nil,
				count: 2,
				name:  "Test Product",
				price: 500,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(s.ctx, (*string)(nil), args.req.CategoryGUID, args.req.MinPrice, args.req.MaxPrice).
					Return([]entity.Product{
						{
							GUID:         uuid.Must(uuid.NewV4()),
							Name:         "Test Product",
							Price:        500,
							CategoryGUID: categoryGUID,
						},
						{
							GUID:         uuid.Must(uuid.NewV4()),
							Name:         "Another Product",
							Price:        900,
							CategoryGUID: categoryGUID,
						},
					}, nil).
					Once()
			},
		},
		{
			name: "empty result",
			args: args{
				req: entity.RequestProductList{},
			},
			want: want{
				err:   nil,
				count: 0,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(s.ctx, (*string)(nil), (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			result, err := s.srv.List(s.ctx, tc.args.req)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(result)

				return
			}

			s.NoError(err)
			s.Len(result, tc.want.count)

			if tc.want.count == 0 {
				return
			}

			s.Equal(tc.want.name, result[0].Name)
			s.Equal(tc.want.price, result[0].Price)
		})
	}
}

type updateProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *updateProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestUpdateProductSuite(t *testing.T) {
	suite.Run(t, new(updateProductSuite))
}

func (s *updateProductSuite) TestUpdate() {
	type args struct {
		guid uuid.UUID
		req  entity.RequestProductUpdate
	}
	type want struct {
		err   error
		name  string
		price int64
	}

	guid := uuid.Must(uuid.NewV4())
	categoryGUID := uuid.Must(uuid.NewV4())
	newCategoryGUID := uuid.Must(uuid.NewV4())
	anotherGUID := uuid.Must(uuid.NewV4())

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "full update",
			args: args{
				guid: guid,
				req: entity.RequestProductUpdate{
					Name:         testutil.PtrString("Updated Product"),
					Description:  testutil.PtrString("Updated description"),
					Price:        testutil.PtrInt64(2000),
					CategoryGUID: testutil.PtrUUID(newCategoryGUID),
				},
			},
			want: want{
				err:   nil,
				name:  "Updated Product",
				price: 2000,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{{
						GUID:         guid,
						Name:         "Old Product",
						Price:        1000,
						CategoryGUID: categoryGUID,
					}}, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{*args.req.CategoryGUID}).
					Return([]entity.Category{{GUID: newCategoryGUID}}, nil).
					Once()

				s.productRepo.EXPECT().
					Update(s.ctx, mock.MatchedBy(func(p entity.Product) bool {
						return p.GUID == guid &&
							p.Name == *args.req.Name &&
							p.Description == args.req.Description &&
							p.Price == *args.req.Price &&
							p.CategoryGUID == *args.req.CategoryGUID &&
							!p.UpdatedAt.IsZero()
					})).
					Return(nil).
					Once()
			},
		},
		{
			name: "partial update - name only",
			args: args{
				guid: guid,
				req: entity.RequestProductUpdate{
					Name: testutil.PtrString("Renamed Product"),
				},
			},
			want: want{
				err:   nil,
				name:  "Renamed Product",
				price: 1000,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{{
						GUID:         guid,
						Name:         "Old Product",
						Price:        1000,
						CategoryGUID: categoryGUID,
					}}, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.productRepo.EXPECT().
					Update(s.ctx, mock.MatchedBy(func(p entity.Product) bool {
						return p.GUID == guid &&
							p.Name == *args.req.Name &&
							p.Price == 1000 &&
							p.CategoryGUID == categoryGUID
					})).
					Return(nil).
					Once()
			},
		},
		{
			name: "not found",
			args: args{
				guid: guid,
				req:  entity.RequestProductUpdate{},
			},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
		{
			name: "duplicate name",
			args: args{
				guid: guid,
				req: entity.RequestProductUpdate{
					Name: testutil.PtrString("Duplicate Product"),
				},
			},
			want: want{err: entity.ErrAlreadyExists},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{{
						GUID:         guid,
						Name:         "Old Product",
						Price:        1000,
						CategoryGUID: categoryGUID,
					}}, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{{
						GUID: anotherGUID,
						Name: "Duplicate Product",
					}}, nil).
					Once()
			},
		},
		{
			name: "category not found",
			args: args{
				guid: guid,
				req: entity.RequestProductUpdate{
					CategoryGUID: testutil.PtrUUID(newCategoryGUID),
				},
			},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{{
						GUID:         guid,
						Name:         "Old Product",
						Price:        1000,
						CategoryGUID: categoryGUID,
					}}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{*args.req.CategoryGUID}).
					Return([]entity.Category{}, nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			result, err := s.srv.Update(s.ctx, tc.args.guid, tc.args.req)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(result.GUID)

				return
			}

			s.NoError(err)
			s.Equal(tc.want.name, result.Name)
			s.Equal(tc.want.price, result.Price)
		})
	}
}
