package repository

import (
	"context"
	"database/sql"
)

type TxManagerRepo struct {
	Db *DBObject
}

func NewTxManagerRepo(db *DBObject) *TxManagerRepo {
	return &TxManagerRepo{Db: db}
}

func (r *TxManagerRepo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.Db.DB.BeginTx(ctx, nil)
}

func (r *TxManagerRepo) RollbackTx(tx *sql.Tx) error {
	return tx.Rollback()
}

func (r *TxManagerRepo) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}
