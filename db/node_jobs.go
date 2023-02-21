package db

import (
	"bytes"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	capella "github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	ethutil "github.com/wealdtech/go-eth2-util"
)

func GetNodeJob(id string) (*types.NodeJob, error) {
	if len(id) > 40 {
		return nil, fmt.Errorf("invalid id")
	}
	job := types.NodeJob{}
	err := WriterDb.Get(&job, `select id, type, status, created_time, submitted_to_node_time, completed_time, data from node_jobs where id = $1`, id)
	if err != nil {
		return nil, err
	}
	err = job.ParseData()
	return &job, err
}

func GetNodeJobValidatorInfos(job *types.NodeJob) ([]types.NodeJobValidatorInfo, error) {

	indicesArr := []uint64{}

	if job.Type == types.BLSToExecutionChangesNodeJobType {
		jobData, ok := job.GetBLSToExecutionChangesNodeJobData()
		if !ok {
			return nil, fmt.Errorf("invalid bls to execution job-data")
		}
		for _, op := range *jobData {
			indicesArr = append(indicesArr, uint64(op.Message.ValidatorIndex))
		}
	} else if job.Type == types.VoluntaryExitsNodeJobType {
		jobData, ok := job.GetVoluntaryExitsNodeJobData()
		if !ok {
			return nil, fmt.Errorf("invalid voluntary exit job-data")
		}

		indicesArr = append(indicesArr, uint64(jobData.ValidatorIndex))
	} else {
		return []types.NodeJobValidatorInfo{}, nil
	}

	dbValis := []types.NodeJobValidatorInfo{}
	err := WriterDb.Select(&dbValis, `select validatorindex, pubkey, withdrawalcredentials, exitepoch, status from validators where validatorindex = any($1)`, pq.Array(indicesArr))
	if err != nil {
		return nil, err
	}
	for i, info := range dbValis {
		status := "-"
		if strings.Contains(info.Status, "exit") {
			status = "Exit"
		} else if job.Status == types.PendingNodeJobStatus {
			status = "Pending"
		} else if info.WithdrawCredentials[0] == 1 {
			status = "Withdrawal credentials set"
		}
		dbValis[i].Status = status
	}
	return dbValis, nil
}

func CreateNodeJob(data []byte) (*types.NodeJob, error) {
	j, err := types.NewNodeJob(data)
	if err != nil {
		return nil, err
	}
	fmt.Printf("db.CreateNodeJob: %+v\n", j)
	switch j.Type {
	default:
		return nil, fmt.Errorf("unknown job-type %v", j.Type)
	case types.BLSToExecutionChangesNodeJobType:
		return CreateBLSToExecutionChangesNodeJob(j)
	case types.VoluntaryExitsNodeJobType:
		return CreateVoluntaryExitNodeJob(j)
	}
}

func UpdateNodeJobs() error {
	var err error
	err = UpdateBLSToExecutionChangesNodeJobs()
	if err != nil {
		return err
	}
	err = UpdateVoluntaryExitNodeJobs()
	if err != nil {
		return err
	}
	return nil
}

func SubmitNodeJobs() error {
	var err error
	err = SubmitBLSToExecutionChangesNodeJobs()
	if err != nil {
		return err
	}
	err = SubmitVoluntaryExitNodeJobs()
	if err != nil {
		return err
	}
	return nil
}

func CreateBLSToExecutionChangesNodeJob(nj *types.NodeJob) (*types.NodeJob, error) {
	if len(nj.RawData) > 1e6 {
		return nil, fmt.Errorf("data size exceeds maximum")
	}
	nj.ID = uuid.New().String()
	nj.Status = types.PendingNodeJobStatus

	opsByIndex := map[uint64]*capella.SignedBLSToExecutionChange{}
	opsToCheck := map[uint64]bool{}
	indicesArr := []uint64{}
	d, _ := nj.GetBLSToExecutionChangesNodeJobData()
	for _, op := range *d {
		err := utils.VerifyBlsToExecutionChangeSignature(op)
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
	err := WriterDb.Select(&dbValis, `select validatorindex, pubkey, withdrawalcredentials from validators where validatorindex = any($1)`, pq.Array(indicesArr))
	if err != nil {
		return nil, err
	}

	for _, v := range dbValis {
		op := opsByIndex[v.Index]
		withdrawalCredentials := ethutil.SHA256(op.Message.FromBLSPubkey[:])
		withdrawalCredentials[0] = byte(0) // BLS_WITHDRAWAL_PREFIX
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

	_, err = WriterDb.Exec(`insert into node_jobs (id, type, status, data, created_time) values ($1, $2, $3, $4, now())`, nj.ID, nj.Type, nj.Status, nj.RawData)
	if err != nil {
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"id": nj.ID, "type": nj.Type}).Infof("created node_job")
	return nj, nil
}

func UpdateBLSToExecutionChangesNodeJobs() error {
	jobs := []*types.NodeJob{}
	err := WriterDb.Select(&jobs, `select id, type, status, created_time, submitted_to_node_time, completed_time, data from node_jobs where type = $1 and status = $2`, types.BLSToExecutionChangesNodeJobType, types.SubmittedToNodeNodeJobStatus)
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

func UpdateBLSToExecutionChangesNodeJob(job *types.NodeJob) error {
	toCheck := map[uint64]bool{}
	indicesArr := []uint64{}
	jobData, ok := job.GetBLSToExecutionChangesNodeJobData()
	if !ok {
		return fmt.Errorf("invalid job-data")
	}
	for _, op := range *jobData {
		indicesArr = append(indicesArr, uint64(op.Message.ValidatorIndex))
		toCheck[uint64(op.Message.ValidatorIndex)] = true
	}
	dbValis := []struct {
		Index                 uint64 `db:"validatorindex"`
		WithdrawalCredentials []byte `db:"withdrawalcredentials"`
	}{}
	logrus.Infof("checking valis %v", indicesArr)
	err := WriterDb.Select(&dbValis, `select validatorindex, withdrawalcredentials from validators where validatorindex = any($1)`, pq.Array(indicesArr))
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
			_, err = WriterDb.Exec(`update node_jobs set status = $1 completed_time = now() where id = $2`, types.CompletedNodeJobStatus, job.ID)
			if err != nil {
				return err
			}
			logrus.WithFields(logrus.Fields{"id": job.ID, "type": job.Type, "status": types.CompletedNodeJobStatus, "validatorIndices": indicesArr}).Infof("updated node_job")
		}
	}
	return nil
}

func SubmitBLSToExecutionChangesNodeJobs() error {
	maxSubmittedJobs := 100
	jobs := []*types.NodeJob{}
	err := WriterDb.Select(&jobs, `select id, type, status, created_time, submitted_to_node_time, completed_time, data from node_jobs where type = $1 and status = $2 limit $4-(select count(*) from node_jobs where type = $1 and status = $3)`, types.BLSToExecutionChangesNodeJobType, types.PendingNodeJobStatus, types.SubmittedToNodeNodeJobStatus, maxSubmittedJobs)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		err = job.ParseData()
		if err != nil {
			return err
		}
		if job.Type != types.BLSToExecutionChangesNodeJobType {
			return fmt.Errorf("job.Type != %v", types.BLSToExecutionChangesNodeJobType)
		}
		err = SubmitBLSToExecutionChangesNodeJob(job)
		if err != nil {
			return err
		}
	}
	return nil
}

func SubmitBLSToExecutionChangesNodeJob(job *types.NodeJob) error {
	client := &http.Client{Timeout: time.Second * 10}
	url := fmt.Sprintf("%s/eth/v1/beacon/pool/bls_to_execution_changes", utils.Config.NodeJobsProcessor.ClEndpoint)
	resp, err := client.Post(url, "application/json", bytes.NewReader(job.RawData))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		d, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("http request error: %s: %s, data: %s", resp.Status, d, job.RawData)
	}
	job.SubmittedToNodeTime.Time = time.Now()
	job.SubmittedToNodeTime.Valid = true
	_, err = WriterDb.Exec(`update node_jobs set status = $1 where id = $2`, types.SubmittedToNodeNodeJobStatus, job.ID)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{"id": job.ID, "type": job.Type}).Infof("submitted node_job")
	return nil
}

func CreateVoluntaryExitNodeJob(job *types.NodeJob) (*types.NodeJob, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func UpdateVoluntaryExitNodeJobs() error {
	return nil
}

func SubmitVoluntaryExitNodeJobs() error {
	return nil
}
