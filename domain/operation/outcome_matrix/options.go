package outcome_matrix

import "strings"

// Options — app-configurable row presentation for the matrix view, set by the
// consuming app through the block's EngineBlock option (view option block).
// Each field names a row-level data source using the generic reference form
// "client_attributes.<code>" (the entity-attribute family; <code> is
// attribute.code). An empty field disables that behavior; an unrecognized
// reference is ignored fail-safe — the grid renders exactly as without it.
type Options struct {
	// RowSortField orders the roster rows by the referenced attribute value
	// (rows without a value last), then by row label.
	RowSortField string
	// RowSortDirection is "asc"|"desc" (default "" == "asc"). Added for grammar
	// symmetry with outcome_summary.Options; the empty default preserves the
	// current ascending behavior exactly (see applyRowOptions).
	RowSortDirection string
	// RowDescriptionField renders the referenced attribute value as the
	// secondary line under each row's label.
	RowDescriptionField string
	// RowGroupByField partitions the roster into labeled band groups by the
	// referenced attribute value (bands ordered ascending by value, the
	// no-value band last; rows keep the sort order within a band).
	RowGroupByField string
}

// ClientAttributeFieldPrefix is the reference-form prefix for the entity
// client-attribute family.
const ClientAttributeFieldPrefix = "client_attributes."

// ClientAttributeCode extracts <code> from a "client_attributes.<code>"
// reference. ok is false for an empty or foreign-form reference.
func ClientAttributeCode(field string) (code string, ok bool) {
	rest, found := strings.CutPrefix(field, ClientAttributeFieldPrefix)
	if !found || rest == "" {
		return "", false
	}
	return rest, true
}

// RowDirection returns the normalized row sort direction ("asc"|"desc"). The
// empty default is "asc" — identical to the pre-symmetry behavior.
func (o Options) RowDirection() string {
	if strings.EqualFold(strings.TrimSpace(o.RowSortDirection), "desc") {
		return "desc"
	}
	return "asc"
}

// AttributeCodes returns the distinct client-attribute codes referenced by
// the configured options, in first-use order (sort, description, group_by).
func (o Options) AttributeCodes() []string {
	var out []string
	seen := map[string]bool{}
	for _, f := range []string{o.RowSortField, o.RowDescriptionField, o.RowGroupByField} {
		if code, ok := ClientAttributeCode(f); ok && !seen[code] {
			seen[code] = true
			out = append(out, code)
		}
	}
	return out
}
