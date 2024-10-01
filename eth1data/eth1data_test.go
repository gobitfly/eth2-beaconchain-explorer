package eth1data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/slices"

	"github.com/gobitfly/eth2-beaconchain-explorer/cache"
	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/price"
	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

func TestGetEth1Transaction(t *testing.T) {
	node, exists := os.LookupEnv("ERIGON_NODE")
	if !exists {
		t.Skip()
	}

	erigon, err := rpc.NewErigonClient(node)
	if err != nil {
		t.Fatal(err)
	}
	rpc.CurrentErigonClient = erigon

	utils.Config = &types.Config{
		Chain: types.Chain{
			ClConfig: types.ClChainConfig{
				DepositChainID: 1,
			},
		},
		Bigtable: types.Bigtable{
			Project:      "test",
			Instance:     "instanceTest",
			Emulator:     true,
			EmulatorPort: 8086,
			EmulatorHost: "127.0.0.1",
		},
		Frontend: types.Frontend{
			ElCurrencyDivisor: 1e18,
		},
	}
	cache.TieredCache = &noCache{}
	db.ReaderDb = noSQLReaderDb{}

	price.Init(1, node, "ETH", "ETH")

	bt, err := db.InitBigtableWithCache(context.Background(), "test", "instanceTest", "1", noRedis{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.InitBigtableSchema(); err != nil {
		if !errors.Is(err, db.TableAlreadyExistErr) {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name          string
		block         int64
		txHash        string
		revertIndexes []int
	}{
		{
			name:   "no revert",
			block:  20870689,
			txHash: "0x45ae8f94592cd0fb20cd6b50c8def1b5d478c1968806f129dda881d3eb7b968e",
		},
		{
			name:          "recursive revert",
			block:         20183291,
			txHash:        "0xf7d385f000250c073dfef9a36327c5d30a4c77a0c50588ce3eded29f6829a4cd",
			revertIndexes: []int{0, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, _, err := erigon.GetBlock(tt.block, "parity/geth")
			if err != nil {
				t.Fatal(err)
			}

			if err := bt.SaveBlock(block); err != nil {
				t.Fatal(err)
			}

			res, err := GetEth1Transaction(common.HexToHash(tt.txHash), "ETH")
			if err != nil {
				t.Fatal(err)
			}

			for i, internal := range res.InternalTxns {
				if !slices.Contains(tt.revertIndexes, i) {
					if strings.Contains(string(internal.TracePath), "Transaction failed") {
						t.Errorf("internal transaction should not be flagged as failed")
					}
					continue
				}
				if !strings.Contains(string(internal.TracePath), "Transaction failed") {
					t.Errorf("internal transaction should be flagged as failed")
				}
			}
		})
	}
}

type noRedis struct {
}

func (n noRedis) SCard(ctx context.Context, key string) *redis.IntCmd {
	return redis.NewIntCmd(ctx)
}

func (n noRedis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return redis.NewBoolCmd(ctx)
}

func (n noRedis) Pipeline() redis.Pipeliner {
	//TODO implement me
	panic("implement me")
}

func (n noRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	cmd.SetErr(redis.Nil)
	return cmd
}

func (n noRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return redis.NewStatusCmd(ctx)
}

type noCache struct {
}

func (n noCache) Set(key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (n noCache) SetString(key string, value string, expiration time.Duration) error {
	return nil

}

func (n noCache) SetUint64(key string, value uint64, expiration time.Duration) error {
	return nil

}

func (n noCache) SetBool(key string, value bool, expiration time.Duration) error {
	return nil

}

func (n noCache) Get(ctx context.Context, key string, returnValue any) (any, error) {
	return nil, fmt.Errorf("no cache")
}

func (n noCache) GetString(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("no cache")
}

func (n noCache) GetUint64(ctx context.Context, key string) (uint64, error) {
	return 0, fmt.Errorf("no cache")
}

func (n noCache) GetBool(ctx context.Context, key string) (bool, error) {
	return false, fmt.Errorf("no cache")
}

func (n noCache) GetStringWithLocalTimeout(key string, localExpiration time.Duration) (string, error) {
	return "", fmt.Errorf("no cache")
}

func (n noCache) GetUint64WithLocalTimeout(key string, localExpiration time.Duration) (uint64, error) {
	return 0, fmt.Errorf("no cache")
}

func (n noCache) GetBoolWithLocalTimeout(key string, localExpiration time.Duration) (bool, error) {
	return false, fmt.Errorf("no cache")
}

func (n noCache) GetWithLocalTimeout(key string, localExpiration time.Duration, returnValue interface{}) (interface{}, error) {
	return nil, fmt.Errorf("no cache")
}

type noSQLReaderDb struct {
}

func (n noSQLReaderDb) Close() error {
	//TODO implement me
	panic("implement me")
}

func (n noSQLReaderDb) Get(dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (n noSQLReaderDb) Select(dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (n noSQLReaderDb) Query(query string, args ...any) (*sql.Rows, error) {
	//TODO implement me
	panic("implement me")
}

func (n noSQLReaderDb) Preparex(query string) (*sqlx.Stmt, error) {
	//TODO implement me
	panic("implement me")
}
