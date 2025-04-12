package repository

import (
	"context"
	"database/sql"
)

type TxManagerRepo struct {
	Db *sql.DB
}

func NewTxManagerRepo(db *sql.DB) *TxManagerRepo {
	return &TxManagerRepo{Db: db}
}

func (r *TxManagerRepo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.Db.BeginTx(ctx, nil)
}

func (r *TxManagerRepo) RollbackTx(ctx context.Context, tx *sql.Tx) error {
	return tx.Rollback()
}

func (r *TxManagerRepo) CommitTx(ctx context.Context, tx *sql.Tx) error {
	return tx.Commit()
}
