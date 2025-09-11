package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/services"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/lib/pq"

	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
)

var ErrTooManyValidators = errors.New("too many validators")

func handleValidatorsQuery(w http.ResponseWriter, r *http.Request, checkValidatorLimit bool) ([]uint64, [][]byte, bool, error) {
	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	// Parse all the validator indices and pubkeys from the query string
	queryValidatorIndices, queryValidatorPubkeys, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil && (checkValidatorLimit || err != ErrTooManyValidators) {
		logger.Warnf("could not parse validators from query string: %v; Route: %v", err, r.URL.String())
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return nil, nil, false, err
	}

	// Check whether pubkeys can be converted to indices and redirect if necessary
	redirect, err := updateValidatorsQueryString(w, r, queryValidatorIndices, queryValidatorPubkeys)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error finding validators in database for dashboard query update"), 0, errFieldMap)
		http.Error(w, "Not found", http.StatusNotFound)
		return nil, nil, false, err
	}

	if !redirect {
		// Check after the redirect whether all validators are correct
		err = checkValidatorsQuery(queryValidatorIndices, queryValidatorPubkeys)
		if err != nil {
			logger.Warnf("could not find validators in database from query string: %v; Route: %v", err, r.URL.String())
			http.Error(w, "Not found", http.StatusNotFound)
			return nil, nil, false, err
		}
	}

	return queryValidatorIndices, queryValidatorPubkeys, redirect, nil
}

// parseValidatorsFromQueryString returns a slice of validator indices and a slice of validator pubkeys from a parsed query string
func parseValidatorsFromQueryString(str string, validatorLimit int) ([]uint64, [][]byte, error) {
	if str == "" {
		return []uint64{}, [][]byte{}, nil
	}

	strSplit := strings.Split(str, ",")
	strSplitLen := len(strSplit)

	// we only support up to [validatorLimit] validators
	if strSplitLen > validatorLimit {
		return []uint64{}, [][]byte{}, ErrTooManyValidators
	}

	var validatorIndices []uint64
	var validatorPubkeys [][]byte
	keys := make(map[interface{}]bool, strSplitLen)

	// Find all pubkeys
	for _, vStr := range strSplit {
		if !searchPubkeyExactRE.MatchString(vStr) {
			continue
		}
		if !strings.HasPrefix(vStr, "0x") {
			// Query string public keys have to have 0x prefix
			return []uint64{}, [][]byte{}, fmt.Errorf("invalid pubkey")
		}
		// make sure keys are unique
		if exists := keys[vStr]; exists {
			continue
		}
		keys[vStr] = true
		validatorPubkeys = append(validatorPubkeys, common.FromHex(vStr))

	}

	// Find all indices
	for _, vStr := range strSplit {
		if searchPubkeyExactRE.MatchString(vStr) {
			continue
		}
		v, err := strconv.ParseUint(vStr, 10, 64)
		if err != nil {
			return []uint64{}, [][]byte{}, err
		}
		// make sure keys are unique
		if exists := keys[v]; exists {
			continue
		}
		keys[v] = true
		validatorIndices = append(validatorIndices, v)
	}

	return validatorIndices, validatorPubkeys, nil
}

func updateValidatorsQueryString(w http.ResponseWriter, r *http.Request, validatorIndices []uint64, validatorPubkeys [][]byte) (bool, error) {
	validatorsCount := len(validatorIndices) + len(validatorPubkeys)
	if validatorsCount == 0 {
		return false, nil
	}

	// Convert pubkeys to indices if possible
	// validatorsCount stays the same after conversion
	redirect := false
	if len(validatorPubkeys) > 0 {
		validatorInfos := []struct {
			Index  uint64
			Pubkey []byte
		}{}
		err := db.ReaderDb.Select(&validatorInfos, `SELECT validatorindex as index, pubkey FROM validators WHERE pubkey = ANY($1)`, validatorPubkeys)
		if err != nil {
			return false, err
		}

		for _, info := range validatorInfos {
			// Having duplicates of validator indices is not a problem so we don't need to check for that
			validatorIndices = append(validatorIndices, info.Index)

			redirect = true
			for idx, pubkey := range validatorPubkeys {
				if bytes.Contains(pubkey, info.Pubkey) {
					validatorPubkeys = append(validatorPubkeys[:idx], validatorPubkeys[idx+1:]...)
					break
				}
			}
		}
	}

	if redirect {
		strValidators := make([]string, validatorsCount)
		for i, n := range validatorIndices {
			strValidators[i] = fmt.Sprintf("%v", n)
		}
		for i, n := range validatorPubkeys {
			strValidators[i+len(validatorIndices)] = fmt.Sprintf("%#x", n)
		}

		q := r.URL.Query()
		q.Set("validators", strings.Join(strValidators, ","))
		r.URL.RawQuery = q.Encode()

		http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
	}
	return redirect, nil
}

func checkValidatorsQuery(validatorIndices []uint64, validatorPubkeys [][]byte) error {
	validatorCount := 0

	if len(validatorIndices) > 0 {
		err := db.ReaderDb.Get(&validatorCount, `SELECT COUNT(*) FROM validators WHERE validatorindex = ANY($1)`, validatorIndices)
		if err != nil {
			return err
		}
		if validatorCount != len(validatorIndices) {
			return fmt.Errorf("invalid validator index")
		}
	}

	if len(validatorPubkeys) > 0 {
		err := db.ReaderDb.Get(&validatorCount, `SELECT COUNT(DISTINCT publickey) AS distinct_count FROM eth1_deposits WHERE publickey = ANY($1)`, validatorPubkeys)
		if err != nil {
			return err
		}
		if validatorCount != len(validatorPubkeys) {
			return fmt.Errorf("invalid validator public key")
		}
	}

	return nil
}

func DashboardDataProposals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filterArr, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	filter := pq.Array(filterArr)

	proposals := []struct {
		Slot   uint64
		Status uint64
	}{}

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	err = db.ReaderDb.Select(&proposals, `
		SELECT slot, status
		FROM blocks
		WHERE proposer = ANY($1)
		ORDER BY slot`, filter)
	if err != nil {
		utils.LogError(err, "error retrieving block-proposals", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	proposalsResult := make([][]uint64, len(proposals))
	for i, b := range proposals {
		proposalsResult[i] = []uint64{
			uint64(utils.SlotToTime(b.Slot).Unix()),
			b.Status,
		}
	}

	err = json.NewEncoder(w).Encode(proposalsResult)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Dashboard Chart that combines balance data and
func DashboardDataBalanceCombined(w http.ResponseWriter, r *http.Request) {
	var lowerBoundDay uint64
	param := r.URL.Query().Get("days")
	if len(param) != 0 {
		days, err := strconv.ParseUint(param, 10, 32)
		if err != nil {
			logger.Warnf("error parsing days: %v", err)
			http.Error(w, "Error: invalid parameter days", http.StatusBadRequest)
			return
		}
		lastStatsDay, err := services.LatestExportedStatisticDay()
		if days < lastStatsDay && err == nil {
			lowerBoundDay = lastStatsDay - days + 1
		}
	}

	currency := GetCurrency(r)
	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	w.Header().Set("Content-Type", "application/json")

	queryValidatorIndices, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	if len(queryValidatorIndices) < 1 {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}

	g, _ := errgroup.WithContext(context.Background())
	var incomeHistoryChartData []*types.ChartDataPoint
	var executionChartData []*types.ChartDataPoint
	g.Go(func() error {
		incomeHistoryChartData, err = db.GetValidatorIncomeHistoryChart(queryValidatorIndices, currency, services.LatestFinalizedEpoch(), lowerBoundDay)
		if err != nil {
			return fmt.Errorf("error in GetValidatorIncomeHistoryChart: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		executionChartData, err = getExecutionChartData(queryValidatorIndices, currency, lowerBoundDay)
		if err != nil {
			return fmt.Errorf("error in getExecutionChartData: %w", err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		utils.LogError(err, "error while combining balance chart", 0, errFieldMap)
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	var response struct {
		ConsensusChartData []*types.ChartDataPoint `json:"consensusChartData"`
		ExecutionChartData []*types.ChartDataPoint `json:"executionChartData"`
	}
	response.ConsensusChartData = incomeHistoryChartData
	response.ExecutionChartData = executionChartData

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
