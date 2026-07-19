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
	// CategoryFilter, when set to a job_category CODE (e.g. "academic"),
	// restricts every grade surface (view-2 section grid, view-3 client card,
	// report-card document) to jobs of that category — dropping same-origin jobs
	// of another category (e.g. deportment) that would otherwise render as
	// academic subjects (gate H2). The code is resolved to its id once per
	// request via ResolveCategoryID; jobs are then kept by KeepJobInCategory
	// (matching id OR NULL). Empty = no filter (today's behavior; service-admin,
	// which sets no options, is unaffected). Generic — the vertical code value is
	// supplied by the consuming app (school-admin sets "academic").
	CategoryFilter string
	// Document configures the report-card document render (the per-client
	// DOCX/PDF download). Zero value disables every document enrichment — the
	// download renders exactly as before (service-admin unaffected).
	Document DocumentOptions
	// ClientCard configures view-3 (the per-student client card) presentation —
	// a DEDICATED job-category banding option, NOT the section grid's Row (which
	// is occupied by client-attribute gender bands and gated behind the global
	// academic-only CategoryFilter that would gut category bands, codex §5 /
	// Q-R9-8). Zero value → today's flat client card, byte-identical
	// (service-admin, which sets no options, is unaffected).
	ClientCard ClientCardOptions
}

// DocumentOptions — app-configurable knobs for the report-card document
// render. Same generic-reference contract as the sibling option structs: the
// vertical code values are supplied by the consuming app, never by this
// package.
type DocumentOptions struct {
	// GroupCategoryFilter is the job_category CODE (e.g. "homeroom_deportment")
	// whose single job represents the client's GROUP band on the document — its
	// task assignee renders as the group lead ("Adviser") and its per-phase
	// summaries as the group conduct row. Empty = no group band (blank fields).
	GroupCategoryFilter string
	// ClientReferenceAttributeCode is the client_attributes.<code> whose value
	// prints as the client's reference number on the document identity line
	// (e.g. "lrn"). Empty = blank reference.
	ClientReferenceAttributeCode string
	// ClientAttributeCodes are the attribute CODES exposed generically to the
	// document as the {{client_attributes.<code>}} map (e.g. ["lrn","gender"]).
	// Each configured, non-empty, dot-free code is ALWAYS present in the map
	// (blank when the client has no value) so the placeholder never leaks.
	// Codes are per-workspace data supplied by the consuming app; the package
	// treats them as opaque attribute.code lookups (ResolveAttributeIDByCode).
	ClientAttributeCodes []string
	// TemplateVariant selects the EMBEDDED fallback template the document
	// handler renders when no operator-uploaded binding resolves.
	// "" (zero value) = the original v1 summary layout — tiers that set no
	// options keep their exact prior fallback. TemplateVariantBlock = the
	// block-layout artifact (per-subject blocks, cover/boundary/formation
	// pages). The artifact CONTENT is operator template material; only the
	// selection knob lives in code.
	TemplateVariant string
}

// TemplateVariantBlock selects the block-layout embedded template (the
// faithful per-subject-block document) as the no-binding fallback.
const TemplateVariantBlock = "block"

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
	// ColumnsByField — view-1 landing count-column axis (R9 W-A2). When set to
	// the job_category entity ref ("job_category"), the landing's single static
	// count column is replaced by ONE count column per ACTIVE job_category row
	// (ordered by sort_order NULLS LAST, name ASC), each cell carrying that
	// (section × category) subject count plus an eye deep-link into the
	// section's category view (?jc=<category id>). Zero value (service-admin)
	// or an unrecognized ref keeps today's static column set byte-identical;
	// even when set, a nil/empty/denied/failed category read degrades to the
	// SAME static columns (the two-layer degrade contract, plan §3.7). Generic —
	// the category display names are per-workspace job_category.name DATA.
	ColumnsByField string
	// ScopeByServicingGrant, when true, confines the section landing to the
	// sections the ACTING principal holds an active servicing grant
	// (subscription_group_workspace_user, sgwu) on — fail-closed section
	// visibility (a principal with no grant sees zero sections). A principal
	// holding the workspace:list capability (operator/superadmin) BYPASSES the
	// filter and sees every section. Empty/false = today's unscoped landing
	// (service-admin, zero-valued, is unaffected). Generic: the grant family is
	// the cross-vertical `*_workspace_user` ACCESS axis (visibility resolver),
	// distinct from the delivery/StaffScope row axis that already scopes grades.
	ScopeByServicingGrant bool
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

// ClientCardOptions — view-3 (per-student client card) row presentation. A
// DEDICATED banding knob so the card can group its subject rows into
// job-category bands independently of the section grid's Row (gender) bands and
// the global academic-only CategoryFilter — the section's Row string cannot
// serve both, and CategoryFilter would drop every non-academic job before bands
// could form (codex §5 / Q-R9-8). Zero value → flat card. Generic — the band
// titles are per-workspace job_category.name DATA, never code vocabulary.
type ClientCardOptions struct {
	// Row.GroupByField, when set to the job_category entity ref ("job_category"),
	// groups the card's subject rows into one native TableRowGroup band per
	// distinct job_category (band title = job_category.name DATA, ordered by the
	// category sort contract; a NULL/foreign effective category folds into a
	// single trailing Uncategorized band, never dropped/duplicated). Any other
	// value → flat rows. Reuses the section's RowOptions shape for grammar
	// symmetry (Options.Row vs Options.ClientCard.Row); only GroupByField is
	// consulted here — the ordering follows the category's own sort_order, not
	// GroupValueOrder.
	Row RowOptions
	// IncludeAllCategories, when true (with banding on), LIFTS the card's H2
	// academic-only job filter FOR BANDING so same-origin deportment subjects
	// render under their own band. The lift is LOCAL to the card's own table:
	// the report-card DOCUMENT download and the section grid keep H2 (separate
	// fetches/handlers — the document view has its own job read + CategoryFilter).
	// Without it the card keeps H2 and, on academic-only seeded data, the
	// ≥2-category branch is unreachable (bands would never render). Zero value
	// (false) preserves today's H2-filtered card.
	IncludeAllCategories bool
}

// Entity reference constants (the implemented values).
const (
	// TabEntityPriceSchedule is the entity ref that turns each price_schedule
	// row into a view-1 tab.
	TabEntityPriceSchedule = "price_schedule"
	// ListEntitySubscriptionGroup is the single implemented List.Entity value.
	ListEntitySubscriptionGroup = "subscription_group"
	// ListColumnsJobCategory is the entity ref that turns each ACTIVE
	// job_category row into a view-1 landing count column (R9 W-A2).
	ListColumnsJobCategory = "job_category"
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

// CategoryColumns reports whether the landing should render one count column
// per active job_category (the single implemented ColumnsByField value). Any
// other value keeps the static column set (the fail-safe default).
func (l ListOptions) CategoryColumns() bool {
	return strings.TrimSpace(l.ColumnsByField) == ListColumnsJobCategory
}

// --- Client-card helpers -------------------------------------------------

// BandByCategory reports whether the client card should group its subject rows
// into job_category bands (the single implemented ClientCard.Row.GroupByField
// value — the shared job_category entity ref, the same const the landing count
// columns key on). Any other value keeps flat rows (the fail-safe default).
func (c ClientCardOptions) BandByCategory() bool {
	return strings.TrimSpace(c.Row.GroupByField) == ListColumnsJobCategory
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
