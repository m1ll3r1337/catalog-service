package rcpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/m1ll3r1337/catalog-service/internal/app/config/section"
	"github.com/m1ll3r1337/catalog-service/migration"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
)

type (
	Client struct {
		_bunDB
		rawBunDB *bun.DB

		cfg section.RepositoryPostgres
	}

	_bunDB = bun.IDB
)

func (c *Client) GetRawBunDB() *bun.DB {
	return c.rawBunDB
}

func NewClient(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	u := url.URL{
		Scheme: "postgres",
		Host:   cfg.Address,
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Path:   cfg.Name,
	}

	args := make(url.Values)
	args.Set("sslmode", "disable")
	u.RawQuery = args.Encode()

	dsn := u.String()

	log.Printf("ReadTimeout: %v | WriteTimeout: %v", cfg.ReadTimeout, cfg.WriteTimeout)

	conn := pgdriver.NewConnector(pgdriver.WithDSN(dsn),
		pgdriver.WithReadTimeout(cfg.ReadTimeout), pgdriver.WithWriteTimeout(cfg.WriteTimeout))

	sqlDB := sql.OpenDB(conn)

	sqlDB.SetMaxOpenConns(10)

	bunDB := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := bunDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Postgres database: %w", err)
	}

	return &Client{
		_bunDB:   newTxInjector(bunDB),
		rawBunDB: bunDB,
		cfg:      cfg,
	}, nil
}

func (c *Client) Migrate(ctx context.Context) (oldVer, newVer int64, err error) {
	migrations := migrate.NewMigrations()

	if err := migrations.Discover(migration.Postgres); err != nil {
		return 0, 0, fmt.Errorf("discover embedded postgres migrations: %w", err)
	}

	migrator := migrate.NewMigrator(c.rawBunDB, migrations,
		migrate.WithTableName(c.cfg.MigrationTable),
		migrate.WithLocksTableName(c.cfg.MigrationTable+"_locks"),
		migrate.WithMarkAppliedOnSuccess(true),
	)

	if err := migrator.Init(ctx); err != nil {
		return 0, 0, fmt.Errorf("ensure migration tables exist: %w", err)
	}

	oldVer, err = getLatestVersion(ctx, migrator)
	if err != nil {
		return 0, 0, fmt.Errorf("get current db version: %w", err)
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return oldVer, oldVer, fmt.Errorf("apply pending migrations: %w", err)
	}

	newVer = oldVer
	for _, m := range group.Migrations {
		v, parseErr := strconv.ParseInt(m.Name, 10, 64)
		if parseErr != nil {
			return oldVer, oldVer, fmt.Errorf("parse version from migration name %q: %w", m.Name, parseErr)
		}
		if v > newVer {
			newVer = v
		}
	}

	return oldVer, newVer, nil
}

func getLatestVersion(ctx context.Context, m *migrate.Migrator) (int64, error) {
	applied, err := m.AppliedMigrations(ctx)
	if err != nil {
		return 0, fmt.Errorf("list applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return 0, nil
	}

	var maxVer int64 = 0
	for _, mg := range applied {
		v, err := strconv.ParseInt(mg.Name, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse version from migration name %q: %w", mg.Name, err)
		}
		if v > maxVer {
			maxVer = v
		}
	}

	return maxVer, nil
}

func (c *Client) InsideTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if tx := getTxFromContext(ctx); tx.Tx != nil {
		return fn(ctx)
	}

	tx, err := c.rawBunDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	done := false
	defer func() {
		if !done {
			_ = tx.Rollback()
		}
	}()

	txCtx := setTxToContext(ctx, tx)
	err = fn(txCtx)
	if err != nil {
		return err
	}

	done = true
	return tx.Commit()
}
