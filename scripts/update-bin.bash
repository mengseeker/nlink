#!/bin/bash

OS=("linux" "windows" "darwin")
ARCH=("amd64" "arm64")

# 使用for循环遍历数组
for GOOS in "${OS[@]}"; do
    for GOARCH in "${ARCH[@]}"; do
        binName=nlink-$GOOS-$GOARCH
        if [[ "$GOOS" == windows ]]; then binName=$binName.exe; fi
        out=build/bin/$binName
        CGO_ENABLED='0' GOOS=$GOOS GOARCH=$GOARCH go build -o $out cmd/main/main.go
        curl -X PUT -T $out http://hugohome.codenative.net:9000/public/nlink/$binName

        # binName=nlink-gui-$GOOS-$GOARCH
        # if [[ "$GOOS" == windows ]]; then binName=$binName.exe; fi
        # out=build/bin/$binName
        # CGO_ENABLED='1' GOOS=$GOOS GOARCH=$GOARCH wails build -debug -noPackage -o nlink-gui-$GOOS-$GOARCH
        # curl -X PUT -T $out http://hugohome.codenative.net:9000/public/nlink/$binName
    done
done
