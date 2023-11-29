package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"eth2-exporter/db"
	"eth2-exporter/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gethRPC "github.com/ethereum/go-ethereum/rpc"
	"github.com/hashicorp/go-version"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/klauspost/compress/zstd"
	"github.com/lib/pq"
	rpDAO "github.com/rocket-pool/rocketpool-go/dao"
	rpDAOTrustedNode "github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	rpTypes "github.com/rocket-pool/rocketpool-go/types"
	rputil "github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	smartnodeCfg "github.com/rocket-pool/smartnode/shared/services/config"
	smartnodeRewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	smartnodeNetwork "github.com/rocket-pool/smartnode/shared/types/config"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var rpEth1RPRCClient *gethRPC.Client
var rpEth1Client *ethclient.Client

const GethEventLogInterval = 25000

var RP_CONFIG *smartnodeCfg.SmartnodeConfig
var firstBlockOfRedstone = map[string]uint64{
	"mainnet": 15451165,
	"prater":  7287326,
	"holesky": 0,
}

var leb16, _ = big.NewInt(0).SetString("16000000000000000000", 10)

func rocketpoolExporter() {
	RP_CONFIG = initRPConfig()
	endpoint := utils.Config.Eth1GethEndpoint
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		endpoint = "ws" + endpoint[4:]
	}

	var err error
	rpEth1RPRCClient, err = gethRPC.Dial(endpoint)
	if err != nil {
		utils.LogFatal(err, "new rocketpool geth client error", 0)
	}
	rpEth1Client = ethclient.NewClient(rpEth1RPRCClient)
	rpExporter, err := NewRocketpoolExporter(
		rpEth1Client,
		RP_CONFIG.GetStorageAddress(),
		db.WriterDb,
	)
	if err != nil {
		utils.LogFatal(err, "new rocketpool exporter error", 0)
	}
	rpExporter.Run()
}

func initRPConfig() *smartnodeCfg.SmartnodeConfig {
	config := smartnodeCfg.NewSmartnodeConfig(&smartnodeCfg.RocketPoolConfig{
		RocketPoolDirectory: "/tmp/rocketpool",
	})
	if utils.Config.Chain.Name == "mainnet" {
		config.Network.Value = smartnodeNetwork.Network_Mainnet
	} else if utils.Config.Chain.Name == "prater" {
		config.Network.Value = smartnodeNetwork.Network_Prater
	} else if utils.Config.Chain.Name == "holesky" {
		config.Network.Value = smartnodeNetwork.Network_Holesky
	} else {
		logrus.Warnf("unknown network")
	}
	return config
}

type RocketpoolNetworkStats struct {
	RPLPrice               *big.Int
	ClaimIntervalTime      time.Duration
	ClaimIntervalTimeStart time.Time
	CurrentNodeFee         float64
	CurrentNodeDemand      *big.Int
	RETHSupply             *big.Int
	NodeOperatorRewards    *big.Int
	RETHPrice              float64
	TotalEthStaking        *big.Int
	TotalEthBalance        *big.Int
}

type RocketpoolExporter struct {
	Eth1Client                         *ethclient.Client
	API                                *rocketpool.RocketPool
	DB                                 *sqlx.DB
	UpdateInterval                     time.Duration
	MinipoolsByAddress                 map[string]*RocketpoolMinipool
	NodesByAddress                     map[string]*RocketpoolNode
	DAOProposalsByID                   map[uint64]*RocketpoolDAOProposal
	DAOMembersByAddress                map[string]*RocketpoolDAOMember
	NodeRPLCumulative                  map[string]*big.Int
	NetworkStats                       RocketpoolNetworkStats
	LastRewardTree                     uint64
	RocketpoolRewardTreesDownloadQueue []RocketpoolRewardTreeDownloadable
	RocketpoolRewardTreeData           map[uint64]RewardsFile
}

type RocketpoolRewardTreeDownloadable struct {
	ID   uint64
	Data []byte
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
	rpe.UpdateInterval = time.Minute
	rpe.MinipoolsByAddress = map[string]*RocketpoolMinipool{}
	rpe.NodesByAddress = map[string]*RocketpoolNode{}
	rpe.DAOProposalsByID = map[uint64]*RocketpoolDAOProposal{}
	rpe.DAOMembersByAddress = map[string]*RocketpoolDAOMember{}
	rpe.LastRewardTree = 0
	rpe.RocketpoolRewardTreesDownloadQueue = []RocketpoolRewardTreeDownloadable{}
	rpe.RocketpoolRewardTreeData = map[uint64]RewardsFile{}
	return rpe, nil
}

func (rp *RocketpoolExporter) Init() error {
	var err error
	err = rp.InitMinipools()
	if err != nil {
		return err
	}
	err = rp.InitNodes()
	if err != nil {
		return err
	}
	err = rp.InitDAOProposals()
	if err != nil {
		return err
	}
	err = rp.InitDAOMembers()
	if err != nil {
		return err
	}
	return nil
}

func (rp *RocketpoolExporter) InitMinipools() error {
	dbRes := []RocketpoolMinipool{}
	err := rp.DB.Select(&dbRes, `select address, pubkey, node_address, node_fee, deposit_type, status,status_time, penalty_count from rocketpool_minipools`)
	if err != nil {
		return err
	}
	for _, val := range dbRes {
		rp.MinipoolsByAddress[fmt.Sprintf("%x", val.Address)] = &val
	}
	return nil
}

func (rp *RocketpoolExporter) InitNodes() error {
	dbRes := []RocketpoolNode{}
	err := rp.DB.Select(&dbRes, `select address, timezone_location, rpl_stake, min_rpl_stake, max_rpl_stake, rpl_cumulative_rewards, smoothing_pool_opted_in, claimed_smoothing_pool, unclaimed_smoothing_pool, unclaimed_rpl_rewards from rocketpool_nodes`)
	if err != nil {
		return err
	}
	for _, val := range dbRes {
		rp.NodesByAddress[fmt.Sprintf("%x", val.Address)] = &val
	}
	return nil
}

func (rp *RocketpoolExporter) InitDAOProposals() error {
	dbRes := []RocketpoolDAOProposal{}
	err := rp.DB.Select(&dbRes, `select id, dao, proposer_address, message, created_time, start_time, end_time, expiry_time, votes_required, votes_for, votes_against, member_voted, member_supported, is_cancelled, is_executed, payload, state from rocketpool_proposals`)
	if err != nil {
		return err
	}
	for _, val := range dbRes {
		rp.DAOProposalsByID[val.ID] = &val
	}
	return nil
}

func (rp *RocketpoolExporter) InitDAOMembers() error {
	dbRes := []RocketpoolDAOMember{}
	err := rp.DB.Select(&dbRes, `select url, address, id, joined_time, last_proposal_time, rpl_bond_amount, unbonded_validator_count from rocketpool_dao_members`)
	if err != nil {
		return err
	}
	for _, val := range dbRes {
		rp.DAOMembersByAddress[fmt.Sprintf("%x", val.Address)] = &val
	}
	return nil
}

func (rp *RocketpoolExporter) Run() error {
	errorInterval := time.Minute
	t := time.NewTicker(rp.UpdateInterval)
	defer t.Stop()
	var count int64 = 0

	isMergeUpdateDeployed, err := IsMergeUpdateDeployed(rp.API)
	if err != nil {
		logger.WithError(err).Errorf("error retrieving rocketpool redstone deploy status")
		return err
	}

	if isMergeUpdateDeployed {
		rp.RocketpoolRewardTreeData, err = rp.getRocketpoolRewardTrees()
		if err != nil {
			logger.WithError(err).Errorf("error retrieving known rocketpool reward tree data from db")
			return err
		}

		for _, data := range rp.RocketpoolRewardTreeData {
			if data.Index > rp.LastRewardTree {
				rp.LastRewardTree = data.Index
			}
		}
	}

	logger.Infof("rocketpool exporter initialized")

	for {
		t0 := time.Now()
		var err error
		err = rp.Update(count)
		if err != nil {
			logger.WithError(err).Errorf("error updating rocketpool-data")
			time.Sleep(errorInterval)
			continue
		}
		err = rp.Save(count)
		if err != nil {
			logger.WithError(err).Errorf("error saving rocketpool-data")
			time.Sleep(errorInterval)
			continue
		}

		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("exported rocketpool-data")
		count++
		<-t.C
	}
}

func (rp *RocketpoolExporter) DownloadMissingRewardTrees() error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("updated rocketpool-reward-trees")
	}(t0)

	isMergeUpdateDeployed, err := IsMergeUpdateDeployed(rp.API)
	if err != nil {
		return err
	}

	if !isMergeUpdateDeployed {
		return nil
	}

	missingIntervals := []rewards.RewardsEvent{}
	for interval := rp.LastRewardTree; ; interval++ {
		var event rewards.RewardsEvent
		logger.Infof("retrieving reward tree %v", interval)
		event, err := smartnodeRewards.GetRewardSnapshotEvent(
			rp.API,
			&smartnodeCfg.RocketPoolConfig{
				Smartnode:    RP_CONFIG,
				IsNativeMode: true,
			},
			interval,
			nil,
		)
		if err != nil {
			if strings.Contains(err.Error(), "found") { // could not be found && not found
				logger.Infof("retrieving reward tree not found %v", interval)
				break
			} else {
				logger.WithError(err).Errorf("retrieving reward tree not found %v", interval)
				return err
			}
		}

		_, exists := rp.RocketpoolRewardTreeData[event.Index.Uint64()]
		if !exists {
			missingIntervals = append(missingIntervals, event)
		} else {
			rp.LastRewardTree = interval + 1
		}

	}

	logger.Infof("downloading %v reward trees", len(missingIntervals))
	if len(missingIntervals) == 0 {
		return nil
	}

	for _, missingInterval := range missingIntervals {
		if contains(rp.RocketpoolRewardTreesDownloadQueue, missingInterval.Index.Uint64()) {
			continue
		}

		bytes, err := DownloadRewardsFile(fmt.Sprintf("rp-rewards-%v-%v.json", utils.Config.Chain.Name, missingInterval.Index), missingInterval.Index.Uint64(), missingInterval.MerkleTreeCID, true)
		if err != nil {
			return fmt.Errorf("can not download reward file %v, error: %w", missingInterval.Index, err)
		}

		proofWrapper, err := getRewardsData(bytes)

		merkleRootFromFile := common.HexToHash(proofWrapper.MerkleRoot)
		if missingInterval.MerkleRoot != merkleRootFromFile {
			return fmt.Errorf("invalid merkle root value : %w", err)
		}

		rp.RocketpoolRewardTreesDownloadQueue = append(rp.RocketpoolRewardTreesDownloadQueue, RocketpoolRewardTreeDownloadable{
			ID:   missingInterval.Index.Uint64(),
			Data: bytes,
		})

		logrus.Infof("Downloaded rocketpool reward tree %v", missingInterval.Index)

		if missingInterval.Index.Uint64() > rp.LastRewardTree {
			rp.LastRewardTree = missingInterval.Index.Uint64()
		}
	}

	return nil
}

func contains(s []RocketpoolRewardTreeDownloadable, e uint64) bool {
	for _, a := range s {
		if a.ID == e {
			return true
		}
	}
	return false
}

func (rp *RocketpoolExporter) Update(count int64) error {
	var wg errgroup.Group
	wg.Go(func() error {
		if count == 0 || count%5 == 4 { // run download one iteration before we update nodes
			return rp.DownloadMissingRewardTrees()
		}
		return nil
	})
	wg.Go(func() error { return rp.UpdateMinipools() })
	wg.Go(func() error { return rp.UpdateNodes(count%5 == 0) })
	wg.Go(func() error { return rp.UpdateDAOProposals() })
	wg.Go(func() error { return rp.UpdateDAOMembers() })
	wg.Go(func() error { return rp.UpdateNetworkStats() })
	return wg.Wait()
}

func (rp *RocketpoolExporter) Save(count int64) error {
	var err error
	err = rp.SaveMinipools()
	if err != nil {
		return err
	}
	err = rp.SaveNodes()
	if err != nil {
		return err
	}
	err = rp.SaveDAOProposals()
	if err != nil {
		return err
	}
	err = rp.SaveDAOProposalsMemberVotes()
	if err != nil {
		return err
	}
	err = rp.SaveDAOMembers()
	if err != nil {
		return err
	}
	err = rp.TagValidators()
	if err != nil {
		return err
	}
	if count%5 == 0 { // smart contracts aren't updated that often, so lets save it less often
		err = rp.SaveNetworkStats()
		if err != nil {
			return err
		}
	}
	err = rp.SaveRewardTrees()
	if err != nil {
		return err
	}

	return nil
}

func (rp *RocketpoolExporter) UpdateMinipools() error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("updated rocketpool-minipools")
	}(t0)

	minipoolAddresses, err := minipool.GetMinipoolAddresses(rp.API, nil)
	if err != nil {
		return err
	}

	atlasDeployed, err := IsAtlasDeployed(rp.API)
	if err != nil {
		return err
	}

	for _, a := range minipoolAddresses {
		addrHex := a.Hex()
		if mp, exists := rp.MinipoolsByAddress[addrHex]; exists {
			err = mp.Update(rp.API, atlasDeployed)
			if err != nil {
				return err
			}
			continue
		}
		mp, err := NewRocketpoolMinipool(rp.API, a.Bytes(), atlasDeployed)
		if err != nil {
			return err
		}
		rp.MinipoolsByAddress[addrHex] = mp
	}
	return nil
}

func (rp *RocketpoolExporter) UpdateNodes(includeCumulativeRpl bool) error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("updated rocketpool-nodes")
	}(t0)

	nodeAddresses, err := node.GetNodeAddresses(rp.API, nil)
	if err != nil {
		return err
	}

	atlasDeployed, err := IsAtlasDeployed(rp.API)
	if err != nil {
		return err
	}

	if includeCumulativeRpl {
		legacyRewardsPool := RP_CONFIG.GetV100RewardsPoolAddress()
		legacyClaimNode := RP_CONFIG.GetV100ClaimNodeAddress()
		rp.NodeRPLCumulative, err = CalculateLifetimeNodeRewardsAllLegacy(
			rp.API,
			big.NewInt(GethEventLogInterval),
			&legacyRewardsPool,
			&legacyClaimNode,
		)
		if err != nil {
			return err
		}
	}

	for _, a := range nodeAddresses {
		addrHex := a.Hex()
		if node, exists := rp.NodesByAddress[addrHex]; exists {
			err = node.Update(rp.API, rp.RocketpoolRewardTreeData, includeCumulativeRpl, rp.NodeRPLCumulative, atlasDeployed)
			if err != nil {
				return err
			}
			continue
		}
		node, err := NewRocketpoolNode(rp.API, a.Bytes(), rp.RocketpoolRewardTreeData, rp.NodeRPLCumulative, atlasDeployed)
		if err != nil {
			return err
		}
		rp.NodesByAddress[addrHex] = node
	}

	return nil
}

func (rp *RocketpoolExporter) getRocketpoolRewardTrees() (map[uint64]RewardsFile, error) {
	var allRewards map[uint64]RewardsFile = map[uint64]RewardsFile{}

	type Data struct {
		ID   uint64 `db:"id"`
		Data []byte `db:"data"`
	}

	logger.Infof("rocketpool refreshing all reward tree data...")

	var jsonData []Data
	err := rp.DB.Select(&jsonData, `SELECT id, data FROM rocketpool_reward_tree`)
	if err != nil {
		return allRewards, fmt.Errorf("can not load claimedInterval tree from database, is it exported? %v", err)
	}

	for _, data := range jsonData {
		allRewards[data.ID], err = getRewardsData(data.Data)
		if err != nil {
			return allRewards, fmt.Errorf("can parsing reward tree data to struct for interval %v. Error %w", data.ID, err)
		}
	}
	return allRewards, nil
}

func (rp *RocketpoolExporter) UpdateDAOProposals() error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("updated rocketpool-dao-proposals")
	}(t0)

	pc, err := rpDAO.GetProposalCount(rp.API, nil)
	if err != nil {
		return err
	}
	for i := uint64(0); i < pc; i++ {
		p, err := NewRocketpoolDAOProposal(rp.API, i+1)
		if err != nil {
			return err
		}
		rp.DAOProposalsByID[i] = p
	}
	return nil
}

func (rp *RocketpoolExporter) UpdateDAOMembers() error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("updated rocketpool-dao-members")
	}(t0)

	members, err := rpDAOTrustedNode.GetMembers(rp.API, nil)
	if err != nil {
		return err
	}
	for _, m := range members {
		addrHex := m.Address.Hex()
		if member, exists := rp.DAOMembersByAddress[addrHex]; exists {
			err = member.Update(rp.API)
			if err != nil {
				return err
			}
			continue
		}

		m, err := NewRocketpoolDAOMember(rp.API, m.Address.Bytes())
		if err != nil {
			return err
		}
		rp.DAOMembersByAddress[addrHex] = m
	}
	return nil
}

func (rp *RocketpoolExporter) UpdateNetworkStats() error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("updated rocketpool-network-stats")
	}(t0)

	price, err := network.GetRPLPrice(rp.API, nil)
	if err != nil {
		return err
	}
	claimIntervalTime, err := rewards.GetClaimIntervalTime(rp.API, nil)
	if err != nil {
		return err
	}
	claimIntervalTimeStart, err := rewards.GetClaimIntervalTimeStart(rp.API, nil)
	if err != nil {
		return err
	}

	currentNodeFee, err := network.GetNodeFee(rp.API, nil)
	if err != nil {
		return err
	}

	currentNodeDemand, err := network.GetNodeDemand(rp.API, nil)
	if err != nil {
		return err
	}

	exchangeRate, err := tokens.GetRETHExchangeRate(rp.API, nil)
	if err != nil {
		return err
	}

	rethSupply, err := network.GetTotalRETHSupply(rp.API, nil)
	if err != nil {
		return err
	}

	isMergeUpdateDeployed, err := IsMergeUpdateDeployed(rp.API)
	if err != nil {
		return err
	}

	var nodeOperatorRewards *big.Int
	if !isMergeUpdateDeployed {
		nodeOperatorRewards, err = getBigIntFrom(rp.API, "rocketRewardsPool", "getClaimingContractAllowance", "rocketClaimNode")
		if err != nil {
			return err
		}
	} else {
		inflationInterval, err := tokens.GetRPLInflationIntervalRate(rp.API, nil)
		if err != nil {
			return err
		}

		totalRplSupply, err := tokens.GetRPLTotalSupply(rp.API, nil)
		if err != nil {
			return err
		}

		nodeOperatorRewardsPercentRaw, err := rewards.GetNodeOperatorRewardsPercent(rp.API, nil)
		if err != nil {
			return err
		}
		nodeOperatorRewardsPercent := eth.WeiToEth(nodeOperatorRewardsPercentRaw)

		rewardsIntervalDays := claimIntervalTime.Seconds() / (60 * 60 * 24)
		inflationPerDay := eth.WeiToEth(inflationInterval)
		totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * eth.WeiToEth(totalRplSupply)
		if totalRplAtNextCheckpoint < 0 {
			totalRplAtNextCheckpoint = 0
		}

		nodeOperatorRewards = eth.EthToWei(totalRplAtNextCheckpoint * nodeOperatorRewardsPercent)
	}

	totalEthStaking, err := network.GetStakingETHBalance(rp.API, nil)
	if err != nil {
		return err
	}

	totalEthBalance, err := network.GetTotalETHBalance(rp.API, nil)
	if err != nil {
		return err
	}

	rp.NetworkStats = RocketpoolNetworkStats{
		RPLPrice:               price,
		ClaimIntervalTime:      claimIntervalTime,
		ClaimIntervalTimeStart: claimIntervalTimeStart,
		CurrentNodeFee:         currentNodeFee,
		CurrentNodeDemand:      currentNodeDemand,
		RETHSupply:             rethSupply,
		NodeOperatorRewards:    nodeOperatorRewards,
		RETHPrice:              exchangeRate,
		TotalEthStaking:        totalEthStaking,
		TotalEthBalance:        totalEthBalance,
	}
	return err
}

// Redstone activation check
// Credit https://github.com/rocket-pool/smartnode/blob/4fd78852a331a7ec7a7e462fef2bcd49d1f0b0af/shared/utils/rp/update-checks.go
func IsMergeUpdateDeployed(rp *rocketpool.RocketPool) (bool, error) {
	currentVersion, err := rputil.GetCurrentVersion(rp, nil)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.1.0")
	return constraint.Check(currentVersion), nil
}

func IsAtlasDeployed(rp *rocketpool.RocketPool) (bool, error) {
	currentVersion, err := rputil.GetCurrentVersion(rp, nil)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.2.0")
	return constraint.Check(currentVersion), nil
}

func getBigIntFrom(rp *rocketpool.RocketPool, contract string, method string, args ...interface{}) (*big.Int, error) {
	rocketRewardsPool, err := rp.GetContract(contract, nil)
	if err != nil {
		return nil, err
	}
	perc := new(*big.Int)
	if err = rocketRewardsPool.Call(nil, perc, method, args...); err != nil {
		return nil, err
	}
	return *perc, err
}

func (rp *RocketpoolExporter) SaveMinipools() error {
	if len(rp.MinipoolsByAddress) == 0 {
		return nil
	}

	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Debugf("saved rocketpool-minipools")
	}(t0)

	data := make([]*RocketpoolMinipool, len(rp.MinipoolsByAddress))
	i := 0
	for _, mp := range rp.MinipoolsByAddress {
		data[i] = mp
		i++
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nArgs := 14
	valueStringsArr := make([]string, nArgs)
	for i := range valueStringsArr {
		valueStringsArr[i] = "$%d"
	}
	valueStringsTpl := "(" + strings.Join(valueStringsArr, ",") + ")"
	valueStringsArgs := make([]interface{}, nArgs)

	batchSize := 1000
	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*nArgs)
		for i, d := range data[start:end] {
			for j := 0; j < nArgs; j++ {
				valueStringsArgs[j] = i*nArgs + j + 1
			}
			valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
			valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
			valueArgs = append(valueArgs, d.Address)
			valueArgs = append(valueArgs, d.Pubkey)
			valueArgs = append(valueArgs, d.Status)
			valueArgs = append(valueArgs, d.StatusTime)
			valueArgs = append(valueArgs, d.NodeAddress)
			valueArgs = append(valueArgs, d.NodeFee)
			valueArgs = append(valueArgs, d.DepositType)
			valueArgs = append(valueArgs, d.PenaltyCount)
			valueArgs = append(valueArgs, d.NodeDepositBalance.String())
			valueArgs = append(valueArgs, d.NodeRefundBalance.String())
			valueArgs = append(valueArgs, d.UserDepositBalance.String())
			valueArgs = append(valueArgs, d.IsVacant)
			valueArgs = append(valueArgs, d.Version)
		}
		stmt := fmt.Sprintf(`
			insert into rocketpool_minipools (
				rocketpool_storage_address, address, pubkey, status, status_time, node_address, node_fee, 
				deposit_type, penalty_count, node_deposit_balance, node_refund_balance,
				user_deposit_balance, is_vacant, version
			) values %s on conflict (rocketpool_storage_address, address) do update set 
				pubkey = excluded.pubkey, 
				status = excluded.status, 
				status_time = excluded.status_time, 
				node_address = excluded.node_address, 
				node_fee = excluded.node_fee, 
				deposit_type = excluded.deposit_type, 
				penalty_count = excluded.penalty_count,
				node_deposit_balance = excluded.node_deposit_balance,
				node_refund_balance = excluded.node_refund_balance,
				user_deposit_balance = excluded.user_deposit_balance,
				is_vacant = excluded.is_vacant,
				version = excluded.version`,
			strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting into rocketpool_minipools: %w", err)
		}
	}

	return tx.Commit()
}

func (rp *RocketpoolExporter) SaveNodes() error {
	if len(rp.NodesByAddress) == 0 {
		return nil
	}

	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Debugf("saved rocketpool-nodes")
	}(t0)

	data := make([]*RocketpoolNode, len(rp.NodesByAddress))
	i := 0
	for _, node := range rp.NodesByAddress {
		data[i] = node
		i++
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nArgs := 13

	valueStringsArr := make([]string, nArgs)
	for i := range valueStringsArr {
		valueStringsArr[i] = "$%d"
	}
	valueStringsTpl := "(" + strings.Join(valueStringsArr, ",") + ")"
	valueStringsArgs := make([]interface{}, nArgs)

	var stmt string
	batchSize := 1000
	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*nArgs)
		for i, d := range data[start:end] {
			for j := 0; j < nArgs; j++ {
				valueStringsArgs[j] = i*nArgs + j + 1
			}
			valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
			valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
			valueArgs = append(valueArgs, d.Address)
			valueArgs = append(valueArgs, d.TimezoneLocation)
			valueArgs = append(valueArgs, d.RPLStake.String())
			valueArgs = append(valueArgs, d.MinRPLStake.String())
			valueArgs = append(valueArgs, d.MaxRPLStake.String())
			valueArgs = append(valueArgs, d.RPLCumulativeRewards.String())
			valueArgs = append(valueArgs, d.SmoothingPoolOptedIn)
			valueArgs = append(valueArgs, d.ClaimedSmoothingPool.String())
			valueArgs = append(valueArgs, d.UnclaimedSmoothingPool.String())
			valueArgs = append(valueArgs, d.UnclaimedRPLRewards.String())
			valueArgs = append(valueArgs, d.EffectiveRPLStake.String())
			valueArgs = append(valueArgs, d.DepositCredit.String())
		}

		stmt = fmt.Sprintf(`
			insert into rocketpool_nodes (
				rocketpool_storage_address, 
				address, 
				timezone_location, 
				rpl_stake, 
				min_rpl_stake, 
				max_rpl_stake, 
				rpl_cumulative_rewards, 
				smoothing_pool_opted_in, 
				claimed_smoothing_pool, 
				unclaimed_smoothing_pool, 
				unclaimed_rpl_rewards,
				effective_rpl_stake,
				deposit_credit
			) 
			values %s 
			on conflict (rocketpool_storage_address, address) do update set 
				rpl_stake = excluded.rpl_stake, 
				min_rpl_stake = excluded.min_rpl_stake, 
				max_rpl_stake = excluded.max_rpl_stake, 
				rpl_cumulative_rewards = excluded.rpl_cumulative_rewards,
				smoothing_pool_opted_in = excluded.smoothing_pool_opted_in,
				claimed_smoothing_pool = excluded.claimed_smoothing_pool,
				unclaimed_smoothing_pool = excluded.unclaimed_smoothing_pool,
				unclaimed_rpl_rewards = excluded.unclaimed_rpl_rewards,
				effective_rpl_stake = excluded.effective_rpl_stake,
				timezone_location = excluded.timezone_location,
				deposit_credit = excluded.deposit_credit
		`, strings.Join(valueStrings, ","))

		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting into rocketpool_nodes: %w", err)
		}
	}

	return tx.Commit()
}

func (rp *RocketpoolExporter) SaveRewardTrees() error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("saved rocketpool reward trees")
	}(t0)

	if len(rp.RocketpoolRewardTreesDownloadQueue) == 0 {
		return nil
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	logger.Infof("saving %v rocketpool reward trees", len(rp.RocketpoolRewardTreesDownloadQueue))

	for _, rewardTree := range rp.RocketpoolRewardTreesDownloadQueue {
		_, err = tx.Exec(`INSERT INTO rocketpool_reward_tree (id, data) VALUES($1, $2) ON CONFLICT DO NOTHING`, rewardTree.ID, rewardTree.Data)
		if err != nil {
			return fmt.Errorf("can not store reward file %v. Error %w", rewardTree.ID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	rp.RocketpoolRewardTreeData, err = rp.getRocketpoolRewardTrees()
	if err != nil {
		return err
	}
	// Delete download queue after refreshing the trees from db in case
	// refreshing throws an error so we try again in the next iteration
	// and always have an up to date tree
	rp.RocketpoolRewardTreesDownloadQueue = []RocketpoolRewardTreeDownloadable{}

	return nil
}

func (rp *RocketpoolExporter) SaveDAOProposals() error {
	if len(rp.DAOProposalsByID) == 0 {
		return nil
	}

	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Debugf("saved rocketpool-dao-proposals")
	}(t0)

	data := make([]*RocketpoolDAOProposal, len(rp.DAOProposalsByID))
	i := 0
	for _, val := range rp.DAOProposalsByID {
		data[i] = val
		i++
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nArgs := 18
	valueStringsArr := make([]string, nArgs)
	for i := range valueStringsArr {
		valueStringsArr[i] = "$%d"
	}
	valueStringsTpl := "(" + strings.Join(valueStringsArr, ",") + ")"
	valueStringsArgs := make([]interface{}, nArgs)

	batchSize := 1000
	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*nArgs)
		for i, d := range data[start:end] {
			for j := 0; j < nArgs; j++ {
				valueStringsArgs[j] = i*nArgs + j + 1
			}
			valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
			valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
			valueArgs = append(valueArgs, d.ID)
			valueArgs = append(valueArgs, d.DAO)
			valueArgs = append(valueArgs, d.ProposerAddress)
			valueArgs = append(valueArgs, d.Message)
			valueArgs = append(valueArgs, d.CreatedTime)
			valueArgs = append(valueArgs, d.StartTime)
			valueArgs = append(valueArgs, d.EndTime)
			valueArgs = append(valueArgs, d.ExpiryTime)
			valueArgs = append(valueArgs, d.VotesRequired)
			valueArgs = append(valueArgs, d.VotesFor)
			valueArgs = append(valueArgs, d.VotesAgainst)
			valueArgs = append(valueArgs, d.MemberVoted)
			valueArgs = append(valueArgs, d.MemberSupported)
			valueArgs = append(valueArgs, d.IsCancelled)
			valueArgs = append(valueArgs, d.IsExecuted)
			valueArgs = append(valueArgs, d.Payload)
			valueArgs = append(valueArgs, d.State)
		}
		stmt := fmt.Sprintf(`insert into rocketpool_dao_proposals (rocketpool_storage_address, id, dao, proposer_address, message, created_time, start_time, end_time, expiry_time, votes_required, votes_for, votes_against, member_voted, member_supported, is_cancelled, is_executed, payload, state) values %s on conflict (rocketpool_storage_address, id) do update set dao = excluded.dao, proposer_address = excluded.proposer_address, message = excluded.message, created_time = excluded.created_time, start_time = excluded.start_time, end_time = excluded.end_time, expiry_time = excluded.expiry_time, votes_required = excluded.votes_required, votes_for = excluded.votes_for, votes_against = excluded.votes_against, member_voted = excluded.member_voted, member_supported = excluded.member_supported, is_cancelled = excluded.is_cancelled, is_executed = excluded.is_executed, payload = excluded.payload, state = excluded.state`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting into rocketpool_dao_proposals: %w", err)
		}
	}

	return tx.Commit()
}

func (rp *RocketpoolExporter) SaveDAOProposalsMemberVotes() error {
	if len(rp.DAOProposalsByID) == 0 {
		return nil
	}

	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Debugf("saved rocketpool-dao-proposals-member-votes")
	}(t0)

	data := []RocketpoolDAOProposalMemberVotes{}
	for _, val := range rp.DAOProposalsByID {
		data = append(data, val.MemberVotes...)
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nArgs := 5
	valueStringsArr := make([]string, nArgs)
	for i := range valueStringsArr {
		valueStringsArr[i] = "$%d"
	}
	valueStringsTpl := "(" + strings.Join(valueStringsArr, ",") + ")"
	valueStringsArgs := make([]interface{}, nArgs)

	batchSize := 1000
	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*nArgs)
		for i, d := range data[start:end] {
			for j := 0; j < nArgs; j++ {
				valueStringsArgs[j] = i*nArgs + j + 1
			}
			valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
			valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
			valueArgs = append(valueArgs, d.ProposalID)
			valueArgs = append(valueArgs, d.Address)
			valueArgs = append(valueArgs, d.Voted)
			valueArgs = append(valueArgs, d.Supported)
		}

		stmt := fmt.Sprintf(`
			insert into rocketpool_dao_proposals_member_votes (rocketpool_storage_address, id, member_address, voted, supported) 
			values %s 
			on conflict (rocketpool_storage_address, id, member_address) do update 
				set voted = excluded.voted, 
				supported = excluded.supported`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting into rocketpool_dao_proposals_member_votes: %w", err)
		}
	}

	return tx.Commit()
}

func (rp *RocketpoolExporter) SaveDAOMembers() error {
	if len(rp.DAOMembersByAddress) == 0 {
		return nil
	}

	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Debugf("saved rocketpool-dao-members")
	}(t0)

	data := make([]*RocketpoolDAOMember, len(rp.DAOMembersByAddress))
	i := 0
	for _, val := range rp.DAOMembersByAddress {
		data[i] = val
		i++
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nArgs := 8
	valueStringsArr := make([]string, nArgs)
	for i := range valueStringsArr {
		valueStringsArr[i] = "$%d"
	}
	valueStringsTpl := "(" + strings.Join(valueStringsArr, ",") + ")"
	valueStringsArgs := make([]interface{}, nArgs)

	batchSize := 1000
	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*nArgs)
		addresses := make([][]byte, 0, batchSize)
		for i, d := range data[start:end] {
			for j := 0; j < nArgs; j++ {
				valueStringsArgs[j] = i*nArgs + j + 1
			}
			valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
			valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
			valueArgs = append(valueArgs, d.Address)
			valueArgs = append(valueArgs, d.ID)
			valueArgs = append(valueArgs, d.URL)
			valueArgs = append(valueArgs, d.JoinedTime)
			valueArgs = append(valueArgs, d.LastProposalTime)
			valueArgs = append(valueArgs, d.RPLBondAmount.String())
			valueArgs = append(valueArgs, d.UnbondedValidatorCount)
			addresses = append(addresses, d.Address)
		}
		stmt := fmt.Sprintf(`
			INSERT INTO rocketpool_dao_members (
				rocketpool_storage_address,
				address,
				id,
				url,
				joined_time,
				last_proposal_time,
				rpl_bond_amount,
				unbonded_validator_count
			)
			values %s
			on conflict (rocketpool_storage_address, address) do update set
				id = excluded.id,
				url = excluded.url,
				joined_time = excluded.joined_time,
				last_proposal_time = excluded.last_proposal_time,
				rpl_bond_amount = excluded.rpl_bond_amount,
				unbonded_validator_count = excluded.unbonded_validator_count
			`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting into rocketpool_dao_members: %w", err)
		}

		_, err = tx.Exec(`
			DELETE FROM rocketpool_dao_members
			WHERE NOT address = ANY($1)`, pq.ByteaArray(addresses))
		if err != nil {
			return fmt.Errorf("error deleting from rocketpool_dao_members: %w", err)
		}
	}

	return tx.Commit()
}

func (rp *RocketpoolExporter) TagValidators() error {
	if len(rp.MinipoolsByAddress) == 0 {
		return nil
	}

	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Debugf("saved rocketpool-validator-tags")
	}(t0)

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	data := make([]*RocketpoolMinipool, len(rp.MinipoolsByAddress))
	i := 0
	for _, mp := range rp.MinipoolsByAddress {
		data[i] = mp
		i++
	}

	batchSize := 5000
	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}
		n := 1
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*n)
		for i, d := range data[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, 'rocketpool')", i*n+1))
			valueArgs = append(valueArgs, d.Pubkey)
		}
		_, err := tx.Exec(fmt.Sprintf(`insert into validator_tags (publickey, tag) values %s on conflict (publickey, tag) do nothing`, strings.Join(valueStrings, ",")), valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting into validator_tags: %w", err)
		}
		_, err = tx.Exec(fmt.Sprintf(`insert into validator_pool (publickey, pool) values %s on conflict (publickey) do nothing`, strings.Join(valueStrings, ",")), valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting into validator_pool: %w", err)
		}
	}

	return tx.Commit()
}

func (rp *RocketpoolExporter) SaveNetworkStats() error {
	_, err := db.WriterDb.Exec(`
		INSERT INTO rocketpool_network_stats 
		(
			ts, rpl_price, claim_interval_time, claim_interval_time_start, current_node_fee, current_node_demand, 
			reth_supply, node_operator_rewards, reth_exchange_rate, node_count, minipool_count, odao_member_count, 
			total_eth_staking, total_eth_balance, effective_rpl_staked
		) 
		VALUES(
			now(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			(SELECT sum(effective_rpl_stake) FROM rocketpool_nodes)
		)`,
		rp.NetworkStats.RPLPrice.String(),
		rp.NetworkStats.ClaimIntervalTime.String(),
		rp.NetworkStats.ClaimIntervalTimeStart,
		rp.NetworkStats.CurrentNodeFee,
		rp.NetworkStats.CurrentNodeDemand.String(),
		rp.NetworkStats.RETHSupply.String(),
		rp.NetworkStats.NodeOperatorRewards.String(),
		rp.NetworkStats.RETHPrice,
		len(rp.NodesByAddress),
		len(rp.MinipoolsByAddress),
		len(rp.DAOMembersByAddress),
		rp.NetworkStats.TotalEthStaking.String(),
		rp.NetworkStats.TotalEthBalance.String(),
	)
	return err
}

type RocketpoolMinipool struct {
	Address            []byte    `db:"address"`
	Pubkey             []byte    `db:"pubkey"`
	NodeAddress        []byte    `db:"node_address"`
	NodeFee            float64   `db:"node_fee"`
	DepositType        string    `db:"deposit_type"`
	Status             string    `db:"status"`
	StatusTime         time.Time `db:"status_time"`
	PenaltyCount       uint64    `db:"penalty_count"`
	NodeDepositBalance *big.Int  `db:"node_deposit_balance"`
	NodeRefundBalance  *big.Int  `db:"node_refund_balance"`
	UserDepositBalance *big.Int  `db:"user_deposit_balance"`
	IsVacant           bool      `db:"is_vacant"`
	Version            uint8     `db:"version"`
}

func NewRocketpoolMinipool(rp *rocketpool.RocketPool, addr []byte, atlasDeployed bool) (*RocketpoolMinipool, error) {
	pubk, err := minipool.GetMinipoolPubkey(rp, common.BytesToAddress(addr), nil)
	if err != nil {
		return nil, err
	}
	mp, err := minipool.NewMinipool(rp, common.BytesToAddress(addr), nil)
	if err != nil {
		return nil, err
	}
	nodeAddr, err := mp.GetNodeAddress(nil)
	if err != nil {
		return nil, err
	}

	rpm := &RocketpoolMinipool{
		Address:     addr,
		Pubkey:      pubk.Bytes(),
		NodeAddress: nodeAddr.Bytes(),
	}
	err = rpm.Update(rp, atlasDeployed)
	if err != nil {
		return nil, err
	}
	return rpm, nil
}

func (r *RocketpoolMinipool) Update(rp *rocketpool.RocketPool, atlasDeployed bool) error {
	mp, err := minipool.NewMinipool(rp, common.BytesToAddress(r.Address), nil)
	if err != nil {
		return err
	}

	var wg errgroup.Group
	var status rpTypes.MinipoolStatus
	var statusTime time.Time
	var penaltyCount uint64
	var nodeFee float64

	var nodeDepositBalance, nodeRefundBalance, userDepositBalance *big.Int = leb16, big.NewInt(0), leb16
	var version uint8
	var statusDetail minipool.StatusDetails = minipool.StatusDetails{
		IsVacant: false,
	}
	var depositType rpTypes.MinipoolDeposit

	// Node fee can change on conversion starting with Atlas
	wg.Go(func() error {
		var err error
		nodeFee, err = mp.GetNodeFee(nil)
		return err
	})

	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		statusTime, err = mp.GetStatusTime(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		penaltyCount, err = minipool.GetMinipoolPenaltyCount(rp, common.BytesToAddress(r.Address), nil)
		return err
	})

	if atlasDeployed {
		wg.Go(func() error {
			var err error
			nodeDepositBalance, err = mp.GetNodeDepositBalance(nil)
			return err
		})

		wg.Go(func() error {
			var err error
			userDepositBalance, err = mp.GetUserDepositBalance(nil)
			return err
		})

		wg.Go(func() error {
			var err error
			nodeRefundBalance, err = mp.GetNodeRefundBalance(nil)
			return err
		})

		wg.Go(func() error {
			var err error
			statusDetail, err = mp.GetStatusDetails(nil)
			return err
		})
	}

	wg.Go(func() error {
		var err error
		version = mp.GetVersion()
		return err
	})

	wg.Go(func() error {
		var err error
		depositType, err = mp.GetDepositType(nil)
		return err
	})

	if err := wg.Wait(); err != nil {
		return err
	}

	r.NodeFee = nodeFee
	r.Status = status.String()
	r.StatusTime = statusTime
	r.PenaltyCount = penaltyCount
	r.Version = version
	r.NodeDepositBalance = nodeDepositBalance
	r.NodeRefundBalance = nodeRefundBalance
	r.UserDepositBalance = userDepositBalance
	r.IsVacant = statusDetail.IsVacant
	r.DepositType = depositType.String()
	return nil
}

type RocketpoolNode struct {
	Address                []byte   `db:"address"`
	TimezoneLocation       string   `db:"timezone_location"`
	RPLStake               *big.Int `db:"rpl_stake"`
	EffectiveRPLStake      *big.Int `db:"effective_rpl_stake"`
	MinRPLStake            *big.Int `db:"min_rpl_stake"`
	MaxRPLStake            *big.Int `db:"max_rpl_stake"`
	RPLCumulativeRewards   *big.Int `db:"rpl_cumulative_rewards"`
	SmoothingPoolOptedIn   bool     `db:"smoothing_pool_opted_in"`
	ClaimedSmoothingPool   *big.Int `db:"claimed_smoothing_pool"`
	UnclaimedSmoothingPool *big.Int `db:"unclaimed_smoothing_pool"`
	UnclaimedRPLRewards    *big.Int `db:"unclaimed_rpl_rewards"`
	DepositCredit          *big.Int `db:"deposit_credit"`
}

func NewRocketpoolNode(rp *rocketpool.RocketPool, addr []byte, rewardTrees map[uint64]RewardsFile, legacyClaims map[string]*big.Int, atlasDeployed bool) (*RocketpoolNode, error) {
	rpn := &RocketpoolNode{
		Address: addr,
	}

	err := rpn.Update(rp, rewardTrees, true, legacyClaims, atlasDeployed)
	if err != nil {
		return nil, err
	}
	return rpn, nil
}

type RocketpoolRewards struct {
	RplColl          *big.Int
	SmoothingPoolEth *big.Int
	OdaoRpl          *big.Int
}

func (r *RocketpoolNode) Update(rp *rocketpool.RocketPool, rewardTrees map[uint64]RewardsFile, includeCumulativeRpl bool, legacyClaims map[string]*big.Int, atlasDeployed bool) error {
	address := common.BytesToAddress(r.Address)

	var wg errgroup.Group
	var err error
	var tl string
	var stake, minStake, maxStake, effectiveStake, depositCredit *big.Int = big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)

	wg.Go(func() error {
		var err error
		tl, err = node.GetNodeTimezoneLocation(rp, address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		stake, err = node.GetNodeRPLStake(rp, address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		minStake, err = node.GetNodeMinimumRPLStake(rp, address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		maxStake, err = node.GetNodeMaximumRPLStake(rp, address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		effectiveStake, err = node.GetNodeEffectiveRPLStake(rp, address, nil)
		return err
	})

	if atlasDeployed {
		wg.Go(func() error {
			var err error
			depositCredit, err = node.GetNodeDepositCredit(rp, address, nil)
			return err
		})
	}

	if err = wg.Wait(); err != nil {
		return err
	}

	if len(rewardTrees) > 0 {
		r.SmoothingPoolOptedIn, err = node.GetSmoothingPoolRegistrationState(rp, address, nil)
		if err != nil {
			return err
		}

		if includeCumulativeRpl {

			var claimedSum RocketpoolRewards = RocketpoolRewards{
				SmoothingPoolEth: big.NewInt(0),
				OdaoRpl:          big.NewInt(0),
				RplColl:          big.NewInt(0),
			}
			var unclaimedSum RocketpoolRewards = RocketpoolRewards{
				SmoothingPoolEth: big.NewInt(0),
				OdaoRpl:          big.NewInt(0),
				RplColl:          big.NewInt(0),
			}

			unclaimed, claimed, err := smartnodeRewards.GetClaimStatus(rp, address)
			if err != nil {
				return err
			}

			// Get the info for each claimed interval
			for _, claimedInterval := range claimed {
				rewardData := rewardTrees[claimedInterval]

				rewards, exists := rewardData.NodeRewards[address]

				if exists {
					claimedSum.RplColl = claimedSum.RplColl.Add(claimedSum.RplColl, &rewards.CollateralRpl.Int)
					claimedSum.SmoothingPoolEth = claimedSum.SmoothingPoolEth.Add(claimedSum.SmoothingPoolEth, &rewards.SmoothingPoolEth.Int)
					claimedSum.OdaoRpl = claimedSum.OdaoRpl.Add(claimedSum.OdaoRpl, &rewards.OracleDaoRpl.Int)
				}
			}

			// Get the unclaimed rewards
			for _, unclaimedInterval := range unclaimed {
				rewardData := rewardTrees[unclaimedInterval]

				rewards, exists := rewardData.NodeRewards[address]

				if exists {
					unclaimedSum.RplColl = unclaimedSum.RplColl.Add(unclaimedSum.RplColl, &rewards.CollateralRpl.Int)
					unclaimedSum.SmoothingPoolEth = unclaimedSum.SmoothingPoolEth.Add(unclaimedSum.SmoothingPoolEth, &rewards.SmoothingPoolEth.Int)
					unclaimedSum.OdaoRpl = unclaimedSum.OdaoRpl.Add(unclaimedSum.OdaoRpl, &rewards.OracleDaoRpl.Int)
				}
			}

			r.RPLCumulativeRewards = claimedSum.RplColl
			if legacyAmount, exists := legacyClaims[address.Hex()]; exists {
				r.RPLCumulativeRewards = r.RPLCumulativeRewards.Add(r.RPLCumulativeRewards, legacyAmount)
			}
			r.ClaimedSmoothingPool = claimedSum.SmoothingPoolEth
			r.UnclaimedSmoothingPool = unclaimedSum.SmoothingPoolEth
			r.UnclaimedRPLRewards = unclaimedSum.RplColl
		}
	}

	if r.RPLCumulativeRewards == nil {
		r.RPLCumulativeRewards = big.NewInt(0)
	}
	if r.UnclaimedRPLRewards == nil {
		r.UnclaimedRPLRewards = big.NewInt(0)
	}
	if r.UnclaimedSmoothingPool == nil {
		r.UnclaimedSmoothingPool = big.NewInt(0)
	}
	if r.ClaimedSmoothingPool == nil {
		r.ClaimedSmoothingPool = big.NewInt(0)
	}

	r.TimezoneLocation = tl
	r.RPLStake = stake
	r.MinRPLStake = minStake
	r.MaxRPLStake = maxStake
	r.EffectiveRPLStake = effectiveStake
	r.DepositCredit = depositCredit

	return nil
}

func getRewardsData(jsonData []byte) (RewardsFile, error) {
	var proofWrapper RewardsFile

	err := json.Unmarshal(jsonData, &proofWrapper)
	if err != nil {
		err = fmt.Errorf("error deserializing : %w", err)
		return proofWrapper, err
	}

	return proofWrapper, err
}

func CalculateLifetimeNodeRewardsAllLegacy(rp *rocketpool.RocketPool, intervalSize *big.Int, legacyRocketRewardsPoolAddress *common.Address, legacyRocketClaimNodeAddress *common.Address) (map[string]*big.Int, error) {

	// Get contracts
	rocketRewardsPool, err := getRocketRewardsPoolLegacy(rp, legacyRocketRewardsPoolAddress)
	if err != nil {
		return nil, err
	}
	rocketClaimNode, err := getRocketClaimNodeLegacy(rp, legacyRocketClaimNodeAddress)
	if err != nil {
		return nil, err
	}
	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketRewardsPool.Address}
	// RPLTokensClaimed(address clamingContract, address claimingAddress, uint256 amount, uint256 time)
	topicFilter := [][]common.Hash{{rocketRewardsPool.ABI.Events["RPLTokensClaimed"].ID}, {rocketClaimNode.Address.Hash()}}

	sumMap := make(map[string]*big.Int)
	prerecordedIntervals, exists := firstBlockOfRedstone[utils.Config.Chain.Name]
	var maxBlockNumber *big.Int = nil
	if prerecordedIntervals == 0 || !exists {
		return sumMap, nil
	}
	// only look for legacy lifetime rewards before the new rewards system went live
	maxBlockNumber = big.NewInt(0).SetUint64(prerecordedIntervals)

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, nil, maxBlockNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("can not load lifetime rewards: %w", err)
	}

	// Iterate over the logs and sum the amount
	for _, log := range logs {
		values := make(map[string]interface{})
		// Decode the event
		if rocketRewardsPool.ABI.Events["RPLTokensClaimed"].Inputs.UnpackIntoMap(values, log.Data) != nil {
			return nil, err
		}
		// Add the amount argument to our sum
		amount := values["amount"].(*big.Int)
		claimAddress := common.BytesToAddress(log.Topics[2].Bytes())
		sum, ok := sumMap[claimAddress.Hex()]
		if !ok {
			sum = big.NewInt(0)
		}
		sumMap[claimAddress.Hex()] = sum.Add(sum, amount)

	}
	// Return the result
	return sumMap, nil
}

// Get contracts
var rocketRewardsPoolLock sync.Mutex

func getRocketRewardsPoolLegacy(rp *rocketpool.RocketPool, address *common.Address) (*rocketpool.Contract, error) {
	rocketRewardsPoolLock.Lock()
	defer rocketRewardsPoolLock.Unlock()
	if address == nil {
		return rp.VersionManager.V1_0_0.GetContract("rocketRewardsPool", nil)
	} else {
		return rp.VersionManager.V1_0_0.GetContractWithAddress("rocketRewardsPool", *address)
	}
}

// Get contracts
var rocketClaimNodeLock sync.Mutex

func getRocketClaimNodeLegacy(rp *rocketpool.RocketPool, address *common.Address) (*rocketpool.Contract, error) {
	rocketClaimNodeLock.Lock()
	defer rocketClaimNodeLock.Unlock()
	if address == nil {
		return rp.VersionManager.V1_0_0.GetContract("rocketClaimNode", nil)
	} else {
		return rp.VersionManager.V1_0_0.GetContractWithAddress("rocketClaimNode", *address)
	}
}

type RocketpoolDAOProposalMemberVotes struct {
	ProposalID uint64 `db:"id"`
	Address    []byte `db:"member_address"`
	Voted      bool   `db:"voted"`
	Supported  bool   `db:"supported"`
}

type RocketpoolDAOProposal struct {
	ID              uint64    `db:"id"`
	DAO             string    `db:"dao"`
	ProposerAddress []byte    `db:"proposer_address"`
	Message         string    `db:"message"`
	CreatedTime     time.Time `db:"created_time"`
	StartTime       time.Time `db:"start_time"`
	EndTime         time.Time `db:"end_time"`
	ExpiryTime      time.Time `db:"expiry_time"`
	VotesRequired   float64   `db:"votes_required"`
	VotesFor        float64   `db:"votes_for"`
	VotesAgainst    float64   `db:"votes_against"`
	MemberVoted     bool      `db:"member_voted"`
	MemberSupported bool      `db:"member_supported"`
	IsCancelled     bool      `db:"is_cancelled"`
	IsExecuted      bool      `db:"is_executed"`
	Payload         []byte    `db:"payload"`
	State           string    `db:"state"`
	MemberVotes     []RocketpoolDAOProposalMemberVotes
}

func NewRocketpoolDAOProposal(rp *rocketpool.RocketPool, pid uint64) (*RocketpoolDAOProposal, error) {
	p := &RocketpoolDAOProposal{ID: pid}
	err := p.Update(rp)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *RocketpoolDAOProposal) Update(rp *rocketpool.RocketPool) error {
	pd, err := rpDAO.GetProposalDetails(rp, r.ID, nil)
	if err != nil {
		return err
	}
	r.ID = pd.ID
	r.DAO = pd.DAO
	r.ProposerAddress = pd.ProposerAddress.Bytes()
	r.Message = pd.Message
	r.CreatedTime = time.Unix(int64(pd.CreatedTime), 0)
	r.StartTime = time.Unix(int64(pd.StartTime), 0)
	r.EndTime = time.Unix(int64(pd.EndTime), 0)
	r.ExpiryTime = time.Unix(int64(pd.ExpiryTime), 0)
	r.VotesRequired = pd.VotesRequired
	r.VotesFor = pd.VotesFor
	r.VotesAgainst = pd.VotesAgainst
	r.MemberVoted = pd.MemberVoted
	r.MemberSupported = pd.MemberSupported
	r.IsCancelled = pd.IsCancelled
	r.IsExecuted = pd.IsExecuted
	r.Payload = pd.Payload
	r.State = pd.State.String()

	// Update member votes
	r.MemberVotes = []RocketpoolDAOProposalMemberVotes{}
	members, err := rpDAOTrustedNode.GetMembers(rp, nil)
	if err != nil {
		return err
	}
	for _, m := range members {
		memberVoted, err := rpDAO.GetProposalMemberVoted(rp, r.ID, m.Address, nil)
		if err != nil {
			return err
		}

		memberSupported, err := rpDAO.GetProposalMemberSupported(rp, r.ID, m.Address, nil)
		if err != nil {
			return err
		}

		r.MemberVotes = append(r.MemberVotes, RocketpoolDAOProposalMemberVotes{
			ProposalID: r.ID,
			Address:    m.Address.Bytes(),
			Voted:      memberVoted,
			Supported:  memberSupported,
		})
	}

	return nil
}

type RocketpoolDAOMember struct {
	Address                []byte    `db:"address"`
	ID                     string    `db:"id"`
	URL                    string    `url:"url"`
	JoinedTime             time.Time `db:"joined_time"`
	LastProposalTime       time.Time `db:"last_proposal_time"`
	RPLBondAmount          *big.Int  `db:"rpl_bond_amount"`
	UnbondedValidatorCount uint64    `db:"unbonded_validator_count"`
}

func NewRocketpoolDAOMember(rp *rocketpool.RocketPool, addr []byte) (*RocketpoolDAOMember, error) {
	m := &RocketpoolDAOMember{}
	m.Address = addr
	err := m.Update(rp)
	if err != nil {
		return m, err
	}
	return m, nil
}

func (r *RocketpoolDAOMember) Update(rp *rocketpool.RocketPool) error {
	d, err := rpDAOTrustedNode.GetMemberDetails(rp, common.BytesToAddress(r.Address), nil)
	if err != nil {
		return err
	}
	r.ID = d.ID
	r.URL = d.Url
	r.JoinedTime = time.Unix(int64(d.JoinedTime), 0)
	r.LastProposalTime = time.Unix(int64(d.LastProposalTime), 0)
	r.RPLBondAmount = d.RPLBondAmount
	r.UnbondedValidatorCount = d.UnbondedValidatorCount
	return nil
}

type MinipoolPerformanceFile struct {
	Index               uint64                                               `json:"index"`
	Network             string                                               `json:"network"`
	MinipoolPerformance map[common.Address]*SmoothingPoolMinipoolPerformance `json:"minipoolPerformance"`
}

// Minipool stats
type SmoothingPoolMinipoolPerformance struct {
	Pubkey                  string   `json:"pubkey"`
	SuccessfulAttestations  uint64   `json:"successfulAttestations"`
	MissedAttestations      uint64   `json:"missedAttestations"`
	ParticipationRate       float64  `json:"participationRate"`
	MissingAttestationSlots []uint64 `json:"missingAttestationSlots"`
	EthEarned               float64  `json:"ethEarned"`
}

// Node operator rewards
type NodeRewardsInfo struct {
	RewardNetwork                uint64        `json:"rewardNetwork"`
	CollateralRpl                *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl                 *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth             *QuotedBigInt `json:"smoothingPoolEth"`
	SmoothingPoolEligibilityRate float64       `json:"smoothingPoolEligibilityRate"`
	MerkleData                   []byte        `json:"-"`
	MerkleProof                  []string      `json:"merkleProof"`
}

// Rewards per network
type NetworkRewardsInfo struct {
	CollateralRpl    *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl     *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth *QuotedBigInt `json:"smoothingPoolEth"`
}

// Total cumulative rewards for an interval
type TotalRewards struct {
	ProtocolDaoRpl               *QuotedBigInt `json:"protocolDaoRpl"`
	TotalCollateralRpl           *QuotedBigInt `json:"totalCollateralRpl"`
	TotalOracleDaoRpl            *QuotedBigInt `json:"totalOracleDaoRpl"`
	TotalSmoothingPoolEth        *QuotedBigInt `json:"totalSmoothingPoolEth"`
	PoolStakerSmoothingPoolEth   *QuotedBigInt `json:"poolStakerSmoothingPoolEth"`
	NodeOperatorSmoothingPoolEth *QuotedBigInt `json:"nodeOperatorSmoothingPoolEth"`
}

// JSON struct for a complete rewards file
type RewardsFile struct {
	// Serialized fields
	RewardsFileVersion         uint64                              `json:"rewardsFileVersion"`
	Index                      uint64                              `json:"index"`
	Network                    string                              `json:"network"`
	StartTime                  time.Time                           `json:"startTime,omitempty"`
	EndTime                    time.Time                           `json:"endTime"`
	ConsensusStartBlock        uint64                              `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock          uint64                              `json:"consensusEndBlock"`
	ExecutionStartBlock        uint64                              `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock          uint64                              `json:"executionEndBlock"`
	IntervalsPassed            uint64                              `json:"intervalsPassed"`
	MerkleRoot                 string                              `json:"merkleRoot,omitempty"`
	MinipoolPerformanceFileCID string                              `json:"minipoolPerformanceFileCid,omitempty"`
	TotalRewards               *TotalRewards                       `json:"totalRewards"`
	NetworkRewards             map[uint64]*NetworkRewardsInfo      `json:"networkRewards"`
	NodeRewards                map[common.Address]*NodeRewardsInfo `json:"nodeRewards"`
	MinipoolPerformanceFile    MinipoolPerformanceFile             `json:"-"`
}

type QuotedBigInt struct {
	big.Int
}

func NewQuotedBigInt(x int64) *QuotedBigInt {
	q := QuotedBigInt{}
	native := big.NewInt(x)
	q.Int = *native
	return &q
}

func NewQuotedBigIntFromBigInt(x *big.Int) *QuotedBigInt {
	q := QuotedBigInt{}
	q.Int = *x
	return &q
}

func (b *QuotedBigInt) MarshalJSON() ([]byte, error) {
	return []byte("\"" + b.String() + "\""), nil
}

func (b *QuotedBigInt) UnmarshalJSON(p []byte) error {
	strippedString := strings.Trim(string(p), "\"")
	nativeInt, success := big.NewInt(0).SetString(strippedString, 0)
	if !success {
		return fmt.Errorf("%s is not a valid big integer", strippedString)
	}

	b.Int = *nativeInt
	return nil
}

func DownloadRewardsFile(fileName string, interval uint64, cid string, isDaemon bool) ([]byte, error) {

	ipfsFilename := fileName + ".zst"

	// Create URL list
	urls := []string{
		fmt.Sprintf("https://%s.ipfs.dweb.link/%s", cid, ipfsFilename),
		fmt.Sprintf("https://ipfs.io/ipfs/%s/%s", cid, ipfsFilename),
	}

	// Attempt downloads
	errBuilder := strings.Builder{}
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			errBuilder.WriteString(fmt.Sprintf("Downloading %s failed (%s)\n", url, err.Error()))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errBuilder.WriteString(fmt.Sprintf("Downloading %s failed with status %s\n", url, resp.Status))
			continue
		} else {
			// If we got here, we have a successful download
			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				errBuilder.WriteString(fmt.Sprintf("Error reading response bytes from %s: %s\n", url, err.Error()))
				continue
			}

			// Decompress it
			decompressedBytes, err := decompressFile(bytes)
			if err != nil {
				errBuilder.WriteString(fmt.Sprintf("Error decompressing %s: %s\n", url, err.Error()))
				continue
			}

			return decompressedBytes, nil
		}
	}

	return nil, fmt.Errorf(errBuilder.String())

}

// Decompresses a rewards file
func decompressFile(compressedBytes []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating compression decoder: %w", err)
	}

	decompressedBytes, err := decoder.DecodeAll(compressedBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("error decompressing rewards file: %w", err)
	}

	return decompressedBytes, nil
}
