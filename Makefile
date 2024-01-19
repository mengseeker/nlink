
.PHONY: build proto
proto:
	lib/bin/protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		core/api/*.proto

build:
	mkdir -p build/bin
	go build -o build/bin/nlink main.go
