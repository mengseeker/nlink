.PHONY: all build build-image push-image

VERSION ?= latest
IMAGE ?= mengseeker/nlink:${VERSION}

UPLOAD_DIR=http://hugohome.codenative.net:9000/public/nlink

all: build build-image push-image push

build:
	go build -o build/nlink main.go
	GOOS=linux GOARCH=amd64 go build -o build/nlink-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o build/nlink-darwin-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o build/nlink-windows-amd64.exe main.go
	GOOS=linux GOARCH=arm64 go build -o build/nlink-linux-arm64 main.go
	GOOS=windows GOARCH=arm64 go build -o build/nlink-windows-arm64.exe main.go
	GOOS=darwin GOARCH=arm64 go build -o build/nlink-darwin-arm64 main.go

push:
	curl -X PUT -T build/nlink-linux-amd64 ${UPLOAD_DIR}/nlink-linux-amd64
	curl -X PUT -T build/nlink-darwin-amd64 ${UPLOAD_DIR}/nlink-darwin-amd64
	curl -X PUT -T build/nlink-windows-amd64.exe ${UPLOAD_DIR}/nlink-windows-amd64.exe
	curl -X PUT -T build/nlink-linux-arm64 ${UPLOAD_DIR}/nlink-linux-arm64
	curl -X PUT -T build/nlink-windows-arm64.exe ${UPLOAD_DIR}/nlink-windows-arm64.exe
	curl -X PUT -T build/nlink-darwin-arm64 ${UPLOAD_DIR}/nlink-darwin-arm64

build-image:
	docker build -t ${IMAGE} .

push-image:
	docker push ${IMAGE}