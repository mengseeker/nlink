#!/bin/bash

set -ex

cd .dev/tls
for file in *; do
    if [ -f "$file" ]; then
        # echo "http://hugohome.codenative.net:9000/public/nlink/config/$file"
        # 处理文件
        curl -X PUT -T $file http://hugohome.codenative.net:9000/public/nlink/config/$file
    fi
done
