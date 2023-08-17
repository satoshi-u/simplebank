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

type JWTPayload struct {
	*Payload
	*jwt.RegisteredClaims
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
		RegisteredClaims: &jwt.RegisteredClaims{
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
