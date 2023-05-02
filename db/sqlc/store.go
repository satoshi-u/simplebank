package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
type Store struct {
	*Queries // extend struct functionality in golang - inheritance equivalent
	db       *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db, Queries: New(db)}
}

// execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil) // &sql.TxOptions{} - todo later
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
	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountId int64 `json:"from_account_id"`
	ToAccountId   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult contains the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var txKey = struct{}{}

// TransferTx performs a money transfer from one account to other
// It creates a transfer record, add account entries, and update accounts' balance within a single db tx
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// get tx name from ctx
		txName := ctx.Value(txKey)

		// transfer
		fmt.Println(txName, "create Transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountId,
			ToAccountID:   arg.ToAccountId,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// from entry
		fmt.Println(txName, "create FromEntry")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountId,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// to entry
		fmt.Println(txName, "create ToEntry")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountId,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// get account ->  update from accounts' balance
		// fmt.Println(txName, "get accountFrom")
		// accountFrom, err := q.GetAccountForUpdate(ctx, arg.FromAccountId)
		// if err != nil {
		// 	return err
		// }
		// fmt.Println(txName, "update accountFrom")
		// result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
		// 	ID:      arg.FromAccountId,
		// 	Balance: accountFrom.Balance - arg.Amount,
		// })
		// if err != nil {
		// 	return err
		// }

		// get account ->  update to accounts' balance
		// fmt.Println(txName, "get accountTo")
		// accountTo, err := q.GetAccountForUpdate(ctx, arg.ToAccountId)
		// if err != nil {
		// 	return err
		// }
		// fmt.Println(txName, "update accountTo")
		// result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
		// 	ID:      arg.ToAccountId,
		// 	Balance: accountTo.Balance + arg.Amount,
		// })
		// if err != nil {
		// 	return err
		// }

		// update account_from balance with 1 single query
		fmt.Println(txName, "update accountFrom")
		result.FromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			ID:     arg.FromAccountId,
			Amount: -arg.Amount,
		})
		if err != nil {
			return err
		}

		// update account_to balance with 1 single query
		fmt.Println(txName, "update accountTo")
		result.ToAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			ID:     arg.ToAccountId,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}
