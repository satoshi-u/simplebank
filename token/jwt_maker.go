package token

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

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
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	jwtPayload, err := NewJWTPayload(username, duration)
	if err != nil {
		return "", jwtPayload.Payload, err
	}

	// Declare the token with the algorithm used for signing, and the jwtPayload (which has an embedded JWT claim)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtPayload)

	// Create the JWT string
	// JWTs are commonly signed using one of two algorithms: HS256 (HMAC using SHA256) and RS256 (RSA using SHA256).
	// Here we sign with HS256
	tokenString, err := jwtToken.SignedString([]byte(maker.secretKey))
	if err != nil {
		// fmt.Println(err)
		// If there is an error in signing the JWT,  return that error
		return "", jwtPayload.Payload, err
	}

	return tokenString, jwtPayload.Payload, nil
}

// VerifyToken checks if the token is valid or not, if yes, return payload data in body of token
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// Initialize a new instance of `JWTPayload`
	jwtPayload := &JWTPayload{
		Payload: &Payload{},
		// RegisteredClaims: &jwt.RegisteredClaims{},
	}

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
	jwtToken, err := jwt.ParseWithClaims(token, jwtPayload, keyfunc)
	if err != nil {
		// fmt.Println(err)
		// If there is an error in signing the JWT, return that error
		return nil, err
	}
	if !jwtToken.Valid {
		return nil, ErrInvalidToken
	}

	// get an instance of a validator
	v := validator.New()
	// call the `Struct` function passing in your payload
	err = v.Struct(jwtPayload)
	if err != nil {
		// If there is any error in type JWTPayload struct validation, return ErrInvalidPayload
		return nil, ErrInvalidPayload
	}

	err = jwtPayload.Valid()
	if err != nil {
		// fmt.Println(err)
		// If there is an error in validating the payload(check expiry), return that error
		return nil, err
	}

	// return payload
	return &Payload{
		ID:        jwtPayload.Payload.ID,
		Username:  jwtPayload.Payload.Username,
		IssuedAt:  jwtPayload.Payload.IssuedAt,
		ExpiresAt: jwtPayload.Payload.ExpiresAt,
	}, nil
}
