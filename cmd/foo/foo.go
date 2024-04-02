package main

import (
	"context"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/jackc/pgx/v5/stdlib"
	go_ens "github.com/wealdtech/go-ens/v3"
	"github.com/wealdtech/go-ens/v3/contracts/resolver"
	"github.com/wealdtech/go-ens/v3/contracts/reverseresolver"
	"golang.org/x/crypto/sha3"
	"golang.org/x/sync/errgroup"

	"flag"

	"github.com/sirupsen/logrus"
)

var opts = struct {
	Command             string
	User                uint64
	TargetVersion       int64
	StartEpoch          uint64
	EndEpoch            uint64
	StartDay            uint64
	EndDay              uint64
	Validator           uint64
	Validators          string
	Epochs              string
	StartBlock          uint64
	EndBlock            uint64
	BatchSize           uint64
	DataConcurrency     uint64
	Transformers        string
	Table               string
	Family              string
	Key                 string
	ValidatorNameRanges string
	Email               string
	Data                string
	DataHex             string
	Name                string
	DryRun              bool
}{}

var bt *db.Bigtable
var erigonClient *rpc.ErigonClient
var rpcClient *rpc.LighthouseClient

func main() {
	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	flag.StringVar(&opts.Command, "command", "", "command to run, available: updateAPIKey, applyDbSchema, initBigtableSchema, epoch-export, debug-rewards, debug-blocks, clear-bigtable, index-old-eth1-blocks, update-aggregation-bits, historic-prices-export, index-missing-blocks, export-epoch-missed-slots, migrate-last-attestation-slot-bigtable, export-genesis-validators, update-block-finalization-sequentially, nameValidatorsByRanges")
	flag.Uint64Var(&opts.StartEpoch, "start-epoch", 0, "start epoch")
	flag.Uint64Var(&opts.EndEpoch, "end-epoch", 0, "end epoch")
	flag.Uint64Var(&opts.User, "user", 0, "user id")
	flag.Uint64Var(&opts.StartDay, "day-start", 0, "start day to debug")
	flag.Uint64Var(&opts.EndDay, "day-end", 0, "end day to debug")
	flag.Uint64Var(&opts.Validator, "validator", 0, "validator to check for")
	flag.StringVar(&opts.Validators, "validators", "", "comma separated list of validators")
	flag.StringVar(&opts.Epochs, "epochs", "", "comma separated list of epochs")
	flag.Int64Var(&opts.TargetVersion, "target-version", -2, "Db migration target version, use -2 to apply up to the latest version, -1 to apply only the next version or the specific versions")
	flag.StringVar(&opts.Table, "table", "", "big table table")
	flag.StringVar(&opts.Family, "family", "", "big table family")
	flag.StringVar(&opts.Key, "key", "", "big table key")
	flag.Uint64Var(&opts.StartBlock, "blocks.start", 0, "Block to start indexing")
	flag.Uint64Var(&opts.EndBlock, "blocks.end", 0, "Block to finish indexing")
	flag.Uint64Var(&opts.DataConcurrency, "data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	flag.Uint64Var(&opts.BatchSize, "data.batchSize", 1000, "Batch size")
	flag.StringVar(&opts.Transformers, "transformers", "", "Comma separated list of transformers used by the eth1 indexer")
	flag.StringVar(&opts.ValidatorNameRanges, "validator-name-ranges", "https://config.dencun-devnet-8.ethpandaops.io/api/v1/nodes/validator-ranges", "url to or json of validator-ranges (format must be: {'ranges':{'X-Y':'name'}})")
	flag.StringVar(&opts.Email, "email", "", "email to debug")
	flag.StringVar(&opts.Data, "data", "", "data to debug")
	flag.StringVar(&opts.DataHex, "data-hex", "", "data to debug")
	flag.StringVar(&opts.Name, "name", "", "name")
	dryRun := flag.String("dry-run", "true", "if 'false' it deletes all rows starting with the key, per default it only logs the rows that would be deleted, but does not really delete them")
	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		fmt.Println(version.GoVersion)
		return
	}

	opts.DryRun = *dryRun != "false"

	logrus.WithField("config", *configPath).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)
		var err error
		rpcClient, err = rpc.NewLighthouseClient("http://"+utils.Config.Indexer.Node.Host+":"+utils.Config.Indexer.Node.Port, chainIDBig)
		if err != nil {
			utils.LogFatal(err, "lighthouse client error", 0)
		}
		wg.Done()
	}()

	go func() {
		var err error
		erigonClient, err = rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
		if err != nil {
			logrus.Fatalf("error initializing erigon client: %v", err)
		}
		wg.Done()
	}()
	wg.Wait()

	switch opts.Command {
	case "foo":
		mustInitDbs()
		err = Foo()
	case "foo2":
		err = Foo2()
	case "foo3":
		err = Foo3()
	case "foo4":
		err = Foo4()
	case "foo5":
		err = Foo5()
	case "bids2555":
		err = bids2555()
	case "debug-ens":
		err = debugEns()
	case "debug-sessions":
		err = debugSessions()
	case "bt-get-keys-by-prefix":
		err = btGetKeysByPrefix()
	case "ens-find-change-block":
		err = ensFindChangeBlock()
	case "keccak256":
		err = keccak256()
	default:
		utils.LogFatal(nil, fmt.Sprintf("unknown command %s", opts.Command), 0)
	}

	if err != nil {
		logrus.WithError(err).Errorf("command failed")
	} else {
		logrus.Infof("command executed successfully")
	}
}

func mustInitDbs() {
	chainIdString := strconv.FormatUint(utils.Config.Chain.ClConfig.DepositChainID, 10)

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		var err error
		bt, err = db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, chainIdString, utils.Config.RedisCacheEndpoint)
		if err != nil {
			utils.LogFatal(err, "error initializing bigtable", 0)
		}
		wg.Done()
	}()

	go func() {
		db.MustInitDB(&types.DatabaseConfig{
			Username:     utils.Config.WriterDatabase.Username,
			Password:     utils.Config.WriterDatabase.Password,
			Name:         utils.Config.WriterDatabase.Name,
			Host:         utils.Config.WriterDatabase.Host,
			Port:         utils.Config.WriterDatabase.Port,
			MaxOpenConns: utils.Config.WriterDatabase.MaxOpenConns,
			MaxIdleConns: utils.Config.WriterDatabase.MaxIdleConns,
		}, &types.DatabaseConfig{
			Username:     utils.Config.ReaderDatabase.Username,
			Password:     utils.Config.ReaderDatabase.Password,
			Name:         utils.Config.ReaderDatabase.Name,
			Host:         utils.Config.ReaderDatabase.Host,
			Port:         utils.Config.ReaderDatabase.Port,
			MaxOpenConns: utils.Config.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: utils.Config.ReaderDatabase.MaxIdleConns,
		})
		wg.Done()
	}()
	go func() {
		db.MustInitFrontendDB(&types.DatabaseConfig{
			Username:     utils.Config.Frontend.WriterDatabase.Username,
			Password:     utils.Config.Frontend.WriterDatabase.Password,
			Name:         utils.Config.Frontend.WriterDatabase.Name,
			Host:         utils.Config.Frontend.WriterDatabase.Host,
			Port:         utils.Config.Frontend.WriterDatabase.Port,
			MaxOpenConns: utils.Config.Frontend.WriterDatabase.MaxOpenConns,
			MaxIdleConns: utils.Config.Frontend.WriterDatabase.MaxIdleConns,
		}, &types.DatabaseConfig{
			Username:     utils.Config.Frontend.ReaderDatabase.Username,
			Password:     utils.Config.Frontend.ReaderDatabase.Password,
			Name:         utils.Config.Frontend.ReaderDatabase.Name,
			Host:         utils.Config.Frontend.ReaderDatabase.Host,
			Port:         utils.Config.Frontend.ReaderDatabase.Port,
			MaxOpenConns: utils.Config.Frontend.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: utils.Config.Frontend.ReaderDatabase.MaxIdleConns,
		})
		wg.Done()
	}()

	wg.Wait()

	go func() {
		utils.WaitForCtrlC()
		logrus.Infof("closing db-connections")

		db.ReaderDb.Close()
		db.WriterDb.Close()

		db.FrontendReaderDB.Close()
		db.FrontendWriterDB.Close()
	}()
}

func keccak256() error {
	var hash [32]byte
	sha := sha3.NewLegacyKeccak256()
	// //nolint:golint,errcheck
	sha.Write([]byte(opts.Data))
	sha.Sum(hash[:0])
	fmt.Printf("%#x\n", hash)
	return nil
}

type ResolveAtBlockResult struct {
	ReverseResolvedName string
	Name                string
	NameHash            [32]byte
	ResolvedAddr        common.Address
	ResolverAddr        common.Address
	Expires             time.Time
	IsPrimary           bool
}

func (r *ResolveAtBlockResult) String() string {
	return fmt.Sprintf("name: %v, nameHash: %#x, resolvedAddr: %v, expires: %v, isPrimary: %v, reverseResolvedName: %v", r.Name, r.NameHash, r.ResolvedAddr, r.Expires, r.IsPrimary, r.ReverseResolvedName)
}

func ResolveAtBlock(name string, block uint64) (*ResolveAtBlockResult, error) {
	callOpts := &bind.CallOpts{
		BlockNumber: new(big.Int).SetUint64(block),
	}

	if !strings.HasSuffix(name, ".eth") {
		name = name + ".eth"
	}

	// Resolve()
	nameHash, err := go_ens.NameHash(name)
	if err != nil {
		return nil, err
	}

	registry, err := go_ens.NewRegistry(erigonClient.GetNativeClient())
	if err != nil {
		return nil, err
	}

	// resolveHash()
	resolverAddr, err := registry.Contract.Resolver(callOpts, nameHash)
	if err != nil {
		return nil, err
	}

	contract, err := resolver.NewContract(resolverAddr, erigonClient.GetNativeClient())
	if err != nil {
		return nil, err
	}

	resolvedAddr, err := contract.Addr(callOpts, nameHash)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(name, ".")
	mainName := strings.Join(parts[len(parts)-2:], ".")
	// ensName, err := go_ens.NewName(erigonClient.GetNativeClient(), mainName)
	// if err != nil {
	// 	return nil, err
	// }

	normName, err := go_ens.NormaliseDomain(mainName)
	if err != nil {
		return nil, err
	}

	domain := go_ens.Domain(normName)
	// label, err := go_ens.DomainPart(normName, 1)
	// if err != nil {
	// 	return nil, err
	// }

	label, err := go_ens.DomainPart(normName, 1)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("normName: %v, hash: %#x, domain: %v, label: %v\n", normName, nameHash, domain, label)

	registrar, err := go_ens.NewBaseRegistrar(erigonClient.GetNativeClient(), domain)
	if err != nil {
		return nil, err
	}

	uqName, err := go_ens.UnqualifiedName(label, domain) // todo
	if err != nil {
		return nil, err
	}
	labelHash, err := go_ens.LabelHash(uqName)
	if err != nil {
		return nil, err
	}
	id := new(big.Int).SetBytes(labelHash[:])
	ts, err := registrar.Contract.NameExpires(callOpts, id)
	if err != nil {
		return nil, err
	}
	expireTime := time.Unix(ts.Int64(), 0)

	reverseDomain := fmt.Sprintf("%x.addr.reverse", resolvedAddr.Bytes())
	reverseNameHash, err := go_ens.NameHash(reverseDomain)
	if err != nil {
		return nil, err
	}
	reverseResolverAddr, err := registry.Contract.Resolver(callOpts, reverseNameHash)
	if err != nil {
		return nil, err
	}
	reverseResolverContract, err := reverseresolver.NewContract(reverseResolverAddr, erigonClient.GetNativeClient())
	if err != nil {
		return nil, err
	}

	// Ensure the contract is a resolver.
	reverseResolvedName := ""
	reverseNameHash0, err := go_ens.NameHash("0.addr.reverse")
	if err != nil {
		return nil, err
	}
	_, err = reverseResolverContract.Name(callOpts, reverseNameHash0)
	if err != nil {
		// return nil, fmt.Errorf("not a resolver")
		// fmt.Printf("contract is not a reslover: %v\n", err)
	} else {
		reverseResolvedName, err = reverseResolverContract.Name(callOpts, reverseNameHash)
		if err != nil {
			fmt.Printf("error getting reverseResolvedName: %v\n", err)
		}
	}

	return &ResolveAtBlockResult{
		Name:                name,
		NameHash:            nameHash,
		ReverseResolvedName: reverseResolvedName,
		ResolvedAddr:        resolvedAddr,
		ResolverAddr:        resolverAddr,
		Expires:             expireTime,
		IsPrimary:           name == reverseResolvedName,
	}, nil
}

func Foo5() error {
	mustInitDbs()
	// for _, blockNumber := range []uint64{19517000} {
	// 	res, err := ResolveAtBlock("vitalik.eth", blockNumber)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if res != nil {
	// 		logrus.Infof("block %v: %+v", blockNumber, *res)
	// 	} else {
	// 		logrus.Infof("block %v: nil", blockNumber)
	// 	}
	// }
	// if true {
	// 	return nil
	// }

	for _, blockNumber := range []uint64{14060476, 14060507, 17563418, 19517650} {
		res, err := ResolveAtBlock("mar-a-frogo.eth", blockNumber)
		if err != nil {
			return err
		}
		if res != nil {
			logrus.Infof("block %v: %+v", blockNumber, *res)
		} else {
			logrus.Infof("block %v: nil", blockNumber)
		}
	}

	if true {
		return nil
	}

	for _, blockNumber := range []uint64{19015796, 13452504, 11407775, 9465947, 9413106} {
		res, err := ResolveAtBlock("deficoin.eth", blockNumber)
		if err != nil {
			return err
		}
		if res != nil {
			logrus.Infof("block %v: %+v", blockNumber, *res)
		} else {
			logrus.Infof("block %v: nil", blockNumber)
		}
	}
	if true {
		return nil
	}

	// we will debug the history of the ens-name "deficoin"
	// relevant blocks: 19015796, 13452504, 11407775, 9465947, 9413106

	/*
		go_ens.Resolve(erigonClient.GetNativeClient(), "deficoin")
			Resolve
				NewResolver
					NewRegistry().ResolverAddress()
				return NewResolver().Address()

	*/

	nameHash, err := go_ens.NameHash("deficoin.eth")
	if err != nil {
		return err
	}
	fmt.Printf("nameHash: %x\n", nameHash)
	registry, err := go_ens.NewRegistry(erigonClient.GetNativeClient())
	if err != nil {
		return err
	}
	for _, blockNumber := range []uint64{19015796, 13452504, 11407775, 9465947, 9413106} {
		// owner, err := registry.Contract.Owner(&bind.CallOpts{
		// 	BlockNumber: new(big.Int).SetUint64(blockNumber),
		// }, nameHash)

		addr, err := registry.Contract.Resolver(&bind.CallOpts{
			BlockNumber: new(big.Int).SetUint64(blockNumber),
		}, nameHash)
		fmt.Printf("%v: resolver: %s, err: %v\n", blockNumber, addr, err)
	}

	return nil
}

func btGetKeysByPrefix() error {
	rows, err := bt.GetRowsByPrefix(opts.Key)
	if err != nil {
		return err
	}
	for _, row := range rows {
		logrus.Infof("row: %v", row)
	}
	return nil
}

func debugSessions() error {
	mustInitDbs()
	if opts.Email == "" {
		return errors.New("no email specified")
	}

	if utils.Config.Frontend.SessionSecret == "" {
		return fmt.Errorf("session secret is empty, please provide a secure random string.")
	}

	logrus.Infof("initializing session store: %v", utils.Config.RedisSessionStoreEndpoint)

	utils.InitSessionStore(utils.Config.Frontend.SessionSecret)

	user := struct {
		ID    uint64 `db:"id"`
		Email string `db:"email"`
	}{}
	err := db.FrontendWriterDB.Get(&user, `select id, email from users where email = $1`, opts.Email)
	if err != nil {
		return err
	}

	logrus.Infof("iterating over all sessions of user %v", user.ID)
	allSessionsCounter := 0
	ctx := context.Background()
	t0 := time.Now()
	// invalidate all sessions for this user
	err = utils.SessionStore.SCS.Iterate(ctx, func(ctx context.Context) error {
		allSessionsCounter++
		sessionUserID, ok := utils.SessionStore.SCS.Get(ctx, "user_id").(uint64)
		if !ok {
			return nil
		}

		if user.ID == sessionUserID {
			logrus.Infof("found a session of user %v", user.ID)
			// return utils.SessionStore.SCS.Destroy(ctx)
		}

		return nil
	})
	if err != nil {
		return err
	}
	logrus.Infof("iterated over all sessions (%v) in %v", allSessionsCounter, time.Since(t0))

	return nil
}

func debugEns() error {
	mustInitDbs()
	fmt.Printf("ctrcts: %v\n", utils.Config.Indexer.EnsTransformer.ValidRegistrarContracts)

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit
	transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)
	transforms = append(transforms, bt.TransformEnsNameRegistered)

	// 9380471
	batchSize := int64(1000)
	firstEnsBlock := int64(9380471)
	firstBlock := firstEnsBlock
	lastBlock := int64(19290977)
	if true {
		firstBlock = 13452504
		firstBlock = 11407775
		lastBlock = firstBlock
	}
	_ = lastBlock
	//start := int64(19287750)
	//for i := start - batchSize; i > firstEnsBlock; i -= batchSize {
	if false {
		for i := firstBlock; i < lastBlock; i += batchSize {
			keys, err := bt.IndexEventsWithTransformersDry(i, i+batchSize, transforms, 2, cache)
			if err != nil {
				utils.LogError(err, "error indexing from bigtable", 0)
				return err
			}
			for _, key := range keys {
				if !strings.HasPrefix(key, "1:ENS:V:") {
					rows, err := bt.GetRowsByPrefix(key)
					if err != nil {
						utils.LogError(err, fmt.Sprintf("error getting rows by prefix for key %v", key), 0)
						return err
					}
					if len(rows) == 0 {
						fmt.Println(key)
					} else {
						fmt.Printf("got rows for key %v\n", key)
					}
				}
			}
		}
	}

	for _, i := range []int64{19015796, 13452504, 11407775, 9465947, 9413106} {
		// for _, i := range []int64{9465947} {
		// for _, i := range []int64{11407775} {
		logrus.Infof("------- %v", i)
		keys, err := bt.IndexEventsWithTransformersDry(i, i, transforms, 2, cache)
		if err != nil {
			utils.LogError(err, "error indexing from bigtable", 0)
			return err
		}
		if len(keys) == 0 {
			logrus.Infof("no keys")
		}
		for _, key := range keys {
			fmt.Println(key)
			if !strings.HasPrefix(key, "1:ENS:V:") {
				rows, err := bt.GetRowsByPrefix(key)
				if err != nil {
					utils.LogError(err, fmt.Sprintf("error getting rows by prefix for key %v", key), 0)
					return err
				}
				if len(rows) != 0 {
					fmt.Printf("got rows for key %v\n", key)
				}
			}

			strings.Split(key, ":")
		}
	}

	return nil
}

func extractGaps(data []uint64) []string {
	sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
	gaps := []string{}
	a := data[0]
	foundGap := false
	for i := 1; i < len(data); i++ {
		if data[i] == data[i-1]+1 {
			if foundGap {
				gaps = append(gaps, fmt.Sprintf("%v-%v", a, data[i]))
			}
			a = data[i]
			foundGap = false
		} else {
			foundGap = true
		}
	}
	return gaps
}

func bids2555() error {
	mustInitDbs()
	/*
		gnosis
			v2SchemaCutOffEpoch: 725500
			missing: 557463

	*/
	valis := []uint64{1}
	startEpoch := uint64(557000)
	endEpoch := uint64(557999)
	history, err := db.BigtableClient.GetValidatorIncomeDetailsHistory(valis, startEpoch, endEpoch)
	if err != nil {
		return err
	}
	missing := make(map[uint64]bool)
	for e := startEpoch; e <= endEpoch; e++ {
		missing[e] = true
	}
	for _, epochs := range history {
		for epoch, _ := range epochs {
			delete(missing, epoch)
		}
	}
	missingUint64 := []uint64{}
	for e := range missing {
		missingUint64 = append(missingUint64, e)
	}
	gaps := extractGaps(missingUint64)
	logrus.Infof("missing (%v): %v", len(missingUint64), missingUint64)
	logrus.Infof("gaps (%v): %v", len(gaps), gaps)
	return nil
}

func Foo() error {
	/*
		mustInitDbs()
		cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit
		transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)
		transforms = append(transforms, bt.TransformEnsNameRegistered)

		type DebugEnsName struct {
			Name       string
			AddressHex string
			TxBlocks   []int64
		}

		debugEnsNames := []DebugEnsName{
			// wrong because internal txs
			// {"deficoin.eth", "0x41c259847ca4bbd0129687eeeee17c689d91a96d", []int64{9413106, 9465947, 11407775, 13452504, 19015796}},
			{"straightupbussin.eth", "0x3b29c500ee9752f0638d2a0420042168d2d79e95", []int64{17815297}},

			// failed resolve: no address
			//{"xianchina.eth", "0xd9a4f8ddf9c5482fe0278198826a584ff251f277", []int64{14658607, 18903906}},
			// {"gasper.eth", "0x00284d2e4b7201f070f7d02bf6c899be33b0aa56", []int64{17976531, 14489320, 11193441, 11193422, 9417934}},
		}

		for _, debugEnsName := range debugEnsNames {
			logrus.WithFields(logrus.Fields{"name": debugEnsName.Name, "addr": debugEnsName.AddressHex, "blocks": debugEnsName.TxBlocks}).Infof("==== debugging ens")
			name := debugEnsName.Name
			if !strings.HasSuffix(name, ".eth") {
				name += ".eth"
			}
			nameHash, err := go_ens.NameHash(name)
			if err != nil {
				return err
			}
			logrus.Infof("nameHash: %x", nameHash)
			rAddr, err := go_ens.Resolve(erigonClient.GetNativeClient(), name)
			if err != nil {
				logrus.WithFields(logrus.Fields{"name": name, "err": err}).Warnf("error resolving name")
			} else {
				logrus.Infof("resolved address: %v", rAddr)
			}

			rrName, err := go_ens.ReverseResolve(erigonClient.GetNativeClient(), common.HexToAddress(debugEnsName.AddressHex))
			if err != nil {
				logrus.WithFields(logrus.Fields{"name": name, "err": err, "addr": debugEnsName.AddressHex}).Warnf("error reverse-resolving name")
			} else {
				logrus.Infof("reverse resolved name: %v", rrName)
			}

			keys := []string{
				fmt.Sprintf("%d:ENS:I:A:%s", utils.Config.Chain.ClConfig.DepositChainID, strings.Replace(debugEnsName.AddressHex, "0x", "", -1)),
				fmt.Sprintf("%d:ENS:I:H:%x", utils.Config.Chain.ClConfig.DepositChainID, nameHash),
			}
			for _, key := range keys {
				logrus.Infof("getting rows with prefix %v", key)
				rows, err := bt.GetRowsByPrefix(key)
				if err != nil {
					return err
				}
				for _, row := range rows {
					logrus.Infof("row: %v", row)
				}
			}

			for _, blockNumber := range debugEnsName.TxBlocks {
				logrus.Infof("debugging transform for block %v", blockNumber)
				_, err = bt.IndexEventsWithTransformersDry(blockNumber, blockNumber, transforms, int64(2), cache)
				if err != nil {
					utils.LogError(err, "error indexing from bigtable", 0)
				}
			}
		}

		rows, err := bt.GetRowsByPrefix(fmt.Sprintf("%d:ENS:V", utils.Config.Chain.ClConfig.DepositChainID))
		if err != nil {
			return err
		}
		logrus.Infof("rows that need verification: %v", len(rows))
		for _, row := range rows {
			logrus.Infof("row: %v", row)
		}
	*/
	return nil
}

func ensFindChangeBlock() error {
	if opts.Name == "" {
		return fmt.Errorf("no name specified")
	}
	if opts.StartBlock == 0 {
		return fmt.Errorf("no start block specified")
	}

	latestBlockNumber, err := erigonClient.GetNativeClient().BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting latest block number: %w", err)
	}

	latestBlockRes, err := ResolveAtBlock(opts.Name, uint64(latestBlockNumber))
	if err != nil {
		return err
	}

	startBlockRes, err := ResolveAtBlock(opts.Name, opts.StartBlock)
	if err != nil {
		return err
	}

	check := func(a, b *ResolveAtBlockResult) bool {
		return a.Expires.Equal(b.Expires)
		// return a.IsPrimary == b.IsPrimary
	}

	if check(startBlockRes, latestBlockRes) {
		return fmt.Errorf("no change between start block %v and latest block %v", opts.StartBlock, latestBlockNumber)
	}

	fmt.Printf("start block: %v (primary: %v), latest block: %v\n", opts.StartBlock, startBlockRes.IsPrimary, latestBlockNumber)

	upperBound := latestBlockNumber
	lowerBound := opts.StartBlock
	n := lowerBound + (upperBound-lowerBound)/2
	for {
		res, err := ResolveAtBlock(opts.Name, n)
		if err != nil {
			return fmt.Errorf("error resolving at block %v: %w", upperBound, err)
		}
		if check(res, startBlockRes) {
			lowerBound = lowerBound + (upperBound-lowerBound)/2
			n = lowerBound + (upperBound-lowerBound)/2
		} else {
			upperBound = lowerBound + (upperBound-lowerBound)/2
			n = lowerBound + (upperBound-lowerBound)/2
		}
		fmt.Printf("checked block %v: lower/upper: %v/%v: %v\n", n, lowerBound, upperBound, *res)
		if upperBound-lowerBound <= 1 {
			fmt.Printf("primary changed at block %v\n", upperBound)
			return nil
		}
	}
	return nil
}

func Foo4() error {
	mustInitDbs()

	// for i := 19294310; i < 19517650; i += 1 {
	// 	res, err := ResolveAtBlock("mar-a-frogo.eth", uint64(i))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if !res.IsPrimary {
	// 		logrus.Infof("block %v: %+v", i, *res)
	// 		return nil
	// 	}
	// 	logrus.Infof("block %v: %+v", i, *res)
	// }
	// if true {
	// 	return nil
	// }

	/* */
	//for _, name := range []string{
	//	"",
	//} {
	//
	//}
	// eventSignature := []byte("ItemSet(bytes32,bytes32)")
	// hash := crypto.Keccak256Hash(eventSignature)

	logrus.Infof("ens-contracts: %v", utils.Config.Indexer.EnsTransformer.ValidRegistrarContracts)

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit
	transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)
	transforms = append(transforms, bt.TransformEnsNameRegistered)

	latestBlockNumber, err := erigonClient.GetNativeClient().BlockNumber(context.Background())
	if err != nil {
		return err
	}

	type DebugEnsName struct {
		Name       string
		AddressHex string
		TxBlocks   []int64
	}

	debugEnsNames := []DebugEnsName{
		// time="2024-03-25T22:02:43Z" level=warning msg="updating ens entry: is_primary_name = false" addr=0xd613cb815902a1a6c469b899a18ab9b1901e8c30 name=alexneto.eth reason="failed reverse-resolve: no resolution"
		// {"alexneto.eth", "0xD613cb815902a1A6C469b899a18aB9B1901E8c30", []int64{16393744, 16393755, 17741613, 18185315, int64(latestBlockNumber)}},

		// time="2024-03-25T18:38:24Z" level=warning msg="updating ens entry: is_primary_name = false" addr=0x1aa571bb84242a6acfb6df1a6f232266d0a753b8 name=mysets.eth reason="failed reverse-resolve: not a resolver"
		{"mysets.eth", "0x1AA571bb84242A6ACfB6Df1A6f232266D0A753b8", []int64{8754994, 9432092, 15283509, int64(latestBlockNumber)}},

		// {"mar-a-frogo.eth", "0x028B0363ffD8c9DC545A4eB60C085177cd31436B", []int64{14060476, 14060507, 17563418, 19294312, 19517650}},
		// {"mar-a-frogo.eth", "0x028B0363ffD8c9DC545A4eB60C085177cd31436B", []int64{19294312}},
		// {"mar-a-frogo.eth", "0x028B0363ffD8c9DC545A4eB60C085177cd31436B", []int64{19400000, 19517650}},

		/*
			// failed resolve: no address
			{"xianchina.eth", "0xd9a4f8ddf9c5482fe0278198826a584ff251f277", []int64{14658607, 18903906}},
			{"qhsailing.eth", "0x3ccfc95726606575a873369b9a9f90e2c3e7e349", []int64{16281917, 18874384}},
			{"nomelgibsonisacasinosbiglemon.eth", "0x41839c9236b6c40cb1d7992b9f58137573571fb0", []int64{15116351, 18876284}},
			{"bulgaa.eth", "0x458f9e629f378f7cd78cfe08145c5d0807ba24c2", []int64{15419858, 18939200}},
			{"0xtakashi.eth", "0x6fb30da1ca2ed9261730f4a86dfaea237403cf9a", []int64{14294043, 18909760}},
			{"pgpg.eth", "0x739579406156648512b7029291c1aabddd95bb29", []int64{13354289, 15479607, 15816255, 18924611, 18924719}},
			{"georgejetty", "0x7d69254b25717382fe04ac82a416fdbc1c4ee410", []int64{13700886, 18874113}},

			// failed resolve: no resolver
			{"yijia11.eth", "0x17c861dbdd61e6e57d4db6d452726a5b68898b46", []int64{17517454, 18911023}},
			{"wsw50.eth", "0x76c3d80b9a67698dede9a5f80a5fbb55033a4b89", []int64{14802830, 18925441}},

			// failed resolve: unregistered name
			{"pouch.codemonkey.eth", "0x5b7aa47a8ae661db9e63030f57f68134317ba205", []int64{15131766}},
			{"vault.moby.eth", "0x905d13367f3bb940072a133c81563f1b1a6779ad", []int64{}},
			{"bids.cryptopunks.eth", "0xd678aa3afaa7c145b9d96574d5c97a977703d8cd", []int64{}},
			{"lucaskohorst.argent.xyz.eth", "0xda7a203806a6be3c3c4357c38e7b3aaac47f5dd2", []int64{}},
			{"edusr.argent.xyz.eth", "0xb939d8dfa6de6741153c19ce3fc4565c2f6d1b3d", []int64{}},
			{"c.privato.eth", "0x1386e50c0547a7cb7232bce79c1db7f19927208f", []int64{}},

			// failed resolve: execution reverted
			{"siher.eth", "0x59653a323561a719bbcca3534d2f0619977ad3fa", []int64{}},

			// failed reverse-resolve: no resolution
			{"sloshtastic.eth", "0x37ae899bdeec1205840ab5250091824435b186bb", []int64{}},

			// failed reverse-resolve: not a resolver
			{"dappydev.eth", "0x358f6260f1f90cd11a10e251ce16ea526f131b02", []int64{}},

			// resolved != reverseResolved
			{"🦄🐉✨🐲🦄.eth", "0x000000001a3494ed0ee59623ffaa46995a6c99a9", []int64{}},
			{"strikefighter.eth", "0x8fabb4f765d69a37074aa30d550bc5f64351f1a6", []int64{}},
		*/
	}

	for _, debugEnsName := range debugEnsNames {
		logrus.WithFields(logrus.Fields{"name": debugEnsName.Name, "addr": debugEnsName.AddressHex, "blocks": debugEnsName.TxBlocks}).Infof("==== debugging ens")
		name := debugEnsName.Name
		if !strings.HasSuffix(name, ".eth") {
			name += ".eth"
		}
		nameHash, err := go_ens.NameHash(name)
		if err != nil {
			return err
		}
		logrus.Infof("nameHash: %x", nameHash)
		// rAddr, err := go_ens.Resolve(erigonClient.GetNativeClient(), name)
		// if err != nil {
		// 	logrus.WithFields(logrus.Fields{"name": name, "err": err}).Warnf("error resolving name")
		// } else {
		// 	logrus.Infof("resolved name: %v", rAddr)
		// }
		// rrName, err := go_ens.ReverseResolve(erigonClient.GetNativeClient(), common.HexToAddress(debugEnsName.AddressHex))
		// if err != nil {
		// 	return err
		// }
		// logrus.Infof("reverse resolved name: %v", rrName)

		keys := []string{
			fmt.Sprintf("%d:ENS:I:A:%s", utils.Config.Chain.ClConfig.DepositChainID, strings.Replace(debugEnsName.AddressHex, "0x", "", -1)),
			fmt.Sprintf("%d:ENS:I:H:%x", utils.Config.Chain.ClConfig.DepositChainID, nameHash),
		}
		for _, key := range keys {
			logrus.Infof("getting rows with prefix %v", key)
			rows, err := bt.GetRowsByPrefix(key)
			if err != nil {
				return err
			}
			for _, row := range rows {
				logrus.Infof("row: %v", row)
			}
		}

		for _, blockNumber := range debugEnsName.TxBlocks {
			logrus.Infof("== debugging block %v", blockNumber)
			res, err := ResolveAtBlock(debugEnsName.Name, uint64(blockNumber))
			if err != nil {
				logrus.Errorf("error resolving at block %v: %v", blockNumber, err)
				continue
			}
			logrus.Infof("resolved at block %v: %+v", blockNumber, res)
			keys, err = bt.IndexEventsWithTransformersDry(blockNumber, blockNumber, transforms, int64(2), cache)
			if err != nil {
				utils.LogError(err, "error indexing from bigtable", 0)
			} else {
				logrus.Infof("dry keys: %v", keys)
			}
		}
	}
	/* */
	return nil
}

func Foo3() error {
	mustInitDbs()
	logrus.WithFields(logrus.Fields{"dry": opts.DryRun}).Infof("command: fix-ens")

	addrs := []struct {
		Address []byte    `db:"address"`
		EnsName string    `db:"ens_name"`
		ValidTo time.Time `db:"valid_to"`
	}{}
	err := db.WriterDb.Select(&addrs, `select address, ens_name, valid_to from ens where is_primary_name = true and valid_to > now()`)
	if err != nil {
		return err
	}

	logrus.Infof("found %v ens entries", len(addrs))

	g := new(errgroup.Group)
	g.SetLimit(10) // limit load on the node

	batchSize := 500
	total := len(addrs)
	for i := 0; i < total; i += batchSize {
		to := i + batchSize
		if to > total {
			to = total
		}
		batch := addrs[i:to]

		logrus.Infof("processing batch %v-%v / %v", i, to, total)
		for _, addr := range batch {
			addr := addr
			g.Go(func() error {
				ensAddr, err := go_ens.Resolve(erigonClient.GetNativeClient(), addr.EnsName)
				if err != nil {
					if err.Error() == "unregistered name" ||
						err.Error() == "no address" ||
						err.Error() == "no resolver" ||
						err.Error() == "abi: attempting to unmarshall an empty string while arguments are expected" ||
						strings.Contains(err.Error(), "execution reverted") ||
						err.Error() == "invalid jump destination" {
						logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("failed resolve: %v", err.Error())}).Warnf("deleting ens entry")
						if !opts.DryRun {
							_, err = db.WriterDb.Exec(`delete from ens where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
							if err != nil {
								return err
							}
						}
						return nil
					}
					return err
				}

				dbAddr := common.BytesToAddress(addr.Address)
				if dbAddr.Cmp(ensAddr) != 0 {
					logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("dbAddr != resolved ensAddr: %#x != %#x", addr.Address, ensAddr.Bytes())}).Warnf("deleting ens entry")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`delete from ens where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
						if err != nil {
							return err
						}
					}
				}

				reverseName, err := go_ens.ReverseResolve(erigonClient.GetNativeClient(), dbAddr)
				if err != nil {
					if err.Error() == "not a resolver" || err.Error() == "no resolution" {
						logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("failed reverse-resolve: %v", err.Error())}).Warnf("updating ens entry: is_primary_name = false")
						if !opts.DryRun {
							_, err = db.WriterDb.Exec(`update ens set is_primary_name = false where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
							if err != nil {
								return err
							}
						}
						return nil
					}
					return err
				}

				if reverseName != addr.EnsName {
					logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("resolved != reverseResolved: %v != %v", addr.EnsName, reverseName)}).Warnf("updating ens entry: is_primary_name = false")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`update ens set is_primary_name = false where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
						if err != nil {
							return err
						}
					}
				}

				return nil
			})
		}

		err = g.Wait()
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func Foo2() error {
	mustInitDbs()
	for _, name := range []string{
		"kozzmozzg.eth",
	} {
		addr, err := go_ens.Resolve(erigonClient.GetNativeClient(), name)
		if err != nil {
			if err.Error() == "unregistered name" ||
				err.Error() == "no address" ||
				err.Error() == "no resolver" ||
				err.Error() == "abi: attempting to unmarshall an empty string while arguments are expected" ||
				strings.Contains(err.Error(), "execution reverted") ||
				err.Error() == "invalid jump destination" {
				logrus.WithFields(logrus.Fields{"addr": addr.Hex()}).Warnf("error resolving name: %v", err)
			} else {
				return fmt.Errorf("error go_ens.Resolve(%v): %w", name, err)
			}
		}
		logrus.WithFields(logrus.Fields{"addr": addr.Hex(), "name": name}).Infof("checked ens entry")
	}

	for _, addrHex := range []string{
		"0x4976fb03c32e5b8cfe2b6ccb31c09ba78ebaba41",
		"0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
		"0x7919Cc11472899A67dF23206A50dDEb61555BD5a",
		"0x45D964B0E218ffD89958F25B6D02AfeB87BaFd90",
		"0xe4ad0f95b36248208ef22ae01c47f872c8b37392",
		"0x5d17a3635ab4d5526fcfaae1670a21f2201202c4",
		"0xf14955B6f701A4BFD422dCC324cf1F4b5A466265",
	} {
		addr := common.HexToAddress(addrHex)

		name, err := go_ens.ReverseResolve(erigonClient.GetNativeClient(), addr)
		if err != nil {
			if err.Error() == "not a resolver" ||
				err.Error() == "no resolution" {
				logrus.WithFields(logrus.Fields{"addr": addr.Hex()}).Warnf("error reverse-resolving name: %v", err)
				continue
			} else {
				return fmt.Errorf("error go_ens.ReverseResolve: %w", err)
			}
		}

		// dbName, err := db.GetEnsNameForAddress(common.HexToAddress(addrHex))
		// if err != nil && err != sql.ErrNoRows {
		// 	return err
		// }
		// if dbName == nil {
		// 	logrus.WithFields(logrus.Fields{"addr": addr.Hex(), "name": name}).Warnf("no dbName found for address %v", addrHex)
		// 	continue
		// }

		resolvedAddr, err := go_ens.Resolve(erigonClient.GetNativeClient(), name)
		if err != nil {
			if err.Error() == "unregistered name" ||
				err.Error() == "no address" ||
				err.Error() == "no resolver" ||
				err.Error() == "abi: attempting to unmarshall an empty string while arguments are expected" ||
				strings.Contains(err.Error(), "execution reverted") ||
				err.Error() == "invalid jump destination" {
				logrus.WithFields(logrus.Fields{"addr": addr.Hex()}).Warnf("error resolving name: %v", err)
			} else {
				return fmt.Errorf("error go_ens.Resolve(%v): %w", name, err)
			}
		}

		nameHash, err := go_ens.NameHash(name)
		if err != nil {
			return err
		}
		parts := strings.Split(name, ".")
		mainName := strings.Join(parts[len(parts)-2:], ".")
		ensName, err := go_ens.NewName(erigonClient.GetNativeClient(), mainName)
		if err != nil {
			return fmt.Errorf("error could not create name via go_ens.NewName for [%v]: %w", name, err)
		}
		expires, err := ensName.Expires()
		if err != nil {
			return fmt.Errorf("error could not get ens expire date for [%v]: %w", name, err)
		}

		logrus.WithFields(logrus.Fields{"resolvedAddr": resolvedAddr, "addr": addr.Hex(), "name": name, "nameHash": fmt.Sprintf("%#x", nameHash), "expires": expires}).Infof("checked ens entry")
	}

	return nil
}

func TransformEns(blk *types.Eth1Block, cache *freecache.Cache) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error) {
	mustInitDbs()
	// filterer, err := go_ens_contracts_resolver.NewContractFilterer(common.Address{}, nil)
	// if err != nil {
	// 	return nil, nil, err
	// }
	for i, tx := range blk.GetTransactions() {
		if i > 9999 {
			return nil, nil, fmt.Errorf("unexpected number of transactions in block expected at most 9999 but got: %v, tx: %x", i, tx.GetHash())
		}
		logs := tx.GetLogs()
		for _, log := range logs {
			for _, topic := range log.GetTopics() {
				_ = topic
			}
		}
	}
	return nil, nil, nil
}

func exportHistoricPrices(dayStart uint64, dayEnd uint64) {
	mustInitDbs()
	logrus.Infof("exporting historic prices for days %v - %v", dayStart, dayEnd)
	for day := dayStart; day <= dayEnd; day++ {
		timeStart := time.Now()
		ts := utils.DayToTime(int64(day)).UTC().Truncate(utils.Day)
		err := services.WriteHistoricPricesForDay(ts)
		if err != nil {
			errMsg := fmt.Sprintf("error exporting historic prices for day %v", day)
			utils.LogError(err, errMsg, 0)
			return
		}
		logrus.Printf("finished export for day %v, took %v", day, time.Since(timeStart))

		if day < dayEnd {
			// Wait to not overload the API
			time.Sleep(5 * time.Second)
		}
	}

	logrus.Info("historic price update run completed")
}

// ParseUint64Ranges parses a string of comma separated values and/or ranges into a sorted slice of unique uint64 values (e.g. "1,4,7,6-9" -> []uint64{1,4,6,7,8,9}).
func ParseUint64Ranges(ranges string) ([]uint64, error) {
	res := []uint64{}
	if ranges == "" {
		return res, nil
	}
	for _, s := range strings.Split(ranges, ",") {
		ss := strings.Split(s, "-")
		if len(ss) == 2 {
			u64a, err := strconv.ParseUint(ss[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid ranges: %v: %w", ranges, err)
			}
			u64b, err := strconv.ParseUint(ss[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid ranges: %v: %w", ranges, err)
			}
			if u64b < u64a {
				return nil, fmt.Errorf("invalid ranges: %v", ranges)
			}
			for i := u64a; i <= u64b; i++ {
				res = append(res, i)
			}
			continue
		}
		if len(ss) > 2 {
			return nil, fmt.Errorf("invalid ranges: %v", ranges)
		}
		u64, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		res = append(res, u64)
	}
	return SortedUniqueUint64(res), nil
}

// UniqueStrings returns an array of strings containing each value of s only once
func UniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := seen[str]; !ok {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

func SortedUniqueUint64(arr []uint64) []uint64 {
	if len(arr) <= 1 {
		return arr
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	result := make([]uint64, 1, len(arr))
	result[0] = arr[0]
	for i := 1; i < len(arr); i++ {
		if arr[i-1] != arr[i] {
			result = append(result, arr[i])
		}
	}

	return result
}

type ContractEvent []byte

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
var FilterApprovalContractEvent = []byte{140, 91, 225, 229, 235, 236, 125, 91, 209, 79, 113, 66, 125, 30, 132, 243, 221, 3, 20, 192, 247, 178, 41, 30, 91, 32, 10, 200, 199, 195, 185, 37}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
var FilterApprovalForAllContractEvent = []byte{23, 48, 126, 171, 57, 171, 97, 7, 232, 137, 152, 69, 173, 61, 89, 189, 150, 83, 242, 0, 242, 32, 146, 4, 137, 202, 43, 89, 55, 105, 108, 49}

// FilterControllerAdded is a free log retrieval operation binding the contract event 0x0a8bb31534c0ed46f380cb867bd5c803a189ced9a764e30b3a4991a9901d7474.
// FilterControllerRemoved is a free log retrieval operation binding the contract event 0x33d83959be2573f5453b12eb9d43b3499bc57d96bd2f067ba44803c859e81113.
// FilterNameMigrated is a free log retrieval operation binding the contract event 0xea3d7e1195a15d2ddcd859b01abd4c6b960fa9f9264e499a70a90c7f0c64b717.
// FilterNameRegistered is a free log retrieval operation binding the contract event 0xb3d987963d01b2f68493b4bdb130988f157ea43070d4ad840fee0466ed9370d9.
// FilterNameRenewed is a free log retrieval operation binding the contract event 0x9b87a00e30f1ac65d898f070f8a3488fe60517182d0a2098e1b4b93a54aa9bd6.
var FilterNameRenewedContractEvent = []byte{155, 135, 160, 14, 48, 241, 172, 101, 216, 152, 240, 112, 248, 163, 72, 143, 229, 5, 23, 24, 45, 10, 32, 152, 225, 180, 185, 58, 84, 170, 155, 214}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
