package api

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	db "github.com/web3dev6/simplebank/db/sqlc"
	token "github.com/web3dev6/simplebank/token"
	"github.com/web3dev6/simplebank/util"
	"github.com/web3dev6/simplebank/worker"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store           db.Store               // do the transfer_tx
	tokenMaker      token.Maker            // manage tokens for users
	router          *gin.Engine            // send to correct handler for processing
	config          util.Config            // store config used to start the server
	taskDistributor worker.TaskDistributor // To create tasks in redis queue
}

// NewServer creates a new HTTP server and setup routing for service
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
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
		store:           store,
		tokenMaker:      tokenMaker,
		config:          config,
		taskDistributor: taskDistributor,
	}
	// 	Gin Validator binding - register "currency" as a validator tag
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	// setup router with routes
	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	//  Default Gin router
	// router := gin.Default()

	// Initialize custom Gin for enabling structured logger
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultErrorWriter = os.Stderr
	router := gin.New()            // empty engine
	router.Use(loggerMiddleware()) // adds our new middleware
	router.Use(gin.Recovery())     // adds the default recovery middleware

	// authRoutes filter requests through our authMiddleware returned authHandler first
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	// add public routes to router
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	// add protected routes to authRoutes
	authRoutes.GET("/users", server.getUserDetails)
	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccounts)
	authRoutes.PATCH("/users", server.updateUser)

	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router
}

// Start runs the http server on a specified address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func abortWithErrorResponse(ctx *gin.Context, httpErrCode int, err error) {
	ctx.AbortWithError(httpErrCode, err)
	ctx.JSON(httpErrCode, errorResponse(err))
}
