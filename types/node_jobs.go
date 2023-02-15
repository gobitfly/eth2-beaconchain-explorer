package types

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

type NodeJobInfo struct {
	ID            string        `db:"id"`
	CreatedTime   time.Time     `db:"created_time"`
	SubmittedTime sql.NullTime  `db:"submitted_time"`
	CompletedTime sql.NullTime  `db:"completed_time"`
	Type          NodeJobType   `db:"type"`
	Status        NodeJobStatus `db:"status"`
}

type NodeJob interface {
	GetInfo() *NodeJobInfo
	GetData() interface{}
}

type RawNodeJob struct {
	Info    *NodeJobInfo
	RawData []byte `db:"data"`
}

func (nj RawNodeJob) ToNodeJob() (NodeJob, error) {
	switch nj.Info.Type {
	case BLSToExecutionChangesNodeJobType:
		jj := &BLSToExecutionChangesNodeJob{
			Info: nj.Info,
		}
		err := json.Unmarshal(nj.RawData, &jj.Data)
		return jj, err
	default:
		return nil, fmt.Errorf("unknown job-type %v", nj.Info.Type)
	}
}

type BLSToExecutionChangesNodeJob struct {
	Info *NodeJobInfo
	Data []*capella.SignedBLSToExecutionChange `db:"data"`
}

func (nj BLSToExecutionChangesNodeJob) GetInfo() *NodeJobInfo {
	return nj.Info
}

func (nj BLSToExecutionChangesNodeJob) GetData() interface{} {
	return nj.Data
}

type VoluntaryExitsNodeJob struct {
	Info *NodeJobInfo
	Data *phase0.VoluntaryExit `db:"data"`
}

func (nj VoluntaryExitsNodeJob) GetInfo() *NodeJobInfo {
	return nj.Info
}

func (nj VoluntaryExitsNodeJob) GetData() interface{} {
	return nj.Data
}
