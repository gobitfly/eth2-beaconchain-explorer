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
const UnknownNodeJobType NodeJobType = "UNKNOWN"

var NodeJobTypes = []NodeJobType{
	BLSToExecutionChangesNodeJobType,
	VoluntaryExitsNodeJobType,
}

type BLSToExecutionChangesNodeJobData []*capella.SignedBLSToExecutionChange
type VoluntaryExitsNodeJobData phase0.VoluntaryExit

func NewNodeJob(data []byte) (*NodeJob, error) {
	j := &NodeJob{}
	j.RawData = data
	err := j.ParseData()
	if err != nil {
		return nil, err
	}
	return j, nil
}

type NodeJob struct {
	ID                  string        `db:"id"`
	CreatedTime         time.Time     `db:"created_time"`
	SubmittedToNodeTime sql.NullTime  `db:"submitted_to_node_time"`
	CompletedTime       sql.NullTime  `db:"completed_time"`
	Type                NodeJobType   `db:"type"`
	Status              NodeJobStatus `db:"status"`
	RawData             []byte        `db:"data"`
	Data                interface{}   `db:"-"`
}

type NodeJobValidatorInfo struct {
	ValidatorIndex      uint64 `db:"validatorindex"`
	PublicKey           []byte `db:"pubkey"`
	WithdrawCredentials []byte `db:"withdrawalcredentials"`
	ExitEpoch           uint64 `db:"exitepoch"`
	Status              string
}

func (nj *NodeJob) ParseData() error {
	{
		d := BLSToExecutionChangesNodeJobData{}
		err := json.Unmarshal(nj.RawData, &d)
		if err == nil {
			nj.Type = BLSToExecutionChangesNodeJobType
			nj.Data = d
			return nj.SanitizeRawData()
		}
	}
	{
		d := VoluntaryExitsNodeJobData{}
		err := json.Unmarshal(nj.RawData, &d)
		if err == nil && d.Epoch != 0 {
			nj.Type = VoluntaryExitsNodeJobType
			nj.Data = d
			return nj.SanitizeRawData()
		}
	}
	nj.Type = UnknownNodeJobType
	return fmt.Errorf("invalid data")
}

func (nj *NodeJob) SanitizeRawData() error {
	d, err := json.Marshal(nj.Data)
	if err != nil {
		return err
	}
	nj.RawData = d
	return nil
}

func (nj NodeJob) GetBLSToExecutionChangesNodeJobData() (*BLSToExecutionChangesNodeJobData, bool) {
	d, ok := nj.Data.(BLSToExecutionChangesNodeJobData)
	return &d, ok
}

func (nj NodeJob) GetVoluntaryExitsNodeJobData() (*VoluntaryExitsNodeJobData, bool) {
	d, ok := nj.Data.(VoluntaryExitsNodeJobData)
	return &d, ok
}
