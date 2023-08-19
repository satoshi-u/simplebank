package token

import "errors"

var ErrInvalidSigningMethod = errors.New("token is not signed with HS256")
var ErrInvalidPayload = errors.New("token payload not in expecteed format")
var ErrInvalidToken = errors.New("token is invalid")
var ErrExpiredToken = errors.New("token has expired")
