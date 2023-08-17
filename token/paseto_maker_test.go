package token

import (
	"strings"
	"testing"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/require"
	"github.com/web3dev6/simplebank/util"
)

func TestPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(chacha20poly1305.KeySize))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiresAt, time.Second)
}

func TestExpiredPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(minSecretKeySize))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute
	token, err := maker.CreateToken(username, -duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

// invalid payload in paseto token body
func TestInvalidPasetoTokenInvalidPayload(t *testing.T) {
	// create payload
	username := util.RandomOwner()
	duration := time.Minute
	tokenID, err := uuid.NewRandom()
	require.NoError(t, err)
	invalidPayload := struct {
		Id        string    `json:"id"`
		User      string    `json:"user"`
		ExpiresAt time.Time `json:"expires_at"`
	}{
		Id:        tokenID.String(),
		User:      username,
		ExpiresAt: time.Now().Add(duration),
	}

	skey := util.RandomString(chacha20poly1305.KeySize)

	// create token with an invalid payload
	tokenString, err := paseto.NewV2().Encrypt([]byte(skey), invalidPayload, nil)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	maker, err := NewPasetoMaker(skey)
	require.NoError(t, err)
	// try to verify the above created token with invalid payload
	payload, err := maker.VerifyToken(tokenString)
	require.Error(t, err)
	require.EqualError(t, err, strings.Join(
		[]string{
			ErrInvalidPayload.Error(),
		}, ": "))
	require.Nil(t, payload)
}
