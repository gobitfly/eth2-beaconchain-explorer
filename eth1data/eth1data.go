package eth1data

import (
	"bytes"
	"context"
	"encoding/hex"
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "eth1data")

func GetEth1Transaction(hash common.Hash) (*types.Eth1TxData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cacheKey := fmt.Sprintf("%d:tx:%s", utils.Config.Chain.Config.DepositChainID, hash.String())
	if wanted, err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Hour, new(types.Eth1TxData)); err == nil {
		logger.Infof("retrieved data for tx %v from cache", hash)
		logger.Trace(wanted)

		data := wanted.(*types.Eth1TxData)
		if data.BlockNumber != 0 {
			err := db.ReaderDb.Get(&data.Epoch,
				`select epochs.finalized, epochs.globalparticipationrate from blocks left join epochs on blocks.epoch = epochs.epoch where blocks.exec_block_number = $1 and blocks.status='1';`,
				data.BlockNumber)
			if err != nil {
				logger.Warningf("failed to get finalization stats for block %v", data.BlockNumber)
				data.Epoch.Finalized = false
				data.Epoch.Participation = -1
			}
		}

		return data, nil
	}
	tx, pending, err := rpc.CurrentErigonClient.GetNativeClient().TransactionByHash(ctx, hash)

	if err != nil {
		return nil, fmt.Errorf("error retrieving data for tx %v: %v", hash, err)
	}

	if pending {
		return nil, fmt.Errorf("error retrieving data for tx %v: tx is still pending", hash)
	}

	txPageData := &types.Eth1TxData{
		Hash:      tx.Hash(),
		CallData:  fmt.Sprintf("0x%x", tx.Data()),
		Value:     tx.Value().Bytes(),
		IsPending: pending,
		Events:    make([]*types.Eth1EventData, 0, 10),
	}

	receipt, err := GetTransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving receipt data for tx %v: %v", hash, err)
	}

	txPageData.Receipt = receipt

	txPageData.To = tx.To()

	if txPageData.To == nil {
		txPageData.To = &receipt.ContractAddress
		txPageData.IsContractCreation = true
	}
	code, err := GetCodeAt(ctx, *txPageData.To)
	if err != nil {
		return nil, fmt.Errorf("error retrieving code data for tx %v recipient %v: %v", hash, tx.To(), err)
	}
	txPageData.TargetIsContract = len(code) != 0

	header, err := GetBlockHeaderByHash(ctx, receipt.BlockHash)
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
	txPageData.Nonce = msg.Nonce()
	txPageData.Type = receipt.Type
	txPageData.TypeFormatted = utils.FormatTransactionType(receipt.Type)
	txPageData.TxnPosition = receipt.TransactionIndex

	txPageData.Gas.MaxPriorityFee = msg.GasTipCap().Bytes()
	txPageData.Gas.MaxFee = msg.GasFeeCap().Bytes()
	if header.BaseFee != nil {
		txPageData.Gas.BlockBaseFee = header.BaseFee.Bytes()
	}
	txPageData.Gas.Used = receipt.GasUsed
	txPageData.Gas.Limit = msg.Gas()
	txPageData.Gas.UsedPerc = float64(receipt.GasUsed) / float64(msg.Gas())
	if receipt.Type >= 2 {
		tmp := new(big.Int)
		tmp.Add(tmp, header.BaseFee)
		if t := *new(big.Int).Sub(msg.GasFeeCap(), tmp); t.Cmp(msg.GasTipCap()) == -1 {
			tmp.Add(tmp, &t)
		} else {
			tmp.Add(tmp, msg.GasTipCap())
		}
		txPageData.Gas.EffectiveFee = tmp.Bytes()
		txPageData.Gas.TxFee = tmp.Mul(tmp, big.NewInt(int64(receipt.GasUsed))).Bytes()
	} else {
		txPageData.Gas.EffectiveFee = msg.GasFeeCap().Bytes()
		txPageData.Gas.TxFee = msg.GasFeeCap().Mul(msg.GasFeeCap(), big.NewInt(int64(receipt.GasUsed))).Bytes()
	}

	if receipt.Status != 1 {
		data, err := rpc.CurrentErigonClient.TraceParityTx(tx.Hash().Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get parity trace for revert reason: %v", err)
		}
		errorMsg, err := abi.UnpackRevert(utils.MustParseHex(data[0].Result.Output))
		if err == nil {
			txPageData.ErrorMsg = errorMsg
		}
	}
	if receipt.Status == 1 {
		txPageData.Transfers, err = db.BigtableClient.GetArbitraryTokenTransfersForTransaction(tx.Hash().Bytes())
		if err != nil {
			return nil, fmt.Errorf("error loading token transfers from tx %v: %v", hash, err)
		}
		txPageData.InternalTxns, err = db.BigtableClient.GetInternalTransfersForTransaction(tx.Hash().Bytes(), msg.From().Bytes())
		if err != nil {
			return nil, fmt.Errorf("error loading internal transfers from tx %v: %v", hash, err)
		}
	}
	txPageData.FromName, err = db.BigtableClient.GetAddressName(msg.From().Bytes())
	if err != nil {
		return nil, fmt.Errorf("error retrieveing from name for tx %v: %v", hash, err)
	}
	if msg.To() != nil {
		txPageData.ToName, err = db.BigtableClient.GetAddressName(msg.To().Bytes())
		if err != nil {
			return nil, fmt.Errorf("error retrieveing to name for tx %v: %v", hash, err)
		}
	}

	if len(receipt.Logs) > 0 {
		var wasContractMetadataCached bool
		type contractMetadataMapEntry struct {
			err  error
			meta *types.ContractMetadata
		}
		var cmEntry contractMetadataMapEntry
		contractMetadataCache := make(map[common.Address]contractMetadataMapEntry)

		for _, log := range receipt.Logs {
			if cmEntry, wasContractMetadataCached = contractMetadataCache[log.Address]; !wasContractMetadataCached {
				cmEntry.meta, cmEntry.err = db.BigtableClient.GetContractMetadata(log.Address.Bytes())
				contractMetadataCache[log.Address] = cmEntry
			}

			if cmEntry.err != nil || cmEntry.meta == nil {
				if !wasContractMetadataCached {
					logger.Warnf("error retrieving abi for contract %v: %v", tx.To(), cmEntry.err)
				}
				eth1Event := &types.Eth1EventData{
					Address: log.Address,
					Name:    "",
					Topics:  log.Topics,
					Data:    log.Data,
				}

				txPageData.Events = append(txPageData.Events, eth1Event)
			} else {
				boundContract := bind.NewBoundContract(*txPageData.To, *cmEntry.meta.ABI, nil, nil, nil)

				for name, event := range cmEntry.meta.ABI.Events {
					if bytes.Equal(event.ID.Bytes(), log.Topics[0].Bytes()) {
						logData := make(map[string]interface{})
						err := boundContract.UnpackLogIntoMap(logData, name, *log)

						if err != nil {
							logger.Errorf("error decoding event %v", name)
						}

						eth1Event := &types.Eth1EventData{
							Address:     log.Address,
							Name:        strings.Replace(event.String(), "event ", "", 1),
							Topics:      log.Topics,
							Data:        log.Data,
							DecodedData: map[string]types.Eth1DecodedEventData{},
						}
						typeMap := make(map[string]string)
						for _, input := range cmEntry.meta.ABI.Events[name].Inputs {
							typeMap[input.Name] = input.Type.String()
						}

						for lName, val := range logData {
							a := types.Eth1DecodedEventData{
								Type:  typeMap[lName],
								Raw:   fmt.Sprintf("0x%x", val),
								Value: fmt.Sprintf("%s", val),
							}
							b := typeMap[lName]
							if b == "address" {
								a.Address = val.(common.Address)
							}
							if strings.HasPrefix(b, "byte") {
								a.Value = a.Raw
							}
							eth1Event.DecodedData[lName] = a
						}

						txPageData.Events = append(txPageData.Events, eth1Event)
					}
				}
			}
		}
	}

	if txPageData.BlockNumber != 0 {
		err := db.ReaderDb.Get(&txPageData.Epoch,
			`select epochs.finalized, epochs.globalparticipationrate from blocks left join epochs on blocks.epoch = epochs.epoch where blocks.exec_block_number = $1 and blocks.status='1';`,
			&txPageData.BlockNumber)
		if err != nil {
			logger.Warningf("failed to get finalization stats for block %v: %v", txPageData.BlockNumber, err)
			txPageData.Epoch.Finalized = false
			txPageData.Epoch.Participation = -1
		}
	}

	// staking deposit information (only add complete events if any)
	for _, v := range txPageData.Events {
		if v.Address == common.HexToAddress(utils.Config.Chain.Config.DepositContractAddress) && strings.HasPrefix(v.Name, "DepositEvent") {
			var d types.DepositContractInteraction

			if pubkey, found := v.DecodedData["pubkey"]; found {
				d.ValidatorPubkey, err = hex.DecodeString(pubkey.Raw[2:])
				if err != nil {
					continue
				}
			} else {
				continue
			}

			if wcreds, found := v.DecodedData["withdrawal_credentials"]; found {
				d.WithdrawalCreds, err = hex.DecodeString(wcreds.Raw[2:])
				if err != nil {
					continue
				}
			} else {
				continue
			}

			if amount, found := v.DecodedData["amount"]; found {
				// amount is a little endian hex denominated in GEwei so we have to decode and reverse it and then convert to ETH
				ba, err := hex.DecodeString(amount.Raw[2:])
				if err != nil {
					continue
				}
				utils.ReverseSlice(ba)
				amount := new(big.Int).Mul(new(big.Int).SetBytes(ba), big.NewInt(1000000000))

				d.Amount = amount.Bytes()
			} else {
				continue
			}

			txPageData.DepositContractInteractions = append(txPageData.DepositContractInteractions, d)
		}
	}

	err = cache.TieredCache.Set(cacheKey, txPageData, time.Hour*24)
	if err != nil {
		return nil, fmt.Errorf("error writing data for tx %v to cache: %v", hash, err)
	}

	return txPageData, nil
}

func GetCodeAt(ctx context.Context, address common.Address) ([]byte, error) {
	cacheKey := fmt.Sprintf("%d:a:%s", utils.Config.Chain.Config.DepositChainID, address.String())
	if wanted, err := cache.TieredCache.GetStringWithLocalTimeout(cacheKey, time.Hour); err == nil {
		logger.Infof("retrieved code data for address %v from cache", address)

		return []byte(wanted), nil
	}

	code, err := rpc.CurrentErigonClient.GetNativeClient().CodeAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving code data for address %v: %v", address, err)
	}

	err = cache.TieredCache.SetString(cacheKey, string(code), time.Hour*24)
	if err != nil {
		return nil, fmt.Errorf("error writing code data for address %v to cache: %v", address, err)
	}

	return code, nil
}

func GetBlockHeaderByHash(ctx context.Context, hash common.Hash) (*geth_types.Header, error) {
	// cacheKey := fmt.Sprintf("%d:h:%s", utils.Config.Chain.Config.DepositChainID, hash.String())

	// if wanted, err := db.EkoCache.Get(ctx, cacheKey, new(geth_types.Header)); err == nil {
	// 	logger.Infof("retrieved header data for block %v from cache", hash)
	// 	return wanted.(*geth_types.Header), nil
	// }

	header, err := rpc.CurrentErigonClient.GetNativeClient().HeaderByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block header data for tx %v: %v", hash, err)
	}

	// err = db.EkoCache.Set(ctx, cacheKey, header)
	// if err != nil {
	// 	return nil, fmt.Errorf("error writing header data for block %v to cache: %v", hash, err)
	// }

	return header, nil
}

func GetTransactionReceipt(ctx context.Context, hash common.Hash) (*geth_types.Receipt, error) {
	cacheKey := fmt.Sprintf("%d:r:%s", utils.Config.Chain.Config.DepositChainID, hash.String())

	if wanted, err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Hour, new(geth_types.Receipt)); err == nil {
		logger.Infof("retrieved receipt data for tx %v from cache", hash)
		return wanted.(*geth_types.Receipt), nil
	}

	receipt, err := rpc.CurrentErigonClient.GetNativeClient().TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving receipt data for tx %v: %v", hash, err)
	}

	err = cache.TieredCache.Set(cacheKey, receipt, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("error writing receipt data for tx %v to cache: %v", hash, err)
	}

	return receipt, nil
}
