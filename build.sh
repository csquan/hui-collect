#!/bin/bash

set -eu

tag="${1}"
echo "tag is ${tag}"

docker build --pull --force-rm --no-cache -t "reg.huiwang.io/fat/hui-collect:${tag}" .
digest=$(docker push "reg.huiwang.io/fat/hui-collect:${tag}" | awk '/digest/{print $3}')
cosign sign --key ~/cosign.key "reg.huiwang.io/fat/hui-collect@${digest}"
