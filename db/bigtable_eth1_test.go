package db

import (
	"testing"

	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
)

func TestName(t *testing.T) {
	traces := &rpc.GethTraceCallResult{
		Input: "source",
		Calls: []*rpc.GethTraceCallResult{
			{
				// child 1 and its children are reverted
				Input: "child 1",
				Error: "error",
				Calls: []*rpc.GethTraceCallResult{
					{
						Input: "child 1-1",
					},
				},
			},
			{
				Input: "child 2",
			},
		},
	}
	tests := []struct {
		name      string
		internals *rpc.GethTraceCallResult
		expected  bool
		total     int
	}{
		{
			name:      "head",
			internals: traces,
			expected:  false,
			total:     0,
		}, {
			name:      "child 1",
			internals: traces.Calls[0],
			expected:  true,
			total:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			revertedTraces := make(map[*rpc.GethTraceCallResult]struct{})
			reverted := isReverted(tt.internals, revertedTraces)
			if got, want := reverted, tt.expected; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := len(revertedTraces), tt.total; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
