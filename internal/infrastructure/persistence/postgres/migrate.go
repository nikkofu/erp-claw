package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	migrator *migrate.Migrate
	closeDB  func() error
}

func NewMigrator(cfg Config, migrationsPath string) (*Migrator, error) {
	db, err := New(cfg)
	if err != nil {
		return nil, err
	}

	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Migrator{
		migrator: m,
		closeDB:  db.Close,
	}, nil
}

func ApplyUp(ctx context.Context, cfg Config, migrationsPath string) error {
	m, err := NewMigrator(cfg, migrationsPath)
	if err != nil {
		return err
	}
	defer m.Close()
	return m.Up(ctx)
}

func (m *Migrator) Up(_ context.Context) error {
	if err := m.migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (m *Migrator) Down(_ context.Context) error {
	if err := m.migrator.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (m *Migrator) Close() error {
	sourceErr, dbErr := m.migrator.Close()
	closeErr := m.closeDB()
	if sourceErr != nil {
		return sourceErr
	}
	if dbErr != nil {
		return dbErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}
