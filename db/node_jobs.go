package db

import (
	"bytes"
	"database/sql"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/attestantio/go-eth2-client/spec/capella"
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
		for _, op := range jobData {
			indicesArr = append(indicesArr, uint64(op.Message.ValidatorIndex))
		}
	} else if job.Type == types.VoluntaryExitsNodeJobType {
		jobData, ok := job.GetVoluntaryExitsNodeJobData()
		if !ok {
			return nil, fmt.Errorf("invalid voluntary exit job-data")
		}

		indicesArr = append(indicesArr, uint64(jobData.Message.ValidatorIndex))
	} else {
		return []types.NodeJobValidatorInfo{}, nil
	}

	dbValis := []types.NodeJobValidatorInfo{}
	err := WriterDb.Select(&dbValis, `select validatorindex, pubkey, withdrawalcredentials, exitepoch, status from validators where validatorindex = any($1)`, pq.Array(indicesArr))
	if err != nil {
		return nil, err
	}
	jobStatus := "Pending"
	switch job.Status {
	case types.SubmittedToNodeNodeJobStatus:
		jobStatus = "Submitted to node"
	case types.CompletedNodeJobStatus:
		jobStatus = "Processed"
	case types.FailedNodeJobStatus:
		jobStatus = "Failed"
	}
	for i, info := range dbValis {
		status := jobStatus
		if strings.HasPrefix(info.Status, "exit") && job.Type == types.BLSToExecutionChangesNodeJobType {
			status = fmt.Sprintf("%s (Validator Status: Exited)", status)
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
		return fmt.Errorf("error updating bls-to-exec-job: %w", err)
	}
	err = UpdateVoluntaryExitNodeJobs()
	if err != nil {
		return fmt.Errorf("error updating voluntary-exit-job: %w", err)
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
		return nil, types.CreateNodeJobUserError{Message: "data-size exceeds maximum of 1MB"}
	}
	nj.ID = uuid.New().String()
	nj.Status = types.PendingNodeJobStatus

	opsByIndex := map[uint64]*capella.SignedBLSToExecutionChange{}
	opsToCheck := map[uint64]bool{}
	indicesArr := []uint64{}
	d, ok := nj.GetBLSToExecutionChangesNodeJobData()
	if !ok {
		return nil, types.CreateNodeJobUserError{Message: "invalid data"}
	}

	for _, op := range d {
		err := utils.VerifyBlsToExecutionChangeSignature(op)
		if err != nil {
			return nil, types.CreateNodeJobUserError{Message: fmt.Sprintf("can not verify signature: %v", err)}
		}
		_, exists := opsByIndex[uint64(op.Message.ValidatorIndex)]
		if exists {
			return nil, types.CreateNodeJobUserError{Message: fmt.Sprintf("multiple entries for the same validator: %v", uint64(op.Message.ValidatorIndex))}
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
			return nil, types.CreateNodeJobUserError{Message: fmt.Sprintf("fromBLSPubkey do not match withdrawalCredentials for validator with index %v", v.Index)}
		}
		if v.WithdrawalCredentials[0] != 0 {
			return nil, types.CreateNodeJobUserError{Message: fmt.Sprintf("withdrawalCredentials[0] != 0 for validator with index %v", v.Index)}
		}
		delete(opsToCheck, v.Index)
	}
	if len(opsToCheck) > 0 {
		return nil, fmt.Errorf("could not check all validators")
	}

	tx, err := WriterDb.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error starting db transactions: %w", err)
	}
	defer tx.Rollback()

	batchSize := 1000
	for b := 0; b < len(indicesArr); b += batchSize {
		start := b
		end := b + batchSize
		if len(indicesArr) < end {
			end = len(indicesArr)
		}
		size := end - start
		valueStrs := make([]string, 0, size)
		valueArgs := make([]interface{}, 0, size+1)
		valueArgs = append(valueArgs, nj.ID)
		for i, idx := range indicesArr[start:end] {
			valueStrs = append(valueStrs, fmt.Sprintf("($1, $%d)", i+2))
			valueArgs = append(valueArgs, idx)
		}
		stmt := fmt.Sprintf(`insert into node_jobs_bls_changes_validators (node_job_id, validatorindex) values %s on conflict do nothing`, strings.Join(valueStrs, ","))
		res, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return nil, fmt.Errorf("error inserting into node_jobs_bls_changes_validators: %w", err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("error getting rowsAffected: %w", err)
		}
		if rows != int64(size) {
			return nil, types.CreateNodeJobUserError{Message: "there is already a job for some of the validators"}
		}
	}

	_, err = tx.Exec(`insert into node_jobs (id, type, status, data, created_time) values ($1, $2, $3, $4, now())`, nj.ID, nj.Type, nj.Status, nj.RawData)
	if err != nil {
		return nil, fmt.Errorf("error inserting into node_jobs: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error commiting db-tx: %w", err)
	}

	logrus.WithFields(logrus.Fields{"id": nj.ID, "type": nj.Type, "validators": len(indicesArr)}).Infof("created node_job")
	return nj, nil
}

func UpdateBLSToExecutionChangesNodeJobs() error {
	jobs := []*types.NodeJob{}
	err := WriterDb.Select(&jobs, `select id, type, status, created_time, submitted_to_node_time, completed_time, data from node_jobs where type = $1 and status = $2`, types.BLSToExecutionChangesNodeJobType, types.SubmittedToNodeNodeJobStatus)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		err := job.ParseData()
		if err != nil {
			return err
		}
		err = UpdateBLSToExecutionChangesNodeJob(job)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateBLSToExecutionChangesNodeJob(job *types.NodeJob) error {
	jobData, ok := job.GetBLSToExecutionChangesNodeJobData()
	if !ok {
		return fmt.Errorf("invalid job-data")
	}
	toCheck := map[uint64]bool{}
	indicesArr := []uint64{}
	for _, op := range jobData {
		indicesArr = append(indicesArr, uint64(op.Message.ValidatorIndex))
		toCheck[uint64(op.Message.ValidatorIndex)] = true
	}
	dbValis := []struct {
		Index                 uint64 `db:"validatorindex"`
		WithdrawalCredentials []byte `db:"withdrawalcredentials"`
	}{}
	logrus.WithFields(logrus.Fields{"id": job.ID, "type": job.Type, "status": job.Status, "validators": len(indicesArr)}).Infof("checking node_job")
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
			job.CompletedTime.Time = time.Now()
			job.CompletedTime.Valid = true
			_, err = WriterDb.Exec(`update node_jobs set status = $1, completed_time = $2 where id = $3`, types.CompletedNodeJobStatus, job.CompletedTime.Time, job.ID)
			if err != nil {
				return err
			}
			logrus.WithFields(logrus.Fields{"id": job.ID, "type": job.Type, "status": types.CompletedNodeJobStatus, "validators": len(indicesArr)}).Infof("updated node_job")
		}
	}
	return nil
}

func SubmitBLSToExecutionChangesNodeJobs() error {
	maxSubmittedJobs := 1000
	jobs := []*types.NodeJob{}
	err := WriterDb.Select(&jobs, `select id, type, status, created_time, submitted_to_node_time, completed_time, data from node_jobs where type = $1 and status = $2 order by created_time limit $4-(select count(*) from node_jobs where type = $1 and status = $3)`, types.BLSToExecutionChangesNodeJobType, types.PendingNodeJobStatus, types.SubmittedToNodeNodeJobStatus, maxSubmittedJobs)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		err = job.ParseData()
		if err != nil {
			return err
		}
		err = SubmitBLSToExecutionChangesNodeJob(job)
		if err != nil {
			return fmt.Errorf("error calling SubmitBLSToExecutionChangesNodeJob for job %v: %w", job.ID, err)
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
	jobStatus := types.SubmittedToNodeNodeJobStatus
	if resp.StatusCode != 200 {
		d, _ := io.ReadAll(resp.Body)
		if len(d) > 1000 {
			d = d[:1000]
		}
		jobStatus = types.FailedNodeJobStatus
		logrus.WithFields(logrus.Fields{"data": string(d), "status": resp.Status, "jobID": job.ID}).Warnf("failed submitting a job")
	}
	job.Status = jobStatus
	job.SubmittedToNodeTime.Time = time.Now()
	job.SubmittedToNodeTime.Valid = true
	_, err = WriterDb.Exec(`update node_jobs set status = $1, submitted_to_node_time = $2 where id = $3`, job.Status, job.SubmittedToNodeTime.Time, job.ID)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{"id": job.ID, "type": job.Type, "status": jobStatus}).Infof("submitted node_job")
	return nil
}

func CreateVoluntaryExitNodeJob(nj *types.NodeJob) (*types.NodeJob, error) {
	if len(nj.RawData) > 5e3 {
		return nil, fmt.Errorf("data size exceeds maximum")
	}
	nj.ID = uuid.New().String()
	nj.Status = types.PendingNodeJobStatus

	njd, ok := nj.GetVoluntaryExitsNodeJobData()
	if !ok {
		return nil, fmt.Errorf("invalid job")
	}

	vali := struct {
		Pubkey []byte `db:"pubkey"`
		Status string `db:"status"`
	}{}
	err := WriterDb.Get(&vali, `select pubkey, status from validators where validatorindex = $1`, njd.Message.ValidatorIndex)
	if err != nil {
		return nil, err
	}

	switch vali.Status {
	case "exited", "exiting_online", "exiting_offline":
		return nil, fmt.Errorf("validator has exited")
	case "slashed", "slashing_offline", "slashing_online":
		return nil, fmt.Errorf("validator has been slashed")
	default:
	}

	forkVersion := utils.ForkVersionAtEpoch(uint64(njd.Message.Epoch))
	err = utils.VerifyVoluntaryExitSignature(njd, forkVersion.CurrentVersion, vali.Pubkey)
	if err != nil {
		return nil, err
	}

	_, err = WriterDb.Exec(`insert into node_jobs (id, type, status, data, created_time) values ($1, $2, $3, $4, now())`, nj.ID, nj.Type, nj.Status, nj.RawData)
	if err != nil {
		return nil, err
	}
	logrus.WithFields(logrus.Fields{"id": nj.ID, "type": nj.Type}).Infof("created node_job")
	return nj, nil
}

func UpdateVoluntaryExitNodeJobs() error {
	jobs := []*types.NodeJob{}
	err := WriterDb.Select(&jobs, `select id, type, status, created_time, submitted_to_node_time, completed_time, data from node_jobs where type = $1 and status = $2`, types.VoluntaryExitsNodeJobType, types.SubmittedToNodeNodeJobStatus)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		err := job.ParseData()
		if err != nil {
			return err
		}
		err = UpdateVoluntaryExitNodeJob(job)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateVoluntaryExitNodeJob(job *types.NodeJob) error {
	jobData, ok := job.GetVoluntaryExitsNodeJobData()
	if !ok {
		return fmt.Errorf("invalid job-data")
	}
	dbResult := struct {
		Status string `db:"status"`
	}{}
	err := WriterDb.Get(&dbResult, `select status from validators where validatorindex = $1`, uint64(jobData.Message.ValidatorIndex))
	if err == sql.ErrNoRows {
		return fmt.Errorf("validator not found")
	}
	if err != nil {
		return err
	}
	if strings.HasPrefix(dbResult.Status, "exit") {
		job.Status = types.CompletedNodeJobStatus
		job.CompletedTime.Time = time.Now()
		job.CompletedTime.Valid = true
		_, err = WriterDb.Exec(`update node_jobs set status = $1, completed_time = $2 where id = $3`, job.Status, job.CompletedTime.Time, job.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func SubmitVoluntaryExitNodeJobs() error {
	maxSubmittedJobs := 100
	jobs := []*types.NodeJob{}
	err := WriterDb.Select(&jobs, `select id, type, status, created_time, submitted_to_node_time, completed_time, data from node_jobs where type = $1 and status = $2 order by created_time limit $4-(select count(*) from node_jobs where type = $1 and status = $3)`, types.VoluntaryExitsNodeJobType, types.PendingNodeJobStatus, types.SubmittedToNodeNodeJobStatus, maxSubmittedJobs)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		err = job.ParseData()
		if err != nil {
			return err
		}
		err = SubmitVoluntaryExitNodeJob(job)
		if err != nil {
			return err
		}
	}
	return nil
}

func SubmitVoluntaryExitNodeJob(job *types.NodeJob) error {
	client := &http.Client{Timeout: time.Second * 10}
	url := fmt.Sprintf("%s/eth/v1/beacon/pool/voluntary_exits", utils.Config.NodeJobsProcessor.ClEndpoint)
	resp, err := client.Post(url, "application/json", bytes.NewReader(job.RawData))
	if err != nil {
		return err
	}
	jobStatus := types.SubmittedToNodeNodeJobStatus
	if resp.StatusCode != 200 {
		d, _ := io.ReadAll(resp.Body)
		if len(d) > 1000 {
			d = d[:1000]
		}
		jobStatus = types.FailedNodeJobStatus
		logrus.WithFields(logrus.Fields{"res": string(d), "status": resp.Status, "jobID": job.ID, "jobType": job.Type}).Warnf("failed submitting a job")
	}
	job.Status = jobStatus
	job.SubmittedToNodeTime.Time = time.Now()
	job.SubmittedToNodeTime.Valid = true
	_, err = WriterDb.Exec(`update node_jobs set status = $1, submitted_to_node_time = $2 where id = $3`, job.Status, job.SubmittedToNodeTime.Time, job.ID)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{"id": job.ID, "type": job.Type}).Infof("submitted node_job")
	return nil
}
