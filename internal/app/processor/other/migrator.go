package pprocessor

import (
	"context"
	"sync"

	"github.com/m1ll3r1337/catalog-service/internal/app/processor"
	"github.com/m1ll3r1337/catalog-service/internal/app/repository"
	"github.com/rs/zerolog/log"
)

type procMigrate struct {
	migrator repository.Migrate
}

func NewMigrator(migrator repository.Migrate) processor.Processor {
	return &procMigrate{
		migrator: migrator,
	}
}

func (p *procMigrate) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	processor.Wrap(ctx, wg, p.job)
}

func (p *procMigrate) job(ctx context.Context) {
	oldVer, newVer, err := p.migrator.Migrate(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to apply migrations")
		return
	}

	if oldVer != newVer {
		log.Info().Int64("old_version", oldVer).Int64("new_version", newVer).Msg("schema has been updated")
		return
	}

	log.Info().Msg("schema is up to date")
}
