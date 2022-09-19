package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

const (
	FAMILY_TEN_MINUTES = "10_min"
	FAMILY_ONE_HOUR    = "1_hour"
	FAMILY_ONE_DAY     = "1_day"
	COLUMN_DATA        = "d"
)

type BigtableCache struct {
	client *gcp_bigtable.Client

	tableCache   *gcp_bigtable.Table
	localGoCache gocache.Cache

	chainId string
}

func (cache *BigtableCache) Set(key string, value any, expiration time.Duration) error {
	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	family := FAMILY_TEN_MINUTES
	if expiration.Minutes() >= 60 {
		family = FAMILY_ONE_HOUR
	}
	if expiration.Hours() > 1 {
		family = FAMILY_ONE_DAY
	}

	valueMarshal, err := json.Marshal(value)
	if err != nil {
		return err
	}

	ts := gcp_bigtable.Now()
	mut := gcp_bigtable.NewMutation()
	mut.Set(family, COLUMN_DATA, ts, valueMarshal)

	err = cache.tableCache.Apply(ctx, fmt.Sprintf("C:%s:%s", cache.chainId, key), mut)
	if err != nil {
		return err
	}
	return nil
}

func (cache *BigtableCache) setByte(ctx context.Context, key string, value []byte, expiration time.Duration) error {

	family := FAMILY_TEN_MINUTES
	if expiration.Minutes() >= 60 {
		family = FAMILY_ONE_HOUR
	}
	if expiration.Hours() > 1 {
		family = FAMILY_ONE_DAY
	}

	ts := gcp_bigtable.Now()
	mut := gcp_bigtable.NewMutation()
	mut.Set(family, COLUMN_DATA, ts, value)

	err := cache.tableCache.Apply(ctx, fmt.Sprintf("C:%s:%s", cache.chainId, key), mut)
	if err != nil {
		return err
	}
	return nil
}

func (cache *BigtableCache) SetString(key, value string, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cache.localGoCache.Set(key, value, expiration)
	return cache.setByte(ctx, key, []byte(value), expiration)
}

func (cache *BigtableCache) SetUint64(key string, value uint64, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cache.localGoCache.Set(key, value, expiration)
	return cache.setByte(ctx, key, ui64tob(value), expiration)
}

func (cache *BigtableCache) GetWithLocalTimeout(key string, localExpiration time.Duration, returnValue interface{}) (interface{}, error) {
	// try to retrieve the key from the local cache
	wanted, found := cache.localGoCache.Get(key)
	if found {
		return wanted, nil
	}

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	res, err := cache.getByteWithLocalTimeout(ctx, key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(res), returnValue)
	if err != nil {
		// cache.remoteRedisCache.Del(ctx, key).Err()
		logrus.Warnf("error unmarshalling data for key %v: %v", key, err)
		return nil, err
	}

	cache.localGoCache.Set(key, returnValue, localExpiration)
	return returnValue, nil
}

func (cache *BigtableCache) getByteWithLocalTimeout(ctx context.Context, key string) ([]byte, error) {
	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.ColumnFilter("d"),
		gcp_bigtable.LatestNFilter(1),
	)

	row, err := cache.tableCache.ReadRow(ctx, fmt.Sprintf("C:%s:%s", cache.chainId, key), gcp_bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}

	if len(row) == 0 {
		return nil, fmt.Errorf("error getting key: %s no result available, row: %+v", key, row)
	}

	// iterate over all column families and only take the most recent entry
	res := new(gcp_bigtable.ReadItem)
	for _, column := range row {
		if len(column) != 1 {
			return nil, fmt.Errorf("error unexpected number of results returned key: %s no result available, row: %+v", key, row)
		}
		if res == nil {
			res = &column[0]
		}

		if res.Timestamp.Time().Before(column[0].Timestamp.Time()) {
			res = &column[0]
		}
	}

	return res.Value, nil
}

func (cache *BigtableCache) GetUint64WithLocalTimeout(key string, localExpiration time.Duration) (uint64, error) {
	// try to retrieve the key from the local cache
	wanted, found := cache.localGoCache.Get(key)
	if found {
		return wanted.(uint64), nil
	}

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	res, err := cache.getByteWithLocalTimeout(ctx, key)
	if err != nil {
		return 0, err
	}

	returnValue := btoi64(res)

	cache.localGoCache.Set(key, returnValue, localExpiration)
	return returnValue, nil
}

func (cache *BigtableCache) GetStringWithLocalTimeout(key string, localExpiration time.Duration) (string, error) {
	// try to retrieve the key from the local cache
	wanted, found := cache.localGoCache.Get(key)
	if found {
		return wanted.(string), nil
	}

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	res, err := cache.getByteWithLocalTimeout(ctx, key)
	if err != nil {
		return "", err
	}

	returnValue := string(res)

	cache.localGoCache.Set(key, returnValue, localExpiration)
	return returnValue, nil
}

func ui64tob(val uint64) []byte {
	r := make([]byte, 8)
	for i := uint64(0); i < 8; i++ {
		r[i] = byte((val >> (i * 8)) & 0xff)
	}
	return r
}

func btoi64(val []byte) uint64 {
	r := uint64(0)
	for i := uint64(0); i < 8; i++ {
		r |= uint64(val[i]) << (8 * i)
	}
	return r
}
