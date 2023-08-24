package config

import _ "embed"

//go:embed mainnet.preset.yml
var MainnetPresetYml string // https://github.com/ethereum/consensus-specs

//go:embed minimal.preset.yml
var MinimalPresetYml string // https://github.com/ethereum/consensus-specs

//go:embed gnosis.preset.yml
var GnosisPresetYml string // https://github.com/gnosischain/configs

//go:embed mainnet.chain.yml
var MainnetChainYml string // https://github.com/ethereum/consensus-specs

//go:embed mainnet.chain.yml
var MinimalChainYml string // https://github.com/ethereum/consensus-specs

//go:embed prater.chain.yml
var PraterChainYml string // https://github.com/eth-clients/goerli/blob/d6a227e/prater/config.yaml

//go:embed sepolia.chain.yml
var SepoliaChainYml string // https://github.com/eth-clients/sepolia/blob/main/bepolia/config.yaml

//go:embed gnosis.chain.yml
var GnosisChainYml string // https://github.com/gnosischain/configs/blob/main/mainnet/config.yaml

//go:embed dencun-devnet-8.chain.yml
var DencunDevnet8ChainYml string // https://github.com/ethpandaops/dencun-testnet/blob/83e4547/network-configs/devnet-8/config.yaml

//go:embed default.config.yml
var DefaultConfigYml string
