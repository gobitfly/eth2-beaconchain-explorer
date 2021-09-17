package handlers

import (
	"bytes"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

var poolsRocketpoolTemplate = template.Must(template.New("rocketpool").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/pools_rocketpool.html"))

// PoolsRocketpool returns the rocketpool using a go template
func PoolsRocketpool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "pools/rocketpool", "/pools/rocketpool", "Rocketpool")
	data.HeaderAd = true

	err := poolsRocketpoolTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func PoolsRocketpoolDataMinipools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	recordsTotal := uint64(0)
	recordsFiltered := uint64(0)
	var minipools []types.RocketpoolPageDataMinipool
	err = db.DB.Select(&minipools, `
		select 
			rocketpool_minipools.*, 
			coalesce(validator_names.name,'') as validator_name,
			cnt.total_count
		from rocketpool_minipools
		left join validator_names on rocketpool_minipools.pubkey = validator_names.publickey
		left join (select count(*) from rocketpool_minipools) cnt(total_count) ON true
		limit $1
		offset $2`, length, start)
	if err != nil {
		logger.Errorf("error getting rocketpool-minipools from db: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if len(minipools) > 0 {
		recordsTotal = minipools[0].TotalCount
		recordsFiltered = minipools[0].TotalCount
	}

	tableData := make([][]interface{}, 0, len(minipools))
	zeroAddr := make([]byte, 48)

	fmt.Printf("%x\n", zeroAddr)
	for _, row := range minipools {
		entry := []interface{}{}
		entry = append(entry, utils.FormatEth1Address(row.Address))
		if c := bytes.Compare(row.Pubkey, zeroAddr); c == 0 {
			entry = append(entry, "N/A")
		} else {
			entry = append(entry, utils.FormatValidatorWithName(row.Pubkey, row.ValidatorName))
		}
		entry = append(entry, utils.FormatEth1Address(row.NodeAddress))
		entry = append(entry, row.NodeFee)
		entry = append(entry, row.DepositType)
		entry = append(entry, row.Status)
		tableData = append(tableData, entry)
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func PoolsRocketpoolDataNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	recordsTotal := uint64(0)
	recordsFiltered := uint64(0)
	var dbResult []types.RocketpoolPageDataNode
	err = db.DB.Select(&dbResult, `
		select rocketpool_nodes.*, cnt.total_count
		from rocketpool_nodes
		left join (select count(*) from rocketpool_nodes) cnt(total_count) ON true
		limit $1
		offset $2`, length, start)
	if err != nil {
		logger.Errorf("error getting rocketpool-nodes from db: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if len(dbResult) > 0 {
		recordsTotal = dbResult[0].TotalCount
		recordsFiltered = dbResult[0].TotalCount
	}

	tableData := make([][]interface{}, 0, len(dbResult))

	for _, row := range dbResult {
		entry := []interface{}{}
		entry = append(entry, utils.FormatEth1Address(row.Address))
		entry = append(entry, row.TimezoneLocation)
		entry = append(entry, row.RPLStake)
		entry = append(entry, row.MinRPLStake)
		entry = append(entry, row.MaxRPLStake)
		tableData = append(tableData, entry)
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func PoolsRocketpoolDataDAOProposals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	recordsTotal := uint64(0)
	recordsFiltered := uint64(0)
	var dbResult []types.RocketpoolPageDataDAOProposal
	err = db.DB.Select(&dbResult, `
		select rocketpool_dao_proposals.*, cnt.total_count
		from rocketpool_dao_proposals
		left join (select count(*) from rocketpool_dao_proposals) cnt(total_count) ON true
		order by rocketpool_dao_proposals.id desc
		limit $1
		offset $2`, length, start)
	if err != nil {
		logger.Errorf("error getting rocketpool-nodes from db: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if len(dbResult) > 0 {
		recordsTotal = dbResult[0].TotalCount
		recordsFiltered = dbResult[0].TotalCount
	}

	tableData := make([][]interface{}, 0, len(dbResult))

	for _, row := range dbResult {
		entry := []interface{}{}
		entry = append(entry, row.ID)
		entry = append(entry, row.DAO)
		entry = append(entry, utils.FormatEth1Address(row.ProposerAddress))
		entry = append(entry, template.HTMLEscapeString(row.Message))
		entry = append(entry, utils.FormatTimestamp(row.CreatedTime.Unix()))
		entry = append(entry, utils.FormatTimestamp(row.StartTime.Unix()))
		entry = append(entry, utils.FormatTimestamp(row.EndTime.Unix()))
		entry = append(entry, utils.FormatTimestamp(row.ExpiryTime.Unix()))
		entry = append(entry, row.VotesRequired)
		entry = append(entry, row.VotesFor)
		entry = append(entry, row.VotesAgainst)
		entry = append(entry, row.MemberVoted)
		entry = append(entry, row.MemberSupported)
		entry = append(entry, row.IsCancelled)
		entry = append(entry, row.IsExecuted)
		if len(row.Payload) > 4 {
			entry = append(entry, fmt.Sprintf(`<span>%x…%x<span><i class="fa fa-copy text-muted ml-2 p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text="%x"></i>`, row.Payload[:2], row.Payload[len(row.Payload)-2:], row.Payload))
			// entry = append(entry, fmt.Sprintf(`<span>%x…%x<span> <button class="btn btn-dark text-white btn-sm" type="button" data-toggle="tooltip" title="" data-clipboard-text="%x" data-original-title="Copy to clipboard"><i class="fa fa-copy"></i></button>`, row.Payload[:2], row.Payload[len(row.Payload)-2:], row.Payload))
			// entry = append(entry, fmt.Sprintf(`<span id="rocketpool-dao-proposal-payload-%v">%x…%x</span> <button></button>`, i, row.Payload[:2], row.Payload[len(row.Payload)-2:]))
		} else {
			entry = append(entry, fmt.Sprintf("%x", row.Payload))
		}

		entry = append(entry, row.State)
		tableData = append(tableData, entry)
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func PoolsRocketpoolDataDAOMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	recordsTotal := uint64(0)
	recordsFiltered := uint64(0)
	var dbResult []types.RocketpoolPageDataDAOMember
	err = db.DB.Select(&dbResult, `
		select rocketpool_dao_members.*, cnt.total_count
		from rocketpool_dao_members
		left join (select count(*) from rocketpool_dao_members) cnt(total_count) ON true
		limit $1
		offset $2`, length, start)
	if err != nil {
		logger.Errorf("error getting rocketpool-nodes from db: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if len(dbResult) > 0 {
		recordsTotal = dbResult[0].TotalCount
		recordsFiltered = dbResult[0].TotalCount
	}

	tableData := make([][]interface{}, 0, len(dbResult))

	for _, row := range dbResult {
		entry := []interface{}{}
		entry = append(entry, utils.FormatEth1Address(row.Address))
		entry = append(entry, row.ID)
		entry = append(entry, row.URL)
		entry = append(entry, utils.FormatTimestamp(row.JoinedTime.Unix()))
		entry = append(entry, utils.FormatTimestamp(row.LastProposalTime.Unix()))
		entry = append(entry, row.RPLBondAmount)
		entry = append(entry, row.UnbondedValidatorCount)
		tableData = append(tableData, entry)
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
