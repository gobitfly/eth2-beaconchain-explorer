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
