package db

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var lastAttestationCacheDb *leveldb.DB

func InitLastAttestationCache(path string) error {
	if path == "" {
		logger.Infof("no last attestation cache path provided, using temporary directory %v", os.TempDir()+"/lastAttestationCache")
		path = os.TempDir() + "/lastAttestationCache"
	}
	ldb, err := leveldb.OpenFile(path, nil)

	if err != nil {
		return err
	}

	lastAttestationCacheDb = ldb
	return nil
}

func SetLastAttestationSlots(tx *sqlx.Tx, attestedSlots map[uint64]uint64) error {
	start := time.Now()

	defer func() {
		logger.Infof("setting last attestation slots took %v", time.Since(start))
	}()

	if len(attestedSlots) == 0 {
		return nil
	}

	cachedSlots, err := GetLastAttestationSlotsFromCache()
	if err != nil {
		return err
	}

	batch := new(leveldb.Batch)
	dirty := make([][2]uint64, 0)

	for index, slot := range attestedSlots {
		key := fmt.Sprintf("%d", index)

		if slot > cachedSlots[index] {
			batch.Put([]byte(key), []byte(fmt.Sprintf("%d", slot)))
			dirty = append(dirty, [2]uint64{index, slot})
		}
	}

	UpdateLastAttestationSlots(tx, dirty)

	err = lastAttestationCacheDb.Write(batch, nil)
	if err != nil {
		return err
	}

	return nil
}

func GetLastAttestationSlotsFromCache() (map[uint64]uint64, error) {
	ret := make(map[uint64]uint64)
	iter := lastAttestationCacheDb.NewIterator(util.BytesPrefix([]byte("")), nil)
	defer iter.Release()
	for iter.Next() {
		index, err := strconv.ParseUint(string(iter.Key()), 10, 64)
		if err != nil {
			return nil, err
		}
		slot, err := strconv.ParseUint(string(iter.Value()), 10, 64)
		if err != nil {
			return nil, err
		}

		ret[index] = slot
	}

	return ret, iter.Error()
}

func GetLastAttestationSlotFromCache(validatorIndex uint64) (uint64, error) {
	key := fmt.Sprintf("%d", validatorIndex)

	ret, err := lastAttestationCacheDb.Get([]byte(key), nil)

	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return 0, nil
		} else {
			return 0, err
		}
	}
	slot, err := strconv.ParseUint(string(ret), 10, 64)
	if err != nil {
		return 0, err
	}

	return slot, nil
}
