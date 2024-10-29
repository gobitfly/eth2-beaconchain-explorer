package db2

import (
	"encoding/json"
	"sync"
	"time"
)

// var ttl = 20_000 * time.Millisecond
const (
	oneBlockTTL = 1 * time.Second
	blocksTTL   = 30 * time.Second // default ttl, if read it will be deleted sooner
)

type MinimalBlock struct {
	Result struct {
		Hash string `json:"hash"`
	} `json:"result"`
}

type CachedRawStore struct {
	db RawStoreReader
	// sync.Map with manual delete have better perf than freecache because we can handle this way a ttl < 1s
	cache sync.Map

	locks   map[string]*sync.RWMutex
	mapLock sync.Mutex // to make the map safe concurrently
}

func WithCache(reader RawStoreReader) *CachedRawStore {
	return &CachedRawStore{
		db:    reader,
		locks: make(map[string]*sync.RWMutex),
	}
}

func (c *CachedRawStore) lockBy(key string) func() {
	c.mapLock.Lock()
	defer c.mapLock.Unlock()

	lock, found := c.locks[key]
	if !found {
		lock = &sync.RWMutex{}
		c.locks[key] = lock
		lock.Lock()
		return lock.Unlock
	}
	lock.RLock()
	return lock.RUnlock
}

func (c *CachedRawStore) ReadBlockByNumber(chainID uint64, number int64) (*FullBlockRawData, error) {
	key := blockKey(chainID, number)

	unlock := c.lockBy(key)
	defer unlock()

	v, ok := c.cache.Load(key)
	if ok {
		// once read ensure to delete it from the cache
		go c.unCacheBlockAfter(key, "", oneBlockTTL)
		return v.(*FullBlockRawData), nil
	}
	// TODO make warning not found in cache
	block, err := c.db.ReadBlockByNumber(chainID, number)
	if block != nil {
		c.cacheBlock(block, oneBlockTTL)
	}
	return block, err
}

func (c *CachedRawStore) cacheBlock(block *FullBlockRawData, ttl time.Duration) {
	key := blockKey(block.ChainID, block.BlockNumber)
	c.cache.Store(key, block)

	var mini MinimalBlock
	if len(block.Uncles) != 0 {
		// retrieve the block hash for caching but only if the block has uncle(s)
		_ = json.Unmarshal(block.Block, &mini)
		c.cache.Store(mini.Result.Hash, block.BlockNumber)
	}

	go c.unCacheBlockAfter(key, mini.Result.Hash, ttl)
}

func (c *CachedRawStore) unCacheBlockAfter(key, hash string, ttl time.Duration) {
	time.Sleep(ttl)
	c.cache.Delete(key)
	c.mapLock.Lock()
	if hash != "" {
		c.cache.Delete(hash)
	}
	defer c.mapLock.Unlock()
	delete(c.locks, key)
}

func (c *CachedRawStore) ReadBlockByHash(chainID uint64, hash string) (*FullBlockRawData, error) {
	v, ok := c.cache.Load(hash)
	if !ok {
		return c.db.ReadBlockByHash(chainID, hash)
	}

	v, ok = c.cache.Load(blockKey(chainID, v.(int64)))
	if !ok {
		return c.db.ReadBlockByHash(chainID, hash)
	}

	return v.(*FullBlockRawData), nil
}

func (c *CachedRawStore) ReadBlocksByNumber(chainID uint64, start, end int64) ([]*FullBlockRawData, error) {
	blocks, err := c.db.ReadBlocksByNumber(chainID, start, end)
	if err != nil {
		return nil, err
	}
	for _, block := range blocks {
		c.cacheBlock(block, blocksTTL)
	}
	return blocks, nil
}
