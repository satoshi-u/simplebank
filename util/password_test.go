package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := RandomString(6)

	// check if HashPassword works
	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	// check if CheckPassword works
	err = CheckPassword(password, hashedPassword)
	require.NoError(t, err)

	// check if a wrong password throws error and can't generate the stored hash
	wrongPassword := RandomString(6)
	err = CheckPassword(wrongPassword, hashedPassword)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	// check if a new hash is created from the password every time
	_hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, _hashedPassword)
	require.NotEmpty(t, _hashedPassword)
	require.NotEqual(t, hashedPassword, _hashedPassword)
}
