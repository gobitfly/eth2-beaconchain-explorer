package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/templates"
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
	// Row for sub-entities list in the view model
	type SubEntityRow struct {
		SubEntity   string  `db:"sub_entity"`
		NetShare    float64 `db:"net_share"`
		Beaconscore float64 `db:"efficiency"`
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

		// List of sub-entities for the selected entity, sorted by net share desc
		SubEntities []SubEntityRow

		// Pagination for sub-entities table
		Page       int
		TotalPages int

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

	// Pagination params for sub-entities table
	const pageSize = 25
	page := 1
	if pStr := strings.TrimSpace(r.URL.Query().Get("page")); pStr != "" {
		if pv, err := strconv.Atoi(pStr); err == nil && pv > 0 {
			page = pv
		}
	}

	// Count total sub-entities for this entity and period
	totalSubs, err := db.CountSubEntities(entity, period)
	if err != nil {
		logger.WithError(err).WithField("entity", entity).Error("EntityDetail: count sub-entities failed")
	}
	// Compute pagination bounds
	totalPages := (totalSubs + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	offset := (page - 1) * pageSize

	// Load paginated sub-entities for the selected entity and period (sorted by net share desc)
	subDtos, err := db.GetSubEntitiesPaginated(entity, period, pageSize, offset)
	if err != nil {
		logger.WithError(err).WithField("entity", entity).WithField("page", page).Error("EntityDetail: load sub-entities (paged) failed")
	} else {
		subRows := make([]SubEntityRow, 0, len(subDtos))
		for _, s := range subDtos {
			subRows = append(subRows, SubEntityRow{SubEntity: s.SubEntity, NetShare: s.NetShare, Beaconscore: s.Efficiency})
		}
		vm.SubEntities = subRows
	}
	vm.Page = page
	vm.TotalPages = totalPages

	// Derive online/offline validator counts from status_counts JSONB
	var statusCounts map[string]int64
	if len(row24.StatusCountsRaw) > 0 {
		_ = json.Unmarshal(row24.StatusCountsRaw, &statusCounts)
	} else {
		statusCounts = map[string]int64{}
	}

	// Helper: order keys for online/offline breakdowns
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

	// Append any unknown keys not in the desired order arrays
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

	// Number formatting with thousands separators (used in tooltip)
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

	// Build multi-line HTML tooltip content with icons and humanized labels (rendered with CSS white-space: pre-line)
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
		// Join using newline so tooltip shows line breaks via CSS white-space: pre-line
		return template.HTMLAttr("data-tippy-content='" + strings.Join(lines, "<br>") + "'")
	}

	vm.OnlineCount = onlineSum
	vm.OfflineCount = offlineSum
	vm.OnlineBreakdownHTML = buildTooltip(onlineOrder, onlineMap)
	vm.OfflineBreakdownHTML = buildTooltip(offlineOrder, offlineMap)

	data.Data = vm

	if handleTemplateError(w, r, "entity_detail.go", "EntityDetail", "Done", entityTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return
	}
}
