package db

import (
	"bytes"
	"encoding/json"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"sync"
	"time"

	capella "github.com/attestantio/go-eth2-client/spec/capella"
	phase0 "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	ssz "github.com/prysmaticlabs/go-ssz"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func UpdateNodeJobs(elEndpoint, clEndpoint string) error {
	var err error
	err = UpdateBLSToExecutionChangesNodeJobs(elEndpoint, clEndpoint)
	if err != nil {
		return err
	}
	err = UpdateVoluntaryExitNodeJobs(elEndpoint, clEndpoint)
	if err != nil {
		return err
	}
	return nil
}

func SubmitNodeJobs(elEndpoint, clEndpoint string) error {
	var err error
	err = SubmitBLSToExecutionChangesNodeJobs(elEndpoint, clEndpoint)
	if err != nil {
		return err
	}
	err = SubmitVoluntaryExitNodeJobs(elEndpoint, clEndpoint)
	if err != nil {
		return err
	}
	return nil
}

func CreateBLSToExecutionChangesNodeJob(data []byte) (*types.BLSToExecutionChangesNodeJob, error) {

	nj := &types.BLSToExecutionChangesNodeJob{}
	nj.Info = &types.NodeJobInfo{
		ID:     uuid.New().String(),
		Type:   types.BLSToExecutionChangesNodeJobType,
		Status: types.PendingNodeJobStatus,
	}
	err := json.Unmarshal(data, &nj.Data)
	if err != nil {
		return nil, err
	}

	opsByIndex := map[uint64]*capella.SignedBLSToExecutionChange{}
	opsToCheck := map[uint64]bool{}
	indicesArr := []uint64{}
	for _, op := range nj.Data {
		// #TODO(patrick) fix verifyBlsToExecutionChangeSignature
		err = verifyBlsToExecutionChangeSignature(op)
		if err != nil {
			return nil, err
		}
		indicesArr = append(indicesArr, uint64(op.Message.ValidatorIndex))
		opsByIndex[uint64(op.Message.ValidatorIndex)] = op
		opsToCheck[uint64(op.Message.ValidatorIndex)] = true
	}

	dbValis := []struct {
		Index                 uint64 `db:"validatorindex"`
		Pubkey                []byte `db:"pubkey"`
		WithdrawalCredentials []byte `db:"withdrawalcredentials"`
	}{}
	err = WriterDb.Select(&dbValis, `select validatorindex, pubkey, withdrawalcredentials from validators where validatorindex = any($1)`, pq.Array(indicesArr))
	if err != nil {
		return nil, err
	}

	for _, v := range dbValis {
		if v.WithdrawalCredentials[0] == 0x01 {
			return nil, fmt.Errorf("withdrawal credentials for validator %v were already changed, please remove this validator from the batch and try again ", v.Index)
		}
		op := opsByIndex[v.Index]

		withdrawalCredentials := utils.SHA256(op.Message.FromBLSPubkey[:])
		withdrawalCredentials[0] = byte(0)
		if !bytes.Equal(withdrawalCredentials, v.WithdrawalCredentials) {
			return nil, fmt.Errorf("message.FromBLSPubkey != validator.WithdrawalCredentials for validator with index %v", v.Index)
		}
		if v.WithdrawalCredentials[0] != 0 {
			return nil, fmt.Errorf("validator.WithdrawalCredentials[0] != 0 for validator with index %v", v.Index)
		}
		delete(opsToCheck, v.Index)
	}
	if len(opsToCheck) > 0 {
		return nil, fmt.Errorf("could not check all validators")
	}

	dataEncoded, err := json.Marshal(nj.Data)
	if err != nil {
		return nil, err
	}

	_, err = WriterDb.Exec(`insert into node_jobs (id, type, status, data) values ($1, $2, $3, $4)`, nj.Info.ID, nj.Info.Type, nj.Info.Status, dataEncoded)
	if err != nil {
		return nil, err
	}
	return nj, nil
}

func GetNodeJob(jobId string) (*types.NodeJobInfo, error) {
	logger.Info(jobId)
	nji := []*types.NodeJobInfo{}
	err := WriterDb.Select(&nji, "SELECT id, type, status FROM node_jobs WHERE id = $1 LIMIT 1", jobId)

	if err != nil {
		return nil, err
	}

	if len(nji) == 0 {
		return nil, fmt.Errorf("job not found")
	}
	return nji[0], err
}

func UpdateBLSToExecutionChangesNodeJobs(elEndpoint, clEndpoint string) error {
	jobs := []*types.BLSToExecutionChangesNodeJob{}
	jobType := types.BLSToExecutionChangesNodeJobType
	jobStatus := types.SubmittedToNodeNodeJobStatus
	err := WriterDb.Select(&jobs, `select id, type, status, data from node_jobs where type = $1 and status = $2`, jobType, jobStatus)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		err = UpdateBLSToExecutionChangesNodeJob(job)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateBLSToExecutionChangesNodeJob(job *types.BLSToExecutionChangesNodeJob) error {
	toCheck := map[uint64]bool{}
	indicesArr := []uint64{}
	for _, op := range job.Data {
		indicesArr = append(indicesArr, uint64(op.Message.ValidatorIndex))
		toCheck[uint64(op.Message.ValidatorIndex)] = true
	}
	dbValis := []struct {
		Index                 uint64 `db:"validatorindex"`
		WithdrawalCredentials []byte `db:"withdrawalcredentials"`
	}{}
	err := WriterDb.Select(&dbValis, `select validatorindex, withdrawalcredentials from validators where validatorindex any($1)`, pq.Array(indicesArr))
	if err != nil {
		return err
	}

	for _, v := range dbValis {
		if v.WithdrawalCredentials[0] == 1 {
			delete(toCheck, v.Index)
		} else {
			// not all valis have been updated yet
			return nil
		}
		if len(toCheck) == 0 {
			// all valis have been updated
			_, err = WriterDb.Exec(`update node_jobs set status = $1 where id = $2`, types.CompletedNodeJobStatus, job.GetInfo().ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func SubmitBLSToExecutionChangesNodeJobs(elEndpoint, clEndpoint string) error {
	jobs := []*types.BLSToExecutionChangesNodeJob{}
	jobType := types.BLSToExecutionChangesNodeJobType
	err := WriterDb.Select(&jobs, `select id, type, status, data from node_jobs where type = $1 and status = $2 limit 10-(select count(*) from node_jobs where type = $1 and status = $3)`, jobType, types.PendingNodeJobStatus, types.SubmittedToNodeNodeJobStatus)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		err := SubmitBLSToExecutionChangesNodeJob(job, clEndpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

func SubmitBLSToExecutionChangesNodeJob(job *types.BLSToExecutionChangesNodeJob, clEndpoint string) error {
	data, err := json.Marshal(job.GetData())
	if err != nil {
		return err
	}
	if true {
		fmt.Printf("DEBUG: not sending bls_to_execution_change because debugging: %+v\n", job)
		return nil
	}
	client := &http.Client{Timeout: time.Second * 10}
	url := fmt.Sprintf("%s/eth/v1/beacon/pool/bls_to_execution_changes", clEndpoint)
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http request error: %s", resp.Status)
	}
	_, err = WriterDb.Exec(`update node_jobs set status = $1 where id = $2`, types.SubmittedToNodeNodeJobStatus, job.GetInfo().ID)
	if err != nil {
		return err
	}
	return nil
}

// verifyBlsToExecutionChangeSignature verifies the signature of an bls_to_execution_change message
// #TODO(patrick) fix verifyBlsToExecutionChangeSignature
// see: https://github.com/wealdtech/ethdo/blob/master/cmd/validator/credentials/set/process.go
// see: https://github.com/prysmaticlabs/prysm/blob/76ed634f7386609f0d1ee47b703eb0143c995464/beacon-chain/core/blocks/withdrawals.go
var BLSInitialized = sync.Once{}

func verifyBlsToExecutionChangeSignature(op *capella.SignedBLSToExecutionChange) error {

	BLSInitialized.Do(func() {
		e2types.InitBLS() //see https://github.com/wealdtech/go-eth2-types/blob/master/README.md?plain=1#L31
	})

	network := "zhejiang"

	genesisForkVersion := phase0.Version{}
	genesisValidatorsRoot := phase0.Root{}
	domainBLSToExecutionChange := utils.MustParseHex("0x0A000000")
	var forkDataRoot [32]byte

	switch network {
	case "mainnet":
		copy(genesisForkVersion[:], utils.MustParseHex("0x00000000"))
		copy(genesisValidatorsRoot[:], utils.MustParseHex("0x4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95"))
	case "prater":
		copy(genesisForkVersion[:], utils.MustParseHex("0x00000069"))
		copy(genesisValidatorsRoot[:], utils.MustParseHex("0x043db0d9a83813551ee2f33450d23797757d430911a9320530ad8a0eabc43efb"))
	case "sepolia":
		copy(genesisForkVersion[:], utils.MustParseHex("0x00000069"))
		copy(genesisValidatorsRoot[:], utils.MustParseHex("0xd8ea171f3c94aea21ebc42a1ed61052acf3f9209c00e4efbaaddac09ed9b8078"))
	case "zhejiang":
		copy(genesisForkVersion[:], utils.MustParseHex("0x00000069"))
		copy(genesisValidatorsRoot[:], utils.MustParseHex("0x53a92d8f2bb1d85f62d16a156e6ebcd1bcaba652d0900b2c2f387826f3481f6f"))
	default:
		return fmt.Errorf("invalid network: %v", network)
	}

	r, err := (&phase0.ForkData{
		CurrentVersion:        genesisForkVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}).HashTreeRoot()
	if err != nil {
		return err
	}

	forkDataRoot = r
	domain := phase0.Domain{}
	copy(domain[:], domainBLSToExecutionChange[:])
	copy(domain[4:], forkDataRoot[:])

	root, err := op.Message.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "failed to generate message root")
	}

	sigBytes := make([]byte, len(op.Signature))
	copy(sigBytes, op.Signature[:])

	sig, err := e2types.BLSSignatureFromBytes(sigBytes)
	if err != nil {
		return errors.Wrap(err, "invalid signature")
	}

	container := &phase0.SigningData{
		ObjectRoot: root,
		Domain:     domain,
	}
	signingRoot, err := ssz.HashTreeRoot(container)
	if err != nil {
		return errors.Wrap(err, "failed to generate signing root")
	}

	pubkeyBytes := make([]byte, len(op.Message.FromBLSPubkey))
	copy(pubkeyBytes, op.Message.FromBLSPubkey[:])
	pubkey, err := e2types.BLSPublicKeyFromBytes(pubkeyBytes)
	if err != nil {
		return errors.Wrap(err, "invalid public key")
	}
	if !sig.Verify(signingRoot[:], pubkey) {
		return errors.New("signature does not verify")
	}

	return nil
}

func CreateVoluntaryExitNodeJob(data []byte) (*types.VoluntaryExitsNodeJob, error) {
	nj := &types.VoluntaryExitsNodeJob{}
	nj.Info = &types.NodeJobInfo{
		ID:     uuid.New().String(),
		Type:   types.VoluntaryExitsNodeJobType,
		Status: types.PendingNodeJobStatus,
	}

	err := json.Unmarshal(data, &nj.Data)
	if err != nil {
		return nil, err
	}

	d, err := json.Marshal(nj.Data)
	if err != nil {
		return nil, err
	}

	_, err = WriterDb.Exec(`insert into node_jobs (id, type, status, data) values ($1, $2, $3, $4)`, nj.Info.ID, nj.Info.Type, nj.Info.Status, d)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func UpdateVoluntaryExitNodeJobs(elEndpoint, clEndpoint string) error {
	return nil
}

func SubmitVoluntaryExitNodeJobs(elEndpoint, clEndpoint string) error {
	return nil
}
