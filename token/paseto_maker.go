package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/go-playground/validator/v10"
	"github.com/o1egl/paseto"
)

// PasetoMaker is a PASETO Token maker - use symmetric-key algo to sign the function
type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// NewPasetoMaker creates a new PasetoMaker
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be atleast %d characters", chacha20poly1305.KeySize)
	}
	return &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	jwtTokenString, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	if err != nil {
		// fmt.Println(err)
		// If there is an error in encrypting the payload, return that error
		return "", payload, err
	}
	return jwtTokenString, payload, nil
}

// VerifyToken checks if the token is valid or not, if yes, return payload data in body of token
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// get an instance of a validator
	v := validator.New()
	// call the `Struct` function passing in your payload
	err = v.Struct(payload)
	if err != nil {
		// If there is any error in type Payload struct validation, return ErrInvalidPayload
		return nil, ErrInvalidPayload
	}

	err = payload.Valid()
	if err != nil {
		// fmt.Println(err)
		// If there is an error in validating the payload(check expiry), return that error
		return nil, err
	}

	// return payload
	return payload, nil
}
