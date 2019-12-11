package cache

import (
	"context"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/sirupsen/logrus"
	"sync"
)

var logger = logrus.New().WithField("module", "cache")

// LRU Cache to store recently used epoch assignments
var assignmentsCache, _ = lru.New(128)
var client ethpb.BeaconChainClient
var assignmentsCacheMux = &sync.Mutex{}

func Init(chainClient ethpb.BeaconChainClient) {
	client = chainClient
}

func GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error) {
	assignmentsCacheMux.Lock()
	defer assignmentsCacheMux.Unlock()

	var err error

	cachedValue, found := assignmentsCache.Get(epoch)
	if found {
		return cachedValue.(*types.EpochAssignments), nil
	}

	assignments := &types.EpochAssignments{
		ProposerAssignments: make(map[uint64]uint64),
		AttestorAssignments: make(map[string]uint64),
	}

	// Retrieve the currently active validator set in order to map public keys to indexes
	validators := make(map[string]uint64)

	validatorBalancesResponse := &ethpb.ValidatorBalances{}
	for {
		validatorBalancesResponse, err = client.ListValidatorBalances(context.Background(), &ethpb.ListValidatorBalancesRequest{PageToken: validatorBalancesResponse.NextPageToken, PageSize: utils.PageSize, QueryFilter: &ethpb.ListValidatorBalancesRequest_Epoch{Epoch: epoch}})
		if err != nil {
			logger.Printf("error retrieving validator balances response: %v", err)
			break
		}
		if validatorBalancesResponse.TotalSize == 0 {
			break
		}

		for _, balance := range validatorBalancesResponse.Balances {
			logger.Debugf("%x - %v", balance.PublicKey, balance.Index)
			validators[fmt.Sprintf("%x", balance.PublicKey)] = balance.Index
		}

		if validatorBalancesResponse.NextPageToken == "" {
			break
		}
	}

	// Retrieve the validator assignments for the epoch
	validatorAssignmentes := make([]*ethpb.ValidatorAssignments_CommitteeAssignment, 0)
	validatorAssignmentResponse := &ethpb.ValidatorAssignments{}
	for validatorAssignmentResponse.NextPageToken == "" || len(validatorAssignmentes) < int(validatorAssignmentResponse.TotalSize) {
		validatorAssignmentResponse, err = client.ListValidatorAssignments(context.Background(), &ethpb.ListValidatorAssignmentsRequest{PageToken: validatorAssignmentResponse.NextPageToken, PageSize: utils.PageSize, QueryFilter: &ethpb.ListValidatorAssignmentsRequest_Epoch{Epoch: epoch}})
		if err != nil {
			return nil, fmt.Errorf("error retrieving validator assignment response for caching: %v", err)
		}
		if validatorAssignmentResponse.TotalSize == 0 || len(validatorAssignmentes) == int(validatorAssignmentResponse.TotalSize) {
			break
		}
		validatorAssignmentes = append(validatorAssignmentes, validatorAssignmentResponse.Assignments...)
		logger.Printf("Retrieved %v assignments of %v for epoch %v", len(validatorAssignmentes), validatorAssignmentResponse.TotalSize, epoch)
	}

	// Extract the proposer & attestation assignments from the response and cache them for later use
	// Proposer assignments are cached by the proposer slot
	// Attestation assignments are cached by the slot & committee key
	for index, assignment := range validatorAssignmentes {
		if assignment.ProposerSlot > 0 {
			logger.Debugf("Slot %v to be proposed by %x - %v - %v", assignment.ProposerSlot, assignment.PublicKey, validators[fmt.Sprintf("%x", assignment.PublicKey)], index)
			assignments.ProposerAssignments[assignment.ProposerSlot] = uint64(index)
		}
		if assignment.AttesterSlot > 0 {
			for memberIndex, validatorIndex := range assignment.BeaconCommittees {
				assignments.AttestorAssignments[FormatAttestorAssignmentKey(assignment.AttesterSlot, assignment.CommitteeIndex, uint64(memberIndex))] = validatorIndex
			}
		}
	}

	if len(assignments.AttestorAssignments) > 0 && len(assignments.ProposerAssignments) > 0 {
		assignmentsCache.Add(epoch, assignments)
	}

	return assignments, nil
}

func FormatAttestorAssignmentKey(AttesterSlot, CommitteeIndex, MemberIndex uint64) string {
	return fmt.Sprintf("%v-%v-%v", AttesterSlot, CommitteeIndex, MemberIndex)
}
