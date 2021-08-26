package exporter

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"eth2-exporter/db"
	"eth2-exporter/utils"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	gethRPC "github.com/ethereum/go-ethereum/rpc"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rocket-pool/rocketpool-go/dao"
	rpDAO "github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rpTypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
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
	rpExporter, err := NewRocketpoolExporter(rpEth1Client, "0xd8Cd47263414aFEca62d6e2a3917d6600abDceB3", db.DB)
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

type RocketpoolMinipool struct {
	Address     []byte    `db:"address"`
	Pubkey      []byte    `db:"Pubkey"`
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
	// fmt.Printf("update node %x %v %v %v\n", this.Address, stake, minStake, maxStake)
	return nil
}

type RocketpoolProposal struct {
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

func NewRocketpoolProposal(rp *rocketpool.RocketPool, pid uint64) (*RocketpoolProposal, error) {
	p := &RocketpoolProposal{ID: pid}
	err := p.Update(rp)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (this *RocketpoolProposal) Update(rp *rocketpool.RocketPool) error {
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

type RocketpoolExporter struct {
	Eth1Client         *ethclient.Client
	API                *rocketpool.RocketPool
	DB                 *sqlx.DB
	MinipoolsByAddress map[string]*RocketpoolMinipool
	NodesByAddress     map[string]*RocketpoolNode
	ProposalsByID      map[uint64]*RocketpoolProposal
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
	rpe.MinipoolsByAddress = map[string]*RocketpoolMinipool{}
	rpe.NodesByAddress = map[string]*RocketpoolNode{}
	rpe.ProposalsByID = map[uint64]*RocketpoolProposal{}
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
	err = rp.InitProposals()
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
	for _, mp := range dbRes {
		rp.MinipoolsByAddress[fmt.Sprintf("%x", mp.Address)] = &mp
	}
	return nil
}

func (rp *RocketpoolExporter) InitNodes() error {
	dbRes := []RocketpoolNode{}
	err := rp.DB.Select(&dbRes, `select * from rocketpool_nodes`)
	if err != nil {
		return err
	}
	for _, node := range dbRes {
		rp.NodesByAddress[fmt.Sprintf("%x", node.Address)] = &node
	}
	return nil
}

func (rp *RocketpoolExporter) InitProposals() error {
	dbRes := []RocketpoolProposal{}
	err := rp.DB.Select(&dbRes, `select * from rocketpool_proposals`)
	if err != nil {
		return err
	}
	for _, proposal := range dbRes {
		rp.ProposalsByID[proposal.ID] = &proposal
	}
	return nil
}

func (rp *RocketpoolExporter) Run() error {
	for {
		t0 := time.Now()
		var err error
		err = rp.Update()
		if err != nil {
			logger.WithError(err).Errorf("error updating rocketpool-data")
			time.Sleep(time.Second * 2)
			continue
		}
		err = rp.Save()
		if err != nil {
			logger.WithError(err).Errorf("error saving rocketpool-data")
			time.Sleep(time.Second * 2)
			continue
		}

		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("exported rocketpool-data")
		time.Sleep(time.Second * 2)
	}
}

func (rp *RocketpoolExporter) Update() error {
	var err error
	err = rp.UpdateMinipools()
	if err != nil {
		return err
	}
	err = rp.UpdateNodes()
	if err != nil {
		return err
	}
	err = rp.UpdateProposals()
	if err != nil {
		return err
	}
	return nil
}

func (rp *RocketpoolExporter) Save() error {
	var err error
	err = rp.SaveMinipools()
	if err != nil {
		return err
	}
	err = rp.SaveNodes()
	if err != nil {
		return err
	}
	err = rp.SaveProposals()
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

func (rp *RocketpoolExporter) UpdateProposals() error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("updated rocketpool-proposals")
	}(t0)

	pc, err := rpDAO.GetProposalCount(rp.API, nil)
	if err != nil {
		return err
	}
	for i := uint64(0); i < pc; i++ {
		p, err := NewRocketpoolProposal(rp.API, i+1)
		if err != nil {
			return err
		}
		rp.ProposalsByID[i] = p
	}
	return nil
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

	tx, err := db.DB.Beginx()
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
			return err
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

	tx, err := db.DB.Beginx()
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
			return err
		}
	}

	return tx.Commit()
}

func (rp *RocketpoolExporter) SaveProposals() error {
	if len(rp.ProposalsByID) == 0 {
		return nil
	}

	t0 := time.Now()
	defer func(t0 time.Time) {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Debugf("saved rocketpool-proposals")
	}(t0)

	data := make([]*RocketpoolProposal, len(rp.ProposalsByID))
	i := 0
	for _, proposal := range rp.ProposalsByID {
		data[i] = proposal
		i++
	}

	tx, err := db.DB.Beginx()
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
		stmt := fmt.Sprintf(`insert into rocketpool_proposals (rocketpool_storage_address, id, dao, proposer_address, message, created_time, start_time, end_time, expiry_time, votes_required, votes_for, votes_against, member_voted, member_supported, is_cancelled, is_executed, payload, state) values %s on conflict (rocketpool_storage_address, id) do update set dao = excluded.dao, proposer_address = excluded.proposer_address, message = excluded.message, created_time = excluded.created_time, start_time = excluded.start_time, end_time = excluded.end_time, expiry_time = excluded.expiry_time, votes_required = excluded.votes_required, votes_for = excluded.votes_for, votes_against = excluded.votes_against, member_voted = excluded.member_voted, member_supported = excluded.member_supported, is_cancelled = excluded.is_cancelled, is_executed = excluded.is_executed, payload = excluded.payload, state = excluded.state`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// func (rp *RocketpoolExporter) Run2() error {
// 	contractCreationBlock := uint64(4443528)
// 	// contractCreationBlock = 5363741
// 	lastQueriedBlock := contractCreationBlock
// 	maxQuerySize := 100000
// 	updatedMinipoolsByAddress := map[string]*RocketpoolMinipool{}
// 	for {
// 		t0 := time.Now()
// 		header, err := rp.Eth1Client.HeaderByNumber(context.Background(), nil)
// 		if err != nil {
// 			logger.WithError(err).Errorf("error getting header from eth1-client")
// 			time.Sleep(time.Second * 10)
// 			continue
// 		}
// 		blockHeight := header.Number.Uint64()
// 		fromBlock := lastQueriedBlock + 1
// 		toBlock := blockHeight
// 		if fromBlock > toBlock {
// 			time.Sleep(time.Second * 10)
// 			continue
// 		}
// 		if toBlock-fromBlock > uint64(maxQuerySize) {
// 			toBlock = fromBlock + uint64(maxQuerySize)
// 		}
// 		if toBlock > blockHeight {
// 			toBlock = blockHeight
// 		}
// 		err = rp.Query(fromBlock, toBlock, updatedMinipoolsByAddress)
// 		if err != nil {
// 			logger.WithFields(logrus.Fields{"error": err, "fromBlock": fromBlock, "toBlock": toBlock, "duration": time.Since(t0)}).Errorf("failed exporting rocketpool-data")
// 			time.Sleep(time.Second * 10)
// 			continue
// 		}
// 		lastQueriedBlock = toBlock
// 		logger.WithFields(logrus.Fields{"fromBlock": fromBlock, "toBlock": toBlock, "duration": time.Since(t0)}).Infof("succeeded exporting rocketpool-data")
//
// 		if toBlock == blockHeight {
// 			// save updatedMinipoolsByAddress
// 		}
//
// 		time.Sleep(time.Second * 60)
// 	}
// 	return nil
// }

// func (rp *RocketpoolExporter) Query(fromBlock, toBlock uint64, updatedMinipoolsByAddress map[string]*RocketpoolMinipool) error {
// 	rp.AddMissingMinipools()
// 	rocketMinipoolABI, err := rp.API.GetABI("rocketMinipool")
// 	if err != nil {
// 		return err
// 	}
// 	logs, err := rp.Eth1Client.FilterLogs(context.Background(), ethereum.FilterQuery{
// 		Addresses: rp.MinipoolAddresses,
// 		Topics: [][]common.Hash{{
// 			rocketMinipoolABI.Events["StatusUpdated"].ID,
// 		}},
// 		FromBlock: big.NewInt(int64(fromBlock)),
// 		ToBlock:   big.NewInt(int64(toBlock)),
// 	})
// 	for _, log := range logs {
// 		if rocketMinipoolABI.Events["StatusUpdated"].ID == log.Topics[0] {
// 			values := make(map[string]interface{})
// 			err := rocketMinipoolABI.Events["StatusUpdated"].Inputs.UnpackIntoMap(values, log.Data)
// 			if err != nil {
// 				return err
// 			}
// 			t := time.Unix(values["time"].(*big.Int).Int64(), 0)
// 			fmt.Printf("status updated %x %x %v %v\n", log.TxHash, log.Address, t, rpTypes.MinipoolStatuses[log.Topics[1].Big().Uint64()])
// 			continue
// 		}
// 	}
// 	return nil
// }

func (rp *RocketpoolExporter) PrintInfo() {
	// https://github.com/rocket-pool/rocketpool/blob/master/migrations/2_deploy_contracts.js#L36
	for _, n := range []string{
		"rocketMinipoolManager",
		"rocketNodeManager",
		"rocketDAONodeTrustedActions",
		"rocketDAONodeTrustedSettingsMembers",
	} {
		c, err := rp.API.GetContract("rocketMinipoolManager")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v: %v\n", n, c.Address.Hex())
	}

	c, err := minipool.GetMinipoolCountPerStatus(rp.API, 0, 0, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("GetMinipoolCountPerStatus: %+v\n", c)
}

func (rp *RocketpoolExporter) Test() {
	rp.Init()
	rp.Run()
	if true {
		return
	}
	rp.PrintInfo()
	contractCreationBlock := uint64(4443528)
	// contractCreationBlock = 5363741
	lastQueriedBlock := contractCreationBlock
	maxQuerySize := 100000
	// lookBack := 200
	t := time.NewTicker(time.Second * 2)
	for range t.C {
		t0 := time.Now()
		header, err := rp.Eth1Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			logger.WithError(err).Errorf("error getting header from eth1-client")
			continue
		}
		blockHeight := header.Number.Uint64()

		fromBlock := lastQueriedBlock + 1
		toBlock := blockHeight

		if fromBlock > toBlock {
			continue
		}
		if toBlock-fromBlock > uint64(maxQuerySize) {
			toBlock = fromBlock + uint64(maxQuerySize)
		}
		if toBlock > blockHeight {
			toBlock = blockHeight
		}
		// if toBlock-fromBlock < uint64(lookBack) {
		// 	fromBlock = toBlock - uint64(lookBack)
		// }

		err = rp.TestRun(fromBlock, toBlock)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err, "fromBlock": fromBlock, "toBlock": toBlock, "duration": time.Since(t0)}).Errorf("failed exporting rocketpool-data")
			continue
		}
		lastQueriedBlock = toBlock
		logger.WithFields(logrus.Fields{"fromBlock": fromBlock, "toBlock": toBlock, "duration": time.Since(t0)}).Infof("succeeded exporting rocketpool-data")

		if toBlock == blockHeight {
			break
		}
	}
	for _, m := range minipoolsArray {
		fmt.Printf("%x\n", m.Pubkey)
	}
}

func (rp *RocketpoolExporter) TestRun(fromBlock, toBlock uint64) error {
	minipoolAddresses, err := minipool.GetMinipoolAddresses(rp.API, nil)
	if err != nil {
		return err
	}
	rocketMinipoolABI, err := rp.API.GetABI("rocketMinipool")
	if err != nil {
		return err
	}
	topicFilter := [][]common.Hash{{
		rocketMinipoolABI.Events["StatusUpdated"].ID,
	}}
	addressFilter := minipoolAddresses
	logs, err := rp.Eth1Client.FilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: addressFilter,
		Topics:    topicFilter,
		// FromBlock: big.NewInt(int64(5356400)),
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
	})
	for _, log := range logs {
		if rocketMinipoolABI.Events["StatusUpdated"].ID == log.Topics[0] {
			values := make(map[string]interface{})
			err := rocketMinipoolABI.Events["StatusUpdated"].Inputs.UnpackIntoMap(values, log.Data)
			if err != nil {
				panic(err)
			}
			t := time.Unix(values["time"].(*big.Int).Int64(), 0)
			fmt.Printf("status updated %x %x %v %v\n", log.TxHash, log.Address, t, rpTypes.MinipoolStatuses[log.Topics[1].Big().Uint64()])
			continue
		}
	}

	return nil
}

func (rp *RocketpoolExporter) Test8() {

	rp.PrintInfo()
	// rp.Test7()
	t0 := time.Now()
	minipools := []RocketpoolMinipool{}
	minipoolAddresses, err := minipool.GetMinipoolAddresses(rp.API, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("got all addresses", len(minipoolAddresses))
	t1 := time.Now()
	for i, a := range minipoolAddresses {
		if i%10 == 0 {
			fmt.Println(i)
		}
		pubk, err := minipool.GetMinipoolPubkey(rp.API, a, nil)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err, "minipool.address": a.Hex()}).Errorf("error getting minipool-pubkey")
			continue
		}
		mp, err := minipool.NewMinipool(rp.API, a)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err, "minipool.address": a.Hex()}).Errorf("error creating minipool-binding")
			continue
		}
		status, err := mp.GetStatus(nil)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err, "minipool.address": a.Hex()}).Errorf("error getting minipool-status")
			continue
		}
		nodeAddr, err := mp.GetNodeAddress(nil)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err, "minipool.address": a.Hex()}).Errorf("error getting minipool-node-address")
			continue
		}
		minipools = append(minipools, RocketpoolMinipool{
			Address:     a.Bytes(),
			NodeAddress: nodeAddr.Bytes(),
			Pubkey:      pubk.Bytes(),
			Status:      status.String(),
		})
	}
	t2 := time.Now()

	for _, mp := range minipools {
		fmt.Printf("%v %v %v %v\n", mp.Address, mp.NodeAddress, mp.Pubkey, mp.Status)
	}
	fmt.Println(t2.Sub(t1), t1.Sub(t0))
}

func (rp *RocketpoolExporter) Test7() {
	t0 := time.Now()
	ms, err := minipool.GetMinipools(rp.API, nil)
	if err != nil {
		panic(err)
	}
	for _, m := range ms {
		fmt.Printf("%+v\n", m)
	}
	t1 := time.Now()
	ns, err := node.GetNodes(rp.API, nil)
	if err != nil {
		panic(err)
	}
	for _, m := range ns {
		fmt.Printf("%+v\n", m)
	}
	t2 := time.Now()
	fmt.Println(t1.Sub(t0), len(ms))
	fmt.Println(t2.Sub(t1), len(ns))
	rp.PrintInfo()
}

type minipoolType struct {
	Address     string
	NodeAddress string
	Pubkey      string
	Status      string
}

var minipoolsArray = []minipoolType{}

func (rp *RocketpoolExporter) Test6() {
	rp.PrintInfo()
	contractCreationBlock := uint64(4443528)
	// contractCreationBlock = 5363741
	lastQueriedBlock := contractCreationBlock
	maxQuerySize := 100000
	// lookBack := 200
	t := time.NewTicker(time.Second * 2)
	for range t.C {
		t0 := time.Now()
		header, err := rp.Eth1Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			logger.WithError(err).Errorf("error getting header from eth1-client")
			continue
		}
		blockHeight := header.Number.Uint64()

		fromBlock := lastQueriedBlock + 1
		toBlock := blockHeight

		if fromBlock > toBlock {
			continue
		}
		if toBlock-fromBlock > uint64(maxQuerySize) {
			toBlock = fromBlock + uint64(maxQuerySize)
		}
		if toBlock > blockHeight {
			toBlock = blockHeight
		}
		// if toBlock-fromBlock < uint64(lookBack) {
		// 	fromBlock = toBlock - uint64(lookBack)
		// }

		err = rp.TestRun6(fromBlock, toBlock)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err, "fromBlock": fromBlock, "toBlock": toBlock, "duration": time.Since(t0)}).Errorf("failed exporting rocketpool-data")
			continue
		}
		lastQueriedBlock = toBlock
		logger.WithFields(logrus.Fields{"fromBlock": fromBlock, "toBlock": toBlock, "duration": time.Since(t0)}).Infof("succeeded exporting rocketpool-data")

		if toBlock == blockHeight {
			break
		}
	}
	for _, m := range minipoolsArray {
		fmt.Printf("%x\n", m.Pubkey)
	}
}

func (rp *RocketpoolExporter) TestRun6(fromBlock, toBlock uint64) error {
	rocketMinipoolManager, err := rp.API.GetContract("rocketMinipoolManager")
	if err != nil {
		return err
	}
	rocketNodeManager, err := rp.API.GetContract("rocketNodeManager")
	if err != nil {
		return err
	}
	_ = rocketNodeManager
	// rocketMinipoolABI, err := rp.API.GetABI("rocketMinipool")
	// if err != nil {
	// 	return err
	// }
	// rocketTokenRPLABI, err := rp.API.GetABI("rocketTokenRPL")
	// if err != nil {
	// 	return err
	// }
	// rocketDAONodeTrustedActionsContract, err := rp.API.GetContract("rocketDAONodeTrustedActions")
	// if err != nil {
	// 	return err
	// }
	rocketDAONodeTrustedActions, err := rp.API.GetContract("rocketDAONodeTrustedActions")
	if err != nil {
		return err
	}
	addressFilter := []common.Address{
		//*rocketNodeManager.Address,
		*rocketMinipoolManager.Address,
		//*rocketDAONodeTrustedActions.Address,
	}
	_ = addressFilter
	topicFilter := [][]common.Hash{
		{
			// rocketNodeManager.ABI.Events["NodeRegistered"].ID,
			rocketMinipoolManager.ABI.Events["MinipoolCreated"].ID,
			// rocketMinipoolManager.ABI.Events["MinipoolDestroyed"].ID,
			// rocketMinipoolABI.Events["StatusUpdated"].ID,
			// rocketTokenRPLABI.Events["RPLInflationLog"].ID,
			rocketDAONodeTrustedActions.ABI.Events["ActionJoined"].ID,
		},
	}
	_ = topicFilter

	logs, err := rp.Eth1Client.FilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: addressFilter,
		Topics:    topicFilter,
		// FromBlock: big.NewInt(int64(5356400)),
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
	})
	for _, log := range logs {
		if rocketDAONodeTrustedActions.ABI.Events["ActionJoined"].ID == log.Topics[0] {
			values := make(map[string]interface{})
			err := rocketDAONodeTrustedActions.ABI.Events["ActionJoined"].Inputs.UnpackIntoMap(values, log.Data)
			if err != nil {
				return err
			}
			nodeAddress := common.BytesToAddress(log.Topics[1].Bytes()).Bytes()
			fmt.Printf("ActionJoined %x %+v\n", nodeAddress, values)
			continue
		}
		if rocketMinipoolManager.ABI.Events["MinipoolCreated"].ID == log.Topics[0] {
			values := make(map[string]interface{})
			err := rocketMinipoolManager.ABI.Events["MinipoolCreated"].Inputs.UnpackIntoMap(values, log.Data)
			if err != nil {
				return err
			}

			// mp, err := minipool.NewMinipool(rp.API, common.BytesToAddress(log.Topics[1].Bytes()))
			// if err != nil {
			// 	return fmt.Errorf("MinipoolCreated error minipool.NewMinipool: %w", err)
			// }
			// exists, err := minipool.GetMinipoolExists(rp.API, common.BytesToAddress(log.Topics[1].Bytes()), nil)
			// if err != nil {
			// 	return fmt.Errorf("MinipoolCreated error minipool.GetMinipoolExists")
			// }
			pubkey, err := minipool.GetMinipoolPubkey(rp.API, common.BytesToAddress(log.Topics[1].Bytes()), nil)
			if err != nil {
				return fmt.Errorf("MinipoolCreated error minipool.GetMinipoolPubkey")
			}
			// statusDetails, err := mp.GetStatusDetails(nil)
			// if err != nil {
			// 	logger.WithError(err).Errorf("error MinipoolCreated minipool.GetStatusDetails for %x (exists: %v)", pubkey, exists)
			// 	continue
			// }
			// details, err := minipool.GetMinipoolDetails(rp.API, common.BytesToAddress(log.Topics[1].Bytes()), nil)
			// if err != nil {
			// 	return fmt.Errorf("error MinipoolCreated minipool.GetMinipoolDetails: %w", err)
			// }

			minipoolsArray = append(minipoolsArray, minipoolType{
				Address:     common.BytesToAddress(log.Topics[1].Bytes()).Hex(),
				NodeAddress: common.BytesToAddress(log.Topics[2].Bytes()).Hex(),
				// CreationTimestamp: values["time"].(*big.Int).Uint64(),
				Pubkey: pubkey.Hex(),
				// Status:            rpTypes.MinipoolStatuses[statusDetails.Status],
			})
			//  fmt.Printf("%v MinipoolCreated: %+v\n", log.BlockNumber, minipool)
		}
		if false && rocketMinipoolManager.ABI.Events["MinipoolDestroyed"].ID == log.Topics[0] {
			values := make(map[string]interface{})
			err := rocketMinipoolManager.ABI.Events["MinipoolDestroyed"].Inputs.UnpackIntoMap(values, log.Data)
			if err != nil {
				continue // return err
			}

			mp, err := minipool.NewMinipool(rp.API, common.BytesToAddress(log.Topics[1].Bytes()))
			if err != nil {
				continue // return fmt.Errorf("error minipool.NewMinipool: %w", err)
			}
			statusDetails, err := mp.GetStatusDetails(nil)
			if err != nil {
				continue // return fmt.Errorf("error minipool.GetStatusDetails: %w", err)
			}
			details, err := minipool.GetMinipoolDetails(rp.API, common.BytesToAddress(log.Topics[1].Bytes()), nil)
			if err != nil {
				continue // return fmt.Errorf("error minipool.GetMinipoolDetails: %w", err)
			}

			minipool := struct {
				Address           string
				NodeAddress       string
				CreationTimestamp uint64
				Pubkey            string
				Status            string
			}{
				Address:           common.BytesToAddress(log.Topics[1].Bytes()).Hex(),
				NodeAddress:       common.BytesToAddress(log.Topics[2].Bytes()).Hex(),
				CreationTimestamp: values["time"].(*big.Int).Uint64(),
				Pubkey:            details.Pubkey.Hex(),
				Status:            rpTypes.MinipoolStatuses[statusDetails.Status],
			}
			fmt.Printf("%v MinipoolDestroyed: %+v\n", log.BlockNumber, minipool)
		}
	}
	return nil
}

func (rp *RocketpoolExporter) Test4() {
	rocketMinipoolManager, err := rp.API.GetContract("rocketMinipoolManager")
	if err != nil {
		panic(err)
	}
	rocketNodeManager, err := rp.API.GetContract("rocketNodeManager")
	if err != nil {
		panic(err)
	}
	rocketMinipoolABI, err := rp.API.GetABI("rocketMinipool")
	if err != nil {
		panic(err)
	}
	rocketTokenRPLABI, err := rp.API.GetABI("rocketTokenRPL")
	if err != nil {
		panic(err)
	}
	rocketDAONodeTrustedActionsContract, err := rp.API.GetContract("rocketDAONodeTrustedActions")
	if err != nil {
		panic(err)
	}
	rocketDAONodeTrustedActionsABI, err := rp.API.GetABI("rocketDAONodeTrustedActions")
	if err != nil {
		panic(err)
	}

	fmt.Println("rocketDAONodeTrustedActionsContract", *rocketDAONodeTrustedActionsContract.Address)
	fmt.Println("rocketNodeManager.Address", *rocketNodeManager.Address)
	fmt.Println("rocketMinipoolManager.Address", *rocketMinipoolManager.Address)
	fmt.Println("rocketNodeManager.Topics.NodeRegistered", rocketNodeManager.ABI.Events["NodeRegistered"].ID)
	fmt.Println("rocketMinipoolManager.Topics.MinipoolCreated", rocketMinipoolManager.ABI.Events["MinipoolCreated"].ID)
	fmt.Println("rocketMinipoolABI.Events[StatusUpdated].ID", rocketMinipoolABI.Events["StatusUpdated"].ID)
	fmt.Println("RPLInflationLog", rocketTokenRPLABI.Events["RPLInflationLog"].ID)

	fmt.Println("--------")

	addressFilter := []common.Address{*rocketNodeManager.Address, *rocketMinipoolManager.Address}
	topicFilter := [][]common.Hash{
		{
			// rocketNodeManager.ABI.Events["NodeRegistered"].ID,
			// rocketMinipoolManager.ABI.Events["MinipoolCreated"].ID,
			// rocketMinipoolABI.Events["StatusUpdated"].ID,
			// rocketTokenRPLABI.Events["RPLInflationLog"].ID,
			rocketDAONodeTrustedActionsABI.Events["ActionJoined"].ID,
		},
	}
	_ = addressFilter
	logs, err := rp.Eth1Client.FilterLogs(context.Background(), ethereum.FilterQuery{
		// Addresses: addressFilter,
		Topics: topicFilter,
		// FromBlock: big.NewInt(int64(5356400)),
		FromBlock: big.NewInt(int64(4443528)),
		ToBlock:   big.NewInt(int64(4443528 + 10000)),
	})
	if err != nil {
		panic(err)
	}

	minipools := []struct {
		Address           []byte
		NodeAddress       []byte
		CreationTimestamp uint64
		Pubkey            []byte
		Status            string
		Log               gethTypes.Log
	}{}

	for _, log := range logs {
		if rocketDAONodeTrustedActionsABI.Events["ActionJoined"].ID == log.Topics[0] {
			values := make(map[string]interface{})
			err := rocketDAONodeTrustedActionsABI.Events["ActionJoined"].Inputs.UnpackIntoMap(values, log.Data)
			if err != nil {
				panic(err)
			}
			nodeAddress := common.BytesToAddress(log.Topics[1].Bytes()).Bytes()
			fmt.Printf("ActionJoined %x %+v\n", nodeAddress, values)
			continue
		}
		if rocketTokenRPLABI.Events["RPLInflationLog"].ID == log.Topics[0] {
			values := make(map[string]interface{})
			err := rocketTokenRPLABI.Events["RPLInflationLog"].Inputs.UnpackIntoMap(values, log.Data)
			if err != nil {
				panic(err)
			}
			fmt.Printf("RPLInflationLog %+v\n", values)
			continue
		}
		if rocketMinipoolABI.Events["StatusUpdated"].ID == log.Topics[0] {
			values := make(map[string]interface{})
			err := rocketMinipoolABI.Events["StatusUpdated"].Inputs.UnpackIntoMap(values, log.Data)
			if err != nil {
				panic(err)
			}
			t := time.Unix(values["time"].(*big.Int).Int64(), 0)
			fmt.Printf("status updated %x %x %v %v\n", log.TxHash, log.Address, t, rpTypes.MinipoolStatuses[log.Topics[1].Big().Uint64()])
			continue
		}
		if true {
			continue
		}
		if rocketMinipoolManager.ABI.Events["MinipoolCreated"].ID == log.Topics[0] {
			values := make(map[string]interface{})
			err := rocketMinipoolManager.ABI.Events["MinipoolCreated"].Inputs.UnpackIntoMap(values, log.Data)
			if err != nil {
				panic(err)
			}

			mp, err := minipool.NewMinipool(rp.API, common.BytesToAddress(log.Topics[1].Bytes()))
			if err != nil {
				panic(err)
			}
			statusDetails, err := mp.GetStatusDetails(nil)
			if err != nil {
				panic(err)
			}
			details, err := minipool.GetMinipoolDetails(rp.API, common.BytesToAddress(log.Topics[1].Bytes()), nil)
			if err != nil {
				panic(err)
			}

			minipools = append(minipools, struct {
				Address           []byte
				NodeAddress       []byte
				CreationTimestamp uint64
				Pubkey            []byte
				Status            string
				Log               gethTypes.Log
			}{
				Address:           common.BytesToAddress(log.Topics[1].Bytes()).Bytes(),
				NodeAddress:       common.BytesToAddress(log.Topics[2].Bytes()).Bytes(),
				CreationTimestamp: values["time"].(*big.Int).Uint64(),
				Pubkey:            details.Pubkey.Bytes(),
				Status:            rpTypes.MinipoolStatuses[statusDetails.Status],
				Log:               log,
			})
		}

		// values := make(map[string]interface{})

		// logger.Infof("%v: %+v: %x", log.TxHash, log.Topics, log.Data)

		// // if rocketMinipoolManager.ABI.EventByID()
		//
		// if rocketNodeManager.ABI.Events["NodeRegistered"].Inputs.UnpackIntoMap(values, log.Data) != nil {
		// 	logger.Error("unable to unpack event: NodeRegistered")
		// 	continue
		// }
		// for k, v := range values {
		// 	logger.Infof("NodeRegistered: %v: %v: %+v", log.Topics, k, v)
		// }
	}
	for _, p := range minipools {
		fmt.Printf("%x %x %x %v %v %v %x\n", p.Address, p.NodeAddress, p.Pubkey, p.Status, time.Unix(int64(p.CreationTimestamp), 0), p.Log.BlockNumber, p.Log.TxHash)
	}
}

func (rp *RocketpoolExporter) Test3() {
	mc, err := minipool.GetMinipoolCount(rp.API, nil)
	if err != nil {
		panic(err)
	}
	nc, err := node.GetNodeCount(rp.API, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(nc, mc)
	// fmt.Println("--")
	// for i := uint64(0); i < nc; i++ {
	// 	addr, err := node.GetNodeAt(rp.API, i, nil)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	nd, err := node.GetNodeDetails(rp.API, addr, nil)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(nd.Exists, nd.Address.Hex())
	// 	if i > 10 {
	// 		break
	// 	}
	// }
	fmt.Println("--")

	for i := uint64(0); i < mc; i++ {
		mAddr, err := minipool.GetMinipoolAt(rp.API, i, nil)
		if err != nil {
			panic(err)
		}
		md, err := minipool.GetMinipoolDetails(rp.API, mAddr, nil)
		if err != nil {
			panic(err)
		}
		m, err := minipool.NewMinipool(rp.API, mAddr)
		if err != nil {
			panic(err)
		}
		nd, err := m.GetNodeDetails(nil)
		if err != nil {
			panic(err)
		}
		rewardsAmountWei, err := rewards.GetNodeClaimRewardsAmount(rp.API, nd.Address, nil)
		if err != nil {
			panic(err)
		}
		fmt.Println(
			md.Exists,
			nd.Address.Hex(),
			md.Address.Hex(),
			md.Pubkey.Hex(),
			nd.Fee,
			eth.WeiToEth(nd.RefundBalance),
			eth.WeiToEth(nd.DepositBalance),
			nd.DepositAssigned,
			eth.WeiToEth(rewardsAmountWei),
		)
		if i > 10 {
			break
		}
	}
}

func (rp *RocketpoolExporter) Test2() {
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
	// n, err := minipool.GetNodeMinipoolAddresses(rp.API, mpd.Address, nil)
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
