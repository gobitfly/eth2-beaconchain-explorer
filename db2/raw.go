package db2

import (
	"fmt"
	"log/slog"

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

func (db RawStore) ReadBlock(chainID uint64, number int64) (*FullBlockRawData, error) {
	key := blockKey(chainID, number)
	data, err := db.store.GetRow(key)
	if err != nil {
		return nil, err
	}
	block, err := db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_BLOCK, BT_COLUMN_BLOCK)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress block %d: %w", number, err)
	}
	receipts, err := db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_RECEIPTS, BT_COLUMN_RECEIPTS)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress block %d: %w", number, err)
	}
	traces, err := db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_TRACES, BT_COLUMN_TRACES)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress block %d: %w", number, err)
	}
	uncles, err := db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_UNCLES, BT_COLUMN_UNCLES)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress block %d: %w", number, err)
	}
	return &FullBlockRawData{
		ChainID:     chainID,
		BlockNumber: number,
		Block:       block,
		Receipts:    receipts,
		Traces:      traces,
		Uncles:      uncles,
	}, nil
}

func blockKey(chainID uint64, number int64) string {
	return fmt.Sprintf("%d:%12d", chainID, MAX_EL_BLOCK_NUMBER-number)
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
