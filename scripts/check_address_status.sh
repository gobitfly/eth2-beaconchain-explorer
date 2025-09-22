#!/bin/bash
# Script to check ETH balance for a given address using Etherscan API
# Usage: ./check_address_status.sh 0xYourAddress

ADDRESS="$1"
API_KEY="YOUR_ETHERSCAN_API_KEY"

if [ -z "$ADDRESS" ]; then
  echo "Usage: $0 0xYourAddress"
  exit 1
fi

RESPONSE=$(curl -s "https://api.etherscan.io/api?module=account&action=balance&address=$ADDRESS&tag=latest&apikey=$API_KEY")
BALANCE_WEI=$(echo $RESPONSE | jq -r '.result')
BALANCE_ETH=$(echo "scale=18; $BALANCE_WEI / 1000000000000000000" | bc)

echo "Address: $ADDRESS"
echo "ETH Balance: $BALANCE_ETH"
