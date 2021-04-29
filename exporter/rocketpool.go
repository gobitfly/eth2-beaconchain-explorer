package exporter

import (
	"context"
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gethRPC "github.com/ethereum/go-ethereum/rpc"
	"github.com/jmoiron/sqlx"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

var rpEth1RPRCClient *gethRPC.Client
var rpEth1Client *ethclient.Client

func rocketpoolExporter() {
	var err error
	rpEth1RPRCClient, err = gethRPC.Dial(utils.Config.Indexer.Eth1Endpoint)
	if err != nil {
		logger.Fatal(err)
	}
	rpEth1Client = ethclient.NewClient(rpEth1RPRCClient)
	rpExporter, err := NewRocketpoolExporter(rpEth1Client, "0xDeE7ec57E6b9f7F9A27e68821F3763e9d2537096", db.DB)
	if err != nil {
		logger.Fatal(err)
	}
	for {
		rpExporter.export()
		time.Sleep(time.Second * 60)
	}
}

type RocketpoolExporter struct {
	Eth1Client *ethclient.Client
	API        *rocketpool.RocketPool
	DB         *sqlx.DB
}

type RocketpoolExporterData struct {
}

func NewRocketpoolExporter(eth1Client *ethclient.Client, storageContractAddressHex string, db *sqlx.DB) (*RocketpoolExporter, error) {
	rpe := &RocketpoolExporter{}
	rp, err := rocketpool.NewRocketPool(eth1Client, common.HexToAddress(storageContractAddressHex))
	if err != nil {
		return nil, err
	}
	rpe.Eth1Client = eth1Client
	rpe.API = rp
	rpe.DB = db
	return rpe, nil
}

func (rp *RocketpoolExporter) export() error {
	latestBlock, err := rp.Eth1Client.BlockNumber(context.Background())
	if err != nil {
		return err
	}
	ab, err := network.GetBalancesBlock(rp.API, nil)
	if err != nil {
		return err
	}
	fmt.Println("ROCKETPOOL", latestBlock, ab)
	return nil
}
