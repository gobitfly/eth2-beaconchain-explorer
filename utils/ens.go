package utils

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/wealdtech/go-ens/v3"
)

func IsValidEnsDomain(text string) bool {
	return strings.HasSuffix(text, ".eth")
}

func ResolveEnsDomain(query string) (string, error) {
	if !IsValidEnsDomain(query) {
		return "", errors.New("not an ens domain")
	}

	// NOTE: could be abused to spam execution node?
	client, err := ethclient.Dial(Config.Indexer.Eth1Endpoint)

	if err != nil {
		logger.Warnf("failed to create ethclient for ens resolve request: %v", err)
		return "", err
	}

	res, err := ens.Resolve(client, query)
	address := res.Hex()

	if err != nil {
		logger.Debugf("failed to resolve ens (%v => %v): %v", query, address, err)
		return "", err
	}
	return address, nil
}
