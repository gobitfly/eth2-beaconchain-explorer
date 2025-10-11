package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/klauspost/pgzip"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// validatorTaggerLogger is the module-scoped logger for the validator tagger service.
var validatorTaggerLogger = logrus.StandardLogger().WithField("module", "services.validator_tagger")

// withStepMetrics returns a finisher function that records duration and errors for a step.
// Usage:
//
//	finish := withStepMetrics("validator_tagger_import")
//	defer finish(err == nil)
func withStepMetrics(step string) func(success bool) {
	start := time.Now()
	metrics.Tasks.WithLabelValues(step).Inc()
	return func(success bool) {
		metrics.TaskDuration.WithLabelValues(step).Observe(time.Since(start).Seconds())
		if !success {
			metrics.Errors.WithLabelValues(step).Inc()
		}
	}
}

func fetchValidatorBalancesFromMapping() (map[int]uint64, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	rdc := redis.NewClient(&redis.Options{
		Addr:        utils.Config.RedisSessionStoreEndpoint,
		ReadTimeout: time.Second * 60,
	})

	if err := rdc.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "failed to ping redis")
	}
	defer func(rdc *redis.Client) {
		err := rdc.Close()
		if err != nil {
			validatorTaggerLogger.Errorf("failed to close redis client: %v", err)
		}
	}(rdc)

	key := fmt.Sprintf("%d:%s", utils.Config.Chain.ClConfig.DepositChainID, "vm")
	compressed, err := rdc.Get(ctx, key).Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get compressed validator mapping from db")
	}
	validatorTaggerLogger.Debugf("reading validator mapping from redis done, took %s", time.Since(start))

	// decompress
	start = time.Now()
	_cachedBufferCompressed := new(bytes.Buffer)
	_cachedBufferCompressed.Write(compressed)
	defer _cachedBufferCompressed.Reset()
	w, err := pgzip.NewReaderN(_cachedBufferCompressed, 1_000_000, 10)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create pgzip reader")
	}
	defer w.Close()
	_cachedBufferDecompressed := new(bytes.Buffer)
	_, err = w.WriteTo(_cachedBufferDecompressed)
	defer _cachedBufferDecompressed.Reset()
	if err != nil {
		return nil, errors.Wrap(err, "failed to decompress validator mapping from redis")
	}
	validatorTaggerLogger.Debugf("decompressing validator mapping using pgzip took %s", time.Since(start))

	// ungob
	start = time.Now()
	dec := gob.NewDecoder(_cachedBufferDecompressed)
	_cachedRedisValidatorMapping := new(redisCachedValidatorsMapping)
	err = dec.Decode(&_cachedRedisValidatorMapping)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding assignments data")
	}
	validatorTaggerLogger.Debugf("decoding validator mapping from gob took %s", time.Since(start))
	balanceMap := make(map[int]uint64, len(_cachedRedisValidatorMapping.Mapping))

	for validatorIndex, validator := range _cachedRedisValidatorMapping.Mapping {
		balanceMap[validatorIndex] = validator.Balance
	}
	return balanceMap, nil
}

type queuesMetadata struct {
	ActivationIndex sql.NullInt64
}

type cachedValidator struct {
	PublicKey                  []byte
	ActivationEligibilityEpoch sql.NullInt64
	ActivationEpoch            sql.NullInt64
	ExitEpoch                  sql.NullInt64
	WithdrawableEpoch          sql.NullInt64
	Status                     string
	WithdrawalCredentials      []byte
	Balance                    uint64
	EffectiveBalance           uint64
	Slashed                    bool
	Queues                     queuesMetadata
}

type epoch uint64

type redisCachedValidatorsMapping struct {
	Epoch   epoch
	Mapping []*cachedValidator
}
