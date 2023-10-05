#! /bin/bash
set -e

clean_up () {
    ARG=$?
    rm -rf /tmp/set_withdrawal_addr
    exit $ARG
} 
trap clean_up EXIT

# default values
bn_endpoint=$(kurtosis port print my-testnet cl-1-lighthouse-geth http)
mnemonic=$(eth2-val-tools mnemonic)

while getopts a:b:i:m: flag
do
    case "${flag}" in
        a) el_address=${OPTARG};;
        b) bn_endpoint=${OPTARG};;
        i) index=${OPTARG};;
        m) mnemonic=${OPTARG};;
    esac
done
echo "Validator Index: $index";
echo "Mnemonic: $mnemonic";
echo "BN Endpoint: $bn_endpoint";
echo "EL Withdrawal Address: $el_address";

mkdir -p /tmp/set_withdrawal_addr
echo "retrieving metadata"
genesis_validators_root=$(curl --silent $bn_endpoint/eth/v1/beacon/genesis | jq -r '.data.genesis_validators_root')
fork_version=$(curl --silent $bn_endpoint/eth/v1/beacon/genesis | jq -r '.data.genesis_fork_version')
deposit_contract_address=$(curl --silent $bn_endpoint/eth/v1/config/spec | jq -r '.data.DEPOSIT_CONTRACT_ADDRESS')
echo "generating bls-el change message"
eth2-val-tools bls-address-change --withdrawals-mnemonic="$mnemonic" --execution-address="$el_address" --source-min="$index" --source-max="$(($index + 1))" --genesis-validators-root="$genesis_validators_root" --fork-version="$fork_version" --as-json-list=true > "/tmp/set_withdrawal_addr/change_operations.json"
cat /tmp/set_withdrawal_addr/change_operations.json
echo "publishing bls-el change message"
curl -X POST $bn_endpoint/eth/v1/beacon/pool/bls_to_execution_changes \
               -H "Content-Type: application/json" \
               --data-binary "@/tmp/set_withdrawal_addr/change_operations.json"
rm -rf /tmp/set_withdrawal_addr