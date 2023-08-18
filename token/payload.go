package token

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrExpiredToken = errors.New("token has expired")

// Payload contains custom payload data of the token
// note* jwt.RegisteredClaims must be added in payload so that payload becomes a valid jwt claim
type Payload struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	Username  string    `json:"username" validate:"required"`
	IssuedAt  time.Time `json:"issued_at" validate:"required"`
	ExpiresAt time.Time `json:"expires_at" validate:"required"`
}

// NewPayload creates a new token payload with specified username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	// create payload
	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}
	return payload, nil
}

// Note: If you provide a custom claim implementation that embeds one of the standard claims (such as RegisteredClaims),
// make sure that a) you either embed a non-pointer version of the claims or b) if you are using a pointer, allocate the
// proper memory for it before passing in the overall claims, otherwise you might run into a panic.
type JWTPayload struct {
	*Payload             // embedded struct pointer-type
	jwt.RegisteredClaims // embedded struct non-pointer-type (recommended)
}

// NewJWTPayload creates a new jwt payload with specified username and duration
func NewJWTPayload(username string, duration time.Duration) (*JWTPayload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	// create jwtPayload
	jwtPayload := &JWTPayload{
		Payload: &Payload{
			ID:        tokenID,
			Username:  username,
			IssuedAt:  time.Now(),
			ExpiresAt: time.Now().Add(duration),
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			Issuer:    "simple_bank",
		},
	}
	return jwtPayload, nil
}

// Valid checks if the token payload is valid or not - write custom token-check logic here
// note* ExpiresAt already verified in claims - this is optional (must in PASETO)
func (payload *Payload) Valid() error {
	// basic check if token is expired
	fmt.Println(time.Now())
	fmt.Println(payload.ExpiresAt)
	if time.Now().After(payload.ExpiresAt) {
		return ErrExpiredToken
	}
	return nil
}
