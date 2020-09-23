package rpc

import (
	"eth2-exporter/eth2api"
	"eth2-exporter/types"
	"time"
)

type Eth2ApiV1Client struct {
	client *eth2api.Client
}

func NewEth2ApiV1Client(endpoint string) (*Eth2ApiV1Client, error) {
	client := &Eth2ApiV1Client{
		client: &eth2api.Client{
			URL:     endpoint,
			Timeout: time.Second * 10,
		},
	}
	return client, nil
}

func (c *Eth2ApiV1Client) GetChainHead() (*types.ChainHead, error) {
	return nil, nil
}

func (c *Eth2ApiV1Client) GetEpochData(epoch uint64) (*types.EpochData, error) {
	return nil, nil
}

func (c *Eth2ApiV1Client) GetValidatorQueue() (*types.ValidatorQueue, error) {
	return nil, nil
}

func (c *Eth2ApiV1Client) GetAttestationPool() ([]*types.Attestation, error) {
	return []*types.Attestation{}, nil
}

func (c *Eth2ApiV1Client) GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error) {
	return nil, nil
}

func (c *Eth2ApiV1Client) GetBlocksBySlot(slot uint64) ([]*types.Block, error) {
	return []*types.Block{}, nil
}

func (c *Eth2ApiV1Client) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {
	return nil, nil
}
