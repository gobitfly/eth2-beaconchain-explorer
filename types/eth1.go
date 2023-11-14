package types

import (
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

type GetBlockTimings struct {
	Headers  time.Duration
	Receipts time.Duration
	Traces   time.Duration
}

type BulkMutations struct {
	Keys []string
	Muts []*gcp_bigtable.Mutation
}

func (bulkMutations *BulkMutations) Add(key string, mut *gcp_bigtable.Mutation) {
	bulkMutations.Keys = append(bulkMutations.Keys, key)
	bulkMutations.Muts = append(bulkMutations.Muts, mut)
}

func (bulkMutations *BulkMutations) Len() int {
	return len(bulkMutations.Keys)
}

func (bulkMutations *BulkMutations) Less(i, j int) bool {
	return bulkMutations.Keys[i] < bulkMutations.Keys[j]
}

func (bulkMutations *BulkMutations) Swap(i, j int) {
	bulkMutations.Keys[i], bulkMutations.Keys[j] = bulkMutations.Keys[j], bulkMutations.Keys[i]
	bulkMutations.Muts[i], bulkMutations.Muts[j] = bulkMutations.Muts[j], bulkMutations.Muts[i]
}

type BulkMutation struct {
	Key string
	Mut *gcp_bigtable.Mutation
}
