package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"github.com/web3dev6/simplebank/api"
	db "github.com/web3dev6/simplebank/db/sqlc"
	_ "github.com/web3dev6/simplebank/doc/statik"
	"github.com/web3dev6/simplebank/gapi"
	"github.com/web3dev6/simplebank/pb"
	"github.com/web3dev6/simplebank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	initDbNumUserAccount = 10
)

func main() {
	// load config from app.env
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	// zerolog config
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		// *** To customize the configuration and formatting:
		// output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		// output.FormatLevel = func(i interface{}) string {
		// 	return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		// }
		// output.FormatMessage = func(i interface{}) string {
		// 	return fmt.Sprintf("MSG %s", i)
		// }
		// output.FormatFieldName = func(i interface{}) string {
		// 	return fmt.Sprintf("| FIELD %s:", i)
		// }
		// output.FormatFieldValue = func(i interface{}) string {
		// 	return strings.ToLower(fmt.Sprintf("%s", i))
		// }
		// log.Logger = zerolog.New(output).With().Timestamp().Logger()
	} else if config.Environment == "production" {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	// open conn to db
	conn, err := sql.Open(config.DbDriver, config.DbSourceMain)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	// run db migrations here, for both main and test
	runDBMigration(config.MigrationUrl, config.DbSourceMain)
	runDBMigration(config.MigrationUrl, config.DbSourceTest)

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
		log.Fatal().Err(err).Msg("cannot create server")
	}

	// start server on a specified http port
	err = server.Start(config.HttpServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start http server")
	}
}

func runGrpcServer(config util.Config, store db.Store) {
	// create a simple_bank server struct which embeds pb.UnimplementedSimpleBankServer
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	// grpc interceptor - logger
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	// grpcServer is a new grpc server instacnce, takes ServerOptions(interceptors like logger  )
	grpcServer := grpc.NewServer(grpcLogger)
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
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	// start server with listener
	log.Info().Msgf("starting gRPC server at %s...", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start grpc server")
	}
}

func runGatewayServer(config util.Config, store db.Store) {
	// create a simple_bank server struct which embeds pb.UnimplementedSimpleBankServer
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler  server")
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
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	// create a http serveMux which takes http requests from client
	mux := http.NewServeMux()
	// to convert the http requests from client to grpcRequest, reroute them to grpcMux
	mux.Handle("/", grpcMux)

	// create a http-fs & serve auto-generated swagger docs for grpc-gateway server
	// fs := http.FileServer(http.Dir("./doc/swagger"))
	// mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs)) // StripPrefix strips the route prefix of the url before passing the request to the static file server

	// create a statik-fs - static data already embedded in binary in `make proto``, no need to read from disk(Dockerfile)
	statikFs, err := fs.New() // alternatively, we can use NewWithNamespace func for custom ns
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs")
	}
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(statikFs)))

	// create listener to listen to client http requests on a specified http-gateway port
	listener, err := net.Listen("tcp", config.HttpServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	// start server with listener and http mux
	log.Info().Msgf("starting HTTP gateway server at %s...", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start HTTP gateway server")
	}
}

func initDbWithMinUsersAccounts(store db.Store, num int64) {
	count, err := store.GetCountForUsers(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("error in getting count for users from db.store")
	}
	if count < num {
		toAdd := num - count
		log.Info().Msgf("store: to add %d users with corresponding funded INR accounts!", toAdd)
		var users = []db.User{}
		var accounts = []db.Account{}
		for i := int64(0); i < toAdd; i++ {
			// create user
			hashedCommonPassword, err := util.HashPassword("secret")
			if err != nil {
				log.Fatal().Err(err).Msg("error in hashing CommonPassword while creating user")
			}
			user, err := store.CreateUser(context.Background(), db.CreateUserParams{
				Username:       util.RandomString(8),
				HashedPassword: hashedCommonPassword,
				FullName:       util.RandomString(4) + util.RandomString(6),
				Email:          util.RandomEmail(),
			})
			if err != nil {
				log.Fatal().Err(err).Msg("error in creating user")
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
				log.Fatal().Err(err).Msg("error in creating account for user")
			}
			accounts = append(accounts, account)
		}
		log.Info().Msgf("num (users created) = %d", len(users))
		log.Info().Msgf("num (accounts created) = %d", len(accounts))
	}
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run the migrate up")
	}
	// migration.Steps(numMigration)

	log.Info().Msgf("db migrate success for : %s", dbSource)
}
