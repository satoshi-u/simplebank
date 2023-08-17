package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

var ErrInvalidSigningMethod = errors.New("token is not signed with HS256")
var ErrInvalidPayload = errors.New("token payload not in expecteed format")
var ErrInvalidToken = errors.New("token is invalid")

// JWTMaker is a JSON Web Token maker - use symmetric-key algo to sign the function
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretkey string) (Maker, error) {
	if len(secretkey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be atleast %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey: secretkey}, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewJWTPayload(username, duration)
	if err != nil {
		return "", err
	}

	// Declare the token with the algorithm used for signing, and the payload (which has an embedded JWT claim)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// Create the JWT string
	// JWTs are commonly signed using one of two algorithms: HS256 (HMAC using SHA256) and RS256 (RSA using SHA256).
	// Here we sign with HS256
	tokenString, err := jwtToken.SignedString([]byte(maker.secretKey))
	if err != nil {
		// fmt.Println(err)
		// If there is an error in signing the JWT, return that error
		return "", err
	}

	return tokenString, nil
}

// VerifyToken checks if the token is valid or not, if yes, return payload data in body of token
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// jwt.KeyFunc to pass the JWTMaker's secret key in jwt.ParseWithClaims
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		// type assertion to check if token was signed with jwt..SigningMethodHS256 by
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			// If the signing method doesn't match, return ErrInvalidSigningMethod
			return nil, ErrInvalidSigningMethod
		}

		return []byte(maker.secretKey), nil
	}

	// Parse the JWT string and store the result in `payload`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	// JWTs are commonly signed using one of two algorithms: HS256 (HMAC using SHA256) and RS256 (RSA using SHA256).
	// Here we verify those only signed with HS256
	jwtToken, err := jwt.ParseWithClaims(token, &JWTPayload{}, keyfunc)
	if err != nil {
		// fmt.Println(err)
		// If there is an error in signing the JWT, return that error
		return nil, err
	}
	if !jwtToken.Valid {
		return nil, ErrInvalidToken
	}

	// get jwtPayload
	jwtPayload, ok := jwtToken.Claims.(JWTPayload)
	if !ok {
		// If there is an error in type assertion of payload from token, return ErrInvalidPayload
		return nil, ErrInvalidPayload
	}

	// return payload
	return &Payload{
		ID:        jwtPayload.Payload.ID,
		Username:  jwtPayload.Payload.Username,
		IssuedAt:  jwtPayload.Payload.IssuedAt,
		ExpiresAt: jwtPayload.Payload.ExpiresAt,
	}, nil
}
