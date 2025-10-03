package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/templates"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/sirupsen/logrus"
)

// Entities renders an overview of staking entities using the precomputed datasets
func Entities(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles,
		"entities.html")
	var entitiesTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/entities", "Staking Entities Overview", templateFiles)

	period := GetRequestedPeriod(r)

	logger := logrus.StandardLogger().WithField("module", "entities")
	logger = logger.WithFields(logrus.Fields{"period": period})
	type entityRow struct {
		Entity     string  `db:"entity"`
		Efficiency float64 `db:"efficiency"`
		NetShare   float64 `db:"net_share"`
	}

	treeMapItems, err := db.GetEntitiesTreemapData(period)
	if err != nil {
		logger.WithError(err).Error("select treemap entity rows")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	treemapRows := make([]entityRow, 0, len(treeMapItems))
	for _, it := range treeMapItems {
		treemapRows = append(treemapRows, entityRow{Entity: it.Entity, Efficiency: it.Efficiency, NetShare: it.NetShare})
	}

	// 6) Attach to template data
	type pagePayload struct {
		Treemap     []entityRow
		Period      string
		PeriodLinks []PeriodLink
	}
	payload := pagePayload{
		Treemap:     treemapRows,
		Period:      period,
		PeriodLinks: BuildPeriodToggleLinks(r, "/entities"),
	}
	data.Data = payload

	if handleTemplateError(w, r, "entitiesOverview.go", "EntitiesOverview", "Done", entitiesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// EntitiesData returns JSON for DataTables server-side processing on the Entities page.
func EntitiesData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	logger := logrus.StandardLogger().WithField("module", "entities-data")

	q := r.URL.Query()

	// DataTables params
	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.WithError(err).Warn("invalid draw parameter")
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.WithError(err).Warn("invalid start parameter")
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.WithError(err).Warn("invalid length parameter")
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}
	if length > 25 {
		length = 25
	}

	period := GetRequestedPeriod(r)
	search := strings.TrimSpace(q.Get("search[value]"))

	// Total entities (entity rows only)
	total, err := db.CountEntities(period)
	if err != nil {
		logger.WithError(err).Error("count entities (total)")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch paged entities ordered by net_share DESC
	var list []db.EntitySummaryData
	if search != "" {
		filtered, err2 := db.CountEntitiesWithSearch(period, search)
		if err2 != nil {
			logger.WithError(err2).Error("count entities (filtered)")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// override total filtered count
		totalFiltered := uint64(filtered)

		list, err = db.GetEntitiesPagedWithSearch(period, search, int(length), int(start))
		if err != nil {
			logger.WithError(err).Error("select paged entities (search)")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// build rows
		rows := buildEntitiesRows(list, period)
		resp := &types.DataTableResponse{
			Draw:            draw,
			RecordsTotal:    uint64(total),
			RecordsFiltered: totalFiltered,
			Data:            rows,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.WithError(err).Error("encode json (search)")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	list, err = db.GetEntitiesPaged(period, int(length), int(start))
	if err != nil {
		logger.WithError(err).Error("select paged entities (no search)")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rows := buildEntitiesRows(list, period)
	resp := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    uint64(total),
		RecordsFiltered: uint64(total),
		Data:            rows,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.WithError(err).Error("encode json")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// buildEntitiesRows converts a slice of EntitySummaryData into DataTables rows for the UI.
// Columns: [Entity link, Sub-entity count (capped display), Beaconscore, Net share]
func buildEntitiesRows(entities []db.EntitySummaryData, period string) [][]interface{} {
	rows := make([][]interface{}, 0, len(entities))
	if len(entities) == 0 {
		return rows
	}
	// Fetch sub-entity counts for the current page in one query
	names := make([]string, 0, len(entities))
	for _, e := range entities {
		names = append(names, e.Entity)
	}
	counts, err := db.GetSubEntityCountsForEntities(names, period)
	if err != nil {
		// On error, fall back to zeros but log it.
		logrus.StandardLogger().WithField("module", "entities-data").WithError(err).Warn("failed to fetch sub-entity counts; defaulting to 0")
		counts = map[string]int{}
	}
	for _, e := range entities {
		entityEsc := html.EscapeString(e.Entity)
		link := fmt.Sprintf("<a href=\"/entity/%s/-\">%s</a>", entityEsc, entityEsc)
		cnt := counts[e.Entity]
		cntStr := fmt.Sprintf("%d", cnt)
		if cnt > 100 {
			cntStr = "100+"
		}
		beaconscore := utils.FormatBeaconscore(e.Efficiency, false)
		netShare := utils.FormatPercentageWithPrecision(e.NetShare, 2) + "%"
		rows = append(rows, []interface{}{link, cntStr, beaconscore, netShare})
	}
	return rows
}
