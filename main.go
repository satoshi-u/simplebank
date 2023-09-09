package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/web3dev6/simplebank/api"
	db "github.com/web3dev6/simplebank/db/sqlc"
	"github.com/web3dev6/simplebank/gapi"
	"github.com/web3dev6/simplebank/pb"
	"github.com/web3dev6/simplebank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

const initDbNumUserAccount = 10

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

	// create store with db conn
	store := db.NewStore(conn)

	// init accounts in db
	initDbWithMinUsersAccounts(store, initDbNumUserAccount)

	if config.ServerType == "HTTP" {
		// run http server on 8080
		runGinServer(config, store)
	} else if config.ServerType == "GRPC" {
		// run grpc server on 9090
		runGrpcServer(config, store)
	} else if config.ServerType == "GRPC_GATEWAY" {
		// run grpc's http gateway server on 8080 as a goroutine without blocking main
		go runGatewayServer(config, store)
		// run grpc server on 9090
		runGrpcServer(config, store)
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	// start server on a specified http port
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

	// create listener to listen to gRPC requests on a specified grpc port
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

func runGatewayServer(config util.Config, store db.Store) {
	// create a simple_bank server struct which embeds pb.UnimplementedSimpleBankServer
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot register handler  server:", err)
	}

	// jsonOptions for snake-case in names of json-fileds in response from gateway
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	// create a grpcMux using grpc-gateway's runtime package with jsonOption
	grpcMux := runtime.NewServeMux(jsonOption)

	// register simple_bank server with above created grpcMux, along with a context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	// create a http serveMux which takes http requests from client
	mux := http.NewServeMux()
	// to convert the http requests from client to grpcRequest, reroute them to grpcMux
	mux.Handle("/", grpcMux)
	// create a http-fs & serve auto-generated swagger docs for grpc-gateway server
	fs := http.FileServer(http.Dir("./doc/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs)) // StripPrefix strips the route prefix of the url before passing the request to the static file server

	// create listener to listen to client http requests on a specified http-gateway port
	listener, err := net.Listen("tcp", config.HttpServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	// start server with listener and http mux
	log.Printf("starting HTTP gateway server at %s...", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway ser  ver:", err)
	}
}

func initDbWithMinUsersAccounts(store db.Store, num int64) {
	count, err := store.GetCountForUsers(context.Background())
	if err != nil {
		log.Fatal("error in getting count for users from db.store")
	}
	if count < num {
		toAdd := num - count
		log.Printf("store: to add %d users with corresponding funded INR accounts!", toAdd)
		var users = []db.User{}
		var accounts = []db.Account{}
		for i := int64(0); i < toAdd; i++ {
			// create user
			hashedCommonPassword, err := util.HashPassword("secret")
			if err != nil {
				log.Fatal("error in hashing CommonPassword while creating user")
			}
			user, err := store.CreateUser(context.Background(), db.CreateUserParams{
				Username:       util.RandomString(8),
				HashedPassword: hashedCommonPassword,
				FullName:       util.RandomString(4) + util.RandomString(6),
				Email:          util.RandomEmail(),
			})
			if err != nil {
				log.Fatal("error in creating user")
			}
			users = append(users, user)
			// create account for user with INR as currency
			arg := db.CreateAccountParams{
				Owner:    user.Username,
				Balance:  util.RandomBalance(),
				Currency: util.INR,
			}
			account, err := store.CreateAccount(context.Background(), arg)
			if err != nil {
				log.Fatal("error in creating account for user")
			}
			accounts = append(accounts, account)
		}
		log.Printf("num (users created) = %d\n", len(users))
		log.Printf("num (accounts created) = %d\n", len(accounts))
	}
}
