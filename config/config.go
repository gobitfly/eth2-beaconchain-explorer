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

//go:embed testnet.chain.yml
var TestnetChainYml string

//go:embed gnosis.chain.yml
var GnosisChainYml string

//go:embed holesky.chain.yml
var HoleskyChainYml string

//go:embed hoodi.chain.yml
var HoodiChainYml string

//go:embed mekong.chain.yml
var MekongChainYml string

//go:embed pectra-devnet-5.chain.yml
var PectraDevnet5ChainYml string
