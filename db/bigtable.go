package db

import (
	"context"
	"errors"
	"eth2-exporter/erc20"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"strconv"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"

	"github.com/ethereum/go-ethereum/common"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

var ErrBlockNotFound = errors.New("block not found")
var BigtableClient *Bigtable

const max_block_number = 1000000000
const (
	DATA_COLUMN    = "data"
	DEFAULT_FAMILY = "default"
	writeRowLimit  = 10000
)

type Bigtable struct {
	client      *gcp_bigtable.Client
	tableData   *gcp_bigtable.Table
	tableBlocks *gcp_bigtable.Table
	tableStefan *gcp_bigtable.Table
	chainId     string
}

func NewBigtable(project, instance, chainId string) (*Bigtable, error) {
	poolSize := 20
	btClient, err := gcp_bigtable.NewClient(context.Background(), project, instance, option.WithGRPCConnectionPool(poolSize))
	// btClient, err := gcp_bigtable.NewClient(context.Background(), project, instance)

	if err != nil {
		return nil, err
	}

	bt := &Bigtable{
		client:      btClient,
		tableData:   btClient.Open("data"),
		tableBlocks: btClient.Open("blocks"),
		tableStefan: btClient.Open("stefan"),
		chainId:     chainId,
	}
	return bt, nil
}

func (bigtable *Bigtable) Close() {
	bigtable.client.Close()
}

func (bigtable *Bigtable) SaveBlock(block *types.Eth1Block) error {

	encodedBc, err := proto.Marshal(block)

	if err != nil {
		return err
	}
	ts := gcp_bigtable.Timestamp(0)

	mut := gcp_bigtable.NewMutation()
	mut.Set(DEFAULT_FAMILY, "data", ts, encodedBc)

	err = bigtable.tableBlocks.Apply(context.Background(), fmt.Sprintf("%s:%s", bigtable.chainId, reversedPaddedBlockNumber(block.Number)), mut)

	if err != nil {
		return err
	}
	return nil
}

func (bigtable *Bigtable) SaveBlocks(block *types.Eth1Block) error {

	encodedBc, err := proto.Marshal(block)

	if err != nil {
		return err
	}
	ts := gcp_bigtable.Timestamp(0)

	mut := gcp_bigtable.NewMutation()
	mut.Set(DEFAULT_FAMILY, "data", ts, encodedBc)

	err = bigtable.tableBlocks.Apply(context.Background(), fmt.Sprintf("%s:%s", bigtable.chainId, reversedPaddedBlockNumber(block.Number)), mut)

	if err != nil {
		return err
	}
	return nil
}

func (bigtable *Bigtable) GetBlock(number uint64) (*types.Eth1Block, error) {

	paddedNumber := reversedPaddedBlockNumber(number)

	row, err := bigtable.tableBlocks.ReadRow(context.Background(), fmt.Sprintf("1:%s", paddedNumber))

	if err != nil {
		return nil, err
	}

	if len(row[DEFAULT_FAMILY]) == 0 { // block not found
		return nil, ErrBlockNotFound
	}

	bc := &types.Eth1Block{}
	err = proto.Unmarshal(row[DEFAULT_FAMILY][0].Value, bc)

	if err != nil {
		return nil, err
	}

	return bc, nil
}

func (bigtable *Bigtable) GetFullBlock(number uint64) (*types.Eth1Block, error) {

	paddedNumber := reversedPaddedBlockNumber(number)

	row, err := bigtable.tableStefan.ReadRow(context.Background(), fmt.Sprintf("1:%s", paddedNumber))

	if err != nil {
		return nil, err
	}

	if len(row[DEFAULT_FAMILY]) == 0 { // block not found
		return nil, ErrBlockNotFound
	}
	blocks := make([]*types.Eth1Block, 0, 1)
	rowHandler := GetFullBlockHandler(&blocks)

	rowHandler(row)

	if err != nil {
		return nil, err
	}

	return blocks[0], nil
}

func (bigtable *Bigtable) GetMostRecentBlock() (*types.Eth1Block, error) {
	ctx, cancle := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancle()

	prefix := fmt.Sprintf("%s:", bigtable.chainId)

	rowRange := gcp_bigtable.PrefixRange(prefix)
	rowFilter := gcp_bigtable.RowFilter(gcp_bigtable.ColumnFilter("block"))
	limit := gcp_bigtable.LimitRows(1)

	block := types.Eth1Block{}

	rowHandler := func(row gcp_bigtable.Row) bool {
		item := row[DEFAULT_FAMILY][0]

		err := proto.Unmarshal(item.Value, &block)
		if err != nil {
			logger.Errorf("error could not unmarschal proto object, err: %v", err)
		}

		return true
	}

	err := bigtable.tableStefan.ReadRows(ctx, rowRange, rowHandler, rowFilter, limit)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func GetBlockHandler(blocks *[]*types.Eth1BlockIndexed) func(gcp_bigtable.Row) bool {
	return func(row gcp_bigtable.Row) bool {
		// startTime := time.Now()
		block := types.Eth1BlockIndexed{}
		err := proto.Unmarshal(row[DEFAULT_FAMILY][0].Value, &block)
		if err != nil {
			logger.Errorf("error could not unmarschal proto object, err: %v", err)
		}
		*blocks = append(*blocks, &block)
		// logger.Infof("finished processing row from table blocks: %v", time.Since(startTime))
		return true
	}
}

func GetFullBlockHandler(blocks *[]*types.Eth1Block) func(gcp_bigtable.Row) bool {
	return func(row gcp_bigtable.Row) bool {
		// startTime := time.Now()
		block := types.Eth1Block{}
		err := proto.Unmarshal(row[DEFAULT_FAMILY][0].Value, &block)
		if err != nil {
			logger.Errorf("error could not unmarschal proto object, err: %v", err)
		}
		if len(row[DEFAULT_FAMILY]) > 1 {
			logs := make(map[string][]*types.Eth1Log, 100)
			itxs := make(map[string][]*types.Eth1InternalTransaction, 100)
			for _, item := range row[DEFAULT_FAMILY][1:] {
				if strings.HasPrefix(item.Column, "default:itx") {
					hash := strings.Split(item.Column, ":")[2]
					itx := types.Eth1InternalTransaction{}
					err := proto.Unmarshal(item.Value, &itx)
					if err != nil {
						logger.Errorf("error could not unmarschal proto object, err: %v", err)
					}
					itxs[hash] = append(itxs[hash], &itx)
				}
				if strings.HasPrefix(item.Column, "default:log") {
					hash := strings.Split(item.Column, ":")[2]
					log := types.Eth1Log{}
					err := proto.Unmarshal(item.Value, &log)
					if err != nil {
						logger.Errorf("error could not unmarschal proto object, err: %v", err)
					}
					logs[hash] = append(logs[hash], &log)
				}
				if strings.HasPrefix(item.Column, "default:tx") {
					hash := strings.Split(item.Column, ":")[3]
					tx := types.Eth1Transaction{}
					err := proto.Unmarshal(item.Value, &tx)
					if err != nil {
						logger.Errorf("error could not unmarschal proto object, err: %v", err)
					}
					tx.Logs = logs[hash]
					tx.Itx = itxs[hash]
					block.Transactions = append(block.Transactions, &tx)
				}
			}

		}

		*blocks = append(*blocks, &block)
		// logger.Infof("finished processing row from table stefan: %v", time.Since(startTime))

		return true
	}
}

// GetFullBlockDescending gets blocks starting at block start
func (bigtable *Bigtable) GetFullBlockDescending(start, limit uint64) ([]*types.Eth1Block, error) {
	startPadded := reversedPaddedBlockNumber(start)
	ctx, cancle := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancle()

	prefix := fmt.Sprintf("%s:%s", bigtable.chainId, startPadded)

	rowRange := gcp_bigtable.InfiniteRange(prefix) //gcp_bigtable.PrefixRange("1:1000000000")

	blocks := make([]*types.Eth1Block, 0, 100)

	rowHandler := GetFullBlockHandler(&blocks)

	startTime := time.Now()
	err := bigtable.tableStefan.ReadRows(ctx, rowRange, rowHandler, gcp_bigtable.LimitRows(int64(limit)))
	if err != nil {
		return nil, err
	}

	logger.Infof("finished getting blocks from table stefan: %v", time.Since(startTime))
	return blocks, nil
}

// GetBlocksDescending gets blocks starting at block start
func (bigtable *Bigtable) GetBlocksDescending(start, limit uint64) ([]*types.Eth1BlockIndexed, error) {
	startPadded := reversedPaddedBlockNumber(start)
	ctx, cancle := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancle()

	prefix := fmt.Sprintf("%s:b:%s", bigtable.chainId, startPadded)

	rowRange := gcp_bigtable.InfiniteRange(prefix) //gcp_bigtable.PrefixRange("1:1000000000")
	rowFilter := gcp_bigtable.RowFilter(gcp_bigtable.ColumnFilter("data"))

	blocks := make([]*types.Eth1BlockIndexed, 0, 100)

	rowHandler := GetBlockHandler(&blocks)

	startTime := time.Now()
	err := bigtable.tableData.ReadRows(ctx, rowRange, rowHandler, rowFilter, gcp_bigtable.LimitRows(int64(limit)))
	if err != nil {
		return nil, err
	}

	logger.Infof("finished getting blocks from table stefan: %v", time.Since(startTime))
	return blocks, nil
}

func reversedPaddedBlockNumber(blockNumber uint64) string {
	return fmt.Sprintf("%09d", max_block_number-blockNumber)
}

func blockFromPaddedBlockNumber(paddedBlockNumber string) uint64 {
	num := strings.Split(paddedBlockNumber, ":")
	paddedNumber, err := strconv.ParseUint(num[1], 10, 64)
	if err != nil {
		logger.WithError(err).Error("error parsing padded block")
		return 0
	}

	return uint64(max_block_number) - paddedNumber
}

func (bigtable *Bigtable) WriteBulk(mutations *types.BulkMutations) error {
	length := 10000
	numMutations := len(mutations.Muts)
	numKeys := len(mutations.Keys)
	iterations := numKeys / length
	g := new(errgroup.Group)
	g.SetLimit(iterations + 1)

	if numKeys != numMutations {
		return fmt.Errorf("error expected same number of keys as mutations keys: %v mutations: %v", numKeys, numMutations)
	}

	for offset := 0; offset < iterations; offset++ {
		g.Go(func() error {
			start := offset * length
			end := offset*length + length
			ctx, done := context.WithTimeout(context.Background(), time.Second*30)
			defer done()
			// startTime := time.Now()
			errs, err := bigtable.tableData.ApplyBulk(ctx, mutations.Keys[start:end], mutations.Muts[start:end])
			for _, e := range errs {
				if e != nil {
					return err
				}
			}
			// logrus.Infof("wrote from %v to %v rows to bigtable in %.1f s", start, end, time.Since(startTime).Seconds())
			if err != nil {
				return err
			}
			return nil
		})
	}

	if (iterations * length) < numKeys {
		g.Go(func() error {
			start := iterations * length

			ctx, done := context.WithTimeout(context.Background(), time.Second*30)
			defer done()
			// startTime := time.Now()
			errs, err := bigtable.tableData.ApplyBulk(ctx, mutations.Keys[start:], mutations.Muts[start:])
			for _, e := range errs {
				if e != nil {
					return err
				}
			}
			// logrus.Infof("wrote from %v to %v rows to bigtable in %.1fs", start, numKeys, time.Since(startTime).Seconds())
			if err != nil {
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err == nil {
		// logrus.Info("Successfully wrote all mutations")
		return nil
	} else {
		return err
	}
}

func (bigtable *Bigtable) TransformBlock(blocks []*types.Eth1Block) (*types.BulkMutations, error) {
	muts := &types.BulkMutations{}

	block := blocks[0]
	previous := blocks[1]

	// logger.Infof("getting blocks block: %v, previous: %v time: %v, time: %v, diff: %v", block.GetNumber(), previous.GetNumber(), block.GetTime().AsTime().Unix(), previous.GetTime().AsTime().Unix(), block.GetTime().AsTime().Unix()-previous.GetTime().AsTime().Unix())

	idx := types.Eth1BlockIndexed{
		Hash:                   block.GetHash(),
		ParentHash:             block.GetParentHash(),
		UncleHash:              block.GetUncleHash(),
		Coinbase:               block.GetCoinbase(),
		Root:                   block.GetRoot(),
		TxHash:                 block.GetTxHash(),
		ReceiptHash:            block.GetReceiptHash(),
		Difficulty:             block.GetDifficulty(),
		Number:                 block.GetNumber(),
		GasLimit:               block.GetGasLimit(),
		GasUsed:                block.GetGasUsed(),
		Time:                   block.GetTime(),
		Extra:                  block.GetExtra(),
		MixDigest:              block.GetMixDigest(),
		Bloom:                  block.GetBloom(),
		BaseFee:                block.GetBaseFee(),
		Duration:               uint64(block.GetTime().AsTime().Unix() - previous.GetTime().AsTime().Unix()),
		UncleCount:             uint64(len(block.GetUncles())),
		TransactionCount:       uint64(len(block.GetTransactions())),
		BaseFeeChange:          new(big.Int).Sub(new(big.Int).SetBytes(block.GetBaseFee()), new(big.Int).SetBytes(previous.GetBaseFee())).Bytes(),
		BlockUtilizationChange: new(big.Int).Sub(new(big.Int).Div(big.NewInt(int64(block.GetGasUsed())), big.NewInt(int64(block.GetGasLimit()))), new(big.Int).Div(big.NewInt(int64(previous.GetGasUsed())), big.NewInt(int64(previous.GetGasLimit())))).Bytes(),
	}

	uncleReward := big.NewInt(0)
	r := new(big.Int)

	for _, uncle := range block.Uncles {
		r.Add(big.NewInt(int64(uncle.GetNumber())), big.NewInt(8))
		r.Sub(r, big.NewInt(int64(block.GetNumber())))
		r.Mul(r, utils.BlockReward(block.GetNumber()))
		r.Div(r, big.NewInt(8))

		r.Div(utils.BlockReward(block.GetNumber()), big.NewInt(32))
		uncleReward.Add(uncleReward, r)
	}

	idx.UncleReward = uncleReward.Bytes()

	var maxGasPrice *big.Int
	var minGasPrice *big.Int
	txReward := big.NewInt(0)

	for _, t := range block.GetTransactions() {
		price := new(big.Int).SetBytes(t.GasPrice)

		if minGasPrice == nil {
			minGasPrice = price
		}
		if maxGasPrice == nil {
			maxGasPrice = price
		}

		if price.Cmp(maxGasPrice) > 0 {
			maxGasPrice = price
		}

		if price.Cmp(minGasPrice) < 0 {
			minGasPrice = price
		}

		txReward.Add(new(big.Int).Mul(big.NewInt(int64(t.GasUsed)), new(big.Int).SetBytes(t.GasPrice)), txReward)
	}

	idx.TxReward = txReward.Bytes()

	if maxGasPrice != nil {
		idx.LowestGasPrice = minGasPrice.Bytes()

	}
	if minGasPrice != nil {
		idx.HighestGasPrice = maxGasPrice.Bytes()
	}

	idx.Mev = CalculateMevFromBlock(block).Bytes()

	// <chainID>:b:<reverse number>
	key := fmt.Sprintf("%s:b:%s", bigtable.chainId, reversedPaddedBlockNumber(block.GetNumber()))
	mut := gcp_bigtable.NewMutation()

	b, err := proto.Marshal(&idx)
	if err != nil {
		return nil, fmt.Errorf("error marshalling proto object err: %w", err)
	}

	mut.Set(DEFAULT_FAMILY, DATA_COLUMN, gcp_bigtable.Timestamp(0), b)

	muts.Keys = append(muts.Keys, key)
	muts.Muts = append(muts.Muts, mut)

	return muts, nil
}

func CalculateMevFromBlock(block *types.Eth1Block) *big.Int {
	mevReward := big.NewInt(0)

	for i, tx := range block.GetTransactions() {

		if strings.ToLower(fmt.Sprintf("0x%x", tx.GetFrom())) == "0xf6da21e95d74767009accb145b96897ac3630bad" {
			if strings.ToLower(fmt.Sprintf("0x%x", tx.GetTo())) == "0x0e09142e36e6dc1d2bb339e02b95bb624f0673c2" || strings.ToLower(fmt.Sprintf("0x%x", tx.GetTo())) == "0xd78a3280085ee846196cb5fab7d510b279486d44" { // ethermine mev arb contract
				for j, l := range tx.GetLogs() {
					if common.BytesToAddress(l.Address) != common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2") {
						continue
					}
					if fmt.Sprintf("0x%x", l.Topics[0]) == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
						filterer, err := erc20.NewErc20Filterer(common.Address{}, nil)
						if err != nil {
							log.Printf("error unpacking log: %v", err)
							break
						}

						topics := make([]common.Hash, 0, len(l.GetTopics()))

						for _, lTopic := range l.GetTopics() {
							topics = append(topics, common.BytesToHash(lTopic))
						}

						log := eth_types.Log{
							Address:     common.BytesToAddress(l.Address),
							Data:        l.Data,
							Topics:      topics,
							BlockNumber: block.GetNumber(),
							TxHash:      common.BytesToHash(tx.GetHash()),
							TxIndex:     uint(i),
							BlockHash:   common.BytesToHash(block.GetHash()),
							Index:       uint(j),
							Removed:     l.GetRemoved(),
						}

						t, err := filterer.ParseTransfer(log)
						if err != nil {
							logrus.Infof("error unpacking log: %v", err)
							break
						}
						if t.From == common.HexToAddress("0xf6da21e95d74767009accb145b96897ac3630bad") {
							logrus.Infof("tx %v subtracting %v to mev profit via mempool arb", tx.Hash, t.Value)
							mevReward = new(big.Int).Sub(mevReward, t.Value)
						}
						if t.To == common.HexToAddress("0xf6da21e95d74767009accb145b96897ac3630bad") {
							logrus.Infof("tx %v adding %v to mev profit via mempool arb", tx.Hash, t.Value)
							mevReward = new(big.Int).Add(mevReward, t.Value)
						}
					}
				}
			}
		}

		for _, itx := range tx.GetItx() {
			//log.Printf("%v - %v", common.HexToAddress(itx.To), common.HexToAddress(block.Miner))
			if common.BytesToAddress(itx.To) == common.BytesToAddress(block.GetCoinbase()) {
				mevReward = new(big.Int).Add(mevReward, new(big.Int).SetBytes(itx.GetValue()))
			}
		}

	}
	return mevReward
}

func (bigtable *Bigtable) TransformTransaction(block *types.Eth1Block) (*types.BulkMutations, error) {
	muts := &types.BulkMutations{}

	// normal tx

	for _, tx := range block.GetTransactions() {
		txHash := tx.GetHash()
		key := fmt.Sprintf("%s:t:%s", bigtable.chainId, tx.GetHash())

		if len(txHash) != 32 {
			return nil, fmt.Errorf("unexpected hash: %x, len: %v, transaction: %+v, from: %x to: %x", txHash, len(txHash), tx, tx.From, tx.To)
		}
		for j, log := range tx.GetLogs() {
			encLog, err := proto.Marshal(log)
			if err != nil {
				return nil, err
			}
			mut := gcp_bigtable.NewMutation()

			// log:<hash>:<position>
			mut.Set(DEFAULT_FAMILY, fmt.Sprintf("log:%03d", j), gcp_bigtable.Timestamp(0), encLog)
			muts.Keys = append(muts.Keys, key)
			muts.Muts = append(muts.Muts, mut)
		}
		tx.Logs = nil

		for k, itx := range tx.GetItx() {
			encItx, err := proto.Marshal(itx)
			if err != nil {
				return nil, err
			}
			mut := gcp_bigtable.NewMutation()
			// itx:<hash>:<position>
			mut.Set(DEFAULT_FAMILY, fmt.Sprintf("itx:%03d", k), gcp_bigtable.Timestamp(0), encItx)
			muts.Keys = append(muts.Keys, key)
			muts.Muts = append(muts.Muts, mut)
		}
		tx.Itx = nil
		// store transaction without logs and internal transactions
		encTx, err := proto.Marshal(tx)
		if err != nil {
			return nil, err
		}
		mut := gcp_bigtable.NewMutation()
		// tx:<position>:<hash>
		mut.Set(DEFAULT_FAMILY, DATA_COLUMN, gcp_bigtable.Timestamp(0), encTx)
		muts.Keys = append(muts.Keys, key)
		muts.Muts = append(muts.Muts, mut)
	}

	return muts, nil
}
