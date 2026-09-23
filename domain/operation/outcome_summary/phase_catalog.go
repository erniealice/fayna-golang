package outcome_summary

import (
	"sort"
	"strings"
)

// ClientReportPhaseEntry is the projection-neutral input to the explicit
// client-report phase catalog. Callers supply only phases belonging to the
// selected client's active jobs.
type ClientReportPhaseEntry struct {
	Code   string
	Name   string
	Order  int32
	Active bool
}

// ClientReportPhase is one distinct active template phase available for an
// explicit client report download.
type ClientReportPhase struct {
	Code  string
	Name  string
	Order int32
}

// BuildClientReportPhaseCatalog returns unique active phase codes in template
// order. If categories use the same code more than once, the first ordered
// entry supplies the display name and order. Sorting ties by name and then code
// keeps the result stable across projection row order.
func BuildClientReportPhaseCatalog(entries []ClientReportPhaseEntry) []ClientReportPhase {
	ordered := make([]ClientReportPhase, 0, len(entries))
	for _, entry := range entries {
		code := strings.TrimSpace(entry.Code)
		if !entry.Active || code == "" {
			continue
		}
		name := strings.TrimSpace(entry.Name)
		if name == "" {
			name = code
		}
		ordered = append(ordered, ClientReportPhase{Code: code, Name: name, Order: entry.Order})
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Order != ordered[j].Order {
			return ordered[i].Order < ordered[j].Order
		}
		if ordered[i].Name != ordered[j].Name {
			return ordered[i].Name < ordered[j].Name
		}
		return ordered[i].Code < ordered[j].Code
	})

	result := make([]ClientReportPhase, 0, len(ordered))
	seen := make(map[string]struct{}, len(ordered))
	for _, phase := range ordered {
		key := strings.ToLower(phase.Code)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, phase)
	}
	return result
}
