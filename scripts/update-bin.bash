#!/bin/bash

OS=("windows" "darwin" "linux")
ARCH=("amd64" "arm64")

# 使用for循环遍历数组
for GOOS in "${OS[@]}"; do
    for GOARCH in "${ARCH[@]}"; do
        binName=nlink-$GOOS-$GOARCH
        if [[ "$GOOS" == windows ]]; then binName=$binName.exe; fi
        out=build/bin/$binName
        GOOS=$GOOS GOARCH=$GOARCH go build -o $out cmd/main/main.go
        curl -X PUT -T $out http://hugohome.codenative.net:9000/public/nlink/$binName
        echo http://hugohome.codenative.net:9000/public/nlink/$binName
    done
done
