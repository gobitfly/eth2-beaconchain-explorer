package types

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
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

type CreateNodeJobUserError struct {
	Message string
}

func (e CreateNodeJobUserError) Error() string {
	return e.Message
}

type NodeJobValidatorInfo struct {
	ValidatorIndex      uint64 `db:"validatorindex"`
	PublicKey           []byte `db:"pubkey"`
	WithdrawCredentials []byte `db:"withdrawalcredentials"`
	ExitEpoch           uint64 `db:"exitepoch"`
	Status              string
}

// ParseData will try to unmarshal NodeJob.RawData into NodeJob.Data and determine NodeJob.Type by doing so. If it is not able to unmarshal any type it will return an error. It will sanitize NodeJob.RawData on success.
func (nj *NodeJob) ParseData() error {
	if len(nj.RawData) == 0 {
		return CreateNodeJobUserError{Message: "data is empty"}
	}
	{
		d := []*capella.SignedBLSToExecutionChange{}
		err := json.Unmarshal(nj.RawData, &d)
		if err == nil {
			if nj.Type != "" && nj.Type != UnknownNodeJobType && nj.Type != BLSToExecutionChangesNodeJobType {
				return fmt.Errorf("nodejob.RawData mismatches nodejob.Type (%v)", nj.Type)
			}
			sort.Slice(d, func(i, j int) bool {
				return d[i].Message.ValidatorIndex < d[j].Message.ValidatorIndex
			})
			nj.Type = BLSToExecutionChangesNodeJobType
			nj.Data = d
			return nj.SanitizeRawData()
		}
	}
	{
		//var d *VoluntaryExitsNodeJobData
		var d *phase0.SignedVoluntaryExit
		err := json.Unmarshal(nj.RawData, &d)
		if err == nil {
			if nj.Type != "" && nj.Type != UnknownNodeJobType && nj.Type != VoluntaryExitsNodeJobType {
				return fmt.Errorf("nodejob.RawData mismatches nodejob.Type (%v)", nj.Type)
			}
			nj.Type = VoluntaryExitsNodeJobType
			nj.Data = d
			return nj.SanitizeRawData()
		}
	}
	return CreateNodeJobUserError{Message: "can not unmarshal data: invalid json"}
}

func (nj *NodeJob) SanitizeRawData() error {
	d, err := json.Marshal(nj.Data)
	if err != nil {
		return err
	}
	nj.RawData = d
	return nil
}

func (nj NodeJob) GetBLSToExecutionChangesNodeJobData() ([]*capella.SignedBLSToExecutionChange, bool) {
	d, ok := nj.Data.([]*capella.SignedBLSToExecutionChange)
	return d, ok
}

func (nj NodeJob) GetVoluntaryExitsNodeJobData() (*phase0.SignedVoluntaryExit, bool) {
	d, ok := nj.Data.(*phase0.SignedVoluntaryExit)
	return d, ok
}
