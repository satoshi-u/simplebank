package api

import (
	"github.com/gin-gonic/gin"

	db "github.com/web3dev6/simplebank/db/sqlc"
)

// Server serves HTTP requests fo r our banking service
type Server struct {
	store  db.Store    // do the transfer_tx
	router *gin.Engine // send to correct handler for processing
}

// NewServer creates a new HTTP server and setup routing for service
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// add routes to router
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)
	server.router = router
	return server
}

// Start runs the http server on a specified address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
