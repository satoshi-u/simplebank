package gapi

import (
	"fmt"

	db "github.com/web3dev6/simplebank/db/sqlc"
	"github.com/web3dev6/simplebank/pb"
	token "github.com/web3dev6/simplebank/token"
	"github.com/web3dev6/simplebank/util"
)

// Server serves gRPC requests for our banking service
type Server struct {
	store                            db.Store    // do the transfer_tx
	tokenMaker                       token.Maker // manage tokens for users
	config                           util.Config // store config used to start the server
	pb.UnimplementedSimpleBankServer             // gRPCs work right away without impl- forward compatibility
}

// NewServer creates a new HTTP server and setup routing for service
func NewServer(config util.Config, store db.Store) (*Server, error) {
	// token maker for auth handling from config
	var tokenMaker token.Maker
	var err error
	switch config.TokenMakerType {
	case "JWT":
		tokenMaker, err = token.NewJWTMaker(config.TokenSymmetricKey)
	case "PASETO":
		tokenMaker, err = token.NewPasetoMaker(config.TokenSymmetricKey)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	// server instance with store, tokenMaker & config
	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}

	return server, nil
}
