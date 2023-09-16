package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/web3dev6/simplebank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

// authorizeUser
// Note: this logic can also be implemented using a gRPC interceptor
// but that won't work with http-gateway and will need to implement a separate HTTP middleware
func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	// inspect metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	// get value stored in authHeader
	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}
	authHeader := values[0]

	// get <authorization-type> <authorization-data::access-token>
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}
	authorizationType := strings.ToLower(fields[0])
	if authorizationType != authorizationBearer {
		return nil, fmt.Errorf("unsupported authorization type: %s", authorizationType)
	}
	accessToken := fields[1]

	// get payload
	payload, err := server.tokenMaker.VerifyToken(accessToken)

	if err != nil {
		return nil, fmt.Errorf("invalid access token %s: %w", accessToken, err)
	}
	return payload, nil
}
