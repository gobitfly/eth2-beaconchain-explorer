package db

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"

	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

func TestTxRevertTransformer(t *testing.T) {
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
	ReaderDb = noSQLReaderDb{}

	bt, err := InitBigtableWithCache(context.Background(), "test", "instanceTest", "1", noRedis{})
	if err != nil {
		t.Fatal(err)
	}
	if err := InitBigtableSchema(); err != nil {
		if !errors.Is(err, ErrTableAlreadyExist) {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		block     int64
		txHash    string
		addresses []string
		expected  []string
	}{
		{
			name:      "partial",
			block:     20183291,
			txHash:    "0xf7d385f000250c073dfef9a36327c5d30a4c77a0c50588ce3eded29f6829a4cd",
			addresses: []string{"0x96abc34501e9fc274f6d4e39cbb4004c0f6e519f"},
			expected:  []string{"Transaction partially executed"},
		},
		{
			name:      "failed",
			block:     20929404,
			txHash:    "0xcce69bddc2b427ecf2f02120b74cde9f5d95f36849f4617fbb31527982daf88c",
			addresses: []string{"0x0d92bC7b13a474937c7C94F882339D68048Af186"},
			expected:  []string{"Transaction failed"},
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
			transformers := []func(blk *types.Eth1Block, cache *freecache.Cache) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error){
				bt.TransformItx,
				bt.TransformTx,
			}
			cache := freecache.NewCache(1 * 1024 * 1024) // 1 MB limit

			if err := bt.IndexEventsWithTransformers(tt.block, tt.block, transformers, 1, cache); err != nil {
				t.Fatal(err)
			}

			for i, address := range tt.addresses {
				res, err := bt.GetAddressTransactionsTableData(common.FromHex(address), "")
				if err != nil {
					t.Fatal(err)
				}
				if got, want := string((res.Data[0][0]).(template.HTML)), tt.expected[i]; !strings.Contains(got, want) {
					t.Errorf("'%s' should contains '%s'", got, want)
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

func (n noSQLReaderDb) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (n noSQLReaderDb) Rebind(query string) string {
	//TODO implement me
	panic("implement me")
}
