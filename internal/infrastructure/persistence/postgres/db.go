package postgres

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.DSN) == "" {
		return errors.New("postgres dsn is required")
	}
	return nil
}

func New(cfg Config) (*sql.DB, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	parsed, err := pgx.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDB(*parsed)
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	return db, nil
}
