postgres:
	docker run --name simple-bank-db -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it simple-bank-db createdb --username=root --owner=root simple_bank
	docker exec -it simple-bank-db createdb --username=root --owner=root simple_bank_test

dropdb:
	docker exec -it simple-bank-db dropdb simple_bank
	docker exec -it simple-bank-db dropdb simple_bank_test

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank_test?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank_test?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -destination db/mock/store.go -package mockdb github.com/web3dev6/simplebank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock