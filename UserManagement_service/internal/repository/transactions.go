package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TxManagerRepo struct {
	db *DBObject
}

func NewTxManagerRepo(db *DBObject) *TxManagerRepo {
	return &TxManagerRepo{db: db}
}

func (r *TxManagerRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.pool.BeginTx(ctx, pgx.TxOptions{})
}

func (r *TxManagerRepo) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}

func (r *TxManagerRepo) CommitTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}
