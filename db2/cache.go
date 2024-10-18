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
		return v.(*FullBlockRawData), nil
	}

	block, err := c.db.ReadBlockByNumber(chainID, number)
	if block != nil {
		c.cache.Store(key, block)

		// retrieve the block hash for caching purpose
		var mini MinimalBlock
		if err := json.Unmarshal(block.Block, &mini); err != nil {
			return nil, fmt.Errorf("cannot unmarshal block: %w", err)
		}
		c.cache.Store(mini.Result.Hash, number)
		go func() {
			time.Sleep(ttl)
			c.cache.Delete(key)
			c.cache.Delete(mini.Result.Hash)
			c.mapLock.Lock()
			defer c.mapLock.Unlock()
			delete(c.locks, key)
		}()
	}
	return block, err
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
