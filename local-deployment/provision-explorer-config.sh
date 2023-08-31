#! /bin/bash
CL_PORT=$(kurtosis enclave inspect my-testnet | grep 4000/tcp | tr -s ' ' | cut -d " " -f 6 | sed -e 's/http\:\/\/127.0.0.1\://' | grep "\S")
echo "CL Node port is $CL_PORT"

EL_PORT=$(kurtosis enclave inspect my-testnet | grep 8545/tcp | tr -s ' ' | cut -d " " -f 5 | sed -e 's/127.0.0.1\://' | grep "\S")
echo "EL Node port is $EL_PORT"

REDIS_PORT=$(kurtosis enclave inspect my-testnet | grep 6379/tcp | tr -s ' ' | cut -d " " -f 6 | sed -e 's/tcp\:\/\/127.0.0.1\://' | grep "\S")
echo "Redis port is $REDIS_PORT"

POSTGRES_PORT=$(kurtosis enclave inspect my-testnet | grep 5432/tcp | tr -s ' ' | cut -d " " -f 6 | sed -e 's/postgresql\:\/\/127.0.0.1\://' | grep "\S")
echo "Postgres port is $POSTGRES_PORT"

LBT_PORT=$(kurtosis enclave inspect my-testnet | grep 9000/tcp | tr -s ' ' | cut -d " " -f 6 | sed -e 's/tcp\:\/\/127.0.0.1\://' | grep "\S")
echo "Little bigtable port is $LBT_PORT"

cat >config.yml <<EOL
chain:
  configPath: 'node'
readerDatabase:
  name: db
  host: 127.0.0.1
  port: "$POSTGRES_PORT"
  user: postgres
  password: "pass"
writerDatabase:
  name: db
  host: 127.0.0.1
  port: "$POSTGRES_PORT"
  user: postgres
  password: "pass"
bigtable:
  project: explorer
  instance: explorer
  emulator: true
  emulatorPort: $LBT_PORT
eth1ErigonEndpoint: 'http://127.0.0.1:$EL_PORT'
eth1GethEndpoint: 'http://127.0.0.1:$EL_PORT'
redisCacheEndpoint: '127.0.0.1:$REDIS_PORT'
tieredCacheProvider: 'redis'
frontend:
  siteDomain: "local-testnet.beaconcha.in"
  siteName: 'Open Source Ethereum (ETH) Testnet Explorer' # Name of the site, displayed in the title tag
  siteSubtitle: "Showing a local testnet."
  server:
    host: '0.0.0.0' # Address to listen on
    port: '8080' # Port to listen on
  readerDatabase:
    name: db
    host: 127.0.0.1
    port: "$POSTGRES_PORT"
    user: postgres
    password: "pass"
  writerDatabase:
    name: db
    host: 127.0.0.1
    port: "$POSTGRES_PORT"
    user: postgres
    password: "pass"
  sessionSecret: "11111111111111111111111111111111"
  jwtSigningSecret: "1111111111111111111111111111111111111111111111111111111111111111"
  jwtIssuer: "localhost"
  jwtValidityInMinutes: 30
  maxMailsPerEmailPerDay: 10
  mail:
    mailgun:
      sender: no-reply@localhost
      domain: mg.localhost
      privateKey: "key-11111111111111111111111111111111"
  csrfAuthKey: '1111111111111111111111111111111111111111111111111111111111111111'
indexer:
  # fullIndexOnStartup: false # Perform a one time full db index on startup
  # indexMissingEpochsOnStartup: true # Check for missing epochs and export them after startup
  node:
    host: 127.0.0.1
    port: '$CL_PORT'
    type: lighthouse
  eth1DepositContractFirstBlock: 0
EOL

echo "generated config written to config.yml"

echo "initializing bigtable schema"
PROJECT="explorer"
INSTANCE="explorer"
HOST="127.0.0.1:$LBT_PORT"

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

echo "bigtable schema initialization completed"

echo "provisioning postgres db schema"
go run ../cmd/misc/main.go -config config.yml -command applyDbSchema
echo "postgres db schema initialization completed"