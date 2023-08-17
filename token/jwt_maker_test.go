package token

import (
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/web3dev6/simplebank/util"
)

func TestJWTToken(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(minSecretKeySize))
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

func TestExpiredJWTToken(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(minSecretKeySize))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute
	token, err := maker.CreateToken(username, -duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, strings.Join(
		[]string{
			jwt.ErrTokenInvalidClaims.Error(),
			jwt.ErrTokenExpired.Error(),
		}, ": "))
	require.Nil(t, payload)
}

// trivial attack with none algo header in jwt
func TestInvalidJWTTokenAlgoNone(t *testing.T) {
	// create payload
	username := util.RandomOwner()
	duration := time.Minute
	jwtPayload, err := NewJWTPayload(username, duration)
	require.NoError(t, err)

	// create invalid token with no signature
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, jwtPayload)            // SigningMethodNone - provided only for testing purposes
	tokenString, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType) // UnsafeAllowNoneSignatureType - provided only for testing purposes
	require.NoError(t, err)

	maker, err := NewJWTMaker(util.RandomString(minSecretKeySize))
	require.NoError(t, err)
	// try to verify the above created invalid token
	payload, err := maker.VerifyToken(tokenString)
	require.Error(t, err)
	require.EqualError(t, err, strings.Join(
		[]string{
			jwt.ErrTokenUnverifiable.Error(),
			"error while executing keyfunc",
			ErrInvalidSigningMethod.Error(),
		}, ": "))
	require.Nil(t, payload)
}

// token signed with a different algorithm than the once server expects in VerifyToken
func TestInvalidJWTTokenAlgoWrong(t *testing.T) {
	// create payload
	username := util.RandomOwner()
	duration := time.Minute
	claims :=
		jwt.RegisteredClaims{
			ID:        username,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		}

	// create invalid token with rsa signature instead of hmac
	rsaPvtKey := util.GenerateRsaPrivateKey(2048)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := jwtToken.SignedString(rsaPvtKey)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	maker, err := NewJWTMaker(util.ConvertRsaPrivateKeyToPemString(rsaPvtKey))
	require.NoError(t, err)
	// try to verify the above created invalid token
	payload, err := maker.VerifyToken(tokenString)
	require.Error(t, err)
	require.EqualError(t, err, strings.Join(
		[]string{
			jwt.ErrTokenUnverifiable.Error(),
			"error while executing keyfunc",
			ErrInvalidSigningMethod.Error(),
		}, ": "))
	require.Nil(t, payload)
}

// invalid payload in jwt token body
func TestInvalidJWTTokenInvalidPayload(t *testing.T) {
	// create payload
	username := util.RandomOwner()
	duration := time.Minute
	tokenID, err := uuid.NewRandom()
	require.NoError(t, err)
	invalidJwtPayload := struct {
		User string
		*jwt.RegisteredClaims
	}{
		User: username,
		RegisteredClaims: &jwt.RegisteredClaims{
			ID:        tokenID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			Issuer:    "simple_bank",
		},
	}

	// create token with an invalid payload
	skey := util.RandomString(minSecretKeySize)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, invalidJwtPayload)
	tokenString, err := jwtToken.SignedString([]byte(skey))
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	maker, err := NewJWTMaker(skey)
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
