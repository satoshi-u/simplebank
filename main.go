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
// 	numTestAccounts = 10
// )

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
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	// create some accounts if none exists in db
	// count, err := store.GetCountForAccounts(context.Background())
	// if err != nil {
	// 	log.Fatal("error in getting count for accounts from db.store")
	// }
	// if count == 0 {
	// 	log.Printf("store empty! Creating some accounts before starting server...")
	// 	var accounts = []db.Account{}
	// 	for i := 0; i < numTestAccounts; i++ {
	// 		account, err := store.CreateAccount(context.Background(), db.CreateAccountParams{
	// 			Owner:    util.RandomOwner(),
	// 			Balance:  util.RandomBalance(),
	// 			Currency: util.RandomCurrency()},
	// 		)
	// 		if err != nil {
	// 			log.Fatal("error in creating accounts")
	// 		}
	// 		accounts = append(accounts, account)
	// 	}
	// 	log.Printf("num (accounts created) = %d\n", len(accounts))
	// }

	// start server
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
