package db

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"eth2-exporter/ens"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"log"
	"strings"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	go_ens "github.com/wealdtech/go-ens/v3"
)

// https://etherscan.io/tx/0x9fec76750a504e5610643d1882e3b07f4fc786acf7b9e6680697bb7165de1165#eventlog
// TransformEnsNameRegistered accepts an eth1 block and creates bigtable mutations for ENS Name events.
// It transforms the logs contained within a block and indexes ens relevant transactions and tags changes (to be verified from the node in a separate process)
// ==================================================
//
// It indexes transactions
//
// - by hashed ens name
// Row:    <chainID>:ENS:I:H<nameHash><txHash>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:I:H:nameHash:txHash"
//
// - by address
// Row:    <chainID>:ENS:I:A<address><txHash>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:I:A:address:txHash"
//
// ==================================================
//
// Track for later verification via the node ("set dirty")
//
// - by name
// Row:    <chainID>:ENS:V:N<name>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:V:N:name"
//
// - by name hash
// Row:    <chainID>:ENS:V:N<nameHash>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:V:N:nameHash"
//
// - by address
// Row:    <chainID>:ENS:V:A<address>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:I:A:address"
//
// ==================================================

func (bigtable *Bigtable) TransformEnsNameRegistered(blk *types.Eth1Block, cache *freecache.Cache) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error) {
	bulkData = &types.BulkMutations{}
	bulkMetadataUpdates = &types.BulkMutations{}

	filterer, err := ens.NewEnsRegistrarFilterer(common.Address{}, nil)
	if err != nil {
		log.Printf("error creating filterer: %v", err)
		return nil, nil, err
	}
	keys := make(map[string]bool)

	for i, tx := range blk.GetTransactions() {
		if i > 9999 {
			return nil, nil, fmt.Errorf("unexpected number of transactions in block expected at most 9999 but got: %v, tx: %x", i, tx.GetHash())
		}

		// We look for the different ENS events,
		// 	most will be triggered by a main registrar contract,
		//  but some are triggered on a different contracts (like a resolver contract), these will be validated when loading the related events
		var isRegistarContract = len(utils.Config.Indexer.EnsTransformer.ValidRegistrarContracts) > 0 && utils.SliceContains(utils.Config.Indexer.EnsTransformer.ValidRegistrarContracts, common.BytesToAddress(tx.To).String())
		foundNameIndex := -1
		foundResolverIndex := -1
		foundNameRenewedIndex := -1
		foundAddressChangedIndices := []int{}
		foundNameChangedIndex := -1
		foundNewOwnerIndex := -1
		logs := tx.GetLogs()
		for j, log := range logs {
			if j > 99999 {
				return nil, nil, fmt.Errorf("unexpected number of logs in block expected at most 99999 but got: %v tx: %x", j, tx.GetHash())
			}
			for _, lTopic := range log.GetTopics() {
				if isRegistarContract {
					if bytes.Equal(lTopic, ens.NameRegisteredTopic) {
						foundNameIndex = j
					} else if bytes.Equal(lTopic, ens.NewResolverTopic) {
						foundResolverIndex = j
					} else if bytes.Equal(lTopic, ens.NameRenewedTopic) {
						foundNameRenewedIndex = j
					}
				} else if bytes.Equal(lTopic, ens.AddressChangedTopic) {
					foundAddressChangedIndices = append(foundAddressChangedIndices, j)
				} else if bytes.Equal(lTopic, ens.NameChangedTopic) {
					foundNameChangedIndex = j
				} else if bytes.Equal(lTopic, ens.NewOwnerTopic) {
					foundNewOwnerIndex = j
				}
			}
		}
		// We found a register name event
		if foundNameIndex > -1 && foundResolverIndex > -1 {

			log := logs[foundNameIndex]
			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			nameLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(foundNameIndex),
				Removed:     log.GetRemoved(),
			}

			log = logs[foundResolverIndex]
			topics = make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			resolverLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(foundResolverIndex),
				Removed:     log.GetRemoved(),
			}

			nameRegistered, err := filterer.ParseNameRegistered(nameLog)
			if err != nil {
				utils.LogError(err, "indexing of register event failed parse register event", 0)
				continue
			}
			resolver, err := filterer.ParseNewResolver(resolverLog)
			if err != nil {
				utils.LogError(err, "indexing of register event failed parse resolver event", 0)
				continue
			}

			keys[fmt.Sprintf("%s:ENS:I:H:%x:%x", bigtable.chainId, resolver.Node, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:I:A:%x:%x", bigtable.chainId, nameRegistered.Owner, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:V:A:%x", bigtable.chainId, nameRegistered.Owner)] = true
			keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, nameRegistered.Name)] = true

		} else if foundNameRenewedIndex > -1 { // We found a renew name event
			log := logs[foundNameRenewedIndex]
			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			nameRenewedLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(foundNameRenewedIndex),
				Removed:     log.GetRemoved(),
			}

			nameRenewed, err := filterer.ParseNameRenewed(nameRenewedLog)
			if err != nil {
				utils.LogError(err, "indexing of renew event failed parse event", 0)
				continue
			}

			nameHash, err := go_ens.NameHash(nameRenewed.Name)
			if err != nil {
				utils.LogError(err, "error hashing ens name", 0)
				continue
			}
			keys[fmt.Sprintf("%s:ENS:I:H:%x:%x", bigtable.chainId, nameHash, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, nameRenewed.Name)] = true

		} else if foundNameChangedIndex > -1 && foundNewOwnerIndex > -1 { // we found a name change event

			log := logs[foundNewOwnerIndex]
			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}
			newOwnerLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(foundNewOwnerIndex),
				Removed:     log.GetRemoved(),
			}

			newOwner, err := filterer.ParseNewOwner(newOwnerLog)
			if err != nil {
				utils.LogError(err, fmt.Errorf("indexing of new owner event failed parse event at index %v", foundNewOwnerIndex), 0)
				continue
			}

			keys[fmt.Sprintf("%s:ENS:I:A:%x:%x", bigtable.chainId, newOwner.Owner, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:V:A:%x", bigtable.chainId, newOwner.Owner)] = true
		}
		// We found a change address event, there can be multiple within one transaction
		for _, addressChangeIndex := range foundAddressChangedIndices {

			log := logs[addressChangeIndex]
			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			addressChangedLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(addressChangeIndex),
				Removed:     log.GetRemoved(),
			}

			addressChanged, err := filterer.ParseAddressChanged(addressChangedLog)
			if err != nil {
				utils.LogError(err, "indexing of address change event failed parse event at index ", 0)
				continue
			}

			keys[fmt.Sprintf("%s:ENS:I:H:%x:%x", bigtable.chainId, addressChanged.Node, tx.GetHash())] = true
			keys[fmt.Sprintf("%s:ENS:V:H:%x", bigtable.chainId, addressChanged.Node)] = true

		}
	}
	for key, _ := range keys {
		mut := gcp_bigtable.NewMutation()
		mut.Set(DEFAULT_FAMILY, key, gcp_bigtable.Timestamp(0), nil)

		bulkData.Keys = append(bulkData.Keys, key)
		bulkData.Muts = append(bulkData.Muts, mut)
	}

	return bulkData, bulkMetadataUpdates, nil
}

func (bigtable *Bigtable) ImportEnsUpdates(client *ethclient.Client) error {
	key := fmt.Sprintf("%s:ENS:V", bigtable.chainId)

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	rowRange := gcp_bigtable.NewRange(key+"\x00", prefixSuccessor(key, 5))
	keys := []string{}

	err := bigtable.tableData.ReadRows(ctx, rowRange, func(row gcp_bigtable.Row) bool {
		row_ := row[DEFAULT_FAMILY][0]
		keys = append(keys, row_.Row)
		return true
	})
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		logger.Infof("Validating %v ENS entries", len(keys))

		for _, key := range keys {
			var name string
			var address *common.Address
			split := strings.Split(key, ":")
			value := split[4]
			switch split[3] {
			case "H":
				// if we have a hash we look if we find a name in the db. If not we can ignore it.
				nameHash, err := hex.DecodeString(value)
				if err != nil {
					utils.LogError(err, fmt.Errorf("name hash could not be decoded: %v", value), 0)
				} else {
					err := ReaderDb.Get(&name, `
					SELECT
						name
					FROM ens_names
					WHERE name_hash = $1
					`, nameHash)
					if err != nil && err != sql.ErrNoRows {
						return err
					}
				}
			case "A":
				addressHash, err := hex.DecodeString(value)
				if err != nil {
					utils.LogError(err, fmt.Errorf("address hash could not be decoded: %v", value), 0)
				} else {
					add := common.BytesToAddress(addressHash)
					address = &add
				}
			case "N":
				name = value
			}

			if name != "" {
				err := validateEnsName(client, name)
				if err != nil {
					return err
				}
			} else if address != nil {
				err := validateEnsAddress(client, *address)
				if err != nil {
					return err
				}
			}

			// After processing the key we remove it from bigtable
			mutsDelete := &types.BulkMutations{
				Keys: make([]string, 0, 1),
				Muts: make([]*gcp_bigtable.Mutation, 0, 1),
			}
			mutDelete := gcp_bigtable.NewMutation()
			mutDelete.DeleteRow()
			mutsDelete.Keys = append(mutsDelete.Keys, key)
			mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
			err = bigtable.WriteBulk(mutsDelete, bigtable.tableData)

			if err != nil {
				return err
			}
		}
	} else {
		logger.Info("No ENS entries to validate")
	}

	return nil
}

func validateEnsAddress(client *ethclient.Client, address common.Address) error {

	name, err := go_ens.ReverseResolve(client, address)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error resolving address: %v", name), 0)
		return removeEnsAddress(client, address)
	}
	nameHash, err := go_ens.NameHash(name)
	if err != nil {
		utils.LogError(err, fmt.Errorf("could not hash name [%v] for address: %x", name, address), 0)
		return nil
	}
	_, err = WriterDb.Exec(`
	INSERT INTO ens_addresses (
		address, 
		name_hash)
	VALUES ($1, $2) 
	ON CONFLICT 
		(address) 
	DO UPDATE SET 
		name_hash = excluded.name_hash
	`, address.Bytes(), nameHash[:])
	if err != nil {
		utils.LogError(err, fmt.Errorf("error writing ens data for address [%x]", address), 0)
		return err
	}
	logger.Infof("Address [%x] resolved -> %v", address, name)
	return nil
}

func validateEnsName(client *ethclient.Client, name string) error {
	if name[:4] != ".eth" {
		name = fmt.Sprintf("%s.eth", name)
	}
	nameHash, err := go_ens.NameHash(name)
	if err != nil {
		utils.LogError(err, fmt.Errorf("could not hash name: %v", name), 0)
		return nil
	}
	addr, err := go_ens.Resolve(client, name)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error resolving name: %v", name), 0)
		return removeEnsName(client, name)
	}
	ensName, err := go_ens.NewName(client, name)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error getting create ens name: %v", name), 0)
		return removeEnsName(client, name)
	}
	expires, err := ensName.Expires()
	if err != nil {
		utils.LogError(err, fmt.Errorf("error get ens expire date: %v", name), 0)
		return removeEnsName(client, name)
	}
	_, err = WriterDb.Exec(`
	INSERT INTO ens_names (
		name_hash, 
		name, 
		address, 
		valid_to)
	VALUES ($1, $2, $3, $4) 
	ON CONFLICT 
		(name_hash) 
	DO UPDATE SET 
		name = excluded.name,
		address = excluded.address,
		valid_to = excluded.valid_to
	`, nameHash[:], name, addr.Bytes(), expires)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error writing ens data for name [%v]", name), 0)
		return err
	}
	logger.Infof("Name [%v] resolved -> %x, expires: %v", name, addr, expires)
	return nil
}

func removeEnsAddress(client *ethclient.Client, address common.Address) error {
	_, err := WriterDb.Exec(`
	DELETE FROM ens_addresses 
	WHERE 
		address = $1;
	`, address.Bytes())
	if err != nil {
		utils.LogError(err, fmt.Errorf("error deleting ens name [%x]", address), 0)
		return err
	}
	logger.Infof("Ens name remove from db: %x", address)
	return nil
}

func removeEnsName(client *ethclient.Client, name string) error {
	_, err := WriterDb.Exec(`
	DELETE FROM ens_names 
	WHERE 
		name = $1;
	`, name)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error deleting ens name [%v]", name), 0)
		return err
	}
	logger.Infof("Ens name remove from db: %v", name)
	return nil
}

func GetAddressForEnsName(name string) (address *common.Address, err error) {
	addressBytes := []byte{}
	err = ReaderDb.Get(&addressBytes, `
	SELECT address 
	FROM ens_names
	WHERE
		name = $1;
	`, name)
	if err == nil && addressBytes != nil {
		add := common.BytesToAddress(addressBytes)
		address = &add
	}
	return address, err
}

func GetEnsNameForAddress(address common.Address) (name *string, err error) {
	err = ReaderDb.Get(&name, `
	SELECT n.name 
	FROM ens_addresses a
	INNER JOIN ens_names n
		on a.name_hash = n.name_hash
	WHERE
		a.address = $1;
	`, address.Bytes())
	return name, err
}
