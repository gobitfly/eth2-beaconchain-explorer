package exporter

import (
	"context"
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gethRPC "github.com/ethereum/go-ethereum/rpc"
	"github.com/jmoiron/sqlx"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rpTypes "github.com/rocket-pool/rocketpool-go/types"
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
		err = rpExporter.Export()
		if err != nil {
			logger.Errorf("error when exporting rocketpool: %v", err)
		}
		time.Sleep(time.Second * 60)
	}
}

type RocketpoolExporter struct {
	Eth1Client *ethclient.Client
	API        *rocketpool.RocketPool
	DB         *sqlx.DB
}

type RocketpoolExporterMinipoolData struct {
	BalancesBlock         uint64
	TotalETHBalance       *big.Int `db:"total_eth_balance"`
	StakingETHBalance     *big.Int `db:"staking_eth_balance"`
	TotalRETHSupply       *big.Int `db:"total_reth_supply"`
	ETHUtilizationRate    float64  `db:"eth_utilization_rate"`
	NodeDemand            *big.Int `db:"node_demand"`
	NodeFee               float64  `db:"node_fee"`
	NodeFeeByDemand       float64  `db:"node_fee_by_demand"`
	WithdrawalBalance     float64  `db:"withdrawal_balance"`
	WithdrawalCredentials float64  `db:"withdrawal_credentials"`
}

type RocketpoolExporterStatsData struct {
	BalancesBlock            uint64
	RocketpoolStorageAddress string
	TotalETHBalance          *big.Int `db:"total_eth_balance"`
	StakingETHBalance        *big.Int `db:"staking_eth_balance"`
	TotalRETHSupply          *big.Int `db:"total_reth_supply"`
	ETHUtilizationRate       float64  `db:"eth_utilization_rate"`
	NodeDemand               *big.Int `db:"node_demand"`
	NodeFee                  float64  `db:"node_fee"`
	NodeFeeByDemand          float64  `db:"node_fee_by_demand"`
	WithdrawalBalance        float64  `db:"withdrawal_balance"`
	WithdrawalCredentials    float64  `db:"withdrawal_credentials"`
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

func checkErrSoft(err error) {
	if err != nil {
		logger.Error(err)
	}
}

func checkErrSoftBench(vs map[string]interface{}, v interface{}, err error, t *time.Time, name string) {
	d := time.Since(*t)
	*t = time.Now()

	vs[name] = v

	if err != nil {
		logger.WithError(err).WithField("duration", d).Error(name)
	} else {
		logger.WithField("duration", d).Info(name)
	}
}

func (rp *RocketpoolExporter) Test() {
	rp.Export()

	addr, err := rpTypes.HexToValidatorPubkey("8c1847cbd2a38e309bfaf79af8d114e896c75034c29bc2a6226db8349110f00e7de6cd8f23d2098c48b9093530611867")
	if err != nil {
		panic(err)
	}
	mp, err := minipool.GetMinipoolByPubkey(rp.API, addr, nil)
	if err != nil {
		panic(err)
	}
	mpd, err := minipool.GetMinipoolDetails(rp.API, mp, nil)
	if err != nil {
		panic(err)
	}
	n, err := minipool.GetNodeMinipoolAddresses()
	fmt.Printf("%+v\n", mpd)
	if true {
		return
	}

	c, err := minipool.GetMinipoolCount(rp.API, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(c)
	for i := uint64(0); i < c; i++ {
		addr, err := minipool.GetMinipoolAt(rp.API, i, nil)
		if err != nil {
			panic(err)
		}
		k, err := minipool.GetMinipoolPubkey(rp.API, addr, nil)
		if err != nil {
			panic(err)
		}
		fmt.Println(k)
		if i != 0 && i%10 == 0 {
			time.Sleep(time.Second * 2)
		}
	}
}

func (rp *RocketpoolExporter) Export() error {
	var err error
	latestBlock, err := rp.Eth1Client.BlockNumber(context.Background())
	if err != nil {
		return err
	}
	_ = latestBlock

	vs := map[string]interface{}{}
	var v interface{}

	t := time.Now()

	v, err = network.GetBalancesBlock(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetBalancesBlock")
	v, err = network.GetTotalETHBalance(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetTotalETHBalance")
	v, err = network.GetStakingETHBalance(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetStakingETHBalance")
	v, err = network.GetTotalRETHSupply(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetTotalRETHSupply")
	v, err = network.GetETHUtilizationRate(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetETHUtilizationRate")
	v, err = network.GetNodeDemand(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetNodeDemand")
	v, err = network.GetNodeFee(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetNodeFee")
	// v, err = network.GetNodeFeeByDemand(rp.API, d.NodeDemand, nil)
	// checkErrSoftBench(vs, v, err, &t, "GetNodeFeeByDemand")
	v, err = network.GetRPLPrice(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetRPLPrice")
	v, err = network.GetPricesBlock(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetRPLPrice")

	v, err = minipool.GetMinipoolCount(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetMinipoolCount")
	v, err = minipool.GetQueueLengths(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetQueueLengths")
	v, err = minipool.GetQueueCapacity(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetQueueCapacity")
	v, err = minipool.GetQueueTotalLength(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetQueueTotalLength")
	v, err = minipool.GetQueueTotalCapacity(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetQueueTotalCapacity")
	v, err = minipool.GetQueueEffectiveCapacity(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetQueueEffectiveCapacity")
	v, err = minipool.GetQueueNextCapacity(rp.API, nil)
	checkErrSoftBench(vs, v, err, &t, "GetQueueNextCapacity")

	for i := uint64(0); i < 5; i++ {
		n, err := node.GetNodeAt(rp.API, i, nil)
		checkErrSoftBench(vs, n, err, &t, fmt.Sprintf("GetNodeAt_%v", i))
		d, err := node.GetNodeDetails(rp.API, n, nil)
		checkErrSoftBench(vs, d, err, &t, fmt.Sprintf("GetNodeDetails_%v", i))
	}

	for i := uint64(0); i < 5; i++ {
		n, err := minipool.GetMinipoolAt(rp.API, i, nil)
		checkErrSoftBench(vs, n, err, &t, fmt.Sprintf("GetMinipoolAt_%v", i))
		d, err := minipool.GetMinipoolDetails(rp.API, n, nil)
		checkErrSoftBench(vs, d, err, &t, fmt.Sprintf("GetMinipoolDetails_%v", i))
	}

	fmt.Printf("%+v\n", vs)
	return nil
}
