package services

import (
	"eth2-exporter/db"
	"fmt"
	"os"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

var pubkeyCacheDb *leveldb.DB

func initPubkeyCache(path string) error {
	if path == "" {
		logger.Infof("no last pubkey cache path provided, using temporary directory %v", os.TempDir()+"/pubkeyCache")
		path = os.TempDir() + "/pubkeyCache"
	}
	db, err := leveldb.OpenFile(path, nil)

	if err != nil {
		return err
	}

	pubkeyCacheDb = db
	return nil
}

// will retrieve the pubkey for a given validatorindex and store it for later use
func GetPubkeyForIndex(index uint64) ([]byte, error) {
	key := []byte(fmt.Sprintf("%d", index))

	pubkey, err := pubkeyCacheDb.Get(key, nil)
	if err == leveldb.ErrNotFound {
		err = db.WriterDb.Get(&pubkey, "SELECT pubkey FROM validators WHERE validatorindex = $1", index)

		if err != nil {
			return nil, err
		}

		err = pubkeyCacheDb.Put(key, pubkey, nil)
		if err != nil {
			return nil, err
		}
		logger.Infof("serving pubkey %x for validator %v from db", pubkey, index)
		return pubkey, nil
	} else if err != nil {
		return nil, err
	}

	// logger.Infof("serving pubkey %x for validator %v from cache", pubkey, index)

	return pubkey, nil
}

func GetIndexForPubkey(pubkey []byte) (uint64, error) {
	var index uint64

	key := []byte(fmt.Sprintf("%x", pubkey))

	indexString, err := pubkeyCacheDb.Get(key, nil)
	if err == leveldb.ErrNotFound {
		err = db.WriterDb.Get(&index, "SELECT validatorindex FROM validators WHERE pubkey = $1", pubkey)

		if err != nil {
			return 0, err
		}

		err = pubkeyCacheDb.Put(key, []byte(fmt.Sprintf("%d", index)), nil)

		if err != nil {
			return 0, err
		}
		logger.Infof("serving index %d for validator %x from db", index, pubkey)
		return index, nil
	} else if err != nil {
		return 0, err
	}

	index, err = strconv.ParseUint(string(indexString), 10, 64)

	if err != nil {
		return 0, err
	}
	// logger.Infof("serving index %d for validator %x from cache", index, pubkey)

	return index, nil
}
