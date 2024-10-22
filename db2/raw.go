package db2

import (
	"fmt"
	"log/slog"
	"math/big"
	"strings"

	"github.com/gobitfly/eth2-beaconchain-explorer/db2/store"
)

type compressor interface {
	compress(src []byte) ([]byte, error)
	decompress(src []byte) ([]byte, error)
}

type RawStore struct {
	store      store.Store
	compressor compressor
}

func NewRawStore(store store.Store) RawStore {
	return RawStore{
		store:      store,
		compressor: gzipCompressor{},
	}
}

func (db RawStore) AddBlocks(blocks []FullBlockRawData) error {
	itemsByKey := make(map[string][]store.Item)
	for _, fullBlock := range blocks {
		if len(fullBlock.Block) == 0 || len(fullBlock.BlockTxs) != 0 && len(fullBlock.Traces) == 0 {
			return fmt.Errorf("block %d: empty data", fullBlock.BlockNumber)
		}
		key := blockKey(fullBlock.ChainID, fullBlock.BlockNumber)

		block, err := db.compressor.compress(fullBlock.Block)
		if err != nil {
			return fmt.Errorf("cannot compress block %d: %w", fullBlock.BlockNumber, err)
		}
		receipts, err := db.compressor.compress(fullBlock.Receipts)
		if err != nil {
			return fmt.Errorf("cannot compress receipts %d: %w", fullBlock.BlockNumber, err)
		}
		traces, err := db.compressor.compress(fullBlock.Traces)
		if err != nil {
			return fmt.Errorf("cannot compress traces %d: %w", fullBlock.BlockNumber, err)
		}
		itemsByKey[key] = []store.Item{
			{
				Family: BT_COLUMNFAMILY_BLOCK,
				Column: BT_COLUMN_BLOCK,
				Data:   block,
			},
			{
				Family: BT_COLUMNFAMILY_RECEIPTS,
				Column: BT_COLUMN_RECEIPTS,
				Data:   receipts,
			},
			{
				Family: BT_COLUMNFAMILY_TRACES,
				Column: BT_COLUMN_TRACES,
				Data:   traces,
			},
		}
		if len(fullBlock.Receipts) < 1 {
			// todo move that log higher up
			slog.Warn(fmt.Sprintf("empty receipts at block %d lRec %d lTxs %d", fullBlock.BlockNumber, len(fullBlock.Receipts), len(fullBlock.BlockTxs)))
		}
		if fullBlock.BlockUnclesCount > 0 {
			uncles, err := db.compressor.compress(fullBlock.Uncles)
			if err != nil {
				return fmt.Errorf("cannot compress block %d: %w", fullBlock.BlockNumber, err)
			}
			itemsByKey[key] = append(itemsByKey[key], store.Item{
				Family: BT_COLUMNFAMILY_UNCLES,
				Column: BT_COLUMN_UNCLES,
				Data:   uncles,
			})
		}
	}
	return db.store.BulkAdd(itemsByKey)
}

func (db RawStore) ReadBlockByNumber(chainID uint64, number int64) (*FullBlockRawData, error) {
	return db.readBlock(chainID, number)
}

func (db RawStore) ReadBlockByHash(chainID uint64, hash string) (*FullBlockRawData, error) {
	// todo use sql db to retrieve hash
	return nil, fmt.Errorf("ReadBlockByHash not implemented")
}

func (db RawStore) readBlock(chainID uint64, number int64) (*FullBlockRawData, error) {
	key := blockKey(chainID, number)
	data, err := db.store.GetRow(key)
	if err != nil {
		return nil, err
	}
	return db.parseRow(chainID, number, data)
}

func (db RawStore) parseRow(chainID uint64, number int64, data map[string][]byte) (*FullBlockRawData, error) {
	block, err := db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_BLOCK, BT_COLUMN_BLOCK)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress block %d: %w", number, err)
	}
	receipts, err := db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_RECEIPTS, BT_COLUMN_RECEIPTS)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress receipts %d: %w", number, err)
	}
	traces, err := db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_TRACES, BT_COLUMN_TRACES)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress traces %d: %w", number, err)
	}
	uncles, err := db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_UNCLES, BT_COLUMN_UNCLES)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress uncles %d: %w", number, err)
	}
	return &FullBlockRawData{
		ChainID:          chainID,
		BlockNumber:      number,
		BlockHash:        nil,
		BlockUnclesCount: 0,
		BlockTxs:         nil,
		Block:            block,
		Receipts:         receipts,
		Traces:           traces,
		Uncles:           uncles,
	}, nil
}

func (db RawStore) ReadBlocksByNumber(chainID uint64, start, end int64) ([]*FullBlockRawData, error) {
	rows, err := db.store.GetRowsRange(blockKey(chainID, start-1), blockKey(chainID, end))
	if err != nil {
		return nil, err
	}
	var blocks []*FullBlockRawData
	for key, data := range rows {
		number := blockKeyToNumber(chainID, key)
		block, err := db.parseRow(chainID, number, data)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func blockKey(chainID uint64, number int64) string {
	return fmt.Sprintf("%d:%12d", chainID, MAX_EL_BLOCK_NUMBER-number)
}

func blockKeyToNumber(chainID uint64, key string) int64 {
	key = strings.TrimPrefix(key, fmt.Sprintf("%d:", chainID))
	reversed, _ := new(big.Int).SetString(key, 10)

	return MAX_EL_BLOCK_NUMBER - reversed.Int64()
}

type FullBlockRawData struct {
	ChainID uint64

	BlockNumber      int64
	BlockHash        Bytes
	BlockUnclesCount int
	BlockTxs         []string

	Block    Bytes
	Receipts Bytes
	Traces   Bytes
	Uncles   Bytes
}
