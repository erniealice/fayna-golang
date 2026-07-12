package outcome_summary

import "strings"

// Options — app-configurable presentation for the outcome-summary surfaces,
// set by the consuming app through the block's EngineBlock option (the view
// option block). It is the level-2 sibling of outcome_matrix.Options: three
// nested sub-structs, one per concern, sharing the same GENERIC reference
// grammar and the same fail-safe contract — an empty field disables that
// behavior and an unrecognized reference is ignored (the surface renders
// exactly as without it, never an error).
//
// Reference forms (all generic — no vertical nouns):
//   - entity ref        "price_schedule", "subscription_group"
//   - entity-field ref  "price_schedule.name", "price_schedule.sort_order"
//   - attribute ref     "client_attributes.<code>"   (attribute.code)
//   - client-column ref "last_name"
//   - direction         "asc" | "desc"               (default "asc")
type Options struct {
	// Tab configures view-1's tabstrip (one tab per referenced entity row).
	Tab TabOptions
	// List configures WHAT view-1 lists.
	List ListOptions
	// Row configures view-2's row presentation (group bands + sort).
	Row RowOptions
}

// TabOptions — view-1 tabstrip. GroupByField names the entity whose rows each
// become a tab (today: "price_schedule"). SortField is an entity-field ref
// ("price_schedule.sort_order" | "price_schedule.name") ordering the tabs;
// SortDirection is "asc"|"desc" (default asc).
type TabOptions struct {
	GroupByField  string
	SortField     string
	SortDirection string
}

// ListOptions — view-1 "what to list". Entity is a single entity ref; only
// "subscription_group" is implemented today (any other value is logged and
// ignored by the view, which then renders its default flat list — Q-LIST-5).
type ListOptions struct {
	Entity string
}

// RowOptions — view-2 row presentation, same semantics as outcome_matrix rows.
// GroupByField partitions rows into value bands ("client_attributes.<code>");
// SortField orders rows within a band (a client-column ref, e.g. "last_name");
// SortDirection is "asc"|"desc" (default asc).
type RowOptions struct {
	GroupByField string
	// GroupValueOrder pins the band order to these attribute values
	// (case-insensitive; listed values lead in list order). Values not listed
	// follow in ascending value order; the no-value band stays last. Empty =
	// ascending value order (the fail-safe default).
	GroupValueOrder []string
	SortField       string
	SortDirection   string
}

// Entity reference constants (the implemented values).
const (
	// TabEntityPriceSchedule is the entity ref that turns each price_schedule
	// row into a view-1 tab.
	TabEntityPriceSchedule = "price_schedule"
	// ListEntitySubscriptionGroup is the single implemented List.Entity value.
	ListEntitySubscriptionGroup = "subscription_group"
	// PriceScheduleSortOrderField is the entity-field ref suffix that selects
	// the explicit sort_order tab ordering (NULLS LAST, name ASC fallback).
	priceScheduleSortOrderField = "sort_order"
)

// ClientAttributeFieldPrefix is the reference-form prefix for the entity
// client-attribute family (shared shape with outcome_matrix.Options).
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

// entityField extracts <field> from an "<entity>.<field>" reference. ok is
// false for an empty or foreign-form reference.
func entityField(ref, entity string) (field string, ok bool) {
	rest, found := strings.CutPrefix(ref, entity+".")
	if !found || rest == "" {
		return "", false
	}
	return rest, true
}

// normalizeDirection maps a direction reference onto "asc"|"desc", defaulting
// to "asc" for an empty or unrecognized value (the fail-safe default).
func normalizeDirection(d string) string {
	if strings.EqualFold(strings.TrimSpace(d), "desc") {
		return "desc"
	}
	return "asc"
}

// --- Tab helpers ---------------------------------------------------------

// Enabled reports whether the tabstrip is configured (GroupByField references
// the price_schedule entity). Any other value disables tabs (flat list).
func (t TabOptions) Enabled() bool {
	return strings.TrimSpace(t.GroupByField) == TabEntityPriceSchedule
}

// SortByOrder reports whether tabs are ordered by price_schedule.sort_order
// (NULLS LAST, name ASC fallback). False → order by name.
func (t TabOptions) SortByOrder() bool {
	field, ok := entityField(t.SortField, TabEntityPriceSchedule)
	return ok && field == priceScheduleSortOrderField
}

// Direction returns the normalized tab sort direction ("asc"|"desc").
func (t TabOptions) Direction() string { return normalizeDirection(t.SortDirection) }

// --- List helpers --------------------------------------------------------

// SubscriptionGroups reports whether view-1 should list subscription_groups
// (the single implemented List.Entity value).
func (l ListOptions) SubscriptionGroups() bool {
	return strings.TrimSpace(l.Entity) == ListEntitySubscriptionGroup
}

// --- Row helpers ---------------------------------------------------------

// Direction returns the normalized row sort direction ("asc"|"desc").
func (r RowOptions) Direction() string { return normalizeDirection(r.SortDirection) }

// GroupValueRank returns the configured position of a band value within
// GroupValueOrder (case-insensitive, trimmed) and whether the value was
// listed. Unlisted values report ok=false and sort after every listed one.
func (r RowOptions) GroupValueRank(value string) (int, bool) {
	v := strings.ToLower(strings.TrimSpace(value))
	for i, want := range r.GroupValueOrder {
		if strings.ToLower(strings.TrimSpace(want)) == v {
			return i, true
		}
	}
	return 0, false
}

// AttributeCodes returns the distinct client-attribute codes referenced by the
// Row options, in first-use order (group_by, sort). SortField is normally a
// plain client column ("last_name"), so this usually yields just the group_by
// code — but an attribute-form SortField is honored too (forward-safe).
func (o Options) AttributeCodes() []string {
	var out []string
	seen := map[string]bool{}
	for _, f := range []string{o.Row.GroupByField, o.Row.SortField} {
		if code, ok := ClientAttributeCode(f); ok && !seen[code] {
			seen[code] = true
			out = append(out, code)
		}
	}
	return out
}
