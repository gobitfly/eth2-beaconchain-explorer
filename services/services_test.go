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

// --- Dummy types for testing suggestGasPrices --- //

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

// --- Test Cases --- //

// TestSuggestGasPrices_SufficientTxs creates 50 transactions with tip values 1..50.
// With gasUsed=10,000,000, gasLimit=20,000,000 (i.e. gasUsage=0.5) and baseFee=100,
// the expected suggestions are:
//
//	rapid = 125, fast = 118, standard = 109, slow = 103.
func TestSuggestGasPrices_SufficientTxs(t *testing.T) {
	// Create 50 dummy transactions (as *big.Int values) with tip values 1, 2, …, 50.
	tips := make([]*big.Int, 50)
	for i := 0; i < 50; i++ {
		tips[i] = big.NewInt(int64(i + 1))
	}
	baseFee := big.NewInt(100)
	gasUsed := uint64(10_000_000)
	gasLimit := uint64(20_000_000)

	result := suggestGasPrices(gasUsed, gasLimit, baseFee, tips)

	// Check basic fields.
	if result.Code != 200 {
		t.Errorf("expected Code 200, got %d", result.Code)
	}
	if result.Data.Timestamp == 0 {
		t.Error("expected nonzero Timestamp")
	}

	// Calculation details:
	// gasUsage = 10,000,000 / 20,000,000 = 0.5 (within [0.10, 0.80]),
	// so the percentiles are computed on 50 values:
	//   rapidTip index = ceil((100*0.5/100)*50) - 1 = ceil(25) - 1 = 24  => tip = 25
	//   fastTip  index = ceil((70*0.5/100)*50) - 1  = ceil(17.5) - 1 = 17 => tip = 18
	//   normalTip index = ceil((35*0.5/100)*50) - 1 = ceil(8.75) - 1 = 8   => tip = 9
	//   slowTip  index = ceil((10*0.5/100)*50) - 1  = ceil(2.5) - 1 = 2   => tip = 3
	// Then the suggestions equal baseFee + tip.
	// Expected: rapid = 100+25 = 125, fast = 100+18 = 118, standard = 100+9 = 109, slow = 100+3 = 103.
	checkSuggestion(t, "Rapid", result.Data.Rapid.Int64(), 125)
	checkSuggestion(t, "Fast", result.Data.Fast.Int64(), 118)
	checkSuggestion(t, "Standard", result.Data.Standard.Int64(), 109)
	checkSuggestion(t, "Slow", result.Data.Slow.Int64(), 103)
}

// TestSuggestGasPrices_EmptyTxs tests the case when no transactions are provided.
// In this situation, all percentile functions return 0 and thus all suggestions equal baseFee.
func TestSuggestGasPrices_EmptyTxs(t *testing.T) {
	var tips []*big.Int // empty slice
	baseFee := big.NewInt(100)
	gasUsed := uint64(10_000_000)
	gasLimit := uint64(20_000_000)

	result := suggestGasPrices(gasUsed, gasLimit, baseFee, tips)

	// Expect all suggestions to equal baseFee.
	checkSuggestion(t, "Rapid", result.Data.Rapid.Int64(), baseFee.Int64())
	checkSuggestion(t, "Fast", result.Data.Fast.Int64(), baseFee.Int64())
	checkSuggestion(t, "Standard", result.Data.Standard.Int64(), baseFee.Int64())
	checkSuggestion(t, "Slow", result.Data.Slow.Int64(), baseFee.Int64())
}

// TestSuggestGasPrices_LowGasUsage verifies that gas usage is clamped to a minimum of 0.10.
// For example, if gasUsed/gasLimit is lower than 0.10, the percentiles are computed as if it were 0.10.
func TestSuggestGasPrices_LowGasUsage(t *testing.T) {
	// Create 50 dummy transactions with tip values 1, 2, …, 50.
	tips := make([]*big.Int, 50)
	for i := 0; i < 50; i++ {
		tips[i] = big.NewInt(int64(i + 1))
	}
	baseFee := big.NewInt(100)
	// Use low gas usage: e.g. gasUsed=1,000,000, gasLimit=20,000,000 gives 0.05,
	// but will be clamped to 0.10.
	gasUsed := uint64(1_000_000)
	gasLimit := uint64(20_000_000)

	result := suggestGasPrices(gasUsed, gasLimit, baseFee, tips)
	// With 50 txs and clamped gasUsage=0.10:
	//   rapid index = ceil((100*0.10/100)*50)-1 = ceil(5)-1 = 4  => tip = 5, suggestion = 100+5 = 105.
	//   fast index  = ceil((70*0.10/100)*50)-1 = ceil(3.5)-1 = 3  => tip = 4, suggestion = 100+4 = 104.
	//   normal index = ceil((35*0.10/100)*50)-1 = ceil(1.75)-1 = 1  => tip = 2, suggestion = 100+2 = 102.
	//   slow index  = ceil((10*0.10/100)*50)-1 = ceil(0.5)-1 = 0   => tip = 1, suggestion = 100+1 = 101.
	checkSuggestion(t, "Rapid", result.Data.Rapid.Int64(), 105)
	checkSuggestion(t, "Fast", result.Data.Fast.Int64(), 104)
	checkSuggestion(t, "Standard", result.Data.Standard.Int64(), 102)
	checkSuggestion(t, "Slow", result.Data.Slow.Int64(), 101)
}

// TestSuggestGasPrices_HighGasUsage verifies that gas usage is clamped to a maximum of 0.80.
// For example, if gasUsed/gasLimit is higher than 0.80, the percentiles are computed as if it were 0.80.
func TestSuggestGasPrices_HighGasUsage(t *testing.T) {
	// Create 50 dummy transactions with tip values 1, 2, …, 50.
	tips := make([]*big.Int, 50)
	for i := 0; i < 50; i++ {
		tips[i] = big.NewInt(int64(i + 1))
	}
	baseFee := big.NewInt(100)
	// Use high gas usage: e.g. gasUsed=18,000,000, gasLimit=20,000,000 gives 0.9,
	// but will be clamped to 0.80.
	gasUsed := uint64(18_000_000)
	gasLimit := uint64(20_000_000)

	result := suggestGasPrices(gasUsed, gasLimit, baseFee, tips)
	// With clamped gasUsage=0.80 and 50 txs:
	//   rapid index = ceil((100*0.80/100)*50)-1 = ceil(40)-1 = 39  => tip = 40, suggestion = 100+40 = 140.
	//   fast index  = ceil((70*0.80/100)*50)-1 = ceil(29)-1 = 28   => tip = 29, suggestion = 100+29 = 129.
	//   normal index = ceil((35*0.80/100)*50)-1 = ceil(15)-1 = 14   => tip = 15, suggestion = 100+15 = 115.
	//   slow index  = ceil((10*0.80/100)*50)-1 = ceil(4)-1 = 3      => tip = 4,  suggestion = 100+4  = 104.
	checkSuggestion(t, "Rapid", result.Data.Rapid.Int64(), 140)
	checkSuggestion(t, "Fast", result.Data.Fast.Int64(), 129)
	checkSuggestion(t, "Standard", result.Data.Standard.Int64(), 115)
	checkSuggestion(t, "Slow", result.Data.Slow.Int64(), 104)
}

// (Optional) TestSuggestGasPrices_Timestamp checks that a timestamp is set and is roughly "now".
func TestSuggestGasPrices_Timestamp(t *testing.T) {
	tips := []*big.Int{
		big.NewInt(10),
	}
	baseFee := big.NewInt(100)
	gasUsed := uint64(10_000_000)
	gasLimit := uint64(20_000_000)

	result := suggestGasPrices(gasUsed, gasLimit, baseFee, tips)
	// Check that the timestamp is roughly the current time (allowing a 1 second skew).
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
	}

	checkSuggestion(t, "Rapid", result.Data.Rapid.Int64(), 3939435494)
	checkSuggestion(t, "Fast", result.Data.Fast.Int64(), 3511919489)
	checkSuggestion(t, "Standard", result.Data.Standard.Int64(), 3123069597)
	checkSuggestion(t, "Slow", result.Data.Slow.Int64(), 3073069597)
}

// func TestGasNow2(t *testing.T) {
// 	cfg := &types.Config{}
// 	utils.Config = cfg
// 	// utils.Config.Eth1GethEndpoint = "http://localhost:18545"
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
