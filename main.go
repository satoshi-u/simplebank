package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/web3dev6/simplebank/api"
	db "github.com/web3dev6/simplebank/db/sqlc"
	"github.com/web3dev6/simplebank/util"
)

// const (
// 	dbDriver      = "postgres"
// 	dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
// 	serverAddress = "0.0.0.0:8080"
// )

func main() {
	// load config from app.env
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// open conn to db
	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}

	// create store, and then server
	store := db.NewStore(conn)
	server := api.NewServer(store)

	// start server
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server")
	}
}
