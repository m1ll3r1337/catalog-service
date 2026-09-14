package builder

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/m1ll3r1337/catalog-service/internal/app/config"
	"github.com/m1ll3r1337/catalog-service/internal/app/processor"
	pprocessor "github.com/m1ll3r1337/catalog-service/internal/app/processor/other"
	"github.com/m1ll3r1337/catalog-service/internal/app/repository"
	pcategory "github.com/m1ll3r1337/catalog-service/internal/app/repository/category"
	rcpostgres "github.com/m1ll3r1337/catalog-service/internal/app/repository/conn/postgres"
	pproduct "github.com/m1ll3r1337/catalog-service/internal/app/repository/product"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type Builder struct {
	cCtx         *cli.Context
	ctx          context.Context
	wg           sync.WaitGroup
	err          error
	cfg          config.Config
	connPostgres *rcpostgres.Client
	categoryRepo repository.Category
	productRepo  repository.Product
	processors   []processor.Processor
}

func NewBuilder(cCtx *cli.Context) *Builder {
	return &Builder{
		cCtx: cCtx,
		ctx:  context.Background(),
	}
}

func (b *Builder) BuildConfig() {
	b.exec(func(b *Builder) {
		b.buildConfig()
	})
}

func (b *Builder) Run() {
	if b.err != nil {
		log.Fatal().Err(b.err).Msg("Failed to initialize application")
	}

	log.Info().Msg("Application initialized")
	defer log.Info().Msg("Application completed")

	for _, p := range b.processors {
		p.StartAsync(b.ctx, &b.wg)
	}

	b.wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORY CONNECTIONS ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildRepoConnPostgres() {
	b.exec(func(b *Builder) {
		conn, err := rcpostgres.NewClient(b.ctx, b.cfg.Repository.Postgres)
		if err != nil {
			b.err = fmt.Errorf("failed to create postgres conn: %w", err)
			return
		}

		b.connPostgres = conn
	})
}

func (b *Builder) BuildRepoConnMigrator() {
	if b.connPostgres == nil {
		return
	}

	b.exec(func(b *Builder) {
		migrator := pprocessor.NewMigrator(b.connPostgres)
		b.processors = append(b.processors, migrator)
	})
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORIES /////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildRepoCategory() {
	b.exec(func(b *Builder) {
		repo := pcategory.NewRepoFromPostgres(b.connPostgres)
		b.categoryRepo = repo
	}, b.connPostgres)
}

func (b *Builder) BuildRepoProduct() {
	b.exec(func(b *Builder) {
		repo := pproduct.NewRepoFromPostgres(b.connPostgres)
		b.productRepo = repo
	}, b.connPostgres)
}

////////////////////////////////////////////////////////////////////////////////
///// PRIVATE //////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) buildConfig() {
	args := config.LoadArgs{
		Output:          b.cCtx.App.Writer,
		EnableSimpleLog: b.cCtx.Bool("no-json"),
	}

	config.Load(args)
	b.cfg = config.Root
}

func (b *Builder) exec(cb func(b *Builder), requiredArgs ...any) {
	if b.err != nil {
		return
	}

	for i, requiredArg := range requiredArgs {
		rv := reflect.ValueOf(requiredArg)
		if !rv.IsValid() {
			b.err = fmt.Errorf("BUG: required argument #%d is nil (check dependencies)", i)
			return
		}
		if rv.Type().Kind() == reflect.Struct || !rv.IsZero() {
			continue
		}
		b.err = fmt.Errorf("BUG: required %s, but empty", rv.Type().String())
		return
	}

	cb(b)
}
