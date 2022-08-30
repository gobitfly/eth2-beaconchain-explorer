package exporter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"eth2-exporter/db"
	"eth2-exporter/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	gethRPC "github.com/ethereum/go-ethereum/rpc"
	"github.com/hashicorp/go-version"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/klauspost/compress/zstd"
	"github.com/rocket-pool/rocketpool-go/dao"
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
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var rpEth1RPRCClient *gethRPC.Client
var rpEth1Client *ethclient.Client

const GethEventLogInterval = 25000

// Previous redstone reward pool addresses
// https://github.com/rocket-pool/smartnode/blob/master/shared/services/config/smartnode-config.go
var previousRewardsPoolAddress = map[string]string{
	"mainnet": "",
	"prater":  "0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1",
	"kiln":    "",
	"ropsten": "",
}

var legacyRewardsPoolAddress = map[string]string{
	"mainnet": "0xA3a18348e6E2d3897B6f2671bb8c120e36554802",
	"prater":  "0xf9aE18eB0CE4930Bc3d7d1A5E33e4286d4FB0f8B",
	"kiln":    "0xFb62F3B5AF8099Bbd19d5d46084Bb152ECDE25A6",
	"ropsten": "0x401e46fA6cBC9e1E6Cc3E9666C10329f938aE1B3",
}

var legacyClaimNodeAddress = map[string]string{
	"mainnet": "0x899336A2a86053705E65dB61f52C686dcFaeF548",
	"prater":  "0xc05b7A2a03A6d2736d1D0ebf4d4a0aFE2cc32cE1",
	"kiln":    "0xF98086202F8F58dad8120055Fdd6e2f36De2c6Fb",
	"ropsten": "0xA55F65219d7254DFde4021E4f534a7a55750C4a1",
}

func rocketpoolExporter() {
	var err error
	rpEth1RPRCClient, err = gethRPC.Dial(utils.Config.Indexer.Eth1Endpoint)
	if err != nil {
		logger.Fatal(err)
	}
	rpEth1Client = ethclient.NewClient(rpEth1RPRCClient)
	rpExporter, err := NewRocketpoolExporter(rpEth1Client, utils.Config.RocketpoolExporter.StorageContractAddress, db.WriterDb)
	if err != nil {
		logger.Fatal(err)
	}
	rpExporter.Run()
}

type RocketpoolNetworkStats struct {
	RPLPrice               *big.Int
	ClaimIntervalTime      time.Duration
	ClaimIntervalTimeStart time.Time
	CurrentNodeFee         float64
	CurrentNodeDemand      *big.Int
	RETHSupply             *big.Int
	EffectiveRPLStake      *big.Int
	NodeOperatorRewards    *big.Int
	RETHPrice              float64
	TotalEthStaking        *big.Int
	TotalEthBalance        *big.Int
}

type RocketpoolExporter struct {
	Eth1Client          *ethclient.Client
	API                 *rocketpool.RocketPool
	DB                  *sqlx.DB
	UpdateInterval      time.Duration
	MinipoolsByAddress  map[string]*RocketpoolMinipool
	NodesByAddress      map[string]*RocketpoolNode
	DAOProposalsByID    map[uint64]*RocketpoolDAOProposal
	DAOMembersByAddress map[string]*RocketpoolDAOMember
	NodeRPLCumulative   map[string]*big.Int
	NetworkStats        RocketpoolNetworkStats
	LastRewardTree      uint64
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
	rpe.UpdateInterval = time.Second * 60
	rpe.MinipoolsByAddress = map[string]*RocketpoolMinipool{}
	rpe.NodesByAddress = map[string]*RocketpoolNode{}
	rpe.DAOProposalsByID = map[uint64]*RocketpoolDAOProposal{}
	rpe.DAOMembersByAddress = map[string]*RocketpoolDAOMember{}
	rpe.LastRewardTree = 0
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
	err := rp.DB.Select(&dbRes, `select * from rocketpool_minipools`)
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
	err := rp.DB.Select(&dbRes, `select * from rocketpool_nodes`)
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
	err := rp.DB.Select(&dbRes, `select * from rocketpool_proposals`)
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
	err := rp.DB.Select(&dbRes, `select * from rocketpool_dao_members`)
	if err != nil {
		return err
	}
	for _, val := range dbRes {
		rp.DAOMembersByAddress[fmt.Sprintf("%x", val.Address)] = &val
	}
	return nil
}

func (rp *RocketpoolExporter) Run() error {
	errorInterval := time.Second * 60
	t := time.NewTicker(rp.UpdateInterval)
	defer t.Stop()
	var count int64 = 0
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

func GetRewardSnapshotEvent(rp *rocketpool.RocketPool, index uint64, intervalSize *big.Int, startBlock *big.Int, rewardPoolAddress *common.Address) (rewards.RewardsEvent, error) {
	// Get contracts
	rocketRewardsPool, err := getRocketRewardsPool(rp, rewardPoolAddress)
	if err != nil {
		return rewards.RewardsEvent{}, err
	}

	// Construct a filter query for relevant logs
	indexBig := big.NewInt(0).SetUint64(index)
	indexBytes := [32]byte{}
	indexBig.FillBytes(indexBytes[:])
	addressFilter := []common.Address{*rocketRewardsPool.Address}
	topicFilter := [][]common.Hash{{rocketRewardsPool.ABI.Events["RewardSnapshot"].ID}, {indexBytes}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, startBlock, nil, nil)
	if err != nil {
		return rewards.RewardsEvent{}, err
	}

	// Get the log info
	values := make(map[string]interface{})
	if len(logs) == 0 {
		return rewards.RewardsEvent{}, fmt.Errorf("reward snapshot for interval %d not found", index)
	}
	if rocketRewardsPool.ABI.Events["RewardSnapshot"].Inputs.UnpackIntoMap(values, logs[0].Data) != nil {
		return rewards.RewardsEvent{}, err
	}

	// Get the decoded data
	var submission RewardSubmission
	if rewardPoolAddress != nil {
		submissionPrototypeTemp := RewardSubmissionLegacy{}
		submissionTypeTemp := reflect.TypeOf(submissionPrototypeTemp)
		submissionTemp := reflect.ValueOf(values["submission"]).Convert(submissionTypeTemp).Interface().(RewardSubmissionLegacy)
		submission = update_v150rc1_to_v150(submissionTemp)
	} else {
		submissionPrototype := RewardSubmission{}
		submissionType := reflect.TypeOf(submissionPrototype)
		submission = reflect.ValueOf(values["submission"]).Convert(submissionType).Interface().(RewardSubmission)
	}

	eventIntervalStartTime := values["intervalStartTime"].(*big.Int)
	eventIntervalEndTime := values["intervalEndTime"].(*big.Int)
	submissionTime := values["time"].(*big.Int)
	eventData := rewards.RewardsEvent{
		Index:             indexBig,
		ExecutionBlock:    submission.ExecutionBlock,
		ConsensusBlock:    submission.ConsensusBlock,
		IntervalsPassed:   submission.IntervalsPassed,
		TreasuryRPL:       submission.TreasuryRPL,
		TrustedNodeRPL:    submission.TrustedNodeRPL,
		NodeRPL:           submission.NodeRPL,
		NodeETH:           submission.NodeETH,
		UserETH:           submission.UserETH,
		MerkleRoot:        common.BytesToHash(submission.MerkleRoot[:]),
		MerkleTreeCID:     submission.MerkleTreeCID,
		IntervalStartTime: time.Unix(eventIntervalStartTime.Int64(), 0),
		IntervalEndTime:   time.Unix(eventIntervalEndTime.Int64(), 0),
		SubmissionTime:    time.Unix(submissionTime.Int64(), 0),
	}

	return eventData, nil

}

func update_v150rc1_to_v150(oldEvent RewardSubmissionLegacy) RewardSubmission {
	newEvent := RewardSubmission{
		RewardIndex:     oldEvent.RewardIndex,
		ExecutionBlock:  oldEvent.ExecutionBlock,
		ConsensusBlock:  oldEvent.ConsensusBlock,
		MerkleRoot:      oldEvent.MerkleRoot,
		MerkleTreeCID:   oldEvent.MerkleTreeCID,
		IntervalsPassed: oldEvent.IntervalsPassed,
		TreasuryRPL:     oldEvent.TreasuryRPL,
		TrustedNodeRPL:  oldEvent.TrustedNodeRPL,
		NodeRPL:         oldEvent.NodeRPL,
		NodeETH:         oldEvent.NodeETH,
		UserETH:         big.NewInt(0),
	}

	return newEvent
}

type RewardSubmissionLegacy struct {
	RewardIndex     *big.Int   `json:"rewardIndex"`
	ExecutionBlock  *big.Int   `json:"executionBlock"`
	ConsensusBlock  *big.Int   `json:"consensusBlock"`
	MerkleRoot      [32]byte   `json:"merkleRoot"`
	MerkleTreeCID   string     `json:"merkleTreeCID"`
	IntervalsPassed *big.Int   `json:"intervalsPassed"`
	TreasuryRPL     *big.Int   `json:"treasuryRPL"`
	TrustedNodeRPL  []*big.Int `json:"trustedNodeRPL"`
	NodeRPL         []*big.Int `json:"nodeRPL"`
	NodeETH         []*big.Int `json:"nodeETH"`
}

type RewardSubmission struct {
	RewardIndex     *big.Int   `json:"rewardIndex"`
	ExecutionBlock  *big.Int   `json:"executionBlock"`
	ConsensusBlock  *big.Int   `json:"consensusBlock"`
	MerkleRoot      [32]byte   `json:"merkleRoot"`
	MerkleTreeCID   string     `json:"merkleTreeCID"`
	IntervalsPassed *big.Int   `json:"intervalsPassed"`
	TreasuryRPL     *big.Int   `json:"treasuryRPL"`
	TrustedNodeRPL  []*big.Int `json:"trustedNodeRPL"`
	NodeRPL         []*big.Int `json:"nodeRPL"`
	NodeETH         []*big.Int `json:"nodeETH"`
	UserETH         *big.Int   `json:"userETH"`
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

		event, err := GetRewardSnapshotEvent(
			rp.API,
			interval,
			big.NewInt(int64(GethEventLogInterval)),
			nil,
			nil, // rewardPoolAddress
		)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				if previousRewardsPoolAddress[utils.Config.Chain.Name] == "" {
					break
				}
				oldAddress := common.HexToAddress(previousRewardsPoolAddress[utils.Config.Chain.Name])
				event, err = GetRewardSnapshotEvent(
					rp.API,
					interval,
					big.NewInt(int64(GethEventLogInterval)),
					nil,
					&oldAddress,
				)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						break
					} else {
						return err
					}
				}
			} else {
				return err
			}

		}

		var dbInterval uint64
		err = db.WriterDb.Get(&dbInterval, `SELECT id FROM rocketpool_reward_tree WHERE id = $1`, interval)
		if err == sql.ErrNoRows {
			missingIntervals = append(missingIntervals, event)
		} else if err != nil {
			return err
		}

	}

	if len(missingIntervals) == 0 {
		return nil
	}

	for _, missingInterval := range missingIntervals {
		logrus.Infof("Downloading interval %d file... ", missingInterval.Index)
		bytes, err := DownloadRewardsFile(fmt.Sprintf("rp-rewards-prater-%v.json", missingInterval.Index), missingInterval.Index.Uint64(), missingInterval.MerkleTreeCID, true)
		if err != nil {
			return fmt.Errorf("can not download reward file %v. Error %v", missingInterval.Index, err)
		}

		proofWrapper, err := getRewardsData(bytes)

		merkleRootFromFile := common.HexToHash(proofWrapper.MerkleRoot)
		if missingInterval.MerkleRoot != merkleRootFromFile {
			return fmt.Errorf("invalid merkle root value : %w", err)
		}

		_, err = db.WriterDb.Exec(`INSERT INTO rocketpool_reward_tree (id, data) VALUES($1, $2)`, missingInterval.Index.Uint64(), bytes)
		if err != nil {
			return fmt.Errorf("can not store reward file %v. Error %v", missingInterval.Index, err)
		}
		logrus.Infof("Downloaded rocketpool rewards tree %v", missingInterval.Index)

		if missingInterval.Index.Uint64() > rp.LastRewardTree {
			rp.LastRewardTree = missingInterval.Index.Uint64()
		}
	}

	return nil
}

func (rp *RocketpoolExporter) Update(count int64) error {
	var wg errgroup.Group
	wg.Go(func() error {
		if count%8 == 0 {
			return rp.DownloadMissingRewardTrees()
		}
		return nil
	})
	wg.Go(func() error { return rp.UpdateMinipools() })
	wg.Go(func() error { return rp.UpdateNodes(count%12 == 0) })
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
	if count%60 == 0 { // every hour (smart contracts aren't updated that often)
		err = rp.SaveNetworkStats()
		if err != nil {
			return err
		}
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
	for _, a := range minipoolAddresses {
		addrHex := a.Hex()
		if mp, exists := rp.MinipoolsByAddress[addrHex]; exists {
			err = mp.Update(rp.API)
			if err != nil {
				return err
			}
			continue
		}
		mp, err := NewRocketpoolMinipool(rp.API, a.Bytes())
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

	isMergeUpdateDeployed, err := IsMergeUpdateDeployed(rp.API)
	if err != nil {
		return err
	}

	var allRocketpoolRewardTrees map[uint64]RewardsFile = nil
	if isMergeUpdateDeployed {
		allRocketpoolRewardTrees, err = getRocketpoolRewardTrees(rp.API)
		if err != nil {
			return err
		}
	}

	if includeCumulativeRpl {
		legacyRewardsPool := common.HexToAddress(legacyRewardsPoolAddress[utils.Config.Chain.Name])
		legacyClaimNode := common.HexToAddress(legacyClaimNodeAddress[utils.Config.Chain.Name])
		rp.NodeRPLCumulative, err = CalculateLifetimeNodeRewardsAllLegacy(
			rp.API,
			big.NewInt(60000),
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
			err = node.Update(rp.API, allRocketpoolRewardTrees, includeCumulativeRpl, rp.NodeRPLCumulative)
			if err != nil {
				return err
			}
			continue
		}
		node, err := NewRocketpoolNode(rp.API, a.Bytes(), allRocketpoolRewardTrees, rp.NodeRPLCumulative)
		if err != nil {
			return err
		}
		rp.NodesByAddress[addrHex] = node
	}

	return nil
}

func getRocketpoolRewardTrees(rp *rocketpool.RocketPool) (map[uint64]RewardsFile, error) {
	var allRewards map[uint64]RewardsFile = map[uint64]RewardsFile{}
	for i := uint64(0); ; i++ {
		var jsonData []byte
		err := db.WriterDb.Get(&jsonData, `SELECT data FROM rocketpool_reward_tree WHERE id = $1`, i)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return allRewards, fmt.Errorf("Can not load claimedInterval %v tree from database, is it exported? %v", i, err)
		}
		allRewards[i], err = getRewardsData(jsonData)
		if err != nil {
			logrus.Infof("err getting reward data %v", err)
			return allRewards, fmt.Errorf("Can parsing reward tree data to struct for interval %v. Error %v", i, err)
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

	effectiveRplStake, err := getBigIntFrom(rp.API, "rocketNetworkPrices", "getEffectiveRPLStake")
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
		nodeOperatorRewards, err = rewards.GetPendingRPLRewards(rp.API, nil)
		if err != nil {
			return err
		}
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
		EffectiveRPLStake:      effectiveRplStake,
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
	currentVersion, err := rputil.GetCurrentVersion(rp)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.1.0")
	return constraint.Check(currentVersion), nil
}

func getBigIntFrom(rp *rocketpool.RocketPool, contract string, method string, args ...interface{}) (*big.Int, error) {
	rocketRewardsPool, err := rp.GetContract(contract)
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
		}
		stmt := fmt.Sprintf(`insert into rocketpool_minipools (rocketpool_storage_address, address, pubkey, status, status_time, node_address, node_fee, deposit_type) values %s on conflict (rocketpool_storage_address, address) do update set pubkey = excluded.pubkey, status = excluded.status, status_time = excluded.status_time, node_address = excluded.node_address, node_fee = excluded.node_fee, deposit_type = excluded.deposit_type`, strings.Join(valueStrings, ","))
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

	nArgs := 11

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
				unclaimed_rpl_rewards
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
				unclaimed_rpl_rewards = excluded.unclaimed_rpl_rewards
		`, strings.Join(valueStrings, ","))

		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting into rocketpool_nodes: %w", err)
		}
	}

	return tx.Commit()
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
		for _, vote := range val.MemberVotes {
			data = append(data, vote)
		}
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
		}
		stmt := fmt.Sprintf(`insert into rocketpool_dao_members (rocketpool_storage_address, address, id, url, joined_time, last_proposal_time, rpl_bond_amount, unbonded_validator_count) values %s on conflict (rocketpool_storage_address, address) do update set id = excluded.id, url = excluded.url, joined_time = excluded.joined_time, last_proposal_time = excluded.last_proposal_time, rpl_bond_amount = excluded.rpl_bond_amount, unbonded_validator_count = excluded.unbonded_validator_count`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting into rocketpool_dao_members: %w", err)
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
	_, err := db.WriterDb.Exec("INSERT INTO rocketpool_network_stats (ts, rpl_price, claim_interval_time, claim_interval_time_start, current_node_fee, current_node_demand, reth_supply, effective_rpl_staked, node_operator_rewards, reth_exchange_rate, node_count, minipool_count, odao_member_count, total_eth_staking, total_eth_balance) VALUES(now(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		rp.NetworkStats.RPLPrice.String(),
		rp.NetworkStats.ClaimIntervalTime.String(),
		rp.NetworkStats.ClaimIntervalTimeStart,
		rp.NetworkStats.CurrentNodeFee,
		rp.NetworkStats.CurrentNodeDemand.String(),
		rp.NetworkStats.RETHSupply.String(),
		rp.NetworkStats.EffectiveRPLStake.String(),
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
	Address     []byte    `db:"address"`
	Pubkey      []byte    `db:"pubkey"`
	NodeAddress []byte    `db:"node_address"`
	NodeFee     float64   `db:"node_fee"`
	DepositType string    `db:"deposit_type"`
	Status      string    `db:"status"`
	StatusTime  time.Time `db:"status_time"`
}

func NewRocketpoolMinipool(rp *rocketpool.RocketPool, addr []byte) (*RocketpoolMinipool, error) {
	pubk, err := minipool.GetMinipoolPubkey(rp, common.BytesToAddress(addr), nil)
	if err != nil {
		return nil, err
	}
	mp, err := minipool.NewMinipool(rp, common.BytesToAddress(addr))
	if err != nil {
		return nil, err
	}
	nodeAddr, err := mp.GetNodeAddress(nil)
	if err != nil {
		return nil, err
	}
	nodeFee, err := mp.GetNodeFee(nil)
	if err != nil {
		return nil, err
	}
	depositType, err := mp.GetDepositType(nil)
	if err != nil {
		return nil, err
	}
	rpm := &RocketpoolMinipool{
		Address:     addr,
		Pubkey:      pubk.Bytes(),
		NodeAddress: nodeAddr.Bytes(),
		NodeFee:     nodeFee,
		DepositType: depositType.String(),
	}
	err = rpm.Update(rp)
	if err != nil {
		return nil, err
	}
	return rpm, nil
}

func (this *RocketpoolMinipool) Update(rp *rocketpool.RocketPool) error {
	mp, err := minipool.NewMinipool(rp, common.BytesToAddress(this.Address))
	if err != nil {
		return err
	}

	var wg errgroup.Group
	var status rpTypes.MinipoolStatus
	var statusTime time.Time

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

	if err := wg.Wait(); err != nil {
		return err
	}

	this.Status = status.String()
	this.StatusTime = statusTime

	return nil
}

type RocketpoolNode struct {
	Address                []byte   `db:"address"`
	TimezoneLocation       string   `db:"timezone_location"`
	RPLStake               *big.Int `db:"rpl_stake"`
	MinRPLStake            *big.Int `db:"min_rpl_stake"`
	MaxRPLStake            *big.Int `db:"max_rpl_stake"`
	RPLCumulativeRewards   *big.Int `db:"rpl_cumulative_rewards"`
	SmoothingPoolOptedIn   bool     `db:"smoothing_pool_opted_in"`
	ClaimedSmoothingPool   *big.Int `db:"claimed_smoothing_pool"`
	UnclaimedSmoothingPool *big.Int `db:"unclaimed_smoothing_pool"`
	UnclaimedRPLRewards    *big.Int `db:"unclaimed_rpl_rewards"`
}

func NewRocketpoolNode(rp *rocketpool.RocketPool, addr []byte, rewardTrees map[uint64]RewardsFile, legacyClaims map[string]*big.Int) (*RocketpoolNode, error) {
	rpn := &RocketpoolNode{
		Address: addr,
	}
	tl, err := node.GetNodeTimezoneLocation(rp, common.BytesToAddress(addr), nil)
	if err != nil {
		return nil, err
	}
	rpn.TimezoneLocation = tl
	err = rpn.Update(rp, rewardTrees, true, legacyClaims)
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

func (this *RocketpoolNode) Update(rp *rocketpool.RocketPool, rewardTrees map[uint64]RewardsFile, includeCumulativeRpl bool, legacyClaims map[string]*big.Int) error {
	stake, err := node.GetNodeRPLStake(rp, common.BytesToAddress(this.Address), nil)
	if err != nil {
		return err
	}
	minStake, err := node.GetNodeMinimumRPLStake(rp, common.BytesToAddress(this.Address), nil)
	if err != nil {
		return err
	}
	maxStake, err := node.GetNodeMaximumRPLStake(rp, common.BytesToAddress(this.Address), nil)
	if err != nil {
		return err
	}

	if rewardTrees != nil {
		nodeAddress := common.BytesToAddress(this.Address)

		this.SmoothingPoolOptedIn, err = node.GetSmoothingPoolRegistrationState(rp, nodeAddress, nil)
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

			unclaimed, claimed, err := GetClaimStatus(rp, nodeAddress)
			if err != nil {
				return err
			}

			// Get the info for each claimed interval
			for _, claimedInterval := range claimed {
				rewardData := rewardTrees[claimedInterval]

				rewards, exists := rewardData.NodeRewards[nodeAddress]

				if exists {
					claimedSum.RplColl = claimedSum.RplColl.Add(claimedSum.RplColl, &rewards.CollateralRpl.Int)
					claimedSum.SmoothingPoolEth = claimedSum.SmoothingPoolEth.Add(claimedSum.SmoothingPoolEth, &rewards.SmoothingPoolEth.Int)
					claimedSum.OdaoRpl = claimedSum.OdaoRpl.Add(claimedSum.OdaoRpl, &rewards.OracleDaoRpl.Int)
				}
			}

			// Get the unclaimed rewards
			for _, unclaimedInterval := range unclaimed {
				rewardData, exists := rewardTrees[unclaimedInterval]

				rewards, exists := rewardData.NodeRewards[nodeAddress]

				if exists {
					unclaimedSum.RplColl = unclaimedSum.RplColl.Add(unclaimedSum.RplColl, &rewards.CollateralRpl.Int)
					unclaimedSum.SmoothingPoolEth = unclaimedSum.SmoothingPoolEth.Add(unclaimedSum.SmoothingPoolEth, &rewards.SmoothingPoolEth.Int)
					unclaimedSum.OdaoRpl = unclaimedSum.OdaoRpl.Add(unclaimedSum.OdaoRpl, &rewards.OracleDaoRpl.Int)
				}
			}

			this.RPLCumulativeRewards = claimedSum.RplColl
			if legacyAmount, exists := legacyClaims[nodeAddress.Hex()]; exists {
				this.RPLCumulativeRewards = this.RPLCumulativeRewards.Add(this.RPLCumulativeRewards, legacyAmount)
			}
			this.ClaimedSmoothingPool = claimedSum.SmoothingPoolEth
			this.UnclaimedSmoothingPool = unclaimedSum.SmoothingPoolEth
			this.UnclaimedRPLRewards = unclaimedSum.RplColl
		}
	}

	this.RPLStake = stake
	this.MinRPLStake = minStake
	this.MaxRPLStake = maxStake

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

// Gets the intervals the node can claim and the intervals that have already been claimed
func GetClaimStatus(rp *rocketpool.RocketPool, nodeAddress common.Address) (unclaimed []uint64, claimed []uint64, err error) {
	// Get the current interval
	currentIndexBig, err := rewards.GetRewardIndex(rp, nil)
	if err != nil {
		return
	}

	currentIndex := currentIndexBig.Uint64() // This is guaranteed to be from 0 to 65535 so the conversion is legal
	if currentIndex == 0 {
		// If we're still in the first interval, there's nothing to report.
		return
	}

	// Get the claim status of every interval that's happened so far
	one := big.NewInt(1)
	bucket := currentIndex / 256
	for i := uint64(0); i <= bucket; i++ {
		bucketBig := big.NewInt(int64(i))
		bucketBytes := [32]byte{}
		bucketBig.FillBytes(bucketBytes[:])

		var bitmap *big.Int
		bitmap, err = rp.RocketStorage.GetUint(nil, crypto.Keccak256Hash([]byte("rewards.interval.claimed"), nodeAddress.Bytes(), bucketBytes[:]))

		for j := uint64(0); j < 256; j++ {
			targetIndex := i*256 + j
			if targetIndex >= currentIndex {
				// End once we've hit the current interval
				break
			}

			mask := big.NewInt(0)
			mask.Lsh(one, uint(j))
			maskedBitmap := big.NewInt(0)
			maskedBitmap.And(bitmap, mask)

			if maskedBitmap.Cmp(mask) == 0 {
				// This bit was flipped, so it's been claimed already
				claimed = append(claimed, targetIndex)
			} else {
				// This bit was not flipped, so it hasn't been claimed yet
				unclaimed = append(unclaimed, targetIndex)
			}
		}
	}

	return
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

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("can not load lifetime rewards: $1", err)
	}

	// Iterate over the logs and sum the amount
	sumMap := make(map[string]*big.Int)
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

func getRocketRewardsPool(rp *rocketpool.RocketPool, address *common.Address) (*rocketpool.Contract, error) {
	rocketRewardsPoolLock.Lock()
	defer rocketRewardsPoolLock.Unlock()

	if address == nil {
		return rp.GetContract("rocketRewardsPool")
	} else {
		return rp.VersionManager.V1_5_0_RC1.GetContractWithAddress("rocketRewardsPool", *address)
	}
}

func getRocketRewardsPoolLegacy(rp *rocketpool.RocketPool, address *common.Address) (*rocketpool.Contract, error) {
	rocketRewardsPoolLock.Lock()
	defer rocketRewardsPoolLock.Unlock()
	if address == nil {
		return rp.VersionManager.V1_0_0.GetContract("rocketRewardsPool")
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
		return rp.VersionManager.V1_0_0.GetContract("rocketClaimNode")
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

func (this *RocketpoolDAOProposal) Update(rp *rocketpool.RocketPool) error {
	pd, err := dao.GetProposalDetails(rp, this.ID, nil)
	if err != nil {
		return err
	}
	this.ID = pd.ID
	this.DAO = pd.DAO
	this.ProposerAddress = pd.ProposerAddress.Bytes()
	this.Message = pd.Message
	this.CreatedTime = time.Unix(int64(pd.CreatedTime), 0)
	this.StartTime = time.Unix(int64(pd.StartTime), 0)
	this.EndTime = time.Unix(int64(pd.EndTime), 0)
	this.ExpiryTime = time.Unix(int64(pd.ExpiryTime), 0)
	this.VotesRequired = pd.VotesRequired
	this.VotesFor = pd.VotesFor
	this.VotesAgainst = pd.VotesAgainst
	this.MemberVoted = pd.MemberVoted
	this.MemberSupported = pd.MemberSupported
	this.IsCancelled = pd.IsCancelled
	this.IsExecuted = pd.IsExecuted
	this.Payload = pd.Payload
	this.State = pd.State.String()

	// Update member votes
	this.MemberVotes = []RocketpoolDAOProposalMemberVotes{}
	members, err := rpDAOTrustedNode.GetMembers(rp, nil)
	if err != nil {
		return err
	}
	for _, m := range members {
		memberVoted, err := dao.GetProposalMemberVoted(rp, this.ID, m.Address, nil)
		if err != nil {
			return err
		}

		memberSupported, err := dao.GetProposalMemberSupported(rp, this.ID, m.Address, nil)
		if err != nil {
			return err
		}

		this.MemberVotes = append(this.MemberVotes, RocketpoolDAOProposalMemberVotes{
			ProposalID: this.ID,
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

func (this *RocketpoolDAOMember) Update(rp *rocketpool.RocketPool) error {
	d, err := rpDAOTrustedNode.GetMemberDetails(rp, common.BytesToAddress(this.Address), nil)
	if err != nil {
		return err
	}
	this.ID = d.ID
	this.URL = d.Url
	this.JoinedTime = time.Unix(int64(d.JoinedTime), 0)
	this.LastProposalTime = time.Unix(int64(d.LastProposalTime), 0)
	this.RPLBondAmount = d.RPLBondAmount
	this.UnbondedValidatorCount = d.UnbondedValidatorCount
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
			bytes, err := ioutil.ReadAll(resp.Body)
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
