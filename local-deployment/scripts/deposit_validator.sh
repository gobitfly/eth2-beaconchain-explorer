#! /bin/bash
set -e

clean_up () {
    ARG=$?
    rm -rf /tmp/deposit
    exit $ARG
} 
trap clean_up EXIT

# default values
bn_endpoint=$(kurtosis port print my-testnet cl-1-lighthouse-geth http)
el_endpoint="http://$(kurtosis port print my-testnet el-1-geth-lighthouse rpc)"
mnemonic=$(eth2-val-tools mnemonic)
count=1

while getopts a:b:e:i:m: flag
do
    case "${flag}" in
        b) bn_endpoint=${OPTARG};;
        e) el_endpoint=${OPTARG};;
        i) index=${OPTARG};;
        m) mnemonic=${OPTARG};;
        c) count=${OPTARG};;
    esac
done

echo "Validator Index: $index";
echo "Mnemonic: $mnemonic";
echo "BN Endpoint: $bn_endpoint";
echo "EL Endpoint: $el_endpoint";
echo "Deposit count: $count";

mkdir -p /tmp/deposit
deposit_path="m/44'/60'/0'/0/3"
privatekey="ef5177cd0b6b21c87db5a0bf35d4084a8a57a9d6a064f86d51ac85f2b873a4e2"
publickey="0x878705ba3f8Bc32FCf7F4CAa1A35E72AF65CF766"
fork_version=$(curl -s $bn_endpoint/eth/v1/beacon/genesis | jq -r '.data.genesis_fork_version')
deposit_contract_address=$(curl -s $bn_endpoint/eth/v1/config/spec | jq -r '.data.DEPOSIT_CONTRACT_ADDRESS')
eth2-val-tools deposit-data --source-min=192 --source-max=$((192 + count)) --amount=32000000000 --fork-version=$fork_version --withdrawals-mnemonic="$mnemonic" --validators-mnemonic="$mnemonic" > /tmp/deposit/deposits_0-9.txt
while read x; do
    account_name="$(echo "$x" | jq '.account')"
    pubkey="$(echo "$x" | jq '.pubkey')"
    echo "Sending deposit for validator $account_name $pubkey"
    ethereal beacon deposit \
        --allow-unknown-contract=true \
        --address="$deposit_contract_address" \
        --connection=$el_endpoint \
        --data="$x" \
        --value="32000000000" \
        --from="$publickey" \
        --privatekey="$privatekey"
    echo "Sent deposit for validator $account_name $pubkey"
    sleep 3
done < /tmp/deposit/deposits_0-9.txt
exit;
rm -rf /tmp/deposit