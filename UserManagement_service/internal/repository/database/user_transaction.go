package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type UserTxManager struct {
	databaseclient *DBObject
}

func NewUserTxManager(db *DBObject) *UserTxManager {
	return &UserTxManager{databaseclient: db}
}

func (r *UserTxManager) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.databaseclient.pool.BeginTx(ctx, pgx.TxOptions{})
}

func (r *UserTxManager) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}

func (r *UserTxManager) CommitTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}
