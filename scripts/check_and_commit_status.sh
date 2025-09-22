#!/bin/bash
# check_and_commit_status.sh
# Checks ETH balance for a given address, logs the result, and commits the log to git

ADDRESS="0x06EE840642a33367ee59fCA237F270d5119d1356"
ETHERSCAN_API_KEY="YOUR_ETHERSCAN_API_KEY" # <-- Set your API key here
LOG_FILE="address_status.log"

# Get current timestamp
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

# Query Etherscan for balance
BALANCE_WEI=$(curl -s "https://api.etherscan.io/api?module=account&action=balance&address=$ADDRESS&tag=latest&apikey=$ETHERSCAN_API_KEY" | jq -r '.result')

# Convert to ETH
BALANCE_ETH=$(awk "BEGIN {print $BALANCE_WEI/1000000000000000000}")

# Log result
echo "$TIMESTAMP | Address: $ADDRESS | Balance: $BALANCE_ETH ETH" >> $LOG_FILE

git add $LOG_FILE
GIT_STATUS=$(git status --porcelain $LOG_FILE)
if [ -n "$GIT_STATUS" ]; then
  git commit -m "Update address status log: $TIMESTAMP"
  git push origin master
fi
