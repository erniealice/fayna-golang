// Package home is the generic, composition-configurable home dashboard block
// surface (docs/plan/20260801-persona-home-dashboard, Phase 4).
//
// The block follows the "apps declare, blocks render" typed-option philosophy
// (the job.Options / outcome_summary.Options seam): the consuming app declares
// per-persona-slot widget rosters through block.WithHomeOptions; the zero
// value registers NOTHING (inert), so an app that does not configure it
// (service-admin) keeps its app-local home module byte-identical (Q6-B lock).
//
// NAMING: every identifier here is generic. Vertical vocabulary ("grading",
// "classes", "semester") enters ONLY through lyngua tier overlays — never a Go
// identifier, route, CSS class, or testid (universal-job-model C1).
package home

import (
	"sort"
	"strings"
)

// Slot is a persona slot keyed by principal kind. It mirrors the esqyma
// domain.entity.v1.PrincipalType integer values as re-declared by pyeza
// render/types.go:33-43 — fayna re-declares the constants and NEVER imports
// espyna/shared/identity (the dependency-direction rule; persona arrives via
// a composition-injected closure).
type Slot = int32

// Principal-kind slot constants (the subset this surface defines behavior
// for; other kinds resolve through Options.VariantFor's fail-closed path).
const (
	SlotOperatorOwner Slot = 1
	SlotOperatorStaff Slot = 2
	SlotStaff         Slot = 7
)

// WidgetRef names one widget in a persona variant's ordered roster. Unknown
// refs are skipped fail-safe (never half-rendered) — the job.Options
// unrecognized-reference grammar.
type WidgetRef string

// The known widget roster.
const (
	WidgetCompletionSummary WidgetRef = "completion_summary"
	WidgetApprovalProgress  WidgetRef = "approval_progress"
	WidgetAttention         WidgetRef = "attention"
	WidgetQuickLinks        WidgetRef = "quick_links"

	// WidgetReviewPipeline renders the review-status buckets as an ordered
	// pipeline (funnel) plus the returned COUNT — the Pulse section framing.
	// It reads the SAME aggregate as the other data widgets (no new backend;
	// the response has a returned count, not a queue of rows — 20260809 D1/H1).
	WidgetReviewPipeline WidgetRef = "review_pipeline"
	// WidgetPerformanceTable renders per-category completion rate ranked
	// against an app-supplied target — the Performance section framing. The
	// target is app policy passed as basis points; this package never sets it.
	WidgetPerformanceTable WidgetRef = "performance_table"
)

// KnownWidget reports whether ref names a widget this surface can render.
// Unknown refs are dropped by the view (fail-safe skip).
func KnownWidget(ref WidgetRef) bool {
	switch ref {
	case WidgetCompletionSummary, WidgetApprovalProgress, WidgetAttention, WidgetQuickLinks,
		WidgetReviewPipeline, WidgetPerformanceTable:
		return true
	}
	return false
}

// WidgetBundle returns the permission-code bundle a widget's data read
// requires. Per the Q4 all-or-nothing lock a widget renders data chrome ONLY
// when its ENTIRE bundle resolves; otherwise it renders the designed,
// visible, labelled denied state (never chrome-over-0-of-0, never silently
// absent). quick_links performs no data read — its bundle is empty and each
// link is individually gated by its own target permission (the sidebar
// permission-filter mirror).
func WidgetBundle(ref WidgetRef) []string {
	switch ref {
	case WidgetCompletionSummary, WidgetApprovalProgress, WidgetAttention,
		WidgetReviewPipeline, WidgetPerformanceTable:
		// The ONE dedicated aggregate capability (Q5/Q7 locks): the category
		// tabstrip and ALL data widgets — including the routed-section Pulse
		// and Performance framings (20260809 H2) — share this single read, so
		// a half-granted multi-code state is structurally impossible AND a
		// section built only from a new widget still performs the gated read
		// and renders the denied/unavailable degrade (never an ungated blank).
		return []string{"outcome_completion:read"}
	}
	return nil
}

// SectionKey names one routed home surface (20260809 D2/D4). The four keys map
// to the /home/{key} routes; the first declared section is a variant's default
// (the /home redirect target and the fail-safe for an unknown section). Generic
// identifiers — no vertical vocabulary (C1).
type SectionKey string

const (
	SectionOverview    SectionKey = "overview"
	SectionPulse       SectionKey = "pulse"
	SectionPerformance SectionKey = "performance"
	SectionAttention   SectionKey = "attention"
)

// KnownSection reports whether key names a routed section this surface serves.
func KnownSection(key SectionKey) bool {
	switch key {
	case SectionOverview, SectionPulse, SectionPerformance, SectionAttention:
		return true
	}
	return false
}

// Section is one routed home surface: an ordered widget roster plus whether it
// carries the category tabstrip (20260809 D6: overview only).
type Section struct {
	Key     SectionKey
	Widgets []WidgetRef
	HasTabs bool
}

// KnownWidgets returns the section's roster with unknown refs dropped.
func (s Section) KnownWidgets() []WidgetRef { return knownWidgets(s.Widgets) }

// Variant is one persona slot's home surface. It KEEPS the legacy single-page
// Widgets roster (20260809 H3: fayna is a shared public module — the exported
// API is preserved) and ADDS the routed Sections axis. A variant that declares
// only Widgets resolves as a single implicit overview section (back-compat).
type Variant struct {
	// Widgets is the legacy single-surface ordered roster; unknown refs are
	// skipped fail-safe. When Sections is empty it becomes one overview section.
	Widgets []WidgetRef
	// Sections is the ordered routed-section roster (20260809). The first entry
	// is the variant's default section.
	Sections []Section
}

// EffectiveSections returns the variant's routed sections with unknown widgets
// dropped. Back-compat: a variant with only Widgets yields one implicit
// overview section carrying those widgets.
func (v Variant) EffectiveSections() []Section {
	if len(v.Sections) > 0 {
		out := make([]Section, 0, len(v.Sections))
		for _, s := range v.Sections {
			out = append(out, Section{Key: s.Key, Widgets: knownWidgets(s.Widgets), HasTabs: s.HasTabs})
		}
		return out
	}
	if len(v.Widgets) > 0 {
		// Back-compat: a legacy Widgets-only variant is the pre-20260809 single
		// page, which carried the category tabstrip — so the synthesized
		// overview section keeps HasTabs true (the tabs still gate on
		// Options.Tab.Enabled() at render).
		return []Section{{Key: SectionOverview, Widgets: knownWidgets(v.Widgets), HasTabs: true}}
	}
	return nil
}

// knownWidgets returns refs with unknown entries dropped, preserving order.
func knownWidgets(refs []WidgetRef) []WidgetRef {
	out := make([]WidgetRef, 0, len(refs))
	for _, ref := range refs {
		if KnownWidget(ref) {
			out = append(out, ref)
		}
	}
	return out
}

// TabOptions configures the dashboard tab axis (Q2 lock) with the SAME
// grammar as job.TabOptions: GroupByField "job_category" renders one tab per
// category row plus an All tab; "" (or any unrecognized reference) renders no
// tabs. Sorting of the category rows is adapter-side (job_category.sort_order
// asc — the aggregate response arrives pre-ordered), so SortField /
// SortDirection are accepted for grammar parity but the response order wins.
type TabOptions struct {
	GroupByField  string
	SortField     string
	SortDirection string
}

// TabEntityJobCategory is the entity ref that turns each category row of the
// aggregate response into a dashboard tab.
const TabEntityJobCategory = "job_category"

// Enabled reports whether the tabstrip is configured. Any value other than
// the known entity ref disables tabs (fail-safe).
func (t TabOptions) Enabled() bool {
	return strings.TrimSpace(t.GroupByField) == TabEntityJobCategory
}

// QuickLink is one declared quick-links row. Labels come from the consuming
// app's EXISTING lyngua-loaded sidebar/route label system (no new label keys
// — lyngua.md §L-2); hrefs resolve per-request through the {{route}} /
// {{routeWith}} template funcs so workspace-slug rewriting applies for free.
type QuickLink struct {
	// RouteKey is the merged-route-map key, e.g. "job.list".
	RouteKey string
	// ParamName/ParamValue optionally resolve one {param} placeholder in the
	// route pattern (e.g. "status" → "active"). Empty = plain {{route}}.
	ParamName  string
	ParamValue string
	// Label is the display text (sourced from the app's lyngua-loaded label
	// set at composition time — never a hardcoded literal in this package).
	Label string
	// Permission gates the link's visibility (HasCode on the session
	// permission set). Empty renders ungated — mirroring the sidebar
	// permission filter's `Permission == ""` passthrough.
	Permission string
}

// Options is the deployment-declared home-surface config, set by the
// consuming app through block.WithHomeOptions. The zero value means the
// block registers NOTHING (inert) — the fail-safe grammar contract.
type Options struct {
	// Variants maps persona slot → widget roster; empty = inert.
	Variants map[Slot]Variant

	// DefaultSlot is used when the session kind matches no Variants key
	// (including the unresolved kind-0 sentinel). Fail-closed: when unset or
	// absent from Variants, the LOWEST-capability declared variant renders
	// (see VariantFor).
	DefaultSlot Slot

	// Tab is the dashboard tab axis (Q2).
	Tab TabOptions

	// QuickLinks is the declared quick-links roster (rendered by variants
	// whose roster includes WidgetQuickLinks).
	QuickLinks []QuickLink

	// PerformanceTargetBasisPoints is the app-declared completion target the
	// performance_table widget ranks against (20260809 H1: the target is APP
	// policy, passed in basis points — 10000 = 100%; this package never sets a
	// default). Zero renders the table with no target line.
	PerformanceTargetBasisPoints int32
}

// Enabled reports whether the app declared any variant. False = the block is
// inert: no routes, no templates, byte-identical consumer behavior.
func (o Options) Enabled() bool {
	return len(o.Variants) > 0
}

// VariantFor resolves the session's principal kind to a declared variant,
// fail-closed:
//
//  1. an exact Variants[kind] match wins;
//  2. else Variants[DefaultSlot] when DefaultSlot is set and declared;
//  3. else the LOWEST-CAPABILITY declared variant — the one with the fewest
//     effective widgets (ties broken toward the numerically highest slot, which
//     in the shipped kind vocabulary is the least-privileged persona), so an
//     unresolved kind can never see a roster broader than the narrowest one
//     the app declared.
//
// ok is false only when no variant is declared at all (the inert case — the
// block never mounts then, so views treat it as unreachable-but-safe).
func (o Options) VariantFor(kind int32) (v Variant, slot Slot, ok bool) {
	if len(o.Variants) == 0 {
		return Variant{}, 0, false
	}
	if v, hit := o.Variants[kind]; hit {
		return v, kind, true
	}
	if o.DefaultSlot != 0 {
		if v, hit := o.Variants[o.DefaultSlot]; hit {
			return v, o.DefaultSlot, true
		}
	}
	slots := make([]int, 0, len(o.Variants))
	for s := range o.Variants {
		slots = append(slots, int(s))
	}
	// Deterministic scan order: highest slot first, so equal capability ties
	// favor the least-privileged persona.
	sort.Sort(sort.Reverse(sort.IntSlice(slots)))
	best := Slot(int32(slots[0]))
	bestCap := sectionCapability(o.Variants[best])
	for _, s := range slots[1:] {
		candidate := Slot(int32(s))
		candidateCap := sectionCapability(o.Variants[candidate])
		if candidateCap.less(bestCap) {
			best = candidate
			bestCap = candidateCap
		}
	}
	return o.Variants[best], best, true
}

type widgetCapability struct {
	unique int
	total  int
}

func (a widgetCapability) less(b widgetCapability) bool {
	if a.unique != b.unique {
		return a.unique < b.unique
	}
	return a.total < b.total
}

func sectionCapability(v Variant) widgetCapability {
	sections := v.EffectiveSections()
	seen := make(map[WidgetRef]struct{}, 4)
	total := 0
	for _, section := range sections {
		for _, ref := range section.Widgets {
			total++
			seen[ref] = struct{}{}
		}
	}
	return widgetCapability{unique: len(seen), total: total}
}

// KnownWidgets returns the variant's legacy roster with unknown refs dropped
// (fail-safe skip), preserving declared order.
func (v Variant) KnownWidgets() []WidgetRef { return knownWidgets(v.Widgets) }

// SectionsFor resolves a session kind to its ordered routed sections
// (fail-closed via VariantFor). ok is false only in the inert case.
func (o Options) SectionsFor(kind int32) (sections []Section, slot Slot, ok bool) {
	v, slot, ok := o.VariantFor(kind)
	if !ok {
		return nil, 0, false
	}
	return v.EffectiveSections(), slot, true
}

// SectionFor resolves ONE declared section for a kind. When a programmatic
// caller supplies an absent or unknown key it fails SAFE to the variant's first
// (default) section. HTTP section routes remain explicitly registered, so an
// undeclared /home/{garbage} path never reaches this resolver and returns 404.
// ok is false only in the inert case (no variant declared).
func (o Options) SectionFor(kind int32, key SectionKey) (section Section, slot Slot, ok bool) {
	sections, slot, ok := o.SectionsFor(kind)
	if !ok || len(sections) == 0 {
		return Section{}, slot, ok && len(sections) > 0
	}
	for _, s := range sections {
		if s.Key == key {
			return s, slot, true
		}
	}
	return sections[0], slot, true
}
