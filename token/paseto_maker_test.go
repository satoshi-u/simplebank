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

// Test VerifyToken for token whose tokenString, symmetricKey, and payload is known
// func TestPasetoVerifyToken(t *testing.T) {
// 	maker, err := NewPasetoMaker("12345678901234567890123456789012")
// 	require.NoError(t, err)
// 	token := "v2.local.TOH11_leZ5mRXv-B_M6nlV8rG3QLkYxvVjNpJAnSsEmR49YF-pVDmMhOMoji747OOQSyG--g3py6jdraruNeHLFcwV1bkPimNIM3IMca6AeIa67BgJe0MqZrPemvvHAcOdWq8UjPWW86KftQ9DZOZZnkIv5m-gpZudVqBzIrWJPvC3IyF3TDl9O5qJLizQ0oLAIgp8Furd64_i3iVCceG9u3jes6xmfNQA6guKTl0yrt7JH_urODN24c3G-mfAwP88jHMQ4m38e5KuySxYc.bnVsbA"
// 	payload, err := maker.VerifyToken(token)

// 	require.NoError(t, err)
// 	require.NotEmpty(t, payload)

// 	require.Equal(t, payload.ID.String(), "1dd8bf50-5971-404c-bb0c-9a946d226071")
// 	require.Equal(t, payload.Username, "sarthakxalts")
// 	require.Equal(t, payload.IssuedAt.String(), "2023-08-18 20:28:15.998132 +0530 IST")
// 	require.Equal(t, payload.ExpiresAt.String(), "2023-08-18 20:43:15.998132 +0530 IST")
// }
