package services

import (
	_ "embed"
	"encoding/json"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

//go:embed services_test_block.json
var pendingBlock string

// checkSuggestion compares a suggestion value with an expected value.
func checkSuggestion(t *testing.T, name string, got, expected int64) {
	if got != expected {
		t.Errorf("expected %s suggestion %d, got %d", name, expected, got)
	}
}

func newFakeRPCServer(pendingBlock, latestBlock string) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     interface{}   `json:"id"`
			Method string        `json:"method"`
			Params []interface{} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var block json.RawMessage // Use RawMessage to handle JSON correctly
		// Choose the block based on the first parameter.
		if req.Method == "eth_getBlockByNumber" && len(req.Params) > 0 {
			if param, ok := req.Params[0].(string); ok {
				if param == "pending" {
					block = json.RawMessage(pendingBlock)
				} else if param == "latest" {
					block = json.RawMessage(latestBlock)
				}
			}
		}

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  block,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	return httptest.NewServer(handler)
}

// TestSuggestGasPrices_TableDriven consolidates various test cases into one table-driven test.
func TestSuggestGasPrices_TableDriven(t *testing.T) {
	// Helper to create 50 dummy transactions with tip values 1..50.
	makeDummyTips := func() []*big.Int {
		tips := make([]*big.Int, 50)
		for i := 0; i < 50; i++ {
			tips[i] = big.NewInt(int64(i + 1))
		}
		return tips
	}

	testCases := []struct {
		name              string
		tips              []*big.Int
		baseFee           *big.Int
		gasUsed, gasLimit uint64
		expectedRapid     int64
		expectedFast      int64
		expectedStandard  int64
		expectedSlow      int64
	}{
		{
			name:             "SufficientTxs",
			tips:             makeDummyTips(),
			baseFee:          big.NewInt(100),
			gasUsed:          10_000_000,
			gasLimit:         20_000_000,
			expectedRapid:    125,
			expectedFast:     118,
			expectedStandard: 109,
			expectedSlow:     103,
		},
		{
			name:             "EmptyTxs",
			tips:             []*big.Int{}, // no transactions
			baseFee:          big.NewInt(100),
			gasUsed:          10_000_000,
			gasLimit:         20_000_000,
			expectedRapid:    100,
			expectedFast:     100,
			expectedStandard: 100,
			expectedSlow:     100,
		},
		{
			name:             "LowGasUsage",
			tips:             makeDummyTips(),
			baseFee:          big.NewInt(100),
			gasUsed:          1_000_000, // 0.05 usage, clamped to 0.10
			gasLimit:         20_000_000,
			expectedRapid:    105,
			expectedFast:     104,
			expectedStandard: 102,
			expectedSlow:     101,
		},
		{
			name:             "HighGasUsage",
			tips:             makeDummyTips(),
			baseFee:          big.NewInt(100),
			gasUsed:          18_000_000, // 0.9 usage, clamped to 0.80
			gasLimit:         20_000_000,
			expectedRapid:    140,
			expectedFast:     129,
			expectedStandard: 115,
			expectedSlow:     104,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := suggestGasPrices(tc.gasUsed, tc.gasLimit, tc.baseFee, tc.tips)

			// Basic checks.
			if result.Code != 200 {
				t.Errorf("expected Code 200, got %d", result.Code)
			}
			if result.Data.Timestamp == 0 {
				t.Error("expected nonzero Timestamp")
			}

			// Validate the suggestions.
			checkSuggestion(t, "Rapid", result.Data.Rapid.Int64(), tc.expectedRapid)
			checkSuggestion(t, "Fast", result.Data.Fast.Int64(), tc.expectedFast)
			checkSuggestion(t, "Standard", result.Data.Standard.Int64(), tc.expectedStandard)
			checkSuggestion(t, "Slow", result.Data.Slow.Int64(), tc.expectedSlow)
		})
	}
}

// TestSuggestGasPrices_Timestamp checks that the timestamp is set and roughly "now".
func TestSuggestGasPrices_Timestamp(t *testing.T) {
	tips := []*big.Int{big.NewInt(10)}
	baseFee := big.NewInt(100)
	gasUsed := uint64(10_000_000)
	gasLimit := uint64(20_000_000)

	result := suggestGasPrices(gasUsed, gasLimit, baseFee, tips)
	now := time.Now().UnixNano() / 1e6
	if math.Abs(float64(result.Data.Timestamp-now)) > 1000 {
		t.Errorf("expected timestamp close to now, got %d (now=%d)", result.Data.Timestamp, now)
	}
}

func TestViaFakeRPC(t *testing.T) {
	cfg := &types.Config{}
	utils.Config = cfg

	server := newFakeRPCServer(pendingBlock, "")
	defer server.Close()
	utils.Config.Eth1GethEndpoint = server.URL

	result, err := getGasNowData()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if result == nil {
		t.Errorf("Error: data is nil")
		return
	}

	checkSuggestion(t, "Rapid", result.Data.Rapid.Int64(), 3939435494)
	checkSuggestion(t, "Fast", result.Data.Fast.Int64(), 3511919489)
	checkSuggestion(t, "Standard", result.Data.Standard.Int64(), 3123069597)
	checkSuggestion(t, "Slow", result.Data.Slow.Int64(), 3073069597)
}

// func TestGasNow2(t *testing.T) {
// 	cfg := &types.Config{}
// 	utils.Config = cfg
// 	utils.Config.Eth1GethEndpoint = "http://localhost:18545"
// 	data, err := getGasNowData()
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	if data == nil {
// 		t.Errorf("Error: data is nil")
// 	}

// 	fmt.Printf("Rapid: %.1f gwei\n", new(big.Float).Quo(new(big.Float).SetInt(data.Data.Rapid), big.NewFloat(1e9)))
// 	fmt.Printf("Fast: %.1f gwei\n", new(big.Float).Quo(new(big.Float).SetInt(data.Data.Fast), big.NewFloat(1e9)))
// 	fmt.Printf("Standard: %.1f gwei\n", new(big.Float).Quo(new(big.Float).SetInt(data.Data.Standard), big.NewFloat(1e9)))
// 	fmt.Printf("Slow: %.1f gwei\n", new(big.Float).Quo(new(big.Float).SetInt(data.Data.Slow), big.NewFloat(1e9)))
// }
