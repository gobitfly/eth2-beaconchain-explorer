#! /bin/bash
set -e

clean_up () {
    ARG=$?
    rm -rf /tmp/full_withdrawal
    exit $ARG
} 
trap clean_up EXIT

# default values
bn_endpoint=$(kurtosis port print my-testnet cl-1-lighthouse-geth http)
mnemonic=$(eth2-val-tools mnemonic)

while getopts b:i:m: flag
do
    case "${flag}" in
        b) bn_endpoint=${OPTARG};;
        i) index=${OPTARG};;
        m) mnemonic=${OPTARG};;
    esac
done
echo "Validator Index: $index";
echo "Mnemonic: $mnemonic";
echo "BN Endpoint: $bn_endpoint";

mkdir -p /tmp/full_withdrawal
echo "creating wallet"
ethdo wallet create --base-dir=/tmp/full_withdrawal --type=hd --wallet=withdrawal-validators --mnemonic="$mnemonic" --wallet-passphrase="superSecure" --allow-weak-passphrases
echo "deriving account wallet"
ethdo account create --base-dir=/tmp/full_withdrawal --account=withdrawal-validators/$index --wallet-passphrase="superSecure" --passphrase="superSecure" --allow-weak-passphrases --path="m/12381/3600/$index/0/0"
echo "creating exit message"
ethdo validator exit --base-dir=/tmp/full_withdrawal --json --account=withdrawal-validators/$index --passphrase="superSecure" --connection=$bn_endpoint > /tmp/full_withdrawal/withdrawal-$index.json
echo "submitting exit message"
ethdo validator exit --signed-operations=$(cat /tmp/full_withdrawal/withdrawal-$index.json) --connection=$bn_endpoint
ethdo wallet delete --base-dir=/tmp/full_withdrawal --wallet=withdrawal-validators
rm -rf /tmp/full_withdrawal