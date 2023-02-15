package db

import (
	"bytes"
	"encoding/json"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"time"

	capella "github.com/attestantio/go-eth2-client/spec/capella"
	phase0 "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	ssz "github.com/prysmaticlabs/go-ssz"
	"github.com/sirupsen/logrus"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func init() {
	err := e2types.InitBLS()
	if err != nil {
		logrus.Fatalf("error in e2types.InitBLS(): %v", err)
	}
}

func GetNodeJob(id string) (types.NodeJob, error) {
	if len(id) > 40 {
		return nil, fmt.Errorf("invalid id")
	}
	var rawJob types.RawNodeJob
	err := WriterDb.Get(&rawJob, `select id, type, status, created_time, submitted_to_node_time, completed_time, data from node_jobs where id = $1`, id)
	if err != nil {
		return nil, err
	}
	return rawJob.ToNodeJob()
}

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
	id := uuid.New().String()
	t := types.BLSToExecutionChangesNodeJobType
	status := types.PendingNodeJobStatus
	nj := &types.BLSToExecutionChangesNodeJob{}
	nj.Info = &types.NodeJobInfo{
		ID:     id,
		Type:   t,
		Status: status,
	}
	err := json.Unmarshal(data, &nj.Data)
	if err != nil {
		return nil, err
	}

	opsByIndex := map[uint64]*capella.SignedBLSToExecutionChange{}
	opsToCheck := map[uint64]bool{}
	indicesArr := []uint64{}
	for _, op := range nj.Data {
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
	err = WriterDb.Select(&dbValis, `select validatorindex, pubkey, withdrawalcredentials from validators where validatorindex any($1)`, pq.Array(indicesArr))
	if err != nil {
		return nil, err
	}

	for _, v := range dbValis {
		op := opsByIndex[v.Index]
		if !bytes.Equal(op.Message.FromBLSPubkey[:], v.WithdrawalCredentials) {
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

	d, err := json.Marshal(nj.Data)
	if err != nil {
		return nil, err
	}

	_, err = WriterDb.Exec(`insert into node_jobs (id, type, status, data) values ($1, $2, $3, $4)`, id, t, status, d)
	if err != nil {
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"id": nj.Info.ID, "type": nj.Info.Type}).Infof("created job")
	return nj, nil
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
			// all validatrors have been completed
			_, err = WriterDb.Exec(`update node_jobs set status = $1 where id = $2`, types.CompletedNodeJobStatus, job.GetInfo().ID)
			if err != nil {
				return err
			}
			logrus.WithFields(logrus.Fields{"id": job.Info.ID, "type": job.Info.Type, "status": types.CompletedNodeJobStatus}).Infof("updated job")
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
	logrus.WithFields(logrus.Fields{"id": job.Info.ID, "type": job.Info.Type}).Infof("submitted job")
	return nil
}

// verifyBlsToExecutionChangeSignature verifies the signature of an bls_to_execution_change message
// see: https://github.com/wealdtech/ethdo/blob/master/cmd/validator/credentials/set/process.go
// see: https://github.com/prysmaticlabs/prysm/blob/76ed634f7386609f0d1ee47b703eb0143c995464/beacon-chain/core/blocks/withdrawals.go
func verifyBlsToExecutionChangeSignature(op *capella.SignedBLSToExecutionChange) error {
	genesisForkVersion := phase0.Version{}
	genesisValidatorsRoot := phase0.Root{}
	copy(genesisForkVersion[:], utils.MustParseHex(utils.Config.Chain.Config.GenesisForkVersion))
	copy(genesisForkVersion[:], utils.MustParseHex(utils.Config.Chain.GenesisValidatorsRoot))

	forkDataRoot, err := (&phase0.ForkData{
		CurrentVersion:        genesisForkVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}).HashTreeRoot()
	if err != nil {
		return err
	}

	domain := phase0.Domain{}
	domainBLSToExecutionChange := utils.MustParseHex(utils.Config.Chain.DomainBLSToExecutionChange)
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
	id := uuid.New().String()
	t := types.VoluntaryExitsNodeJobType
	status := types.PendingNodeJobStatus
	nj := &types.VoluntaryExitsNodeJob{}
	nj.Info = &types.NodeJobInfo{
		ID:     id,
		Type:   t,
		Status: status,
	}

	err := json.Unmarshal(data, &nj.Data)
	if err != nil {
		return nil, err
	}

	d, err := json.Marshal(nj.Data)
	if err != nil {
		return nil, err
	}

	_, err = WriterDb.Exec(`insert into node_jobs (id, type, status, data) values ($1, $2, $3, $4)`, id, t, status, d)
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
