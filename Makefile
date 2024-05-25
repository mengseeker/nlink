.PHONY: build build-image push-image

VERSION ?= v0.0.0
IMAGE ?= mengseeker/nlink:${VERSION}

UPLOAD_DIR=http://hugohome.codenative.net:9000/public/nlink

build:
	go build -o build/nlink main.go
	GOOS=linux GOARCH=amd64 go build -o build/nlink_linux_amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o build/nlink_darwin_amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o build/nlink_windows_amd64.exe main.go
	GOOS=linux GOARCH=arm64 go build -o build/nlink_linux_arm64 main.go
	GOOS=windows GOARCH=arm64 go build -o build/nlink_windows_arm64.exe main.go
	GOOS=darwin GOARCH=arm64 go build -o build/nlink_darwin_arm64 main.go

push:
	curl -X PUT -T build/nlink_linux_amd64 ${UPLOAD_DIR}/nlink_linux_amd64
	curl -X PUT -T build/nlink_darwin_amd64 ${UPLOAD_DIR}/nlink_darwin_amd64
	curl -X PUT -T build/nlink_windows_amd64.exe ${UPLOAD_DIR}/nlink_windows_amd64.exe
	curl -X PUT -T build/nlink_linux_arm64 ${UPLOAD_DIR}/nlink_linux_arm64
	curl -X PUT -T build/nlink_windows_arm64.exe ${UPLOAD_DIR}/nlink_windows_arm64.exe
	curl -X PUT -T build/nlink_darwin_arm64 ${UPLOAD_DIR}/nlink_darwin_arm64

build-image:
	docker build -t ${IMAGE} .

push-image:
	docker push ${IMAGE}