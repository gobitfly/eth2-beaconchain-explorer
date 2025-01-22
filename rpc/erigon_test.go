package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGnosisBlockHash(t *testing.T) {
	gnosisURL := os.Getenv("GNOSIS_NODE_URL")
	if gnosisURL == "" {
		t.Skip("skipping test, set GNOSIS_NODE_URL")
	}

	erigonClient, err := NewErigonClient(gnosisURL)
	if err != nil {
		logrus.Fatalf("error initializing erigon client: %v", err)
	}

	tests := []struct {
		name  string
		block int64
	}{
		{
			name:  "old block with extra fields",
			block: 68140,
		},
		{
			name:  "old block with extra fields #2",
			block: 19187811,
		},
		{
			name:  "without receipts",
			block: 38039324,
		},
		{
			name:  "newest block",
			block: 37591835,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, _, err := erigonClient.GetBlock(tt.block, "parity/geth")
			if err != nil {
				t.Fatal(err)
			}

			type MinimalBlock struct {
				Result struct {
					Hash string `json:"hash"`
				} `json:"result"`
			}

			query := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params": ["0x%x",false],"id":1}`, tt.block)
			resp, err := http.Post(gnosisURL, "application/json", bytes.NewBufferString(query))
			if err != nil {
				t.Fatal(err)
			}

			var res MinimalBlock
			if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
				t.Fatal(err)
			}

			if got, want := fmt.Sprintf("0x%x", parsed.Hash), res.Result.Hash; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
