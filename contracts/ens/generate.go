package ens

//go:generate abigen -abi ens_registry.json -out ens_registry.go -pkg ens -type ENSRegistry
//go:generate abigen -abi ens_base_registrar.json -out ens_base_registrar.go -pkg ens -type ENSBaseRegistrar
//go:generate abigen -abi ens_eth_registrar_controller.json -out ens_eth_registrar_controller.go -pkg ens -type ENSETHRegistrarController
//go:generate abigen -abi ens_dns_registrar.json -out ens_dns_registrar.go -pkg ens -type ENSDNSRegistrar
//go:generate abigen -abi ens_reverse_registrar.json -out ens_reverse_registrar.go -pkg ens -type ENSReverseRegistrar
//go:generate abigen -abi ens_name_wrapper.json -out ens_name_wrapper.go -pkg ens -type ENSNameWrapper
//go:generate abigen -abi ens_public_resolver.json -out ens_public_resolver.go -pkg ens -type ENSPublicResolver
//go:generate abigen -abi ens_old_regstrar_controller.json -out ens_old_regstrar_controller.go -pkg ens -type ENSOldRegistrarController
