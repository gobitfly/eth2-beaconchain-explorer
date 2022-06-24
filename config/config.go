package config

import _ "embed"

//go:embed default.config.yml
var DefaultConfigYml string

//go:embed mainnet.chain.yml
var MainnetChainYml string

//go:embed prater.chain.yml
var PraterChainYml string

//go:embed ropsten.chain.yml
var RopstenChainYml string

//go:embed sepolia.chain.yml
var SepoliaChainYml string
