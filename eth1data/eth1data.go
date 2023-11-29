package eth1data

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "eth1data")
var ErrTxIsPending = errors.New("error retrieving data for tx: tx is still pending")

func GetEth1Transaction(hash common.Hash) (*types.Eth1TxData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cacheKey := fmt.Sprintf("%d:tx:%s", utils.Config.Chain.ClConfig.DepositChainID, hash.String())

	if wanted, err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Hour, new(types.Eth1TxData)); err == nil {
		logger.Infof("retrieved data for tx %v from cache", hash)
		logger.Trace(wanted)

		data := wanted.(*types.Eth1TxData)
		if data.BlockNumber != 0 {
			if err := db.GetBlockStatus(data.BlockNumber, services.LatestFinalizedEpoch(), &data.Epoch); err != nil {
				logger.Warningf("failed to get finalization stats for block %v", data.BlockNumber)
				data.Epoch.Finalized = false
				data.Epoch.Participation = -1
			}
		}
		return data, nil
	}
	tx, pending, err := rpc.CurrentErigonClient.GetNativeClient().TransactionByHash(ctx, hash)

	if err != nil {
		return nil, fmt.Errorf("error retrieving data for tx: %w", err)
	}

	if pending {
		return nil, ErrTxIsPending
	}

	txPageData := &types.Eth1TxData{
		Hash:      tx.Hash(),
		CallData:  fmt.Sprintf("0x%x", tx.Data()),
		Value:     tx.Value().Bytes(),
		IsPending: pending,
		Events:    make([]*types.Eth1EventData, 0, 10),
	}

	receipt, err := getTransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving receipt data for tx: %w", err)
	}

	txPageData.Receipt = receipt

	txPageData.To = tx.To()

	if txPageData.To == nil {
		txPageData.To = &receipt.ContractAddress
		txPageData.IsContractCreation = true
	}
	txPageData.TargetIsContract, err = IsContract(ctx, *txPageData.To)
	if err != nil {
		return nil, fmt.Errorf("error retrieving code data for tx recipient %v: %w", tx.To(), err)
	}

	header, err := getBlockHeaderByHash(ctx, receipt.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block header data for tx: %w", err)
	}
	txPageData.BlockNumber = header.Number.Int64()
	txPageData.Timestamp = time.Unix(int64(header.Time), 0)

	msg, err := core.TransactionToMessage(tx, geth_types.NewCancunSigner(tx.ChainId()), header.BaseFee)
	if err != nil {
		return nil, fmt.Errorf("error getting sender of tx: %w", err)
	}
	txPageData.From = msg.From
	txPageData.Nonce = msg.Nonce
	txPageData.Type = receipt.Type
	txPageData.TypeFormatted = utils.FormatTransactionType(receipt.Type)
	txPageData.TxnPosition = receipt.TransactionIndex

	txPageData.Gas.MaxPriorityFee = msg.GasTipCap.Bytes()
	txPageData.Gas.MaxFee = msg.GasFeeCap.Bytes()
	if header.BaseFee != nil {
		txPageData.Gas.BlockBaseFee = header.BaseFee.Bytes()
	}
	txPageData.Gas.Used = receipt.GasUsed
	txPageData.Gas.Limit = msg.GasLimit
	txPageData.Gas.UsedPerc = float64(receipt.GasUsed) / float64(msg.GasLimit)
	if receipt.Type >= 2 {
		tmp := new(big.Int)
		tmp.Add(tmp, header.BaseFee)
		if t := *new(big.Int).Sub(msg.GasFeeCap, tmp); t.Cmp(msg.GasTipCap) == -1 {
			tmp.Add(tmp, &t)
		} else {
			tmp.Add(tmp, msg.GasTipCap)
		}
		txPageData.Gas.EffectiveFee = tmp.Bytes()
		txPageData.Gas.TxFee = tmp.Mul(tmp, big.NewInt(int64(receipt.GasUsed))).Bytes()
	} else {
		txPageData.Gas.EffectiveFee = msg.GasFeeCap.Bytes()
		txPageData.Gas.TxFee = msg.GasFeeCap.Mul(msg.GasFeeCap, big.NewInt(int64(receipt.GasUsed))).Bytes()
	}

	if receipt.Type == 3 {
		txPageData.Gas.BlobGasPrice = receipt.BlobGasPrice.Bytes()
		txPageData.Gas.BlobGasUsed = receipt.BlobGasUsed
		txPageData.Gas.BlobTxFee = new(big.Int).Mul(receipt.BlobGasPrice, big.NewInt(int64(txPageData.Gas.BlobGasUsed))).Bytes()

		txPageData.BlobHashes = make([][]byte, len(tx.BlobHashes()))
		for i, h := range tx.BlobHashes() {
			txPageData.BlobHashes[i] = h.Bytes()
		}
	}

	if receipt.Status != 1 {
		data, err := rpc.CurrentErigonClient.TraceParityTx(tx.Hash().Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get parity trace for revert reason: %w", err)
		}
		errorMsg, err := abi.UnpackRevert(utils.MustParseHex(data[0].Result.Output))
		if err == nil {
			txPageData.ErrorMsg = errorMsg
		}
	}
	if receipt.Status == 1 {
		txPageData.Transfers, err = db.BigtableClient.GetArbitraryTokenTransfersForTransaction(tx.Hash().Bytes())
		if err != nil {
			return nil, fmt.Errorf("error loading token transfers from tx: %w", err)
		}
		txPageData.InternalTxns, err = db.BigtableClient.GetInternalTransfersForTransaction(tx.Hash().Bytes(), msg.From.Bytes())
		if err != nil {
			return nil, fmt.Errorf("error loading internal transfers from tx: %w", err)
		}
	}
	txPageData.FromName, err = db.BigtableClient.GetAddressName(msg.From.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error retrieveing from name for tx: %w", err)
	}
	if msg.To != nil {
		txPageData.ToName, err = db.BigtableClient.GetAddressName(msg.To.Bytes())
		if err != nil {
			return nil, fmt.Errorf("error retrieveing to name for tx: %w", err)
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
			if cmEntry.err != nil || cmEntry.meta == nil || cmEntry.meta.ABI == nil {
				name := ""
				if len(log.Topics) > 0 {
					name = db.BigtableClient.GetEventLabel(log.Topics[0][:])
				}
				eth1Event := &types.Eth1EventData{
					Address: log.Address,
					Name:    name,
					Topics:  log.Topics,
					Data:    log.Data,
				}

				txPageData.Events = append(txPageData.Events, eth1Event)
			} else {
				boundContract := bind.NewBoundContract(*txPageData.To, *cmEntry.meta.ABI, nil, nil, nil)

				for name, event := range cmEntry.meta.ABI.Events {
					if log != nil && len(log.Topics) > 0 && bytes.Equal(event.ID.Bytes(), log.Topics[0].Bytes()) {
						logData := make(map[string]interface{})
						err := boundContract.UnpackLogIntoMap(logData, name, *log)

						if err != nil {
							logger.Warnf("error decoding event [%v] for tx [0x%x]", name, tx.Hash())
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
								Value: fmt.Sprintf("%v", val),
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
		if err := db.GetBlockStatus(txPageData.BlockNumber, services.LatestFinalizedEpoch(), &txPageData.Epoch); err != nil {
			logger.Warningf("failed to get finalization stats for block %v: %v", txPageData.BlockNumber, err)
			txPageData.Epoch.Finalized = false
			txPageData.Epoch.Participation = -1
		}
	}

	// staking deposit information (only add complete events if any)
	for _, v := range txPageData.Events {
		if v.Address == common.HexToAddress(utils.Config.Chain.ClConfig.DepositContractAddress) && strings.HasPrefix(v.Name, "DepositEvent") {
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

	err = cache.TieredCache.Set(cacheKey, txPageData, utils.Day)
	if err != nil {
		return nil, fmt.Errorf("error writing data for tx to cache: %w", err)
	}

	return txPageData, nil
}

func IsContract(ctx context.Context, address common.Address) (bool, error) {
	cacheKey := fmt.Sprintf("%d:isContract:%s", utils.Config.Chain.ClConfig.DepositChainID, address.String())
	if wanted, err := cache.TieredCache.GetBoolWithLocalTimeout(cacheKey, time.Hour); err == nil {
		return wanted, nil
	}

	code, err := rpc.CurrentErigonClient.GetNativeClient().CodeAt(ctx, address, nil)
	if err != nil {
		return false, fmt.Errorf("error retrieving code data for address %v: %w", address, err)
	}

	isContract := len(code) != 0
	err = cache.TieredCache.SetBool(cacheKey, isContract, utils.Day)
	if err != nil {
		return false, fmt.Errorf("error writing code data for address %v to cache: %w", address, err)
	}

	return isContract, nil
}

func getBlockHeaderByHash(ctx context.Context, hash common.Hash) (*geth_types.Header, error) {
	header, err := rpc.CurrentErigonClient.GetNativeClient().HeaderByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block header data for tx: %w", err)
	}

	return header, nil
}

func getTransactionReceipt(ctx context.Context, hash common.Hash) (*geth_types.Receipt, error) {
	cacheKey := fmt.Sprintf("%d:r:%s", utils.Config.Chain.ClConfig.DepositChainID, hash.String())

	if wanted, err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Hour, new(geth_types.Receipt)); err == nil {
		logger.Infof("retrieved receipt data for tx %v from cache", hash)
		return wanted.(*geth_types.Receipt), nil
	}

	receipt, err := rpc.CurrentErigonClient.GetNativeClient().TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving receipt data for tx: %w", err)
	}

	err = cache.TieredCache.Set(cacheKey, receipt, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("error writing receipt data for tx to cache: %w", err)
	}

	return receipt, nil
}
