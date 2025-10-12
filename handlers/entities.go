package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/templates"
	"github.com/sirupsen/logrus"
)

// Entities renders an overview of staking entities using the precomputed 24h dataset
// with server-side pagination and optional prefix search.
// - Data source: validator_entities_data_24h
// - Fixed page size: 15
// - Search: case-insensitive prefix on entity or sub_entity (ILIKE q%)
// - When searching: only matching sub-entities are included under each entity
// - Treemap: always reflects the full dataset (entity rows only), independent of pagination/search
func Entities(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles,
		"entities.html")
	var entitiesTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/entities", "Staking Entities Overview", templateFiles)

	const pageSize = 15
	pageParam := strings.TrimSpace(r.URL.Query().Get("page"))
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	period := GetRequestedPeriod(r)
	page := 1
	if pageParam != "" {
		if v, err := strconv.Atoi(pageParam); err == nil && v > 0 {
			page = v
		}
	}

	logger := logrus.StandardLogger().WithField("module", "entities")
	logger = logger.WithFields(logrus.Fields{"page": page, "q": q, "period": period})

	// Build search predicate components
	likeTerm := q // prefix match handled in SQL as (term || '%')
	isSearching := likeTerm != ""

	// 1) Count total matching entities (entity rows only)
	var totalEntities int
	var err error
	if isSearching {
		totalEntities, err = db.CountEntitiesWithSearch(period, likeTerm)
		if err != nil {
			logger.WithError(err).Error("count entities (search)")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		totalEntities, err = db.CountEntities(period)
		if err != nil {
			logger.WithError(err).Error("count entities (no search)")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	if totalEntities == 0 {
		totalEntities = 0
	}
	// clamp page to bounds
	totalPages := (totalEntities + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	offset := (page - 1) * pageSize

	// 2) Fetch current page of entity totals (entity row has sub_entity '-' or empty)
	type entityRow struct {
		Entity     string  `db:"entity"`
		Efficiency float64 `db:"efficiency"`
		NetShare   float64 `db:"net_share"`
	}
	pagedEntities := make([]entityRow, 0, pageSize)
	var entRows []db.EntitySummaryData
	if isSearching {
		entRows, err = db.GetEntitiesPagedWithSearch(period, likeTerm, pageSize, offset)
		if err != nil {
			logger.WithError(err).Error("select paged entities (search)")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		entRows, err = db.GetEntitiesPaged(period, pageSize, offset)
		if err != nil {
			logger.WithError(err).Error("select paged entities (no search)")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	for _, e := range entRows {
		pagedEntities = append(pagedEntities, entityRow{Entity: e.Entity, Efficiency: e.Efficiency, NetShare: e.NetShare})
	}

	// Collect entity names for sub-entity fetch
	entityNames := make([]string, 0, len(pagedEntities))
	entityIndex := make(map[string]entityRow, len(pagedEntities))
	for _, er := range pagedEntities {
		entityNames = append(entityNames, er.Entity)
		entityIndex[er.Entity] = er
	}

	// 3) Fetch sub-entities for the paged entities; when searching, include only matching sub-entities
	type row struct {
		Entity     string  `db:"entity"`
		SubEntity  string  `db:"sub_entity"`
		Efficiency float64 `db:"efficiency"`
		NetShare   float64 `db:"net_share"`
		Remark     string  // optional: used to render a truncation note in the template
	}
	subs := make([]row, 0, 256)
	if len(entityNames) > 0 {
		if isSearching {
			// Only sub-entities with prefix match
			subDtos, err2 := db.GetSubEntitiesForEntitiesWithSearch(entityNames, likeTerm, period)
			if err2 != nil {
				logger.WithError(err2).Error("select sub-entities (search)")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			for _, s := range subDtos {
				subs = append(subs, row{Entity: s.Entity, SubEntity: s.SubEntity, Efficiency: s.Efficiency, NetShare: s.NetShare})
			}
		} else {
			subDtos, err2 := db.GetSubEntitiesForEntities(entityNames, period)
			if err2 != nil {
				logger.WithError(err2).Error("select sub-entities (no search)")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			for _, s := range subDtos {
				subs = append(subs, row{Entity: s.Entity, SubEntity: s.SubEntity, Efficiency: s.Efficiency, NetShare: s.NetShare})
			}
		}
	}

	// 4) Build flattened view rows (entity row + its sub-entities already sorted)
	//    Cap sub-entities per entity at 100 and add a remark if truncated.
	const maxSubEntities = 100
	// Group sub-entities by entity to count and truncate efficiently
	entitySubs := make(map[string][]row, len(entityNames))
	for _, s := range subs {
		entitySubs[s.Entity] = append(entitySubs[s.Entity], s)
	}

	view := make([]row, 0, len(pagedEntities)+len(subs))
	for _, er := range pagedEntities {
		// Append the entity header row (no SubEntity)
		view = append(view, row{Entity: er.Entity, SubEntity: "", Efficiency: er.Efficiency, NetShare: er.NetShare})

		subList := entitySubs[er.Entity]
		if len(subList) == 0 {
			continue
		}

		// Append up to maxSubEntities
		limit := len(subList)
		if limit > maxSubEntities {
			limit = maxSubEntities
		}
		for i := 0; i < limit; i++ {
			view = append(view, subList[i])
		}

		// If truncated, append a remark row
		if len(subList) > maxSubEntities {
			logger.WithFields(logrus.Fields{"entity": er.Entity, "sub_entities_total": len(subList), "capped_at": maxSubEntities}).Info("entities: capped sub-entities")
			view = append(view, row{Entity: er.Entity, Remark: "More than 100 sub entities. Use search to filter."})
		}
	}

	// 5) Treemap: load full dataset (entity rows only) to preserve behavior
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
		Rows          []row
		Page          int
		TotalPages    int
		TotalEntities int
		Query         string
		Treemap       []entityRow
		Period        string
		PeriodLinks   []PeriodLink
	}
	payload := pagePayload{
		Rows:          view,
		Page:          page,
		TotalPages:    totalPages,
		TotalEntities: totalEntities,
		Query:         q,
		Treemap:       treemapRows,
		Period:        period,
		PeriodLinks:   BuildPeriodToggleLinks(r, "/entities"),
	}
	data.Data = payload

	if handleTemplateError(w, r, "entitiesOverview.go", "EntitiesOverview", "Done", entitiesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
