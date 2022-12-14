package services

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var lastAttestationCacheDb *leveldb.DB

func InitLastAttestationCache(path string) error {
	if path == "" {
		logger.Infof("no last attestation cache path provided, using temporary directory %v", os.TempDir()+"/lastAttestationCache")
		path = os.TempDir() + "/lastAttestationCache"
	}
	db, err := leveldb.OpenFile(path, nil)

	if err != nil {
		return err
	}

	lastAttestationCacheDb = db
	return nil
}

func SetLastAttestationSlots(attestedSlots map[uint64]uint64) error {

	start := time.Now()

	defer func() {
		logger.Infof("setting last attestation slots took %v", time.Since(start))
	}()

	if len(attestedSlots) == 0 {
		return nil
	}

	cachedSlots, err := GetLastAttestationSlots()
	if err != nil {
		return err
	}

	batch := new(leveldb.Batch)

	for index, slot := range attestedSlots {
		key := fmt.Sprintf("%d", index)

		if slot > cachedSlots[index] {
			batch.Put([]byte(key), []byte(fmt.Sprintf("%d", slot)))
		}
	}

	err = lastAttestationCacheDb.Write(batch, nil)
	if err != nil {
		return err
	}

	return nil
}

func GetLastAttestationSlots() (map[uint64]uint64, error) {
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
