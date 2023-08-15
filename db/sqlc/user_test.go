package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/web3dev6/simplebank/util"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6)) // can be hashed once more and both stored
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user := createRandomUser(t)
	userFromDB, err := testQueries.GetUser(context.Background(), user.Username)

	require.NoError(t, err)
	require.NotEmpty(t, userFromDB)
	require.Equal(t, user.Username, userFromDB.Username)
	require.Equal(t, user.HashedPassword, userFromDB.HashedPassword)
	require.Equal(t, user.FullName, userFromDB.FullName)
	require.Equal(t, user.Email, userFromDB.Email)

	require.WithinDuration(t, user.CreatedAt, userFromDB.CreatedAt, time.Nanosecond)
	require.WithinDuration(t, user.PasswordChangedAt, userFromDB.PasswordChangedAt, time.Nanosecond)
}
