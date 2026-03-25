package postgres

import (
	"context"
	"database/sql"
	"errors"
)

type TxHandler func(ctx context.Context, tx *sql.Tx) error

type TxRunner interface {
	InTx(ctx context.Context, fn TxHandler) error
}

type SQLTxRunner struct {
	db *sql.DB
}

func NewTxRunner(db *sql.DB) (*SQLTxRunner, error) {
	if db == nil {
		return nil, errors.New("postgres db is required")
	}
	return &SQLTxRunner{db: db}, nil
}

func (r *SQLTxRunner) InTx(ctx context.Context, fn TxHandler) (err error) {
	if fn == nil {
		return errors.New("transaction handler is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	err = fn(ctx, tx)
	return err
}
