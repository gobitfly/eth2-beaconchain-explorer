package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/templates"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// EntityDetail renders a dashboard for a specific entity/sub-entity based on precomputed data.
func EntityDetail(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "entity_detail.html")
	var entityTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	vars := mux.Vars(r)
	entity := vars["entity"]
	subEntity := vars["subEntity"]
	if entity == "" || subEntity == "" {
		http.NotFound(w, r)
		return
	}

	data := InitPageData(w, r, "services", r.URL.Path, "Entity Dashboard", templateFiles)

	period := GetRequestedPeriod(r)

	row24, err := db.GetEntityDetailData(entity, subEntity, period)
	if errors.Is(err, sql.ErrNoRows) {
		logger.WithField("entity", entity).WithField("sub_entity", subEntity).Warn("entity detail not found")
		http.NotFound(w, r)
		return
	} else if err != nil {
		logger.WithError(err).WithField("entity", entity).WithField("sub_entity", subEntity).Error("query validator_entities_data_24h failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Prepare series points in ms for Highcharts
	type point struct {
		TsMs  int64
		Value float64
	}
	effSeries := make([]point, 0, len(row24.EfficiencyTimeBucketValues))
	for i := range row24.EfficiencyTimeBucketValues {
		if i < len(row24.EfficiencyTimeBucketTimestampsSec) {
			effSeries = append(effSeries, point{
				TsMs:  row24.EfficiencyTimeBucketTimestampsSec[i] * 1000,
				Value: row24.EfficiencyTimeBucketValues[i],
			})
		}
	}

	// View model tailored for the template
	type viewModel struct {
		Entity        string
		SubEntity     string
		LastUpdatedAt time.Time

		BalanceEndSumGwei int64
		NetShare          float64

		ClRewardsGWei   decimal.Decimal
		ElRewardsWei    decimal.Decimal
		TotalRewardsWei decimal.Decimal
		APR             decimal.Decimal

		Efficiency              float64
		AttestationEfficiency   float64
		ProposalEfficiency      float64
		SyncCommitteeEfficiency float64

		AttestationsScheduledSum         int64
		AttestationsObservedSum          int64
		AttestationsHeadExecutedSum      int64
		AttestationsSourceExecutedSum    int64
		AttestationsTargetExecutedSum    int64
		AttestationsMissedRewardsSum     int64
		AttestationsRewardRewardsOnlySum int64

		BlocksScheduledSum int64
		BlocksProposedSum  int64

		SyncScheduledSum int64
		SyncExecutedSum  int64

		SlashedInPeriodMax            int64
		SlashedAmountSum              int64
		BlocksClMissedMedianRewardSum int64
		SyncLocalizedMaxRewardSum     int64
		SyncRewardRewardsOnlySum      int64

		MissedSyncRewardsSum int64
		TotalMissedClRewards int64

		InclusionDelaySum      int64
		AttestationAvgInclDist float64

		// Online/Offline counts and tooltips
		OnlineCount          uint64
		OfflineCount         uint64
		OnlineBreakdownHTML  template.HTMLAttr
		OfflineBreakdownHTML template.HTMLAttr

		EfficiencySeries []point

		// Controls whether to render sub-entity table/breadcrumbs
		HasRealSubEntities bool

		// Period selection controls for template
		Period      string
		PeriodLinks []PeriodLink
	}

	vm := &viewModel{
		Entity:                           row24.Entity,
		SubEntity:                        row24.SubEntity,
		LastUpdatedAt:                    row24.LastUpdatedAt,
		BalanceEndSumGwei:                row24.BalanceEndSumGwei,
		NetShare:                         row24.NetShare,
		Efficiency:                       row24.Efficiency,
		AttestationEfficiency:            row24.AttestationEfficiency,
		ProposalEfficiency:               row24.ProposalEfficiency,
		SyncCommitteeEfficiency:          row24.SyncCommitteeEfficiency,
		AttestationsScheduledSum:         row24.AttestationsScheduledSum,
		AttestationsObservedSum:          row24.AttestationsObservedSum,
		AttestationsHeadExecutedSum:      row24.AttestationsHeadExecutedSum,
		AttestationsSourceExecutedSum:    row24.AttestationsSourceExecutedSum,
		AttestationsTargetExecutedSum:    row24.AttestationsTargetExecutedSum,
		AttestationsMissedRewardsSum:     row24.AttestationsMissedRewardsSum,
		AttestationsRewardRewardsOnlySum: row24.AttestationsRewardRewardsOnlySum,
		BlocksScheduledSum:               row24.BlocksScheduledSum,
		BlocksProposedSum:                row24.BlocksProposedSum,
		SyncScheduledSum:                 row24.SyncScheduledSum,
		SyncExecutedSum:                  row24.SyncExecutedSum,
		SlashedInPeriodMax:               row24.SlashedInPeriodMax,
		SlashedAmountSum:                 row24.SlashedAmountSum,
		BlocksClMissedMedianRewardSum:    row24.BlocksClMissedMedianRewardSum,
		SyncLocalizedMaxRewardSum:        row24.SyncLocalizedMaxRewardSum,
		SyncRewardRewardsOnlySum:         row24.SyncRewardRewardsOnlySum,
		MissedSyncRewardsSum:             row24.SyncLocalizedMaxRewardSum - row24.SyncRewardRewardsOnlySum,
		TotalMissedClRewards:             row24.AttestationsMissedRewardsSum + row24.BlocksClMissedMedianRewardSum + (row24.SyncLocalizedMaxRewardSum - row24.SyncRewardRewardsOnlySum),
		InclusionDelaySum:                row24.InclusionDelaySum,
		ElRewardsWei:                     row24.ExecutionRewardsSumWei,
		ClRewardsGWei:                    row24.RoiDividend.Sub(row24.RoiDivisor),
		EfficiencySeries:                 effSeries,
	}

	vm.TotalRewardsWei = vm.ClRewardsGWei.Mul(decimal.NewFromInt(1e9)).Add(row24.ExecutionRewardsSumWei)
	vm.Period = period
	vm.PeriodLinks = BuildPeriodToggleLinks(r, r.URL.Path)
	// Compute average inclusion distance: 1 + inclusion_delay_sum / observed_attestations
	if row24.AttestationsObservedSum > 0 {
		vm.AttestationAvgInclDist = 1.0 + float64(row24.InclusionDelaySum)/float64(row24.AttestationsObservedSum)
	}

	// Determine if there are any real sub-entities for this entity (excluding the default '-')
	hasRealSubs, err := db.HasRealSubEntities(entity, period)
	if err != nil {
		logger.WithError(err).WithField("entity", entity).Error("EntityDetail: check real sub-entities failed")
	}
	vm.HasRealSubEntities = hasRealSubs

	// Online/Offline counts and tooltips from status_counts JSONB
	onlineCount, offlineCount, onlineHTML, offlineHTML := buildStatusTooltipsFromRaw(row24.StatusCountsRaw)
	vm.OnlineCount = onlineCount
	vm.OfflineCount = offlineCount
	vm.OnlineBreakdownHTML = onlineHTML
	vm.OfflineBreakdownHTML = offlineHTML

	data.Data = vm

	if handleTemplateError(w, r, "entity_detail.go", "EntityDetail", "Done", entityTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return
	}
}

// buildStatusTooltipsFromRaw produces online/offline counts and HTML tooltips out of the status_counts JSONB bytes.
func buildStatusTooltipsFromRaw(statusCountsRaw []byte) (uint64, uint64, template.HTMLAttr, template.HTMLAttr) {
	var statusCounts map[string]int64
	if len(statusCountsRaw) > 0 {
		_ = json.Unmarshal(statusCountsRaw, &statusCounts)
	} else {
		statusCounts = map[string]int64{}
	}

	onlineOrder := []string{"active_online", "exiting_online", "slashing_online"}
	offlineOrder := []string{"active_offline", "exiting_offline", "slashing_offline", "pending_initialized", "pending", "deposited", "exited", "slashed"}

	var onlineSum uint64
	var offlineSum uint64
	onlineMap := make(map[string]uint64, 8)
	offlineMap := make(map[string]uint64, 16)
	for k, v := range statusCounts {
		if v < 0 {
			continue
		}
		uv := uint64(v)
		if strings.HasSuffix(k, "_online") {
			onlineSum += uv
			onlineMap[k] = uv
		} else {
			offlineSum += uv
			offlineMap[k] = uv
		}
	}

	unknownOnline := make([]string, 0)
	for k := range onlineMap {
		found := false
		for _, okk := range onlineOrder {
			if k == okk {
				found = true
				break
			}
		}
		if !found {
			unknownOnline = append(unknownOnline, k)
		}
	}
	sort.Strings(unknownOnline)
	onlineOrder = append(onlineOrder, unknownOnline...)

	unknownOffline := make([]string, 0)
	for k := range offlineMap {
		found := false
		for _, okk := range offlineOrder {
			if k == okk {
				found = true
				break
			}
		}
		if !found {
			unknownOffline = append(unknownOffline, k)
		}
	}
	sort.Strings(unknownOffline)
	offlineOrder = append(offlineOrder, unknownOffline...)

	formatUintThousands := func(v uint64) string {
		s := fmt.Sprintf("%d", v)
		n := len(s)
		if n <= 3 {
			return s
		}
		rem := n % 3
		if rem == 0 {
			rem = 3
		}
		res := s[:rem]
		for i := rem; i < n; i += 3 {
			res += "," + s[i:i+3]
		}
		return res
	}

	humanizeStatus := func(k string) string {
		switch k {
		case "active_online":
			return "Active (Online)"
		case "exiting_online":
			return "Exiting (Online)"
		case "slashing_online":
			return "Slashing (Online)"
		case "active_offline":
			return "Active (Offline)"
		case "exiting_offline":
			return "Exiting (Offline)"
		case "slashing_offline":
			return "Slashing (Offline)"
		case "pending_initialized":
			return "Pending (Initialized)"
		case "pending":
			return "Pending"
		case "deposited":
			return "Deposited"
		case "exited":
			return "Exited"
		case "slashed":
			return "Slashed"
		default:
			parts := strings.Split(k, "_")
			for i := range parts {
				if len(parts[i]) > 0 {
					parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
				}
			}
			return strings.Join(parts, " ")
		}
	}

	statusIconClass := func(k string) string {
		switch k {
		case "active_online":
			return "fas fa-check-circle text-success"
		case "exiting_online":
			return "fas fa-door-open text-success"
		case "slashing_online":
			return "fas fa-bolt text-success"
		case "active_offline":
			return "fas fa-times-circle text-danger"
		case "exiting_offline":
			return "fas fa-door-open text-danger"
		case "slashing_offline":
			return "fas fa-bolt text-danger"
		case "pending_initialized":
			return "fas fa-hourglass-start text-muted"
		case "pending":
			return "fas fa-hourglass-half text-muted"
		case "deposited":
			return "fas fa-coins text-warning"
		case "exited":
			return "fas fa-sign-out-alt text-secondary"
		case "slashed":
			return "fas fa-cut text-danger"
		default:
			return "fas fa-circle text-muted"
		}
	}

	buildTooltip := func(order []string, src map[string]uint64) template.HTMLAttr {
		lines := make([]string, 0, len(order))
		for _, k := range order {
			if c, ok := src[k]; ok && c > 0 {
				label := humanizeStatus(k)
				icon := statusIconClass(k)
				line := fmt.Sprintf("<i class=\"%s mr-1\" aria-hidden=\"true\"></i>%s: %s", template.HTMLEscapeString(icon), template.HTMLEscapeString(label), formatUintThousands(c))
				lines = append(lines, line)
			}
		}
		if len(lines) == 0 {
			return "No breakdown available"
		}
		return template.HTMLAttr("data-tippy-content='" + strings.Join(lines, "<br>") + "'")
	}

	return onlineSum, offlineSum, buildTooltip(onlineOrder, onlineMap), buildTooltip(offlineOrder, offlineMap)
}

// EntitySubEntitiesData serves DataTables JSON for the sub-entities table on the Entity Detail page.
func EntitySubEntitiesData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	// DataTables params
	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}
	if length > 25 {
		length = 25
	}

	vars := mux.Vars(r)
	entity := vars["entity"]
	// sub-entity path var is ignored for this table; it is always rendered for "-" in UI

	period := GetRequestedPeriod(r)

	total, err := db.CountSubEntities(entity, period)
	if err != nil {
		logger.WithError(err).WithField("entity", entity).Error("entity sub-entities: count failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	dtos, err := db.GetSubEntitiesPaginated(entity, period, int(length), int(start))
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{"entity": entity, "start": start, "length": length}).Error("entity sub-entities: select failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rows := make([][]interface{}, 0, len(dtos))
	for _, s := range dtos {
		link := fmt.Sprintf("<a href=\"/entity/%s/%s?period=%s\">%s</a>", url.PathEscape(entity), url.PathEscape(s.SubEntity), template.HTMLEscapeString(period), template.HTMLEscapeString(s.SubEntity))
		beaconscore := utils.FormatBeaconscore(s.Efficiency, true)
		netShare := utils.FormatPercentageWithPrecision(s.NetShare, 2) + "%"
		rows = append(rows, []interface{}{link, beaconscore, netShare})
	}

	resp := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    uint64(total),
		RecordsFiltered: uint64(total),
		Data:            rows,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.WithError(err).Error("entity sub-entities: encode json")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// EntityValidatorsData serves DataTables JSON for the validators table on the Entity Detail page.
func EntityValidatorsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	// DataTables params
	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	// start is ignored for keyset pagination but kept for compatibility with DataTables
	_, err = strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil && q.Get("start") != "" { // tolerate empty
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}
	if length > 25 {
		length = 25
	}

	vars := mux.Vars(r)
	entity := vars["entity"]
	subEntity := vars["subEntity"]
	if subEntity == "" {
		subEntity = "-"
	}
	period := GetRequestedPeriod(r)

	// Optional paging token for keyset pagination
	pagingToken := q.Get("pagingToken")
	var afterIndex *int
	if pagingToken != "" {
		if idx, decErr := decodeEntityValidatorsToken(pagingToken); decErr == nil {
			afterIndex = &idx
		} else {
			logger.WithError(decErr).WithField("token", pagingToken).Warn("entity validators: invalid pagingToken; falling back to first page")
		}
	}

	total, err := db.CountEntityValidators(entity, subEntity)
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{"entity": entity, "sub_entity": subEntity}).Error("entity validators: count failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	vr, err := db.GetEntityValidatorsByCursor(entity, subEntity, int(length), afterIndex)
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{"entity": entity, "sub_entity": subEntity, "after_index": derefInt(afterIndex), "length": length}).Error("entity validators: select failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	indices := make([]int, 0, len(vr))
	for _, r := range vr {
		indices = append(indices, r.Index)
	}
	eff, err := db.GetValidatorEfficienciesForPeriod(period, indices)
	if err != nil {
		logger.WithError(err).WithField("period", period).Warn("entity validators: efficiencies fetch failed")
	}

	rows := make([][]interface{}, 0, len(vr))
	minIndex := 0
	for i, r := range vr {
		if i == 0 || r.Index < minIndex {
			minIndex = r.Index
		}
		link := fmt.Sprintf("<a href=\"/validator/%d\">%d</a>", r.Index, r.Index)
		status := utils.FormatValidatorStatus(r.Status)
		beaconscore := utils.FormatBeaconscore(eff[r.Index], true)
		rows = append(rows, []interface{}{link, status, beaconscore})
	}

	// Build next-page token if we returned a full page
	nextToken := ""
	if len(vr) > 0 && uint64(len(vr)) == length {
		nextToken = encodeEntityValidatorsToken(minIndex)
	}

	resp := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    uint64(total),
		RecordsFiltered: uint64(total),
		Data:            rows,
		PageLength:      uint64(length),
		DisplayStart:    0,
		PagingToken:     nextToken,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.WithError(err).Error("entity validators: encode json")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ---- Helpers for entity validators keyset pagination ----

type entityValidatorsToken struct {
	After int `json:"after"`
}

func encodeEntityValidatorsToken(index int) string {
	b, err := json.Marshal(entityValidatorsToken{After: index})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeEntityValidatorsToken(token string) (int, error) {
	payload, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, err
	}
	var t entityValidatorsToken
	if err := json.Unmarshal(payload, &t); err != nil {
		return 0, err
	}
	return t.After, nil
}

func derefInt(p *int) interface{} {
	if p == nil {
		return nil
	}
	return *p
}
