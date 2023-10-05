#! /bin/bash
docker start kurtosis-logs-aggregator
echo "Lauchning local network... this may take a while"
kurtosis run --enclave my-testnet local-deployment "$(cat local-deployment/network-params.json)"
bash local-deployment/provision-explorer-config.sh -c ../__gitignore/local.config.yml