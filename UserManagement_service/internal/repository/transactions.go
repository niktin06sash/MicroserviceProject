package repository

import (
	"context"
	"database/sql"
)

type TxManagerRepo struct {
	db *DBObject
}

func NewTxManagerRepo(db *DBObject) *TxManagerRepo {
	return &TxManagerRepo{db: db}
}

func (r *TxManagerRepo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.connect.BeginTx(ctx, nil)
}

func (r *TxManagerRepo) RollbackTx(tx *sql.Tx) error {
	return tx.Rollback()
}

func (r *TxManagerRepo) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}
