SIMPLE_BANK_DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable
SIMPLE_BANK_TEST_DB_URL=postgresql://root:secret@localhost:5431/simple_bank_test?sslmode=disable

postgres:
	docker run --name simple-bank-db -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine
	docker run --name simple-bank-db-test -p 5431:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it simple-bank-db createdb --username=root --owner=root simple_bank
	docker exec -it simple-bank-db-test createdb --username=root --owner=root simple_bank_test
dropdb:
	docker exec -it simple-bank-db dropdb simple_bank
	docker exec -it simple-bank-db-test dropdb simple_bank_test

migrateup:
	migrate -path db/migration -database "$(SIMPLE_BANK_DB_URL)" -verbose up
	migrate -path db/migration -database "$(SIMPLE_BANK_TEST_DB_URL)" -verbose up
migrateup1 :
	migrate -path db/migration -database "$(SIMPLE_BANK_DB_URL)" -verbose up 1
	migrate -path db/migration -database "$(SIMPLE_BANK_TEST_DB_URL)" -verbose up 1
migratedown:
	migrate -path db/migration -database "$(SIMPLE_BANK_DB_URL)" -verbose down
	migrate -path db/migration -database "$(SIMPLE_BANK_TEST_DB_URL)" -verbose down
migratedown1:
	migrate -path db/migration -database "$(SIMPLE_BANK_DB_URL)" -verbose down 1
	migrate -path db/migration -database "$(SIMPLE_BANK_TEST_DB_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go clean -testcache && go test -short -cover ./...
	
test-verbose:
	go clean -testcache && go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -destination db/mock/store.go -package mockdb github.com/web3dev6/simplebank/db/sqlc Store

dbdocs:
	dbdocs build doc/db.dbml

dbschema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml
 
proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank,json_names_for_fields=false \
    proto/*.proto
	statik -src=./doc/swagger -dest=./doc

evans:
	evans --host localhost --port 9090 -r repl

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown sqlc test server mock dbdocs dbschema proto evans redis