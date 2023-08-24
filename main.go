package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/lib/pq"
	"github.com/web3dev6/simplebank/api"
	db "github.com/web3dev6/simplebank/db/sqlc"
	"github.com/web3dev6/simplebank/gapi"
	"github.com/web3dev6/simplebank/pb"
	"github.com/web3dev6/simplebank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// load config from app.env
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// open conn to db
	conn, err := sql.Open(config.DbDriver, config.DbSourceMain)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// create store, and then server
	store := db.NewStore(conn)

	if config.ServerType == "GRPC" {
		// run grpc server on 9090
		runGrpcServer(config, store)
	} else if config.ServerType == "HTTP" {
		// run server on 8080
		runGinServer(config, store)
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	// start server
	err = server.Start(config.HttpServerAddress)
	if err != nil {
		log.Fatal("cannot start http server:", err)
	}
}

func runGrpcServer(config util.Config, store db.Store) {
	// create a simple_bank server struct which embeds pb.UnimplementedSimpleBankServer
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	// grpcServer is a new grpc server instacnce
	grpcServer := grpc.NewServer()
	// register simple_bank server(has unimplemented service) with this grpcServer
	pb.RegisterSimpleBankServer(grpcServer, server)

	// [optional]
	// Register a grpc reflection for server
	// Register registers the server reflection service on the given gRPC server
	// Allows a grpc client to easily explore - what RPCs are available and how to cal them
	reflection.Register(grpcServer)

	// create listener to listen to gRPC requests on a specified port
	listener, err := net.Listen("tcp", config.GrpcServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	// start server with listener
	log.Printf("starting gRPC server at %s...", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start grpc server:", err)
	}
}

// func initDbWithAccounts(store db.Store, numTestAccounts int) {
// 	// create some accounts if none exists in db
// 	count, err := store.GetCountForAccounts(context.Background())
// 	if err != nil {
// 		log.Fatal("error in getting count for accounts from db.store")
// 	}
// 	if count == 0 {
// 		log.Printf("store empty! Creating some accounts before starting server...")
// 		var accounts = []db.Account{}
// 		for i := 0; i < numTestAccounts; i++ {
// 			account, err := store.CreateAccount(context.Background(), db.CreateAccountParams{
// 				Owner:    util.RandomOwner(),
// 				Balance:  util.RandomBalance(),
// 				Currency: util.RandomCurrency()},
// 			)
// 			if err != nil {
// 				log.Fatal("error in creating accounts")
// 			}
// 			accounts = append(accounts, account)
// 		}
// 		log.Printf("num (accounts created) = %d\n", len(accounts))
// 	}
// }
