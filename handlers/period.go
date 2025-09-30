package handlers

import (
	"net/http"
	"sort"
	"strings"
)

// normalizePeriod returns a supported period string or the default "1d".
// Allowed values are: 1d, 7d, 30d, all
func normalizePeriod(raw string) string {
	p := strings.ToLower(strings.TrimSpace(raw))
	switch p {
	case "1d", "7d", "30d", "all":
		return p
	default:
		return "1d"
	}
}

// GetRequestedPeriod extracts the period query parameter from the request and normalizes it.
func GetRequestedPeriod(r *http.Request) string {
	return normalizePeriod(r.URL.Query().Get("period"))
}

// PeriodLink is a simple link model used by templates to render the period toggle.
type PeriodLink struct{ Label, Value, Href string }

// BuildPeriodToggleLinks creates period= links preserving other query params except period.
// It returns a slice ordered as 1d, 7d, 30d, all.
func BuildPeriodToggleLinks(r *http.Request, basePath string) []PeriodLink {
	labels := map[string]string{"1d": "1d", "7d": "7d", "30d": "30d", "all": "All time"}
	values := []string{"1d", "7d", "30d", "all"}

	// copy query and drop any existing period and page to reset pagination when switching
	q := r.URL.Query()
	q.Del("period")
	q.Del("page")

	links := make([]PeriodLink, 0, len(values))
	for _, v := range values {
		q.Set("period", v)
		// ensure deterministic order
		keys := make([]string, 0, len(q))
		for k := range q {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			for _, vv := range q[k] {
				parts = append(parts, k+"="+vv)
			}
		}
		href := basePath
		if len(parts) > 0 {
			href += "?" + strings.Join(parts, "&")
		}
		links = append(links, PeriodLink{Label: labels[v], Value: v, Href: href})
	}
	return links
}
