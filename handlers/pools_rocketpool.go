package handlers

import (
	"bytes"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// PoolsRocketpool returns the rocketpool using a go template
func PoolsRocketpool(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "pools_rocketpool.html")
	var poolsRocketpoolTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "pools/rocketpool", "/pools/rocketpool", "Rocketpool", templateFiles)

	if handleTemplateError(w, r, "pools_rocketpool.go", "PoolsRocketpool", "", poolsRocketpoolTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func PoolsRocketpoolDataMinipools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
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
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}
	if length > 100 {
		length = 100
	}
	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}

	// Search for invalid postgres strings
	if utils.HasProblematicUtfCharacters(search) || strings.HasSuffix(search, "\\") {
		logger.Warnf("error converting search %v to valid UTF-8): %v", search, err)
		http.Error(w, "Error: Invalid parameter search.", http.StatusBadRequest)
		return
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "address",
		"1": "pubkey",
		"2": "node_address",
		"3": "node_fee",
		"4": "node_deposit_balance",
		"5": "deposit_type",
		"6": "status",
		"7": "penalty_count",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "address"
	}
	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}

	recordsTotal := uint64(0)
	recordsFiltered := uint64(0)
	var minipools []types.RocketpoolPageDataMinipool
	if search == "" {
		err = db.ReaderDb.Select(&minipools, fmt.Sprintf(`
			select 
				rocketpool_minipools.rocketpool_storage_address, 
				rocketpool_minipools.address, 
				rocketpool_minipools.pubkey, 
				rocketpool_minipools.node_address, 
				rocketpool_minipools.node_fee, 
				rocketpool_minipools.deposit_type, 
				rocketpool_minipools.status, 
				rocketpool_minipools.status_time, 
				rocketpool_minipools.penalty_count,
				validators.validatorindex as validator_index,
				coalesce(validator_names.name,'') as validator_name,
				coalesce((node_deposit_balance / 1e18)::int, 16) as node_deposit_balance,
				cnt.total_count
			from rocketpool_minipools
			left join validator_names on rocketpool_minipools.pubkey = validator_names.publickey
			left join validators on rocketpool_minipools.pubkey = validators.pubkey
			left join (select count(*) from rocketpool_minipools) cnt(total_count) ON true
			order by %s %s
			limit $1
			offset $2`, orderBy, orderDir), length, start)
		if err != nil {
			logger.Errorf("error getting rocketpool-minipools from db: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		err = db.ReaderDb.Select(&minipools, fmt.Sprintf(`
			with matched_minipools as (
				select address from rocketpool_minipools where encode(pubkey::bytea,'hex') like $3
				union select address from rocketpool_minipools where encode(address::bytea,'hex') like $3
				union (select address from validator_names inner join rocketpool_minipools on rocketpool_minipools.pubkey = validator_names.publickey where name ilike $4)
			)
			select 
				rocketpool_minipools.rocketpool_storage_address, 
				rocketpool_minipools.address, 
				rocketpool_minipools.pubkey, 
				rocketpool_minipools.node_address, 
				rocketpool_minipools.node_fee, 
				rocketpool_minipools.deposit_type, 
				rocketpool_minipools.status, 
				rocketpool_minipools.status_time, 
				rocketpool_minipools.penalty_count,
				validators.validatorindex as validator_index,
				coalesce((node_deposit_balance / 1e18)::int, 16) as node_deposit_balance,
				coalesce(validator_names.name,'') as validator_name,
				cnt.total_count
			from rocketpool_minipools
			inner join matched_minipools on rocketpool_minipools.address = matched_minipools.address
			left join validator_names on rocketpool_minipools.pubkey = validator_names.publickey
			left join validators on rocketpool_minipools.pubkey = validators.pubkey
			left join (select count(*) from matched_minipools) cnt(total_count) ON true
			order by %s %s
			limit $1
			offset $2`, orderBy, orderDir), length, start, search+"%", "%"+search+"%")
		if err != nil {
			logger.Errorf("error getting rocketpool-minipools from db (with search: %v): %v", search, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	if len(minipools) > 0 {
		recordsTotal = minipools[0].TotalCount
		recordsFiltered = minipools[0].TotalCount
	}

	tableData := make([][]interface{}, 0, len(minipools))
	zeroAddr := make([]byte, 48)

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
		entry = append(entry, row.DepositEth)
		entry = append(entry, row.DepositType)
		entry = append(entry, row.Status)
		entry = append(entry, row.PenaltyCount)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func PoolsRocketpoolDataNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
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
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}
	if length > 100 {
		length = 100
	}
	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}

	// Search for invalid postgres strings
	if utils.HasProblematicUtfCharacters(search) || strings.HasSuffix(search, "\\") {
		logger.Warnf("error converting search %v to valid UTF-8): %v", search, err)
		http.Error(w, "Error: Invalid parameter search.", http.StatusBadRequest)
		return
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "address",
		"1": "timezone_location",
		"2": "rpl_stake",
		"3": "min_rpl_stake",
		"4": "max_rpl_stake",
		"5": "rpl_cumulative_rewards",
		"6": "deposit_credit",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "address"
	}
	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}

	recordsTotal := uint64(0)
	recordsFiltered := uint64(0)
	var dbResult []types.RocketpoolPageDataNode

	if search == "" {
		err = db.ReaderDb.Select(&dbResult, fmt.Sprintf(`
			select 
				rocketpool_nodes.rocketpool_storage_address, 
				rocketpool_nodes.address, 
				rocketpool_nodes.timezone_location, 
				rocketpool_nodes.rpl_stake, 
				rocketpool_nodes.min_rpl_stake, 
				rocketpool_nodes.max_rpl_stake, 
				rocketpool_nodes.rpl_cumulative_rewards, 
				rocketpool_nodes.smoothing_pool_opted_in, 
				rocketpool_nodes.claimed_smoothing_pool, 
				rocketpool_nodes.unclaimed_smoothing_pool, 
				rocketpool_nodes.unclaimed_rpl_rewards, 
				rocketpool_nodes.deposit_credit,
				cnt.total_count
			from rocketpool_nodes
			left join (select count(*) from rocketpool_nodes) cnt(total_count) ON true
			order by %s %s
			limit $1
			offset $2`, orderBy, orderDir), length, start)
		if err != nil {
			logger.Errorf("error getting rocketpool-nodes from db: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		err = db.ReaderDb.Select(&dbResult, fmt.Sprintf(`
			with matched_nodes as (
				select address from rocketpool_nodes where encode(address::bytea,'hex') like $3
			)
			select 
				rocketpool_nodes.rocketpool_storage_address, 
				rocketpool_nodes.address, 
				rocketpool_nodes.timezone_location, 
				rocketpool_nodes.rpl_stake, 
				rocketpool_nodes.min_rpl_stake, 
				rocketpool_nodes.max_rpl_stake, 
				rocketpool_nodes.rpl_cumulative_rewards, 
				rocketpool_nodes.smoothing_pool_opted_in, 
				rocketpool_nodes.claimed_smoothing_pool, 
				rocketpool_nodes.unclaimed_smoothing_pool, 
				rocketpool_nodes.unclaimed_rpl_rewards,
				rocketpool_nodes.deposit_credit,
				cnt.total_count
			from rocketpool_nodes
			inner join matched_nodes on matched_nodes.address = rocketpool_nodes.address
			left join (select count(*) from rocketpool_nodes) cnt(total_count) ON true
			order by %s %s
			limit $1
			offset $2`, orderBy, orderDir), length, start, search+"%")
		if err != nil {
			logger.Errorf("error getting rocketpool-nodes from db (with search: %v): %v", search, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
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
		entry = append(entry, row.CumulativeRPL)
		entry = append(entry, row.DepositCredit)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func PoolsRocketpoolDataDAOProposals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
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
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}
	if length > 100 {
		length = 100
	}
	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}

	// Search for invalid postgres strings
	if utils.HasProblematicUtfCharacters(search) || strings.HasSuffix(search, "\\") {
		logger.Warnf("error converting search %v to valid UTF-8): %v", search, err)
		http.Error(w, "Error: Invalid parameter search.", http.StatusBadRequest)
		return
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0":  "id",
		"1":  "dao",
		"2":  "proposer",
		"3":  "message",
		"14": "is_executed",
		"15": "payload",
		"16": "state",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "id"
	}
	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}

	recordsTotal := uint64(0)
	recordsFiltered := uint64(0)
	var dbResult []types.RocketpoolPageDataDAOProposal
	if search == "" {
		err = db.ReaderDb.Select(&dbResult, fmt.Sprintf(`
			select 
				rocketpool_storage_address,
				rocketpool_dao_proposals.id,
				dao,
				proposer_address,
				message,
				created_time,
				start_time,
				end_time,
				expiry_time,
				votes_required,
				votes_for,
				votes_against,
				member_voted,
				member_supported,
				is_cancelled,
				is_executed,
				payload,
				state,
				cnt.total_count, 
				jsonb_agg(t) as member_votes 
			from rocketpool_dao_proposals
			left join (select count(*) from rocketpool_dao_proposals) cnt(total_count) ON true 
			left join (
				SELECT rocketpool_dao_proposals_member_votes.id, encode(member_address::bytea, 'hex') as member_address, voted, supported,rocketpool_dao_members.id as name FROM rocketpool_dao_proposals_member_votes 
				LEFT JOIN rocketpool_dao_members ON member_address = address
			) t ON t.id = rocketpool_dao_proposals.id
			group by rocketpool_dao_proposals.rocketpool_storage_address, rocketpool_dao_proposals.id, cnt.total_count
			order by %s %s
			limit $1
			offset $2`, orderBy, orderDir), length, start)
		if err != nil {
			logger.Errorf("error getting rocketpool-proposals from db: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		err = db.ReaderDb.Select(&dbResult, fmt.Sprintf(`
			with matched_proposals as (
				select id from rocketpool_dao_proposals where cast(id as text) like $3
				union select id from rocketpool_dao_proposals where dao like $5
				union select id from rocketpool_dao_proposals where message ilike $5
				union select id from rocketpool_dao_proposals where state = $3
				union select id from rocketpool_dao_proposals where encode(proposer_address::bytea,'hex') like $4
			)
			select 
				rocketpool_storage_address,
				rocketpool_dao_proposals.id,
				dao,
				proposer_address,
				message,
				created_time,
				start_time,
				end_time,
				expiry_time,
				votes_required,
				votes_for,
				votes_against,
				member_voted,
				member_supported,
				is_cancelled,
				is_executed,
				payload,
				state,
				jsonb_agg(t) as member_votes,
				cnt.total_count
			from rocketpool_dao_proposals
			inner join matched_proposals on matched_proposals.id = rocketpool_dao_proposals.id
			left join (select count(*) from matched_proposals) cnt(total_count) ON true
			left join (
				SELECT rocketpool_dao_proposals_member_votes.id, encode(member_address::bytea, 'hex') as member_address, voted, supported,rocketpool_dao_members.id as name FROM rocketpool_dao_proposals_member_votes 
				LEFT JOIN rocketpool_dao_members ON member_address = address
			) t ON t.id = rocketpool_dao_proposals.id
			group by rocketpool_dao_proposals.rocketpool_storage_address, rocketpool_dao_proposals.id, cnt.total_count
			order by %s %s
			limit $1
			offset $2`, orderBy, orderDir), length, start, search, search+"%", "%"+search+"%")
		if err != nil {
			logger.Errorf("error getting rocketpool-proposals from db (with search: %v): %v", search, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
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
		entry = append(entry, formatVoteTable(row.MemberVotesJSON))
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func formatVoteTable(votes []byte) template.HTML {
	var arr []types.RocketpoolPageDataDAOProposalMemberVotes
	err := json.Unmarshal(votes, &arr)
	if err != nil {
		logger.Warnf("can not parse rocketpool dao proposal member json %v", err)
		return template.HTML("")
	}

	result := `<table style="margin-top: 12px;"><thead><tr><th>Member Name</th><th>Member Address</th><th>Vote</th></tr></thead><tbody>`

	for _, vote := range arr {
		if vote.Voted {
			voted := `👎 <strong style="color: #f82e2e;">nay</strong>`
			if vote.Supported {
				voted = `👍 <strong style="color: #2d7533;">yea</strong>`
			}
			result += fmt.Sprintf("<tr><td>%v</td><td>0x%v</td><td>%v</td></tr>", vote.Name, vote.Address, voted)
		}

	}
	result += "</tbody></table>"
	return template.HTML(result)
}

func PoolsRocketpoolDataDAOMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()
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
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}
	if length > 100 {
		length = 100
	}
	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}

	// Search for invalid postgres strings
	if utils.HasProblematicUtfCharacters(search) || strings.HasSuffix(search, "\\") {
		logger.Warnf("error converting search %v to valid UTF-8): %v", search, err)
		http.Error(w, "Error: Invalid parameter search.", http.StatusBadRequest)
		return
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "address",
		"1": "id",
		"2": "url",
		"3": "joined_time",
		"4": "last_proposal_time",
		"5": "rpl_bond_amount",
		"6": "unbonded_validator_count",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "id"
	}
	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}

	recordsTotal := uint64(0)
	recordsFiltered := uint64(0)
	var dbResult []types.RocketpoolPageDataDAOMember
	if search == "" {
		err = db.ReaderDb.Select(&dbResult, fmt.Sprintf(`
			select 
				rocketpool_dao_members.rocketpool_storage_address, 
				rocketpool_dao_members.address, 
				rocketpool_dao_members.id, 
				rocketpool_dao_members.url, 
				rocketpool_dao_members.joined_time, 
				rocketpool_dao_members.last_proposal_time, 
				rocketpool_dao_members.rpl_bond_amount, 
				rocketpool_dao_members.unbonded_validator_count, 
				cnt.total_count
			from rocketpool_dao_members
			left join (select count(*) from rocketpool_dao_members) cnt(total_count) ON true
			order by %s %s
			limit $1
			offset $2`, orderBy, orderDir), length, start)
		if err != nil {
			logger.Errorf("error getting rocketpool-members from db: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		err = db.ReaderDb.Select(&dbResult, fmt.Sprintf(`
			with matched_members as (
				select address from rocketpool_dao_members where encode(address::bytea,'hex') like $3
				union select address from rocketpool_dao_members where id ilike $4
				union select address from rocketpool_dao_members where url ilike $4
			)
			select 			
				rocketpool_dao_members.rocketpool_storage_address, 
				rocketpool_dao_members.address, 
				rocketpool_dao_members.id, 
				rocketpool_dao_members.url, 
				rocketpool_dao_members.joined_time, 
				rocketpool_dao_members.last_proposal_time, 
				rocketpool_dao_members.rpl_bond_amount, 
				rocketpool_dao_members.unbonded_validator_count, 
				cnt.total_count
			from rocketpool_dao_members
			inner join matched_members on matched_members.address = rocketpool_dao_members.address
			left join (select count(*) from matched_members) cnt(total_count) ON true
			order by %s %s
			limit $1
			offset $2`, orderBy, orderDir), length, start, search+"%", "%"+search+"%")
		if err != nil {
			logger.Errorf("error getting rocketpool-members from db (with search: %v): %v", search, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
