package rpc

import (
	"eth2-exporter/types"

	"github.com/sirupsen/logrus"
)

// Client provides an interface for RPC clients
type Client interface {
	GetChainHead() (*types.ChainHead, error)
	GetEpochData(epoch uint64, skipHistoricBalances bool) (*types.EpochData, error)
	GetValidatorQueue() (*types.ValidatorQueue, error)
	GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error)
	GetBlocksBySlot(slot uint64) ([]*types.Block, error)
	GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error)
	GetNewBlockChan() chan *types.Block
	GetBlockStatusByEpoch(slot uint64) ([]*types.CanonBlock, error)
	GetFinalityCheckpoints(epoch uint64) (*types.FinalityCheckpoints, error)
	GetSyncCommittee(stateID string, epoch uint64) (*StandardSyncCommittee, error)
	GetBalancesForEpoch(epoch int64) (map[uint64]uint64, error)
	GetValidatorState(epoch uint64) (*StandardValidatorsResponse, error)
}

type Eth1Client interface {
	GetBlock(number uint64) (*types.Eth1Block, *types.GetBlockTimings, error)
	GetLatestEth1BlockNumber() (uint64, error)
	Close()
}

var logger = logrus.New().WithField("module", "rpc")
