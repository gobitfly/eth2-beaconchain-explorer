# Eth2 Beacon Chain Explorer
The explorer provides a comprehensive and easy to use interface for the upcoming Eth2 beacon chain. It makes it easy to view proposed blocks, follow attestations and monitor your staking activity.

[![Badge](https://github.com/gobitfly/eth2-beaconchain-explorer/workflows/Publish%20Docker%20images/badge.svg)](https://github.com/gobitfly/eth2-beaconchain-explorer/actions?query=workflow%3A%22Build+%26+Publish+Docker+images%22)
[![Gitter](https://img.shields.io/gitter/room/gobitfly/eth2-beaconchain-explorer?color=%2334D058)](https://gitter.im/gobitfly/beaconchain-explorer?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/gobitfly/eth2-beaconchain-explorer)](https://goreportcard.com/report/github.com/gobitfly/eth2-beaconchain-explorer)
## About
The explorer is built using golang and utilizes a PostgreSQL database for storing and indexing data. In order to avoid the situation we currently have with the Eth1 chain where closed source block explorers dominate the market we decided to make our explorer open source and available for everybody. We run a production instance of the explorer at [beaconcha.in](https://beaconcha.in).

**Warning:** The explorer is still under heavy active development. More or less everything might change without prior notice and we cannot guarantee any backwards compatibility for now. Once the eth2 ecosystem matures we will be able to provide stronger grantees about the updatability of the explorer.

![Site](https://github.com/gobitfly/eth2-beaconchain-explorer/raw/master/static/img/site.png "Beacon Chain Web Interface Screenshot")

## Features
- Bootstrap based and mobile first web interface
- Fast and robust blockchain indexing engine, able to handle missed, duplicate & forked blocks
- Index page
  - Auto refresh - Index page data is automatically updated every 15 seconds
  - Basic chain statistics (current epoch, current slot, active validators, pending validators, staked ether)
  - Information on the 20 most recent blocks (epoch, slot, time, proposer, hash, number of attestations, deposits, slahsings and voluntary exits)
- Epochs page
  - Pageable tabular view of all epochs (epoch, time, blocks, attestations, slashings, exits, finalization status, voting statistics)
- Blocks page
  - Pageable tabular view of all blocks (epoch, time, proposer, hash, attestations, slashings, exits)
- Block page
  - Basic block info (epoch, slot, status, time, proposer, root hash, parent hash, state root hash, signature, randao reveal, graffiti, eth1 data)
  - List of attestations included in the block
  - List of deposits included in the block
  - List of LMD GHOST votes
- Validators page
  - Pageable tabular view of all pending, active and ejected validators
- Validator page
  - Basic validator info (index, current balance, current effective balance, status, slashed, active since, exited on)
  - Historic balance evolution chart
  - List of proposed and missed blocks
- Visualizations
  - Live visualization of blocks being added to the blockchain

## ToDo
- Add chain statistic charts
- Improve design, move away from stock bootstrap 4
- Use a proper open source charting library
- Come up with a smarter exporter logic (the current logic is stupid as it simply dumps the contents of the RPC calls into the database without doing any pre-aggregation or cleanups)

## Getting started
We currently do not provide any pre-built binaries of the explorer. Docker images are available at https://hub.docker.com/repository/docker/gobitfly/eth2-beaconchain-explorer.

- Download the latest version of the Prysm beacon chain client and start it with the `--archive` flag set
- Wait till the client finishes the initial sync
- Setup a PostgreSQL DB and import the `tables.sql` file from the root of this repository
- Install go version 1.13 or higher
- Clone the repository and run `make all` to build the indexer and front-end binaries
- Copy the config-example.yml file an adapt it to your environment
- Start the explorer binary and pass the path to the config file as argument

## Development
Install golint. (see https://github.com/golang/lint)

## Commercial usage
The explorer uses Highsoft charts which are not free for commercial and governmental use. If you plan to use the explorer for commercial purposes you currently need to purchase an appropriate HighSoft license.
We are planning to switch out the Highsoft chart library with a less restrictive charting library (suggestions are welcome).
