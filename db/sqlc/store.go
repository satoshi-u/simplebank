package db

import (
	"context"
	"database/sql"
	"fmt"
)

// a generic interfce for store - Querier stores all generated CRUD, and rest are our custom operations on DB
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions - a real db (postgres in app)
type SQLStore struct {
	*Queries // extend struct functionality in golang - inheritance equivalent
	db       *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{db: db, Queries: New(db)}
}

// execTx executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil) // todo use &sql.TxOptions{}
	if err != nil {
		return err
	}
	q := New(tx) // New can work with either *sql.DB or *sql.Tx - DBTX interface
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error: %v", err, rbErr)
		}
		return err
	}

	// simulate db traffic delay scenarios here
	// time.Sleep(2 * time.Second)

	return tx.Commit()
}
