package api

import "errors"

var ErrMissingAuthHeader = errors.New("missing authorization header")
var ErrInvalidAuthHeaderFormat = errors.New("invalid authorization header format")
var ErrUnsupportedAuthType = errors.New("unsupported authorization type in authorization header")
var ErrFetchingUnauthorizedAccount = errors.New("account doesn't belong to the authenticated user")
var ErrTransferringMoneyFromUnauthorizedAccount = errors.New("account doesn't belong to the authenticated user")
