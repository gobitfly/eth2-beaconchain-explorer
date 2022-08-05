package db

import (
	"context"
	"errors"
	"eth2-exporter/types"
	"fmt"
	"strings"
	"time"

	"strconv"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"

	"github.com/golang/protobuf/proto"
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
