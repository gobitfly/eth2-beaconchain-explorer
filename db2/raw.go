package db2

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
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
	var wg sync.WaitGroup
	errors := make(chan error, len(blocks))
	itemsByKey := make(map[string][]store.Item)
	mu := sync.Mutex{}

	for _, fullBlock := range blocks {
		wg.Add(1)
		go func(fullBlock FullBlockRawData) {
			defer wg.Done()

			if len(fullBlock.Block) == 0 || (len(fullBlock.BlockTxs) != 0 && len(fullBlock.Traces) == 0) {
				errors <- fmt.Errorf("block %d: empty data", fullBlock.BlockNumber)
				return
			}
			key := blockKey(fullBlock.ChainID, fullBlock.BlockNumber)

			block, err := db.compressor.compress(fullBlock.Block)
			if err != nil {
				errors <- fmt.Errorf("cannot compress block %d: %w", fullBlock.BlockNumber, err)
				return
			}
			receipts, err := db.compressor.compress(fullBlock.Receipts)
			if err != nil {
				errors <- fmt.Errorf("cannot compress receipts %d: %w", fullBlock.BlockNumber, err)
				return
			}
			traces, err := db.compressor.compress(fullBlock.Traces)
			if err != nil {
				errors <- fmt.Errorf("cannot compress traces %d: %w", fullBlock.BlockNumber, err)
				return
			}

			mu.Lock()
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
					errors <- fmt.Errorf("cannot compress uncles %d: %w", fullBlock.BlockNumber, err)
					return
				}
				itemsByKey[key] = append(itemsByKey[key], store.Item{
					Family: BT_COLUMNFAMILY_UNCLES,
					Column: BT_COLUMN_UNCLES,
					Data:   uncles,
				})
			}
			mu.Unlock()
		}(fullBlock)
	}

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		return <-errors
	}

	return db.store.BulkAdd(itemsByKey)
}

func (db RawStore) ReadBlockByHash(chainID uint64, hash string) (*FullBlockRawData, error) {
	// todo use sql db to retrieve hash
	return nil, fmt.Errorf("ReadBlockByHash not implemented")
}

func (db RawStore) ReadBlocksByNumbers(chainID uint64, blockNumbers []int64) (map[int64]*FullBlockRawData, error) {
	results := make(map[int64]*FullBlockRawData)
	errorChan := make(chan error, len(blockNumbers))
	workers := make(chan struct{}, 10)

	if len(blockNumbers) == 0 {
		return results, nil
	}

	var wg sync.WaitGroup
	for _, blockNumber := range blockNumbers {
		wg.Add(1)

		go func(number int64) {
			defer wg.Done()
			workers <- struct{}{}

			defer func() { <-workers }()

			block, err := db.ReadBlockByNumber(chainID, number) //@TODO implement batch read from db

			if err != nil {
				errorChan <- fmt.Errorf("error while reading the block %d, error: %v", number, err)
				return
			}

			results[number] = block
		}(blockNumber)
	}
	wg.Wait()
	close(errorChan)

	if len(errorChan) > 0 {
		return nil, <-errorChan
	}

	return results, nil
}

func (db RawStore) ReadBlockByNumber(chainID uint64, number int64) (*FullBlockRawData, error) {
	key := blockKey(chainID, number)
	data, err := db.store.GetRow(key)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	errors := make(chan error, 4)
	var block, receipts, traces, uncles []byte

	wg.Add(4)

	go func() {
		defer wg.Done()
		var err error
		block, err = db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_BLOCK, BT_COLUMN_BLOCK)])
		if err != nil {
			errors <- fmt.Errorf("cannot decompress block %d: %w", number, err)
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		receipts, err = db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_RECEIPTS, BT_COLUMN_RECEIPTS)])
		if err != nil {
			errors <- fmt.Errorf("cannot decompress receipts %d: %w", number, err)
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		traces, err = db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_TRACES, BT_COLUMN_TRACES)])
		if err != nil {
			errors <- fmt.Errorf("cannot decompress traces %d: %w", number, err)
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		uncles, err = db.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_UNCLES, BT_COLUMN_UNCLES)])
		if err != nil {
			errors <- fmt.Errorf("cannot decompress uncles %d: %w", number, err)
		}
	}()

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		return nil, <-errors
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

func blockKey(chainID uint64, number int64) string {
	return fmt.Sprintf("%d:%12d", chainID, MAX_EL_BLOCK_NUMBER-number)
}

type FullBlockRawData struct {
	ChainID uint64

	BlockNumber      int64
	BlockHash        hexutil.Bytes
	BlockUnclesCount int
	BlockTxs         []string
	Block            hexutil.Bytes
	Receipts         hexutil.Bytes
	Traces           hexutil.Bytes
	Uncles           hexutil.Bytes
}
