#!/bin/bash    

PROJECT="explorer"
INSTANCE="explorer"
HOST="127.0.0.1:9000"

BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable beaconchain_validator_balances
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily beaconchain_validator_balances vb
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable beaconchain_validator_attestations
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily beaconchain_validator_attestations at
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable beaconchain_validator_proposals
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily beaconchain_validator_proposals pr
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable beaconchain_validator_sync
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily beaconchain_validator_sync sc
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable beaconchain_validator_income
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily beaconchain_validator_income id
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily beaconchain_validator_income stats
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable beaconchain_validators
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily beaconchain_validators at

BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE setgcpolicy beaconchain_validators at maxversions=1
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable blocks
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily blocks default

BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE setgcpolicy blocks default maxversions=1
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable cache
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily cache 10_min
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily cache 1_day
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily cache 1_hour

BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE setgcpolicy cache 10_min maxage=10m and maxversions=1
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE setgcpolicy cache 1_day maxage=1d and maxversions=1
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE setgcpolicy cache 1_hour maxage=1h and maxversions=1
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable data
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily data c
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily data f

BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE setgcpolicy data c maxage=1d
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable machine_metrics
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily machine_metrics mm

BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE setgcpolicy machine_metrics mm maxage=31d
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable metadata
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily metadata a
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily metadata c
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily metadata erc1155
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily metadata erc20
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily metadata erc721
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily metadata series

BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE setgcpolicy metadata series maxversions=1
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createtable metadata_updates
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily metadata_updates blocks
BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE createfamily metadata_updates f

BIGTABLE_EMULATOR_HOST=$HOST cbt --project $PROJECT --instance $INSTANCE setgcpolicy metadata_updates blocks maxage=1d