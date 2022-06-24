package exporter

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"eth2-exporter/db"
	"eth2-exporter/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gethRPC "github.com/ethereum/go-ethereum/rpc"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
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
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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
	NetworkStats        RocketpoolNetworkStats
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
	errorInterval := time.Second * 10
	t := time.NewTicker(rp.UpdateInterval)
	defer t.Stop()
	var count int64 = 0
	for {
		t0 := time.Now()
		var err error
		err = rp.Update()
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

func (rp *RocketpoolExporter) Update() error {
	var wg errgroup.Group
	wg.Go(func() error { return rp.UpdateMinipools() })
	wg.Go(func() error { return rp.UpdateNodes() })
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

func (rp *RocketpoolExporter) UpdateNodes() error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("updated rocketpool-nodes")
	}(t0)

	nodeAddresses, err := node.GetNodeAddresses(rp.API, nil)
	if err != nil {
		return err
	}
	for _, a := range nodeAddresses {
		addrHex := a.Hex()
		if node, exists := rp.NodesByAddress[addrHex]; exists {
			err = node.Update(rp.API)
			if err != nil {
				return err
			}
			continue
		}
		node, err := NewRocketpoolNode(rp.API, a.Bytes())
		if err != nil {
			return err
		}
		rp.NodesByAddress[addrHex] = node
	}
	return nil
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

	nodeOperatorRewards, err := getBigIntFrom(rp.API, "rocketRewardsPool", "getClaimingContractAllowance", "rocketClaimNode")
	if err != nil {
		return err
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

	nArgs := 6
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
			valueArgs = append(valueArgs, d.TimezoneLocation)
			valueArgs = append(valueArgs, d.RPLStake.String())
			valueArgs = append(valueArgs, d.MinRPLStake.String())
			valueArgs = append(valueArgs, d.MaxRPLStake.String())
		}
		stmt := fmt.Sprintf(`insert into rocketpool_nodes (rocketpool_storage_address, address, timezone_location, rpl_stake, min_rpl_stake, max_rpl_stake) values %s on conflict (rocketpool_storage_address, address) do update set rpl_stake = excluded.rpl_stake, min_rpl_stake = excluded.min_rpl_stake, max_rpl_stake = excluded.max_rpl_stake`, strings.Join(valueStrings, ","))
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
	Address          []byte   `db:"address"`
	TimezoneLocation string   `db:"timezone_location"`
	RPLStake         *big.Int `db:"rpl_stake"`
	MinRPLStake      *big.Int `db:"min_rpl_stake"`
	MaxRPLStake      *big.Int `db:"max_rpl_stake"`
}

func NewRocketpoolNode(rp *rocketpool.RocketPool, addr []byte) (*RocketpoolNode, error) {
	rpn := &RocketpoolNode{
		Address: addr,
	}
	tl, err := node.GetNodeTimezoneLocation(rp, common.BytesToAddress(addr), nil)
	if err != nil {
		return nil, err
	}
	rpn.TimezoneLocation = tl
	err = rpn.Update(rp)
	if err != nil {
		return nil, err
	}
	return rpn, nil
}

func (this *RocketpoolNode) Update(rp *rocketpool.RocketPool) error {
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

	this.RPLStake = stake
	this.MinRPLStake = minStake
	this.MaxRPLStake = maxStake
	return nil
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
