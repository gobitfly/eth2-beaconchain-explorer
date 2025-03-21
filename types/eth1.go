package types

import (
	"encoding/binary"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type GetBlockTimings struct {
	Headers  time.Duration
	Receipts time.Duration
	Traces   time.Duration
}

type ElWithdrawalRequestData []byte

func (e ElWithdrawalRequestData) GetSourceAddressBytes() ([]byte, error) {
	if len(e) < 20 {
		return nil, errors.New("not enough bytes")
	}
	return e[:20], nil
}

func (e ElWithdrawalRequestData) GetSourceAddress() (common.Address, error) {
	adressBytes, err := e.GetSourceAddressBytes()
	if err != nil {
		return common.Address{}, err
	}
	return common.BytesToAddress(adressBytes), nil
}

func (e ElWithdrawalRequestData) GetValidatorPubkey() ([]byte, error) {
	if len(e) < 68 {
		return nil, errors.New("not enough bytes")
	}
	return e[20:68], nil
}

func (e ElWithdrawalRequestData) GetAmount() (*big.Int, error) {
	if len(e) < 76 {
		return nil, errors.New("not enough bytes")
	}

	return new(big.Int).SetBytes(e[68:76]), nil
}

func (e ElWithdrawalRequestData) GetAmountUint64() (uint64, error) {
	if len(e) < 76 {
		return 0, errors.New("not enough bytes")
	}
	amount := binary.BigEndian.Uint64(e[68:76])
	if amount > math.MaxInt64 {
		amount = math.MaxInt64
	}
	return amount, nil
}

type ElConsolidationRequestData []byte

func (e ElConsolidationRequestData) GetSourceAddressBytes() ([]byte, error) {
	if len(e) < 20 {
		return nil, errors.New("not enough bytes")
	}
	return e[:20], nil
}

func (e ElConsolidationRequestData) GetSourceAddress() (common.Address, error) {
	adressBytes, err := e.GetSourceAddressBytes()
	if err != nil {
		return common.Address{}, err
	}
	return common.BytesToAddress(adressBytes), nil
}

func (e ElConsolidationRequestData) GetSourceValidatorPubkey() ([]byte, error) {
	if len(e) < 68 {
		return nil, errors.New("not enough bytes")
	}
	return e[20:68], nil
}

func (e ElConsolidationRequestData) GetTargetValidatorPubkey() ([]byte, error) {
	if len(e) < 116 {
		return nil, errors.New("not enough bytes")
	}
	return e[68:116], nil
}
