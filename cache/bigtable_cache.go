package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/sirupsen/logrus"
)

const (
	TABLE_CACHE        = "cache"
	FAMILY_TEN_MINUTES = "10_min"
	FAMILY_ONE_HOUR    = "1_hour"
	FAMILY_ONE_DAY     = "1_day"
	COLUMN_DATA        = "d"
)

type BigtableCache struct {
	client *gcp_bigtable.Client

	tableCache *gcp_bigtable.Table

	chainId string
}

func InitBigtableCache(client *gcp_bigtable.Client, chainId string) *BigtableCache {
	bt := &BigtableCache{
		client:     client,
		tableCache: client.Open(TABLE_CACHE),
		chainId:    chainId,
	}

	return bt
}

func (cache *BigtableCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {

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

	err = cache.tableCache.Apply(ctx, fmt.Sprintf("C:%s", key), mut)
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

	err := cache.tableCache.Apply(ctx, fmt.Sprintf("C:%s", key), mut)
	if err != nil {
		return err
	}
	return nil
}

func (cache *BigtableCache) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	return cache.setByte(ctx, key, []byte(value), expiration)
}

func (cache *BigtableCache) SetUint64(ctx context.Context, key string, value uint64, expiration time.Duration) error {
	return cache.setByte(ctx, key, ui64tob(value), expiration)
}

func (cache *BigtableCache) SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	return cache.setByte(ctx, key, booltob(value), expiration)
}

func (cache *BigtableCache) Get(ctx context.Context, key string, returnValue any) (any, error) {
	res, err := cache.getByte(ctx, key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(res), returnValue)
	if err != nil {
		// cache.remoteRedisCache.Del(ctx, key).Err()
		logrus.Errorf("error (bigtable_cache.go / Get) unmarshalling data for key %v: %v", key, err)
		return nil, err
	}

	return returnValue, nil
}

func (cache *BigtableCache) getByte(ctx context.Context, key string) ([]byte, error) {
	filter := gcp_bigtable.ChainFilters(
		gcp_bigtable.ColumnFilter("d"),
		gcp_bigtable.LatestNFilter(1),
	)

	row, err := cache.tableCache.ReadRow(ctx, fmt.Sprintf("C:%s", key), gcp_bigtable.RowFilter(filter))
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

func (cache *BigtableCache) GetString(ctx context.Context, key string) (string, error) {

	res, err := cache.getByte(ctx, key)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func (cache *BigtableCache) GetUint64(ctx context.Context, key string) (uint64, error) {

	res, err := cache.getByte(ctx, key)
	if err != nil {
		return 0, err
	}

	return btoi64(res), nil
}

func (cache *BigtableCache) GetBool(ctx context.Context, key string) (bool, error) {

	res, err := cache.getByte(ctx, key)
	if err != nil {
		return false, err
	}

	return btobool(res), nil
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

func booltob(val bool) []byte {
	r := make([]byte, 1)
	if val {
		r[0] = 1
	} else {
		r[0] = 0
	}
	return r
}

func btobool(val []byte) bool {
	return val[0] == 1
}
