package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/web3dev6/simplebank/util"
)

func createRandomTransfer(t *testing.T, account1, account2 Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomAmount(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	accountFrom := createRandomAccount(t)
	accountTo := createRandomAccount(t)
	createRandomTransfer(t, accountFrom, accountTo)
}

func TestGetTransfer(t *testing.T) {
	accountFrom := createRandomAccount(t)
	accountTo := createRandomAccount(t)
	transfer := createRandomTransfer(t, accountFrom, accountTo)

	transferFromDB, err := testQueries.GetTransfer(context.Background(), transfer.ID)

	require.NoError(t, err)
	require.NotEmpty(t, transferFromDB)
	require.Equal(t, transfer.ID, transferFromDB.ID)
	require.Equal(t, transfer.FromAccountID, transferFromDB.FromAccountID)
	require.Equal(t, transfer.ToAccountID, transferFromDB.ToAccountID)
	require.Equal(t, transfer.Amount, transferFromDB.Amount)
	require.WithinDuration(t, transfer.CreatedAt, transferFromDB.CreatedAt, time.Nanosecond)
}

func TestListTransfers(t *testing.T) {
	accountFrom := createRandomAccount(t)
	accountTo := createRandomAccount(t)
	for i := 0; i < 5; i++ {
		createRandomTransfer(t, accountFrom, accountTo)
		createRandomTransfer(t, accountFrom, accountTo)
	}

	arg := ListTransfersParams{
		FromAccountID: accountFrom.ID,
		ToAccountID:   accountTo.ID,
		Limit:         5,
		Offset:        5,
	}
	transfers, err := testQueries.ListTransfers(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, transfers, 5)
	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
		require.True(t, transfer.FromAccountID == accountFrom.ID || transfer.ToAccountID == accountTo.ID)
	}
}
