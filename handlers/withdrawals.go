package handlers

import (
	"context"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// Withdrawals will return information about recent withdrawals
func Withdrawals(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "withdrawals.html", "validator/withdrawalOverviewRow.html", "components/charts.html")
	var withdrawalsTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	pageData := &types.WithdrawalsPageData{}
	pageData.Stats = services.GetLatestStats()

	data := InitPageData(w, r, "validators", "/withdrawals", "Validator Withdrawals", templateFiles)

	latestChartsPageData := services.LatestChartsPageData()
	if len(latestChartsPageData) != 0 {
		for _, c := range latestChartsPageData {
			if c.Path == "withdrawals" {
				pageData.WithdrawalChart = c
				break
			}
		}
	}

	data.Data = pageData

	err := withdrawalsTemplate.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "withdrawals.go", "withdrawals", "", err) != nil {
		return // an error has occurred and was processed
	}
}

// WithdrawalsData will return eth1-deposits as json
func WithdrawalsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	currency := GetCurrency(r)
	q := r.URL.Query()

	search := ReplaceEnsNameWithAddress(q.Get("search[value]"))

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}

	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}
	if start > db.WithdrawalsQueryLimit {
		start = db.WithdrawalsQueryLimit
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}
	if length > 100 {
		length = 100
	}

	orderBy := q.Get("order[0][column]")
	orderDir := q.Get("order[0][dir]")

	data, err := WithdrawalsTableData(draw, search, length, start, orderBy, orderDir, currency)
	if err != nil {
		logger.Errorf("error getting withdrawal table data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func WithdrawalsTableData(draw uint64, search string, length, start uint64, orderBy, orderDir string, currency string) (*types.DataTableResponse, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()

	orderByMap := map[string]string{
		"0": "epoch",
		"1": "slot",
		"2": "index",
		"3": "validator",
		"4": "address",
		"5": "amount",
	}
	orderColumn, exists := orderByMap[orderBy]
	if !exists {
		orderBy = "index"
	}

	if orderDir != "asc" {
		orderDir = "desc"
	}

	g, gCtx := errgroup.WithContext(ctx)
	withdrawalCount := uint64(0)
	g.Go(func() error {
		select {
		case <-gCtx.Done():
			return nil
		default:
		}
		var err error
		withdrawalCount, err = db.GetTotalWithdrawals()
		if err != nil {
			return fmt.Errorf("error getting total withdrawals: %w", err)
		}
		return nil
	})

	filteredCount := uint64(0)
	trimmedSearch := strings.ToLower(strings.TrimPrefix(search, "0x"))
	if trimmedSearch != "" {
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return nil
			default:
			}
			var err error
			filteredCount, err = db.GetWithdrawalsCountForQuery(search)
			if err != nil {
				return fmt.Errorf("error getting withdrwal count for filter [%v]: %w", search, err)
			}
			return nil
		})
	}

	withdrawals := []*types.Withdrawals{}
	g.Go(func() error {
		select {
		case <-gCtx.Done():
			return nil
		default:
		}
		var err error
		withdrawals, err = db.GetWithdrawals(search, length, start, orderColumn, orderDir)
		if err != nil {
			return fmt.Errorf("error getting withdrawals: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if trimmedSearch == "" {
		filteredCount = withdrawalCount
	}

	formatCurrency := currency
	if currency == "ETH" {
		formatCurrency = "Ether"
	}

	var err error
	names := make(map[string]string)
	for _, v := range withdrawals {
		names[string(v.Address)] = ""
	}
	names, _, err = db.BigtableClient.GetAddressesNamesArMetadata(&names, nil)
	if err != nil {
		return nil, err
	}

	tableData := make([][]interface{}, len(withdrawals))
	for i, w := range withdrawals {
		tableData[i] = []interface{}{
			utils.FormatEpoch(utils.EpochOfSlot(w.Slot)),
			utils.FormatBlockSlot(w.Slot),
			template.HTML(fmt.Sprintf("%v", w.Index)),
			utils.FormatValidator(w.ValidatorIndex),
			utils.FormatTimestamp(utils.SlotToTime(w.Slot).Unix()),
			utils.FormatAddressWithLimits(w.Address, names[string(w.Address)], false, "address", visibleDigitsForHash+5, 18, true),
			utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(1e9)), formatCurrency, 6),
		}
	}

	if filteredCount > db.WithdrawalsQueryLimit {
		filteredCount = db.WithdrawalsQueryLimit
	}
	if withdrawalCount > db.WithdrawalsQueryLimit {
		withdrawalCount = db.WithdrawalsQueryLimit
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    withdrawalCount,
		RecordsFiltered: filteredCount,
		Data:            tableData,
		PageLength:      length,
		DisplayStart:    start,
	}
	return data, nil
}

// Eth1DepositsData will return eth1-deposits as json
func BLSChangeData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()

	search := ReplaceEnsNameWithAddress(q.Get("search[value]"))

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}
	if start > db.BlsChangeQueryLimit {
		// limit offset to 10000, otherwise the query will be too slow
		start = db.BlsChangeQueryLimit
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}
	if length > 100 {
		length = 100
	}

	orderBy := q.Get("order[0][column]")
	orderDir := q.Get("order[0][dir]")

	data, err := BLSTableData(draw, search, length, start, orderBy, orderDir)
	if err != nil {
		logger.Errorf("Error getting bls changes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func BLSTableData(draw uint64, search string, length, start uint64, orderBy, orderDir string) (*types.DataTableResponse, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()
	orderByMap := map[string]string{
		"0": "block_slot",
		"1": "block_slot",
		"2": "validatorindex",
	}
	orderVar, exists := orderByMap[orderBy]
	if !exists {
		orderBy = "block_slot"
	}

	if orderDir != "asc" {
		orderDir = "desc"
	}

	g, gCtx := errgroup.WithContext(ctx)
	totalCount := uint64(0)
	g.Go(func() error {
		select {
		case <-gCtx.Done():
			return nil
		default:
		}
		var err error
		totalCount, err = db.GetTotalBLSChanges()
		if err != nil {
			return fmt.Errorf("error getting total bls changes: %w", err)
		}
		return nil
	})

	filteredCount := uint64(0)
	trimmedSearch := strings.ToLower(strings.TrimPrefix(search, "0x"))
	if trimmedSearch != "" {
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return nil
			default:
			}
			var err error
			filteredCount, err = db.GetBLSChangesCountForQuery(search)
			if err != nil {
				return fmt.Errorf("error getting bls changes count for filter [%v]: %w", search, err)
			}
			return nil
		})
	}

	blsChange := []*types.BLSChange{}
	g.Go(func() error {
		select {
		case <-gCtx.Done():
			return nil
		default:
		}
		var err error
		blsChange, err = db.GetBLSChanges(search, length, start, orderVar, orderDir)
		if err != nil {
			return fmt.Errorf("error getting bls changes: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if trimmedSearch == "" {
		filteredCount = totalCount
	}

	var err error
	names := make(map[string]string)
	for _, v := range blsChange {
		names[string(v.Address)] = ""
	}
	names, _, err = db.BigtableClient.GetAddressesNamesArMetadata(&names, nil)
	if err != nil {
		return nil, err
	}

	tableData := make([][]interface{}, len(blsChange))
	for i, bls := range blsChange {
		tableData[i] = []interface{}{
			utils.FormatEpoch(utils.EpochOfSlot(bls.Slot)),
			utils.FormatBlockSlot(bls.Slot),
			utils.FormatValidator(bls.Validatorindex),
			utils.FormatHashWithCopy(bls.Signature),
			utils.FormatHashWithCopy(bls.BlsPubkey),
			utils.FormatAddressWithLimits(bls.Address, names[string(bls.Address)], false, "address", visibleDigitsForHash+5, 18, true),
		}
	}

	if totalCount > db.BlsChangeQueryLimit {
		totalCount = db.BlsChangeQueryLimit
	}
	if filteredCount > db.BlsChangeQueryLimit {
		filteredCount = db.BlsChangeQueryLimit
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: filteredCount,
		Data:            tableData,
		PageLength:      length,
		DisplayStart:    start,
	}
	return data, nil
}
