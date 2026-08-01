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
)

// KnownWidget reports whether ref names a widget this surface can render.
// Unknown refs are dropped by the view (fail-safe skip).
func KnownWidget(ref WidgetRef) bool {
	switch ref {
	case WidgetCompletionSummary, WidgetApprovalProgress, WidgetAttention, WidgetQuickLinks:
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
	case WidgetCompletionSummary, WidgetApprovalProgress, WidgetAttention:
		// The ONE dedicated aggregate capability (Q5/Q7 locks): the category
		// tabstrip and all three data widgets share this single read, so a
		// half-granted multi-code state is structurally impossible.
		return []string{"outcome_completion:read"}
	}
	return nil
}

// Variant is one persona slot's ordered widget roster.
type Variant struct {
	// Widgets is the ordered roster; unknown refs are skipped fail-safe.
	Widgets []WidgetRef
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
//     widgets (ties broken toward the numerically highest slot, which in the
//     shipped kind vocabulary is the least-privileged persona), so an
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
	// Deterministic scan order: highest slot first, so the fewest-widgets
	// comparison ties toward the least-privileged persona.
	sort.Sort(sort.Reverse(sort.IntSlice(slots)))
	best := Slot(int32(slots[0]))
	for _, s := range slots[1:] {
		if len(o.Variants[Slot(int32(s))].Widgets) < len(o.Variants[best].Widgets) {
			best = Slot(int32(s))
		}
	}
	return o.Variants[best], best, true
}

// KnownWidgets returns the variant's roster with unknown refs dropped
// (fail-safe skip), preserving declared order.
func (v Variant) KnownWidgets() []WidgetRef {
	out := make([]WidgetRef, 0, len(v.Widgets))
	for _, ref := range v.Widgets {
		if KnownWidget(ref) {
			out = append(out, ref)
		}
	}
	return out
}
