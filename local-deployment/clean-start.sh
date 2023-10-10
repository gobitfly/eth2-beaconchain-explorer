#! /bin/bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd $DIR
docker compose down
kurtosis clean -a
kurtosis run --enclave my-testnet . "$(cat network-params.json)"
bash provision-explorer-config.sh
cd $DIR/..
make all
cd $DIR
docker compose up -d
