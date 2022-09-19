package cache

import (
	"context"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// Tiered cache is a cache implementation combining a
type tieredCache struct {
	localGoCache gocache.Cache
	remoteCache  RemoteCache
}

type RemoteCache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	SetString(ctx context.Context, key, value string, expiration time.Duration) error
	SetUint64(ctx context.Context, key string, value uint64, expiration time.Duration) error
	GetUint64(ctx context.Context, key string) (uint64, error)
	GetString(ctx context.Context, key string) (string, error)
	Get(ctx context.Context, key string, returnValue any) (any, error)
}

var TieredCache *tieredCache

func MustInitTieredCache(redisAddress string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	remoteCache, err := InitRedisCache(ctx, redisAddress)
	if err != nil {
		logrus.Panicf("error initializing remote redis cache. address: %v", redisAddress)
	}

	TieredCache = &tieredCache{
		remoteCache:  remoteCache,
		localGoCache: *gocache.New(time.Hour, time.Minute),
	}
}

func MustInitTieredCacheBigtable(client *gcp_bigtable.Client, chainId string) {
	localCache := *gocache.New(time.Hour, time.Minute)

	cache := InitBigtableCache(client, chainId)

	TieredCache = &tieredCache{
		remoteCache:  cache,
		localGoCache: localCache,
	}

}

func (cache *tieredCache) SetString(key, value string, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cache.localGoCache.Set(key, value, expiration)
	return cache.remoteCache.SetString(ctx, key, value, expiration)
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

	value, err := cache.remoteCache.GetString(ctx, key)
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
	return cache.remoteCache.SetUint64(ctx, key, value, expiration)
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

	value, err := cache.remoteCache.GetUint64(ctx, key)
	if err != nil {
		return 0, err
	}

	cache.localGoCache.Set(key, value, localExpiration)
	return value, nil
}

func (cache *tieredCache) Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	return cache.remoteCache.Set(ctx, key, value, expiration)
}

func (cache *tieredCache) GetWithLocalTimeout(key string, localExpiration time.Duration, returnValue interface{}) (interface{}, error) {
	// try to retrieve the key from the local cache
	wanted, found := cache.localGoCache.Get(key)
	if found {
		return wanted, nil
	}

	// retrieve the key from the remote cache
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	value, err := cache.remoteCache.Get(ctx, key, returnValue)
	if err != nil {
		return nil, err
	}

	cache.localGoCache.Set(key, value, localExpiration)
	return value, nil
}
