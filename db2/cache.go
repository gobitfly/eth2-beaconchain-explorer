package db2

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

var ttl = 200 * time.Millisecond

type MinimalBlock struct {
	Result struct {
		Hash string `json:"hash"`
	} `json:"result"`
}

type CachedBlock struct {
	Data *FullBlockRawData
	Hash string
}

type CachedRawStore struct {
	db RawStoreReader
	// sync.Map with manual delete have better perf than freecache because we can handle this way a ttl < 1s
	cache sync.Map
}

func WithCache(reader RawStoreReader) *CachedRawStore {
	return &CachedRawStore{
		db: reader,
	}
}

func (c *CachedRawStore) ReadBlockByNumber(chainID uint64, number int64) (*FullBlockRawData, error) {
	key := blockKey(chainID, number)

	v, ok := c.cache.Load(key)
	if ok {
		return v.(*CachedBlock).Data, nil
	}

	block, err := c.db.ReadBlockByNumber(chainID, number)
	if err != nil || block == nil {
		return block, err
	}

	cachedBlock := &CachedBlock{Data: block}

	// retrieve the block hash for caching purpose
	var mini MinimalBlock
	if err := json.Unmarshal(block.Block, &mini); err != nil {
		return nil, fmt.Errorf("cannot unmarshal block: %w", err)
	}

	cachedBlock.Hash = mini.Result.Hash
	c.cache.Store(key, cachedBlock)

	go func() {
		time.Sleep(ttl)
		c.cache.Delete(key)
		c.cache.Delete(mini.Result.Hash)
	}()

	return block, nil
}

func (c *CachedRawStore) ReadBlockByHash(chainID uint64, hash string) (*FullBlockRawData, error) {
	v, ok := c.cache.Load(hash)
	if !ok {
		return v.(*CachedBlock).Data, nil
	}

	blockData, err := c.db.ReadBlockByHash(chainID, hash)
	if err != nil || blockData == nil {
		return blockData, err
	}

	key := blockKey(chainID, blockData.BlockNumber)
	if _, ok := c.cache.Load(key); ok {
		return blockData, nil
	}

	return c.db.ReadBlockByHash(chainID, hash)
}
