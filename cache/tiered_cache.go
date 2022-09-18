package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// Tiered cache is a cache implementation combining a
type tieredCache struct {
	remoteRedisCache *redis.Client
	localGoCache     gocache.Cache
}

var TieredCache *tieredCache

func MustInitTieredCache(redisAddress string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	rdc := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	if err := rdc.Ping(ctx).Err(); err != nil {
		logrus.Fatalf("error initializing tiered cache: %v", err)
	}

	TieredCache = &tieredCache{
		remoteRedisCache: rdc,
		localGoCache:     *gocache.New(time.Hour, time.Minute),
	}
}

func (cache *tieredCache) SetString(key, value string, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cache.localGoCache.Set(key, value, expiration)
	return cache.remoteRedisCache.Set(ctx, key, value, expiration).Err()
}

func (cache *tieredCache) GetStringWithLocalTimeout(key string, localExpiration time.Duration) (string, error) {
	// try to retrieve the key from the local cache
	wanted, found := cache.localGoCache.Get(key)
	if found {
		return wanted.(string), nil
	}

	// retrieve the key from the remote cache
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	value, err := cache.remoteRedisCache.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	cache.localGoCache.Set(key, value, localExpiration)
	return value, nil
}

func (cache *tieredCache) SetUint64(key string, value uint64, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cache.localGoCache.Set(key, value, expiration)
	return cache.remoteRedisCache.Set(ctx, key, fmt.Sprintf("%d", value), expiration).Err()
}

func (cache *tieredCache) GetUint64WithLocalTimeout(key string, localExpiration time.Duration) (uint64, error) {

	// try to retrieve the key from the local cache
	wanted, found := cache.localGoCache.Get(key)
	if found {
		return wanted.(uint64), nil
	}

	// retrieve the key from the remote cache
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	value, err := cache.remoteRedisCache.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	returnValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	cache.localGoCache.Set(key, returnValue, localExpiration)
	return returnValue, nil
}

func (cache *tieredCache) Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	valueMarshal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	cache.localGoCache.Set(key, valueMarshal, expiration)
	return cache.remoteRedisCache.Set(ctx, key, valueMarshal, expiration).Err()
}

func (cache *tieredCache) GetWithLocalTimeout(key string, localExpiration time.Duration, returnValue interface{}) (interface{}, error) {
	// try to retrieve the key from the local cache
	wanted, found := cache.localGoCache.Get(key)
	if found {
		logrus.Infof("retrieved %v from in memory cache", key)
		return wanted, nil
	}

	// retrieve the key from the remote cache
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	value, err := cache.remoteRedisCache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(value), returnValue)
	if err != nil {
		logrus.Fatal(err)
		return nil, err
	}
	logrus.Infof("retrieved %v from redis cache", key)

	cache.localGoCache.Set(key, returnValue, localExpiration)
	return returnValue, nil
}
