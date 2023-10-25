package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	db "github.com/web3dev6/simplebank/db/sqlc"
	"github.com/web3dev6/simplebank/token"
	"github.com/web3dev6/simplebank/util"
	"github.com/web3dev6/simplebank/worker"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
		TokenMakerType:      "PASETO",
	}

	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return server
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}
	return metadata.NewIncomingContext(context.Background(), md)
}
