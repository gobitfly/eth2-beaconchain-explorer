package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/gobitfly/eth2-beaconchain-explorer/cmd/playground/pkg/ens"
	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/ratelimit"
	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/gobitfly/eth2-beaconchain-explorer/version"
)

var opts = struct {
	Command     string
	Version     bool
	Config      string
	Dry         bool
	Blocks      string
	BlocksStart uint64
	BlocksEnd   uint64
	EnsNames    string
}{}

var commands = map[string]func() error{
	"ensFindBlock":  ensFindBlock,
	"testRedis":     ensResolve,
	"findFirstBlob": ensTest,
	"checkS3Blobs":  ratelimitsTest,
	"debugBlock":    debugBlock,
}

var bt *db.Bigtable
var erigonClient *rpc.ErigonClient
var lighthouseClient *rpc.LighthouseClient
var rpcClient *rpc.LighthouseClient

func main() {
	flag.StringVar(&opts.Command, "command", "ratelimitsTest", "command to run")
	flag.BoolVar(&opts.Version, "version", false, "print version")
	flag.StringVar(&opts.Config, "config", "config.yaml", "config file")
	flag.BoolVar(&opts.Dry, "dry", false, "dry run")
	flag.StringVar(&opts.Blocks, "blocks", "", "blocks to process (e.g. 1,4,7,6-9 -> []uint64{1,4,6,7,8,9})")
	flag.Uint64Var(&opts.BlocksStart, "blocks.start", 0, "start")
	flag.Uint64Var(&opts.BlocksEnd, "blocks.end", 0, "end")
	flag.StringVar(&opts.EnsNames, "ensNames", "", "ens names to process")
	flag.Parse()

	if opts.Version {
		fmt.Println(version.Version)
		return
	}

	logrus.Infof("command: %v", opts.Command)

	if opts.Command == "" {
		logrus.Fatal(nil, "no command specified", 0)
	}
	cmd := commands[opts.Command]
	if cmd == nil {
		logrus.Fatal(nil, fmt.Sprintf("unknown command: %s", opts.Command), 0)
	}
	err := cmd()
	if err != nil {
		logrus.Fatal(err, fmt.Sprintf("%v: error: %v", opts.Command, err), 0)
	}
	logrus.Infof("%v: done", opts.Command)
}

func debugBlock() error {
	var err error

	logrus.WithField("config", opts.Config).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err = utils.ReadConfig(cfg, opts.Config)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	erigonClient, err = rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		logrus.Fatalf("error initializing erigon client: %v", err)
	}

	gnosisURL := "http://127.0.0.1:18545"
	gnosisURL = "http://gno-node-mainnet.beaconcha.in:18545"
	tests := []struct {
		block int64
	}{
		{block: 68140},
		{block: 19187811},
		{block: 37591835}, // last block works
		{block: 38039324},
	}
	for _, tt := range tests {
		fmt.Printf("-- block: %v\n", tt.block)
		a, _, err := erigonClient.GetBlock(tt.block, "parity/geth")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%-20s%#x\n", "erigonClient", a.Hash)

		c, _ := ethclient.Dial(gnosisURL)
		block, err := c.BlockByNumber(context.Background(), big.NewInt(tt.block))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%-20s%s\n", "go-eth", block.Hash().String())

		type MinimalBlock struct {
			Result struct {
				Hash string `json:"hash"`
			} `json:"result"`
		}
		query := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params": ["0x%x",false],"id":1}`, tt.block)
		resp, err := http.Post(gnosisURL, "application/json", bytes.NewBufferString(query))
		if err != nil {
			panic(err)
		}
		var res MinimalBlock
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			panic(err)
		}
		fmt.Printf("%-20s%s\n", "rpc", res.Result.Hash)
	}
	return nil
}

/*
func debugBlocks() error {
	cl := consapi.NewClient("http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port)
	nodeImpl, ok := cl.ClientInt.(*consapi.NodeClient)
	if !ok {
		log.Fatal(nil, "lighthouse client can only be used with real node impl", 0)
	}
	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)
	rpcClient, err = rpc.NewLighthouseClient(nodeImpl, chainIDBig)
	if err != nil {
		log.Fatal(err, "lighthouse client error", 0)
	}

	elClient, err := rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		return err
	}

	for i := opts.BlocksStart; i <= opts.BlocksEnd; i++ {
		btBlock, err := db.BigtableClient.GetBlockFromBlocksTable(i)
		if err != nil {
			return err
		}

		elBlock, _, err := elClient.GetBlock(int64(i), "parity/geth")
		if err != nil {
			return err
		}

		slot := utils.TimeToSlot(uint64(elBlock.Time.Seconds))
		clBlock, err := clClient.GetBlockBySlot(slot)
		if err != nil {
			return err
		}
		logFields := log.Fields{
			"block":            i,
			"bt.hash":          fmt.Sprintf("%#x", btBlock.Hash),
			"bt.BlobGasUsed":   btBlock.BlobGasUsed,
			"bt.ExcessBlobGas": btBlock.ExcessBlobGas,
			"bt.txs":           len(btBlock.Transactions),
			"el.BlobGasUsed":   elBlock.BlobGasUsed,
			"el.hash":          fmt.Sprintf("%#x", elBlock.Hash),
			"el.ExcessBlobGas": elBlock.ExcessBlobGas,
			"el.txs":           len(elBlock.Transactions),
		}
		if !bytes.Equal(clBlock.ExecutionPayload.BlockHash, elBlock.Hash) {
			log.Warnf("clBlock.ExecutionPayload.BlockHash != i: %x != %x", clBlock.ExecutionPayload.BlockHash, elBlock.Hash)
		} else if clBlock.ExecutionPayload.BlockNumber != i {
			log.Warnf("clBlock.ExecutionPayload.BlockNumber != i: %v != %v", clBlock.ExecutionPayload.BlockNumber, i)
		} else {
			logFields["cl.txs"] = len(clBlock.ExecutionPayload.Transactions)
		}

		log.InfoWithFields(logFields, "debug block")

		for i := range elBlock.Transactions {
			btx := btBlock.Transactions[i]
			ctx := elBlock.Transactions[i]
			btxH := []string{}
			ctxH := []string{}
			for _, h := range btx.BlobVersionedHashes {
				btxH = append(btxH, fmt.Sprintf("%#x", h))
			}
			for _, h := range ctx.BlobVersionedHashes {
				ctxH = append(ctxH, fmt.Sprintf("%#x", h))
			}

			log.InfoWithFields(log.Fields{
				"b.hash":                 fmt.Sprintf("%#x", btx.Hash),
				"el.hash":                fmt.Sprintf("%#x", ctx.Hash),
				"b.BlobVersionedHashes":  fmt.Sprintf("%+v", btxH),
				"el.BlobVersionedHashes": fmt.Sprintf("%+v", ctxH),
				"b.maxFeePerBlobGas":     btx.MaxFeePerBlobGas,
				"el.maxFeePerBlobGas":    ctx.MaxFeePerBlobGas,
				"b.BlobGasPrice":         btx.BlobGasPrice,
				"el.BlobGasPrice":        ctx.BlobGasPrice,
				"b.BlobGasUsed":          btx.BlobGasUsed,
				"el.BlobGasUsed":         ctx.BlobGasUsed,
			}, "debug tx")

			for ii := range ctx.Itx {
				bitx := btx.Itx[ii]
				citx := ctx.Itx[ii]
				if !bytes.Equal(bitx.Value, citx.Value) {
					log.Warnf("value mismatch at itx %d: %v != %v", ii, bitx.Value, citx.Value)
				}
			}
		}
	}
	return nil
}
*/

func ensFindBlock() error {
	logrus.WithField("config", opts.Config).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, opts.Config)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	erigonClient, err = rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		logrus.Fatalf("error initializing erigon client: %v", err)
	}
	ens.ErigonClient = erigonClient
	err = ens.FindEnsChangeBlock(opts.EnsNames, opts.BlocksStart, opts.BlocksEnd)
	if err != nil {
		logrus.Fatalf("error finding ens change block: %v", err)
	}
	return nil
}

func ensResolve() error {
	logrus.WithField("config", opts.Config).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, opts.Config)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	erigonClient, err = rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		logrus.Fatalf("error initializing erigon client: %v", err)
	}
	ens.ErigonClient = erigonClient
	ens.Resolve(opts.EnsNames)
	return nil
}

func ensTest() error {
	logrus.WithField("config", opts.Config).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, opts.Config)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	chainIdString := strconv.FormatUint(utils.Config.Chain.ClConfig.DepositChainID, 10)
	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)

	wg := &sync.WaitGroup{}
	wg.Add(5)

	go func() {
		defer wg.Done()
		var err error
		bt, err = db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, chainIdString, utils.Config.RedisCacheEndpoint)
		if err != nil {
			utils.LogFatal(err, "error initializing bigtable", 0)
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		rpcClient, err = rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainIDBig)
		if err != nil {
			utils.LogFatal(err, "lighthouse client error", 0)
		}
		lighthouseClient = rpcClient
	}()

	go func() {
		defer wg.Done()
		var err error
		erigonClient, err = rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
		if err != nil {
			logrus.Fatalf("error initializing erigon client: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		db.MustInitDB(&types.DatabaseConfig{
			Username:     cfg.WriterDatabase.Username,
			Password:     cfg.WriterDatabase.Password,
			Name:         cfg.WriterDatabase.Name,
			Host:         cfg.WriterDatabase.Host,
			Port:         cfg.WriterDatabase.Port,
			MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
			SSL:          cfg.WriterDatabase.SSL,
		}, &types.DatabaseConfig{
			Username:     cfg.ReaderDatabase.Username,
			Password:     cfg.ReaderDatabase.Password,
			Name:         cfg.ReaderDatabase.Name,
			Host:         cfg.ReaderDatabase.Host,
			Port:         cfg.ReaderDatabase.Port,
			MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
			SSL:          cfg.ReaderDatabase.SSL,
		}, "pgx", "postgres")
	}()

	go func() {
		defer wg.Done()
		db.MustInitFrontendDB(&types.DatabaseConfig{
			Username:     cfg.Frontend.WriterDatabase.Username,
			Password:     cfg.Frontend.WriterDatabase.Password,
			Name:         cfg.Frontend.WriterDatabase.Name,
			Host:         cfg.Frontend.WriterDatabase.Host,
			Port:         cfg.Frontend.WriterDatabase.Port,
			MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
			SSL:          cfg.Frontend.WriterDatabase.SSL,
		}, &types.DatabaseConfig{
			Username:     cfg.Frontend.ReaderDatabase.Username,
			Password:     cfg.Frontend.ReaderDatabase.Password,
			Name:         cfg.Frontend.ReaderDatabase.Name,
			Host:         cfg.Frontend.ReaderDatabase.Host,
			Port:         cfg.Frontend.ReaderDatabase.Port,
			MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
			SSL:          cfg.Frontend.ReaderDatabase.SSL,
		}, "pgx", "postgres")
	}()

	wg.Wait()

	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	bt, err = db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, chainIdString, utils.Config.RedisCacheEndpoint)
	if err != nil {
		utils.LogFatal(err, "error initializing bigtable", 0)
	}

	if false {
		ensName, err := ens.NewName(erigonClient.GetNativeClient(), "financialfreedom.eth")
		if err != nil {
			utils.LogFatal(err, "error creating ens name", 0)
		}
		fmt.Println(ensName)
		return nil
	}

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit
	transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)
	transforms = append(transforms, bt.TransformEnsNameRegistered)
	transforms2 := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)
	transforms2 = append(transforms2, ens.TransformEnsNameRegistered2)

	blocks, err := ParseUint64Ranges(opts.Blocks)
	if err != nil {
		utils.LogFatal(err, "error parsing blocks", 0)
	}

	batchSize := uint64(1000)

	for i := blocks[0]; i <= blocks[len(blocks)-1]; i += batchSize {
		logrus.Infof("checking blocks %v-%v", i, i+batchSize)
		keys1, err := ens.IndexEventsWithTransformersDry(int64(i), int64(i+batchSize), transforms, 2, cache)
		if err != nil {
			utils.LogFatal(err, "error indexing events [1]", 0)
		}
		if false {
			keys2, err := ens.IndexEventsWithTransformersDry(int64(i), int64(i+batchSize), transforms2, 2, cache)
			if err != nil {
				utils.LogFatal(err, "error indexing events [2]", 0)
			}
			_ = keys2
		}
		for _, k := range keys1 {
			fmt.Println(k)
		}
	}

	return nil
}

func ratelimitsTest() error {
	utils.Config = &types.Config{}
	utils.Config.RedisSessionStoreEndpoint = "redis:6379"
	utils.Config.Frontend.RatelimitUpdateInterval = time.Second * 2
	dbConfig := &types.DatabaseConfig{
		Username:     "user",
		Password:     "pass",
		Name:         "db",
		Host:         "postgres",
		Port:         "5432",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		SSL:          false,
	}

	db.MustInitDB(dbConfig, dbConfig, "pgx", "postgres")
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()
	db.MustInitFrontendDB(dbConfig, dbConfig, "pgx", "postgres")
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	logrus.Infof("applying db schema")
	err := db.ApplyEmbeddedDbSchema(-2)
	if err != nil {
		logrus.WithError(err).Fatal("error applying db schema")
	}
	logrus.Infof("db schema applied successfully")

	_, err = db.FrontendWriterDB.Exec(`
		INSERT INTO api_products (name, bucket, stripe_price_id, second, hour, month, valid_from) VALUES
			('nokey'    , 'default', ''              ,   2,   0,       100, now()),
			('free'     , 'default', ''              ,   5,   0,     30000, now()),
			('unlimited', 'default', ''              , 100,   0,         0, now()),
			('sapphire' , 'default', 'price_sapphire',  10,   0,    500000, now()),
			('emerald'  , 'default', 'price_emerald' ,  10,   0,   1000000, now()),
			('diamond'  , 'default', 'price_diamond' ,  30,   0,   6000000, now()),
			('custom2'  , 'default', 'price_custom2' ,  50,   0,  13000000, now()),
			('custom1'  , 'default', 'price_custom1' ,  50,   0, 500000000, now()),

			('nokey'    , 'machine', ''              ,   1,   0,       100, now()),
			('free'     , 'machine', ''              ,   1,   0,       100, now()),
			('unlimited', 'machine', ''              , 100,   0,         0, now()),
			('whale'    , 'machine', 'price_whale'   ,  25,   0,    700000, now()),
			('goldfish' , 'machine', 'price_goldfish',  20,   0,    200000, now()),
			('plankton' , 'machine', 'price_plankton',  20,   0,    120000, now())
		on conflict (name, bucket, valid_from) do nothing;`)
	if err != nil {
		logrus.WithError(err).Fatal("error inserting api products")
	}

	_, err = db.FrontendWriterDB.Exec(`
		INSERT INTO api_weights (bucket, endpoint, method, params, weight, valid_from) VALUES
			('machine', '/api/v1/client/metrics', 'POST', '', 1, now())
		ON CONFLICT (endpoint, valid_from) DO NOTHING;`)
	if err != nil {
		logrus.WithError(err).Fatal("error inserting into api_weights")
	}

	_, err = db.FrontendWriterDB.Exec(`
		INSERT INTO users (email, password, api_key, stripe_customer_id) 
		SELECT 'user'||i||'@email.com', 'xxx', 'apikey_'||i, 'stripe_customer_'||i
		FROM generate_series(1, 100000-(select count(*) from users)) as s(i)
		ON CONFLICT DO NOTHING;`)
	if err != nil {
		logrus.WithError(err).Fatal("error inserting api products")
	}

	_, err = db.FrontendWriterDB.Exec(`
		INSERT INTO users_stripe_subscriptions (subscription_id, customer_id, price_id, active, payload, purchase_group) 
		select 'stripe_sub_'||i, 'stripe_customer_'||i, 'price_diamond', true, '{}', 'x'
		from generate_series(1, 10000) as s(i)
		ON CONFLICT (customer_id, subscription_id, price_id) DO NOTHING;`)
	if err != nil {
		logrus.WithError(err).Fatal("error inserting api products")
	}

	_, err = db.FrontendWriterDB.Exec(`
		INSERT INTO users_stripe_subscriptions (subscription_id, customer_id, price_id, active, payload, purchase_group) 
		select 'stripe_sub_'||i+10000, 'stripe_customer_'||i, 'price_whale', true, '{}', 'x'
		from generate_series(10000, 10010) as s(i)
		ON CONFLICT (customer_id, subscription_id, price_id) DO NOTHING;`)
	if err != nil {
		logrus.WithError(err).Fatal("error inserting api products")
	}

	router := mux.NewRouter()
	router.HandleFunc("/", defaultHandler).Methods("GET")
	router.HandleFunc("/api/i/a", defaultHandler).Methods("GET")
	router.HandleFunc("/api/v1/a", defaultHandler).Methods("GET")
	router.HandleFunc("/api/v1/client/metrics", defaultHandler).Methods("POST")
	ratelimit.DefaultRequestFilter = func(req *http.Request) bool {
		if req.URL == nil || !strings.HasPrefix(req.URL.Path, "/api") || strings.HasPrefix(req.URL.Path, "/api/i/") || strings.HasPrefix(req.URL.Path, "/api/v1/docs/") || strings.HasPrefix(req.URL.Path, "/api/v2/docs/") {
			return false
		}
		return true
	}
	ratelimit.Init()
	router.Use(ratelimit.HttpMiddleware)
	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second * 10,
		Handler:      router,
	}
	go func() {
		logrus.Infof("listening on %v", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			logrus.WithError(err).Fatal("Error serving frontend")
		}
	}()
	go func() {
		logrus.Infof("starting ratelimit-updater")
		ratelimit.DBUpdater()
	}()

	logrus.Infof("waiting for limits to be get applied")
	for {
		res, err := checkRatelimit("http://localhost:8080/api/v1/a?apikey=apikey_10000", "GET", false)
		if err != nil {
			logrus.Errorf("error checking ratelimit: %v", err)
			time.Sleep(time.Millisecond * 1000)
			continue
		}
		if res.RatelimitLimitMonth != "6000000" {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		break
	}

	checkRatelimit("http://localhost:8080/api/v1/a?apikey=apikey_10000", "GET", true)
	time.Sleep(time.Millisecond * 1200)
	checkRatelimit("http://localhost:8080/api/v1/client/metrics?apikey=apikey_10000", "POST", true)

	checkRatelimit("http://localhost:8080/api/v1/a?apikey=apikey_20000", "GET", true)

	if false {
		checkRatelimit("http://localhost:8080", "GET", true)
		checkRatelimit("http://localhost:8080/api/i/a?apikey=x", "GET", true)
		for i := 0; i < 110; i++ {
			checkRatelimit("http://localhost:8080/api/v1/a?apikey=x", "GET", true)
			time.Sleep(time.Millisecond * 200)
		}
	}

	return nil
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"hello":"world"}`))
}

type checkRatelimitRes struct {
	StatusCode               int
	RatelimitRemainingSecond string
	RatelimitLimitSecond     string
	RatelimitRemainingMonth  string
	RatelimitLimitMonth      string
}

func checkRatelimit(url, method string, debug bool) (*checkRatelimitRes, error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request with params: %w, url: %v", err, url)
	}

	httpClient := &http.Client{Timeout: time.Minute}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to %v: %w", url, err)
	}

	if debug {
		fmt.Printf("%-30s: %v %v:%v/%v se:%v/%v mi:%v/%v ho:%v/%v da:%v/%v mo:%v/%v\n",
			url, res.StatusCode,
			res.Header.Get("ratelimit-window"), res.Header.Get("ratelimit-remaining"), res.Header.Get("ratelimit-limit"),
			res.Header.Get("x-ratelimit-remaining-second"), res.Header.Get("x-ratelimit-limit-second"),
			res.Header.Get("x-ratelimit-remaining-minute"), res.Header.Get("x-ratelimit-limit-minute"),
			res.Header.Get("x-ratelimit-remaining-hour"), res.Header.Get("x-ratelimit-limit-hour"),
			res.Header.Get("x-ratelimit-remaining-day"), res.Header.Get("x-ratelimit-limit-day"),
			res.Header.Get("x-ratelimit-remaining-month"), res.Header.Get("x-ratelimit-limit-month"),
		)
	}

	// for k, v := range res.Header {
	// 	fmt.Println(k, v)
	// }
	return &checkRatelimitRes{
		StatusCode:               res.StatusCode,
		RatelimitRemainingSecond: res.Header.Get("x-ratelimit-remaining-second"),
		RatelimitLimitSecond:     res.Header.Get("x-ratelimit-limit-second"),
		RatelimitRemainingMonth:  res.Header.Get("x-ratelimit-remaining-month"),
		RatelimitLimitMonth:      res.Header.Get("x-ratelimit-limit-month"),
	}, nil
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
