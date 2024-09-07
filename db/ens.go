package db

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	ensContracts "github.com/gobitfly/eth2-beaconchain-explorer/contracts/ens"
	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"golang.org/x/sync/errgroup"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	eth_types "github.com/ethereum/go-ethereum/core/types"
	go_ens "github.com/wealdtech/go-ens/v3"
)

// TransformEnsNameRegistered accepts an eth1 block and creates bigtable mutations for ENS Name events.
// It transforms the logs contained within a block and indexes ens relevant transactions and tags changes (to be verified from the node in a separate process)
// ==================================================
//
// It indexes transactions
//
// - by hashed ens name
// Row:    <chainID>:ENS:I:H:<nameHash>:<txHash>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:I:H:4ae569dd0aa2f6e9207e41423c956d0d27cbc376a499ee8d90fe1d84489ae9d1:e627ae94bd16eb1ed8774cd4003fc25625159f13f8a2612cc1c7f8d2ab11b1d7"
//
// - by address
// Row:    <chainID>:ENS:I:A:<address>:<txHash>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:I:A:05579fadcf7cc6544f7aa018a2726c85251600c5:e627ae94bd16eb1ed8774cd4003fc25625159f13f8a2612cc1c7f8d2ab11b1d7"
//
// ==================================================
//
// Track for later verification via the node ("set dirty")
//
// - by name
// Row:    <chainID>:ENS:V:N:<name>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:V:N:somename"
//
// - by name hash
// Row:    <chainID>:ENS:V:H:<nameHash>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:V:H:6f5d9cc23e60abe836401b4fd386ec9280a1f671d47d9bf3ec75dab76380d845"
//
// - by address
// Row:    <chainID>:ENS:V:A:<address>
// Family: f
// Column: nil
// Cell:   nil
// Example scan: "5:ENS:V:A:27234cb8734d5b1fac0521c6f5dc5aebc6e839b6"
//
// ==================================================

func (bigtable *Bigtable) TransformEnsNameRegistered(blk *types.Eth1Block, cache *freecache.Cache) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error) {
	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("bt_transform_ens").Observe(time.Since(startTime).Seconds())
	}()

	var ensCrontractAddresses map[string]string
	switch bigtable.chainId {
	case "1":
		ensCrontractAddresses = ensContracts.ENSCrontractAddressesEthereum
	case "17000":
		ensCrontractAddresses = ensContracts.ENSCrontractAddressesHolesky
	case "11155111":
		ensCrontractAddresses = ensContracts.ENSCrontractAddressesSepolia
	default:
		return nil, nil, nil
	}

	bulkData = &types.BulkMutations{}
	bulkMetadataUpdates = &types.BulkMutations{}
	keys := make(map[string]bool)
	ethLog := eth_types.Log{}

	for i, tx := range blk.GetTransactions() {
		if i >= TX_PER_BLOCK_LIMIT {
			return nil, nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, tx: %x", TX_PER_BLOCK_LIMIT-1, i, tx.GetHash())
		}
		for j, log := range tx.GetLogs() {
			if j >= ITX_PER_TX_LIMIT {
				return nil, nil, fmt.Errorf("unexpected number of logs in block expected at most %d but got: %v tx: %x", ITX_PER_TX_LIMIT-1, j, tx.GetHash())
			}
			ensContract := ensCrontractAddresses[common.BytesToAddress(log.Address).String()]

			topics := log.GetTopics()
			ethTopics := make([]common.Hash, 0, len(topics))
			for _, t := range topics {
				ethTopics = append(ethTopics, common.BytesToHash(t))
			}

			ethLog.Address = common.BytesToAddress(log.GetAddress())
			ethLog.Data = log.Data
			ethLog.Topics = ethTopics
			ethLog.BlockNumber = blk.GetNumber()
			ethLog.TxHash = common.BytesToHash(tx.GetHash())
			ethLog.TxIndex = uint(i)
			ethLog.BlockHash = common.BytesToHash(blk.GetHash())
			ethLog.Index = uint(j)
			ethLog.Removed = log.GetRemoved()

			for _, lTopic := range topics {
				logFields := map[string]interface{}{
					"block":       blk.GetNumber(),
					"tx":          tx.GetHash(),
					"logIndex":    j,
					"ensContract": ensContract,
				}

				if ensContract == "Registry" {
					if bytes.Equal(lTopic, ensContracts.ENSRegistryParsedABI.Events["NewResolver"].ID.Bytes()) {
						logFields["event"] = "NewResolver"
						r := &ensContracts.ENSRegistryNewResolver{}
						err = ensContracts.ENSRegistryContract.UnpackLog(r, "NewResolver", ethLog)
						if err != nil {
							utils.LogWarn(err, "error unpacking ens-log", 0, logFields)
							continue
						}
						keys[fmt.Sprintf("%s:ENS:V:H:%x", bigtable.chainId, r.Node)] = true
					} else if bytes.Equal(lTopic, ensContracts.ENSRegistryParsedABI.Events["NewOwner"].ID.Bytes()) {
						logFields["event"] = "NewOwner"
						r := &ensContracts.ENSRegistryNewOwner{}
						err = ensContracts.ENSRegistryContract.UnpackLog(r, "NewOwner", ethLog)
						if err != nil {
							utils.LogWarn(err, "error unpacking ens-log", 0, logFields)
							continue
						}
						keys[fmt.Sprintf("%s:ENS:V:A:%x", bigtable.chainId, r.Owner)] = true
					} else if bytes.Equal(lTopic, ensContracts.ENSRegistryParsedABI.Events["NewTTL"].ID.Bytes()) {
						logFields["event"] = "NewTTL"
						r := &ensContracts.ENSRegistryNewTTL{}
						err = ensContracts.ENSRegistryContract.UnpackLog(r, "NewTTL", ethLog)
						if err != nil {
							utils.LogWarn(err, "error unpacking ens-log", 0, logFields)
							continue
						}
						keys[fmt.Sprintf("%s:ENS:V:H:%x", bigtable.chainId, r.Node)] = true
					}
				} else if ensContract == "ETHRegistrarController" {
					if bytes.Equal(lTopic, ensContracts.ENSETHRegistrarControllerParsedABI.Events["NameRegistered"].ID.Bytes()) {
						logFields["event"] = "NameRegistered"
						r := &ensContracts.ENSETHRegistrarControllerNameRegistered{}
						err = ensContracts.ENSETHRegistrarControllerContract.UnpackLog(r, "NameRegistered", ethLog)
						if err != nil {
							utils.LogWarn(err, "error unpacking ens-log", 0, logFields)
							continue
						}
						if err = verifyName(r.Name); err != nil {
							utils.LogWarn(err, "error verifying ens-name", 0, logFields)
							continue
						}
						keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, r.Name)] = true
						keys[fmt.Sprintf("%s:ENS:V:A:%x", bigtable.chainId, r.Owner)] = true
					} else if bytes.Equal(lTopic, ensContracts.ENSETHRegistrarControllerParsedABI.Events["NameRenewed"].ID.Bytes()) {
						logFields["event"] = "NameRenewed"
						r := &ensContracts.ENSETHRegistrarControllerNameRenewed{}
						err = ensContracts.ENSETHRegistrarControllerContract.UnpackLog(r, "NameRenewed", ethLog)
						if err != nil {
							utils.LogWarn(err, "error unpacking ens-log", 0, logFields)
							continue
						}
						if err = verifyName(r.Name); err != nil {
							utils.LogWarn(err, "error verifying ens-name", 0, logFields)
							continue
						}
						keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, r.Name)] = true
					}
				} else if ensContract == "OldEnsRegistrarController" {
					if bytes.Equal(lTopic, ensContracts.ENSOldRegistrarControllerParsedABI.Events["NameRegistered"].ID.Bytes()) {
						logFields["event"] = "NameRegistered"
						r := &ensContracts.ENSOldRegistrarControllerNameRegistered{}
						err = ensContracts.ENSOldRegistrarControllerContract.UnpackLog(r, "NameRegistered", ethLog)
						if err != nil {
							utils.LogWarn(err, "error unpacking ens-log", 0, logFields)
							continue
						}
						if err = verifyName(r.Name); err != nil {
							utils.LogWarn(err, "error verifying ens-name", 0, logFields)
							continue
						}
						keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, r.Name)] = true
						keys[fmt.Sprintf("%s:ENS:V:A:%x", bigtable.chainId, r.Owner)] = true
					} else if bytes.Equal(lTopic, ensContracts.ENSOldRegistrarControllerParsedABI.Events["NameRenewed"].ID.Bytes()) {
						logFields["event"] = "NameRenewed"
						r := &ensContracts.ENSOldRegistrarControllerNameRenewed{}
						err = ensContracts.ENSOldRegistrarControllerContract.UnpackLog(r, "NameRenewed", ethLog)
						if err != nil {
							utils.LogWarn(err, "error unpacking ens-log", 0, logFields)
							continue
						}
						if err = verifyName(r.Name); err != nil {
							utils.LogWarn(err, "error verifying ens-name", 0, logFields)
							continue
						}
						keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, r.Name)] = true
					}
				} else {
					if bytes.Equal(lTopic, ensContracts.ENSPublicResolverParsedABI.Events["NameChanged"].ID.Bytes()) {
						logFields["event"] = "NameChanged"
						r := &ensContracts.ENSPublicResolverNameChanged{}
						err = ensContracts.ENSPublicResolverContract.UnpackLog(r, "NameChanged", ethLog)
						if err != nil {
							utils.LogWarn(err, "error unpacking ens-log", 0, logFields)
							continue
						}
						if err = verifyName(r.Name); err != nil {
							utils.LogWarn(err, "error verifying ens-name", 0, logFields)
							continue
						}
						keys[fmt.Sprintf("%s:ENS:V:N:%s", bigtable.chainId, r.Name)] = true
					} else if bytes.Equal(lTopic, ensContracts.ENSPublicResolverParsedABI.Events["AddressChanged"].ID.Bytes()) {
						logFields["event"] = "AddressChanged"
						r := &ensContracts.ENSPublicResolverAddressChanged{}
						err = ensContracts.ENSPublicResolverContract.UnpackLog(r, "AddressChanged", ethLog)
						if err != nil {
							utils.LogWarn(err, "error unpacking ens-log", 0, logFields)
							continue
						}
						keys[fmt.Sprintf("%s:ENS:V:H:%x", bigtable.chainId, r.Node)] = true
					}
				}
			}
		}
	}
	for key := range keys {
		mut := gcp_bigtable.NewMutation()
		mut.Set(DEFAULT_FAMILY, key, gcp_bigtable.Timestamp(0), nil)

		bulkData.Keys = append(bulkData.Keys, key)
		bulkData.Muts = append(bulkData.Muts, mut)
	}

	return bulkData, bulkMetadataUpdates, nil
}

func verifyName(name string) error {
	// limited by max capacity of db (caused by btrees of indexes); tests showed maximum of 2684 (added buffer)
	if len(name) > 2048 {
		return fmt.Errorf("name too long: %v", name)
	}
	return nil
}

type EnsCheckedDictionary struct {
	mux     sync.Mutex
	address map[common.Address]bool
	name    map[string]bool
}

func (bigtable *Bigtable) GetRowsByPrefix(prefix string) ([]string, error) {
	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	rowRange := gcp_bigtable.PrefixRange(prefix)
	keys := []string{}

	err := bigtable.tableData.ReadRows(ctx, rowRange, func(row gcp_bigtable.Row) bool {
		row_ := row[DEFAULT_FAMILY][0]
		keys = append(keys, row_.Row)
		return true
	}, gcp_bigtable.LimitRows(1000))
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (bigtable *Bigtable) ImportEnsUpdates(client *ethclient.Client, readBatchSize int64) error {
	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("bt_import_ens_updates").Observe(time.Since(startTime).Seconds())
	}()

	key := fmt.Sprintf("%s:ENS:V", bigtable.chainId)

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	rowRange := gcp_bigtable.PrefixRange(key)
	keys := []string{}

	err := bigtable.tableData.ReadRows(ctx, rowRange, func(row gcp_bigtable.Row) bool {
		row_ := row[DEFAULT_FAMILY][0]
		keys = append(keys, row_.Row)
		return true
	}, gcp_bigtable.LimitRows(readBatchSize)) // limit to max 1000 entries to avoid blocking the import of new blocks
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		logger.Info("No ENS entries to validate")
		return nil
	}

	logger.Infof("Validating %v ENS entries", len(keys))
	alreadyChecked := EnsCheckedDictionary{
		address: make(map[common.Address]bool),
		name:    make(map[string]bool),
	}

	mutDelete := gcp_bigtable.NewMutation()
	mutDelete.DeleteRow()

	batchSize := 100
	total := len(keys)
	for i := 0; i < total; i += batchSize {
		to := i + batchSize
		if to > total {
			to = total
		}
		batch := keys[i:to]
		logger.Infof("Batching ENS entries %v:%v of %v", i, to, total)

		g := new(errgroup.Group)
		g.SetLimit(10) // limit load on the node
		mutsDelete := &types.BulkMutations{
			Keys: make([]string, 0, 1),
			Muts: make([]*gcp_bigtable.Mutation, 0, 1),
		}

		for _, k := range batch {
			key := k
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
						ens_name
					FROM ens
					WHERE name_hash = $1
					`, nameHash[:])
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

			g.Go(func() error {
				if name != "" {
					err := validateEnsName(client, name, &alreadyChecked)
					if err != nil {
						return fmt.Errorf("error validating new name [%v]: %w", name, err)
					}
				} else if address != nil {
					err := validateEnsAddress(client, *address, &alreadyChecked)
					if err != nil {
						return fmt.Errorf("error validating new address [%v]: %w", address, err)
					}
				}
				return nil
			})

			mutsDelete.Keys = append(mutsDelete.Keys, key)
			mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
		}

		if err := g.Wait(); err != nil {
			return err
		}

		// After processing a batch of keys we remove them from bigtable
		err = bigtable.WriteBulk(mutsDelete, bigtable.tableData, DEFAULT_BATCH_INSERTS)
		if err != nil {
			return err
		}

		// give node some time for other stuff between batches
		time.Sleep(time.Millisecond * 100)
	}

	logger.WithField("updates", total).Info("Import of ENS updates completed")
	return nil
}

func validateEnsAddress(client *ethclient.Client, address common.Address, alreadyChecked *EnsCheckedDictionary) error {
	alreadyChecked.mux.Lock()
	if alreadyChecked.address[address] {
		alreadyChecked.mux.Unlock()
		return nil
	}
	alreadyChecked.address[address] = true
	alreadyChecked.mux.Unlock()

	names := []string{}
	err := ReaderDb.Select(&names, `SELECT ens_name FROM ens WHERE address = $1 AND is_primary_name AND valid_to >= now()`, address.Bytes())
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	for _, name := range names {
		if name != "" {
			err = validateEnsName(client, name, alreadyChecked)
			if err != nil {
				return err
			}
		}
		reverseName, err := go_ens.ReverseResolve(client, address)
		if err != nil {
			if err.Error() == "not a resolver" ||
				err.Error() == "no resolution" ||
				err.Error() == "execution reverted" ||
				strings.HasPrefix(err.Error(), "name is not valid") {
				// logger.Warnf("reverse resolving address [%v] resulted in a skippable error [%s], skipping it", address, err.Error())
			} else {
				return fmt.Errorf("error could not reverse resolve address [%v]: %w", address, err)
			}
		}

		if reverseName != name {
			err = validateEnsName(client, reverseName, alreadyChecked)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validateEnsName(client *ethclient.Client, name string, alreadyChecked *EnsCheckedDictionary) error {
	if name == "" || name == ".eth" {
		return nil
	}
	// For now only .eth is supported other ens domains use different techniques and require and individual implementation
	if !strings.HasSuffix(name, ".eth") {
		name = fmt.Sprintf("%s.eth", name)
	}
	alreadyChecked.mux.Lock()
	if alreadyChecked.name[name] {
		alreadyChecked.mux.Unlock()
		return nil
	}
	alreadyChecked.name[name] = true
	alreadyChecked.mux.Unlock()

	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ens_validate_ens_name").Observe(time.Since(startTime).Seconds())
	}()

	nameHash, err := go_ens.NameHash(name)
	if err != nil {
		// logger.Warnf("error could not hash name [%v]: %v -> removing ens entry", name, err)
		err = removeEnsName(client, name)
		if err != nil {
			return fmt.Errorf("error removing ens name [%v]: %w", name, err)
		}
		return nil
	}

	addr, err := go_ens.Resolve(client, name)
	if err != nil {
		if err.Error() == "unregistered name" ||
			err.Error() == "no address" ||
			err.Error() == "no resolver" ||
			err.Error() == "abi: attempting to unmarshal an empty string while arguments are expected" ||
			strings.Contains(err.Error(), "execution reverted") ||
			err.Error() == "invalid jump destination" ||
			err.Error() == "invalid opcode: INVALID" {
			// the given name is not available anymore or resolving it did not work properly => we can remove it from the db (if it is there)
			// logger.WithField("error", err).WithField("name", name).Warnf("could not resolve name")
			err = removeEnsName(client, name)
			if err != nil {
				return fmt.Errorf("error removing ens name after resolve failed [%v]: %w", name, err)
			}
			return nil
		}
		return fmt.Errorf("error could not resolve name [%v]: %w", name, err)
	}

	// we need to get the main domain to get the expiration date
	parts := strings.Split(name, ".")
	mainName := strings.Join(parts[len(parts)-2:], ".")

	expires, err := GetEnsExpiration(client, mainName)
	if err != nil {
		return fmt.Errorf("error could not get ens expire date for [%v]: %w", name, err)
	}

	// ensName, err := go_ens.NewName(client, mainName)
	// if err != nil {
	// 	if strings.HasPrefix(err.Error(), "name is not valid") {
	// 		logger.WithField("error", err).WithField("name", name).Warnf("could not create name")
	// 		return nil
	// 	}
	// 	return fmt.Errorf("error could not create name via go_ens.NewName for [%v]: %w", name, err)
	// }
	// expires, err := ensName.Expires()
	// if err != nil {
	// 	return fmt.Errorf("error could not get ens expire date for [%v]: %w", name, err)
	// }

	isPrimary := false
	reverseName, err := go_ens.ReverseResolve(client, addr)
	if err != nil {
		if err.Error() == "not a resolver" || err.Error() == "no resolution" || err.Error() == "execution reverted" {
			// logger.Warnf("reverse resolving address [%v] for name [%v] resulted in an error [%s], marking entry as not primary", addr, name, err.Error())
		} else {
			return fmt.Errorf("error could not reverse resolve address [%v]: %w", addr, err)
		}
	}
	if reverseName == name {
		isPrimary = true
	}

	_, err = WriterDb.Exec(`
	INSERT INTO ens (
		name_hash, 
		ens_name, 
		address,
		is_primary_name, 
		valid_to)
	VALUES ($1, $2, $3, $4, $5) 
	ON CONFLICT 
		(name_hash) 
	DO UPDATE SET 
		ens_name = excluded.ens_name,
		address = excluded.address,
		is_primary_name = excluded.is_primary_name,
		valid_to = excluded.valid_to
	`, nameHash[:], name, addr.Bytes(), isPrimary, expires)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "invalid byte sequence") {
			logger.Warnf("could not insert ens name [%v]: %v", name, err)
			return nil
		}
		return fmt.Errorf("error writing ens data for name [%v]: %w", name, err)
	}

	// logrus.WithFields(logrus.Fields{
	// 	"name":        name,
	// 	"address":     addr,
	// 	"expires":     expires,
	// 	"reverseName": reverseName,
	// }).Infof("validated ens name")
	return nil
}

func GetEnsExpiration(client *ethclient.Client, name string) (time.Time, error) {
	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ens_get_expiration").Observe(time.Since(startTime).Seconds())
	}()

	normName, err := go_ens.NormaliseDomain(name)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.NormaliseDomain: %w", err)
	}
	domain := go_ens.Domain(normName)
	label, err := go_ens.DomainPart(normName, 1)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.DomainPart: %w", err)
	}
	registrar, err := go_ens.NewBaseRegistrar(client, domain)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.NewBaseRegistrar: %w", err)
	}
	uqName, err := go_ens.UnqualifiedName(label, domain)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.UnqualifiedName: %w", err)
	}
	labelHash, err := go_ens.LabelHash(uqName)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.LabelHash: %w", err)
	}
	id := new(big.Int).SetBytes(labelHash[:])
	ts, err := registrar.Contract.NameExpires(&bind.CallOpts{}, id)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts.Int64(), 0), nil
}

func GetAddressForEnsName(name string) (address *common.Address, err error) {
	addressBytes := []byte{}
	err = ReaderDb.Get(&addressBytes, `
	SELECT address 
	FROM ens
	WHERE
		ens_name = $1 AND
		valid_to >= now()
	`, name)
	if err == nil && addressBytes != nil {
		add := common.BytesToAddress(addressBytes)
		address = &add
	}
	return address, err
}

func GetEnsNameForAddress(address common.Address) (name string, err error) {
	err = ReaderDb.Get(&name, `
	SELECT ens_name 
	FROM ens
	WHERE
		address = $1 AND
		is_primary_name AND
		valid_to >= now()
	;`, address.Bytes())
	return name, err
}

func GetEnsNamesForAddress(addressMap map[string]string) error {
	if len(addressMap) == 0 {
		return nil
	}
	type pair struct {
		Address []byte `db:"address"`
		EnsName string `db:"ens_name"`
	}
	dbAddresses := []pair{}
	addresses := make([][]byte, 0, len(addressMap))
	for add := range addressMap {
		addresses = append(addresses, []byte(add))
	}

	err := ReaderDb.Select(&dbAddresses, `
	SELECT address, ens_name 
	FROM ens
	WHERE
		address = ANY($1) AND
		is_primary_name AND
		valid_to >= now()
	;`, addresses)
	if err != nil {
		return err
	}
	for _, foundling := range dbAddresses {
		addressMap[string(foundling.Address)] = foundling.EnsName
	}
	return nil
}

func removeEnsName(client *ethclient.Client, name string) error {
	_, err := WriterDb.Exec(`
	DELETE FROM ens 
	WHERE 
		ens_name = $1
	;`, name)
	if err != nil && strings.Contains(fmt.Sprintf("%v", err), "invalid byte sequence") {
		logger.Warnf("could not delete ens name [%v]: %v", name, err)
		return nil
	} else if err != nil {
		return fmt.Errorf("error deleting ens name [%v]: %v", name, err)
	}
	logger.Infof("Ens name removed from db: %v", name)
	return nil
}
