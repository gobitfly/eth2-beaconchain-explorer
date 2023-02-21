package types

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type NodeJobStatus string

const PendingNodeJobStatus NodeJobStatus = "PENDING"                   // job is waiting to be submitted
const SubmittedToNodeNodeJobStatus NodeJobStatus = "SUBMITTED_TO_NODE" // job has been submitted successfully
const CompletedNodeJobStatus NodeJobStatus = "COMPLETED"               // job has been submitted successfully and result is visible on chain
const FailedNodeJobStatus NodeJobStatus = "FAILED"                     // job has been submitted successfully but something went wrong

type NodeJobType string

const BLSToExecutionChangesNodeJobType NodeJobType = "BLS_TO_EXECUTION_CHANGES"
const VoluntaryExitsNodeJobType NodeJobType = "VOLUNTARY_EXITS"

var NodeJobTypes = []NodeJobType{
	BLSToExecutionChangesNodeJobType,
	VoluntaryExitsNodeJobType,
}

type NodeJob struct {
	ID                  string        `db:"id"`
	CreatedTime         time.Time     `db:"created_time"`
	SubmittedToNodeTime sql.NullTime  `db:"submitted_to_node_time"`
	CompletedTime       sql.NullTime  `db:"completed_time"`
	Type                NodeJobType   `db:"type"`
	Status              NodeJobStatus `db:"status"`
	Data                []byte        `db:"data"`
}

func (nj NodeJob) ToBLSToExecutionChangesNodeJob() (*BLSToExecutionChangesNodeJob, error) {
	j := &BLSToExecutionChangesNodeJob{}
	j.ID = nj.ID
	j.CreatedTime = nj.CreatedTime
	j.SubmittedToNodeTime = nj.SubmittedToNodeTime
	j.CompletedTime = nj.CompletedTime
	j.Type = nj.Type
	j.Status = nj.Status
	err := json.Unmarshal(nj.Data, &j.Data)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (nj NodeJob) ToVoluntaryExitsNodeJob() (*VoluntaryExitsNodeJob, error) {
	j := &VoluntaryExitsNodeJob{}
	j.ID = nj.ID
	j.CreatedTime = nj.CreatedTime
	j.SubmittedToNodeTime = nj.SubmittedToNodeTime
	j.CompletedTime = nj.CompletedTime
	j.Type = nj.Type
	j.Status = nj.Status
	err := json.Unmarshal(nj.Data, &j.Data)
	if err != nil {
		return nil, err
	}
	return j, nil
}

type BLSToExecutionChangesNodeJob struct {
	NodeJob
	Data []*capella.SignedBLSToExecutionChange `db:"data,json"`
}

type VoluntaryExitsNodeJob struct {
	NodeJob
	Data *phase0.VoluntaryExit `db:"data,json"`
}
