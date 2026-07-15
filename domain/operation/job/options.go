package job

import "strings"

// Options — app-configurable presentation for the job list surface ("/classes"
// in the education tier), set by the consuming app through the fayna EngineBlock
// option (block.WithJobListOptions). It mirrors outcome_summary.Options' Tab
// seam: a generic reference grammar with a fail-safe contract — an empty or
// unrecognized GroupByField disables the tabstrip and the list renders as the
// flat (backward-compatible) job list. service-admin leaves it zero-valued, so
// its jobs list is byte-unchanged.
//
// Reference forms (all generic — no vertical nouns):
//   - entity ref        "job_category"
//   - entity-field ref  "job_category.sort_order"
//   - direction         "asc" | "desc"  (default "asc")
type Options struct {
	// Tab configures the list tabstrip (one tab per job_category row).
	Tab TabOptions
}

// TabOptions — list tabstrip. GroupByField names the entity whose rows each
// become a tab (today: "job_category"). SortField is an entity-field ref
// ("job_category.sort_order") ordering the tabs; SortDirection is "asc"|"desc"
// (default asc). Same grammar as outcome_summary.TabOptions.
type TabOptions struct {
	GroupByField  string
	SortField     string
	SortDirection string
}

const (
	// TabEntityJobCategory is the entity ref that turns each job_category row
	// into a list tab.
	TabEntityJobCategory = "job_category"
	// jobCategorySortOrderField is the entity-field ref suffix that selects the
	// explicit sort_order tab ordering (NULLS LAST, name ASC fallback).
	jobCategorySortOrderField = "sort_order"
)

// entityField extracts <field> from an "<entity>.<field>" reference. ok is false
// for an empty or foreign-form reference.
func entityField(ref, entity string) (field string, ok bool) {
	rest, found := strings.CutPrefix(ref, entity+".")
	if !found || rest == "" {
		return "", false
	}
	return rest, true
}

// normalizeDirection maps a direction reference onto "asc"|"desc", defaulting to
// "asc" for an empty or unrecognized value (the fail-safe default).
func normalizeDirection(d string) string {
	if strings.EqualFold(strings.TrimSpace(d), "desc") {
		return "desc"
	}
	return "asc"
}

// Enabled reports whether the tabstrip is configured (GroupByField references
// the job_category entity). Any other value disables tabs (flat list).
func (t TabOptions) Enabled() bool {
	return strings.TrimSpace(t.GroupByField) == TabEntityJobCategory
}

// SortByOrder reports whether tabs are ordered by job_category.sort_order
// (NULLS LAST, name ASC fallback). False → order by name.
func (t TabOptions) SortByOrder() bool {
	field, ok := entityField(t.SortField, TabEntityJobCategory)
	return ok && field == jobCategorySortOrderField
}

// Direction returns the normalized tab sort direction ("asc"|"desc").
func (t TabOptions) Direction() string { return normalizeDirection(t.SortDirection) }
