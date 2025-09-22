#!/bin/bash
# Cron job to check ETH balance for your new address every 10 minutes

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ADDRESS="0x06EE840642a33367ee59fCA237F270d5119d1356"

# Call the check_address_status.sh script
bash "$SCRIPT_DIR/check_address_status.sh" "$ADDRESS"
