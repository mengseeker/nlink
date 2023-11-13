proto:
	lib/bin/protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		core/api/*.proto

build:
	mkdir -p .dev/bin
	export GOOS=windows &&\
		export GOARCH=amd64 &&\
		go build -o  .dev/bin/nlink-$${GOOS}-$${GOARCH}

	export GOOS=linux &&\
		export GOARCH=amd64 &&\
		go build -o  .dev/bin/nlink-$${GOOS}-$${GOARCH}

	export GOOS=darwin &&\
		export GOARCH=amd64 &&\
		go build -o  .dev/bin/nlink-$${GOOS}-$${GOARCH}

	export GOOS=darwin &&\
		export GOARCH=arm64 &&\
		go build -o  .dev/bin/nlink-$${GOOS}-$${GOARCH}