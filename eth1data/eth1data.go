package eth1data

import (
	"bytes"
	"context"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/go-redis/cache/v8"
	"github.com/sirupsen/logrus"
)

func GetEth1Transaction(hash common.Hash) (*types.Eth1TxData, error) {
	cacheKey := fmt.Sprintf("tx:%s", hash.String())
	wanted := &types.Eth1TxData{}
	if err := db.RedisCache.Get(context.Background(), cacheKey, wanted); err == nil {
		logrus.Infof("retrieved data for tx %v from cache", hash)

		return wanted, nil
	}

	tx, pending, err := rpc.CurrentErigonClient.GetNativeClient().TransactionByHash(context.Background(), hash)

	if err != nil {
		return nil, fmt.Errorf("error retrieving data for tx %v: %v", hash, err)
	}

	if pending {
		return nil, fmt.Errorf("error retrieving data for tx %v: tx is still pending", hash)
	}

	txPageData := &types.Eth1TxData{
		GethTx:    tx,
		IsPending: pending,
		Events:    make([]*types.Eth1EventData, 0, 10),
	}

	receipt, err := GetTransactionReceipt(hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving receipt data for tx %v: %v", hash, err)
	}

	txPageData.Receipt = receipt
	txPageData.TxFee = new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(receipt.GasUsed))

	txPageData.To = tx.To()

	if txPageData.To == nil {
		txPageData.To = &receipt.ContractAddress
		txPageData.IsContractCreation = true
	}
	code, err := GetCodeAt(*txPageData.To)
	if err != nil {
		return nil, fmt.Errorf("error retrieving code data for tx %v receipient %v: %v", hash, tx.To(), err)
	}
	txPageData.TargetIsContract = len(code) != 0

	header, err := GetBlockHeaderByHash(receipt.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block header data for tx %v: %v", hash, err)
	}
	txPageData.BlockNumber = header.Number.Int64()
	txPageData.Timestamp = header.Time

	msg, err := tx.AsMessage(geth_types.NewLondonSigner(tx.ChainId()), header.BaseFee)
	if err != nil {
		return nil, fmt.Errorf("error converting tx %v to message: %v", hash, err)
	}
	txPageData.From = msg.From()

	if len(receipt.Logs) > 0 {
		for _, log := range receipt.Logs {
			meta, err := db.BigtableClient.GetContractMetadata(log.Address.Bytes())

			if err != nil {
				logrus.Errorf("error retrieving abi for contract %v: %v", tx.To(), err)
				eth1Event := &types.Eth1EventData{
					Address:     log.Address,
					Name:        "",
					Topics:      log.Topics,
					Data:        log.Data,
					DecodedData: map[string]string{},
				}

				txPageData.Events = append(txPageData.Events, eth1Event)
			} else {
				txPageData.ToName = meta.Name
				boundContract := bind.NewBoundContract(*txPageData.To, *meta.ABI, nil, nil, nil)

				for name, event := range meta.ABI.Events {
					if bytes.Equal(event.ID.Bytes(), log.Topics[0].Bytes()) {
						logData := make(map[string]interface{})
						err := boundContract.UnpackLogIntoMap(logData, name, *log)

						if err != nil {
							logrus.Errorf("error decoding event %v", name)
						}

						eth1Event := &types.Eth1EventData{
							Address:     log.Address,
							Name:        strings.Replace(event.String(), "event ", "", 1),
							Topics:      log.Topics,
							Data:        log.Data,
							DecodedData: map[string]string{},
						}

						for name, val := range logData {
							eth1Event.DecodedData[name] = fmt.Sprintf("0x%x", val)
						}

						txPageData.Events = append(txPageData.Events, eth1Event)
					}
				}
			}
		}

		//

		// for _, log := range receipt.Logs {
		// 	var unpackedLog interface{}
		// 	boundContract.UnpackLog(unpackedLog, )
		// }
	}

	err = db.RedisCache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   cacheKey,
		Value: txPageData,
		TTL:   0,
	})
	if err != nil {
		return nil, fmt.Errorf("error writing data for tx %v to cache: %v", hash, err)
	}

	return txPageData, nil
}

func GetCodeAt(address common.Address) ([]byte, error) {
	cacheKey := fmt.Sprintf("a:%s", address.String())
	wanted := []byte{}
	if err := db.RedisCache.Get(context.Background(), cacheKey, &wanted); err == nil {
		logrus.Infof("retrieved code data for address %v from cache", address)

		return wanted, nil
	}

	code, err := rpc.CurrentErigonClient.GetNativeClient().CodeAt(context.Background(), address, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving code data for address %v: %v", address, err)
	}

	err = db.RedisCache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   cacheKey,
		Value: code,
		TTL:   0,
	})
	if err != nil {
		return nil, fmt.Errorf("error writing code data for address %v to cache: %v", address, err)
	}

	return code, nil
}

func GetBlockHeaderByHash(hash common.Hash) (*geth_types.Header, error) {
	cacheKey := fmt.Sprintf("h:%s", hash.String())

	wanted := &geth_types.Header{}
	if err := db.RedisCache.Get(context.Background(), cacheKey, &wanted); err == nil {
		logrus.Infof("retrieved header data for block %v from cache", hash)
		return wanted, nil
	}

	header, err := rpc.CurrentErigonClient.GetNativeClient().HeaderByHash(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block header data for tx %v: %v", hash, err)
	}

	err = db.RedisCache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   cacheKey,
		Value: header,
		TTL:   0,
	})
	if err != nil {
		return nil, fmt.Errorf("error writing header data for block %v to cache: %v", hash, err)
	}

	return header, nil
}

func GetTransactionReceipt(hash common.Hash) (*geth_types.Receipt, error) {
	cacheKey := fmt.Sprintf("r:%s", hash.String())

	wanted := &geth_types.Receipt{}
	if err := db.RedisCache.Get(context.Background(), cacheKey, &wanted); err == nil {
		logrus.Infof("retrieved receipt data for tx %v from cache", hash)
		return wanted, nil
	}

	receipt, err := rpc.CurrentErigonClient.GetNativeClient().TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving receipt data for tx %v: %v", hash, err)
	}

	err = db.RedisCache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   cacheKey,
		Value: receipt,
		TTL:   0,
	})
	if err != nil {
		return nil, fmt.Errorf("error writing receipt data for tx %v to cache: %v", hash, err)
	}

	return receipt, nil
}
