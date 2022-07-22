package db

import (
	"context"
	"errors"
	"eth2-exporter/types"
	"fmt"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/golang/protobuf/proto"
)

var ErrBlockNotFound = errors.New("block not found")

type Bigtable struct {
	client      *gcp_bigtable.Client
	tableData   *gcp_bigtable.Table
	tableBlocks *gcp_bigtable.Table
	chainId     string
}

func NewBigtable(project, instance, chainId string) (*Bigtable, error) {
	btClient, err := gcp_bigtable.NewClient(context.Background(), project, instance)

	if err != nil {
		return nil, err
	}

	bt := &Bigtable{
		client:      btClient,
		tableData:   btClient.Open("data"),
		tableBlocks: btClient.Open("blocks"),
		chainId:     chainId,
	}
	return bt, nil
}

func (bigtable *Bigtable) Close() {
	bigtable.client.Close()
}

func (bigtable *Bigtable) SaveBlock(block *types.Eth1Block) error {

	encodedBc, err := proto.Marshal(block)

	if err != nil {
		return err
	}
	family := "default"
	ts := gcp_bigtable.Timestamp(0)

	mut := gcp_bigtable.NewMutation()
	mut.Set(family, "data", ts, encodedBc)

	err = bigtable.tableBlocks.Apply(context.Background(), fmt.Sprintf("%s:%s", bigtable.chainId, reversedPaddedBlockNumber(block.Number)), mut)

	if err != nil {
		return err
	}
	return nil
}

const max_block_number = 1000000000

func reversedPaddedBlockNumber(blockNumber uint64) string {

	return fmt.Sprintf("%09d", max_block_number-blockNumber)
}
