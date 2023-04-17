package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RedisCache struct {
	redisRemoteCache *redis.Client
}

func InitRedisCache(ctx context.Context, redisAddress string) (*RedisCache, error) {
	rdc := redis.NewClient(&redis.Options{
		Addr:        redisAddress,
		ReadTimeout: time.Second * 20,
	})

	if err := rdc.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	r := &RedisCache{
		redisRemoteCache: rdc,
	}
	return r, nil
}

func (cache *RedisCache) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	return cache.redisRemoteCache.Set(ctx, key, value, expiration).Err()
}

func (cache *RedisCache) GetString(ctx context.Context, key string) (string, error) {

	value, err := cache.redisRemoteCache.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return value, nil
}

func (cache *RedisCache) SetUint64(ctx context.Context, key string, value uint64, expiration time.Duration) error {
	return cache.redisRemoteCache.Set(ctx, key, fmt.Sprintf("%d", value), expiration).Err()
}

func (cache *RedisCache) GetUint64(ctx context.Context, key string) (uint64, error) {

	value, err := cache.redisRemoteCache.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	returnValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return returnValue, nil
}

func (cache *RedisCache) SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	return cache.redisRemoteCache.Set(ctx, key, fmt.Sprintf("%t", value), expiration).Err()
}

func (cache *RedisCache) GetBool(ctx context.Context, key string) (bool, error) {

	value, err := cache.redisRemoteCache.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}

	returnValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}
	return returnValue, nil
}

func (cache *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	valueMarshal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return cache.redisRemoteCache.Set(ctx, key, valueMarshal, expiration).Err()
}

func (cache *RedisCache) Get(ctx context.Context, key string, returnValue interface{}) (interface{}, error) {
	value, err := cache.redisRemoteCache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(value), returnValue)
	if err != nil {
		cache.redisRemoteCache.Del(ctx, key).Err()
		logrus.Errorf("error (redit_cache / Get) unmarshalling data for key %v: %v", key, err)
		return nil, err
	}

	return returnValue, nil
}
