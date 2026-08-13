package dashboard

// widgets.go — per-widget data mapping from the aggregate response to the
// template payloads. All display strings resolve through lyngua keys
// (viewCtx.T); no hardcoded user-facing literals (label-audit V1–V4).

import (
	"fmt"
	"sort"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	home "github.com/erniealice/fayna-golang/domain/home"

	ocpb "github.com/erniealice/esqyma/pkg/schema/v1/service/dashboard/outcome_completion"
)

// Widget states (template dispatch).
const (
	StateFull   = "full"
	StateEmpty  = "empty"
	StateDenied = "denied"
	// StateUnavailable is the data/infrastructure-failure state (the read
	// errored for a NON-authorization reason): visible card + written reason,
	// distinct from both the denied card and the legitimate empty state.
	StateUnavailable = "unavailable"
)

// WidgetData is one widget card's resolved payload + designed state.
type WidgetData struct {
	Ref   string // home.WidgetRef string — testids home-widget-{ref}[-denied|-empty|-unavailable]
	State string // StateFull | StateEmpty | StateDenied | StateUnavailable

	Title    string
	Subtitle string

	// Denied state (Q4-A).
	DeniedText string
	DeniedHint string

	// Unavailable state (data/infrastructure read failure — T-9).
	UnavailableText string
	UnavailableHint string

	// Empty state text (per-widget lyngua key).
	EmptyText string

	Completion  *CompletionData
	Approval    *ApprovalData
	Attention   *AttentionData
	Pipeline    *PipelineData
	Performance *PerformanceData
	Links       []LinkData
}

// CompletionData is the completion_summary payload, windowed to the A2
// current period (the earliest phase_order with unrecorded work; when every
// slice is fully recorded, the latest one).
type CompletionData struct {
	ScopeText  string // "Assigned to you" / "Across the workspace" (tier-overlaid)
	PeriodText string // visible A2 window: tier-overlaid period label + phase_order (e.g. "Period 2")

	Recorded    string // windowed recorded-cell count
	Expected    string // windowed expected-cell count
	RatePct     string // windowed recorded/expected
	RecordedSub string // honest-scope unit line ("cells recorded", tier-overlaid)

	RecordedLabel string
	ExpectedLabel string
	RateLabel     string

	// ZeroText is set for the legitimate fresh-period zero state
	// (expected > 0, recorded = 0) — distinct from the widget-level empty
	// state (zero EXPECTED cells).
	ZeroText string
}

// ApprovalData is the approval_progress payload (series 2 + the ride-along
// not-started / returned buckets), windowed like the completion widget.
type ApprovalData struct {
	Rows []CountRow
}

// CountRow is one labelled count (text + value — never color-only).
type CountRow struct {
	Key   string // stable slug for testids: home-approval-{key}
	Label string
	Value string
}

// AttentionData is the admin-roster attention payload: categories ranked
// lowest completion first, plus one separately labelled operational-exceptions
// count (returned phase rows — the most actionable exception state).
type AttentionData struct {
	Subtitle        string
	PeriodText      string
	CategoryLabel   string
	RatioLabel      string
	RateLabel       string
	Rows            []AttentionRow
	ExceptionsLabel string // the operational-exceptions summary label
	ExceptionsValue string // returned-count across the workspace window
}

// AttentionRow is one ranked category row.
type AttentionRow struct {
	CategoryID string // testid: home-attention-{category_id}
	Name       string
	Ratio      string // "recorded / expected"
	RatePct    string

	// rateValue backs the lowest-completion-first sort (never rendered).
	rateValue float64
}

// LinkData is one visible quick link.
type LinkData struct {
	Label      string
	RouteKey   string
	ParamName  string
	ParamValue string
}

// PipelineData is the review_pipeline (Pulse) payload: the review ladder as an
// ordered funnel of forward stages plus the returned COUNT surfaced separately
// (20260809 H1 — a count, not a queue of rows). Same aggregate as the other
// data widgets.
type PipelineData struct {
	PeriodText    string     // visible window label
	Stages        []CountRow // not_started → in_progress → for_review → verified → published (funnel order)
	ReturnedLabel string
	ReturnedValue string
}

// PerformanceData is the performance_table payload: per-category completion
// rate scored at the common period (M3), ranked, against the app-supplied
// target (H1). TargetPct is empty when the app declared no target.
type PerformanceData struct {
	PeriodText    string
	TargetLabel   string
	TargetPct     string // "" = no target line
	CategoryLabel string
	RatioLabel    string
	RateLabel     string
	Rows          []PerformanceRow
}

// PerformanceRow is one ranked category row (worst-first, like attention).
type PerformanceRow struct {
	CategoryID  string // testid: home-performance-{category_id}
	Name        string
	Ratio       string // "recorded / expected"
	RatePct     string
	BelowTarget bool

	rateBP int32 // basis points, backs the ranking + target compare (never rendered)
}

// pickPeriodSlice implements the A2 structural window: the earliest
// phase_order slice with unrecorded work (recorded < expected); when every
// slice is fully recorded (or carries no expected cells), the latest slice.
// Slices arrive ordered by phase_order asc. Nil when the row has no slices.
func pickPeriodSlice(row *ocpb.OutcomeCompletionCategoryRow) *ocpb.OutcomeCompletionPeriodSlice {
	if row == nil || len(row.GetPeriodSlices()) == 0 {
		return nil
	}
	for _, s := range row.GetPeriodSlices() {
		if s.GetRecordedCells() < s.GetExpectedCells() {
			return s
		}
	}
	slices := row.GetPeriodSlices()
	return slices[len(slices)-1]
}

// pickCommonPeriodOrder resolves the ONE phase_order that cross-category views
// (performance_table, expanded attention) must share (20260809 M3): the window
// chosen from the All rollup. Ranking categories after selecting each row's own
// period independently would compare one category's period 1 against another's
// period 2, so every row is windowed to this single order instead. Returns
// (order, true) when the rollup has slices; (0, false) otherwise (callers then
// fall back to the row/rollup grain).
func pickCommonPeriodOrder(resp *ocpb.GetOutcomeCompletionSummaryResponse) (int32, bool) {
	if resp == nil {
		return 0, false
	}
	slice := pickPeriodSlice(resp.GetAllRollup())
	if slice == nil {
		return 0, false
	}
	return slice.GetPhaseOrder(), true
}

// sliceForOrder returns the row's slice at the given phase_order, or nil when
// the row has no slice for it (a category that does not run in that period).
func sliceForOrder(row *ocpb.OutcomeCompletionCategoryRow, order int32) *ocpb.OutcomeCompletionPeriodSlice {
	if row == nil {
		return nil
	}
	for _, s := range row.GetPeriodSlices() {
		if s.GetPhaseOrder() == order {
			return s
		}
	}
	return nil
}

// buildCompletionWidget maps the selected row to the completion payload.
func buildCompletionWidget(viewCtx *view.ViewContext, row *ocpb.OutcomeCompletionCategoryRow, slot int32) WidgetData {
	w := WidgetData{
		Ref:      string(home.WidgetCompletionSummary),
		Title:    viewCtx.T("home.widget.completion.title"),
		Subtitle: viewCtx.T("home.widget.completion.subtitle"),
	}

	slice := pickPeriodSlice(row)
	if row == nil || slice == nil || row.GetExpectedCells() == 0 {
		// Zero EXPECTED cells (or no data resolved) — the designed empty
		// state, distinct from the fresh-period zero state below.
		w.State = StateEmpty
		w.EmptyText = viewCtx.T("home.widget.completion.empty")
		return w
	}

	scopeKey := "home.widget.completion.scope_workspace"
	if slot == home.SlotStaff {
		scopeKey = "home.widget.completion.scope_self"
	}

	expected := slice.GetExpectedCells()
	recorded := slice.GetRecordedCells()

	w.State = StateFull
	w.Completion = &CompletionData{
		ScopeText:     viewCtx.T(scopeKey),
		PeriodText:    fmt.Sprintf("%s %d", viewCtx.T("home.period.label"), slice.GetPhaseOrder()),
		Recorded:      formatCount(recorded),
		Expected:      formatCount(expected),
		RatePct:       formatRate(recorded, expected),
		RecordedSub:   viewCtx.T("home.widget.completion.cells"),
		RecordedLabel: viewCtx.T("home.widget.completion.recorded"),
		ExpectedLabel: viewCtx.T("home.widget.completion.expected"),
		RateLabel:     viewCtx.T("home.widget.completion.rate"),
	}
	if expected > 0 && recorded == 0 {
		// The legitimate fresh-period 0% (Q3 empty-vs-zero distinction).
		w.Completion.ZeroText = viewCtx.T("home.widget.completion.zero")
	}
	return w
}

// buildApprovalWidget maps the selected row's windowed approval-ladder
// counts (six disjoint buckets incl. the ride-along not-started/returned).
func buildApprovalWidget(viewCtx *view.ViewContext, row *ocpb.OutcomeCompletionCategoryRow) WidgetData {
	w := WidgetData{
		Ref:   string(home.WidgetApprovalProgress),
		Title: viewCtx.T("home.widget.approval.title"),
	}

	var counts *ocpb.OutcomeCompletionApprovalCounts
	if slice := pickPeriodSlice(row); slice != nil {
		counts = slice.GetApprovalCounts()
		w.Subtitle = fmt.Sprintf("%s %d", viewCtx.T("home.period.label"), slice.GetPhaseOrder())
	} else if row != nil {
		counts = row.GetApprovalCounts()
	}

	total := counts.GetNotStarted() + counts.GetInProgress() + counts.GetForReview() +
		counts.GetVerified() + counts.GetPublished() + counts.GetReturned()
	if counts == nil || total == 0 {
		w.State = StateEmpty
		w.EmptyText = viewCtx.T("home.widget.approval.empty")
		return w
	}

	w.State = StateFull
	w.Approval = &ApprovalData{Rows: []CountRow{
		{Key: "not-started", Label: viewCtx.T("home.widget.approval.not_started"), Value: formatCount(counts.GetNotStarted())},
		{Key: "in-progress", Label: viewCtx.T("home.widget.approval.in_progress"), Value: formatCount(counts.GetInProgress())},
		{Key: "for-review", Label: viewCtx.T("home.widget.approval.for_review"), Value: formatCount(counts.GetForReview())},
		{Key: "verified", Label: viewCtx.T("home.widget.approval.verified"), Value: formatCount(counts.GetVerified())},
		{Key: "published", Label: viewCtx.T("home.widget.approval.published"), Value: formatCount(counts.GetPublished())},
		{Key: "returned", Label: viewCtx.T("home.widget.approval.returned"), Value: formatCount(counts.GetReturned())},
	}}
	return w
}

// buildAttentionWidget ranks the category rows lowest windowed completion
// first and appends the ONE operational-exceptions row (returned phase rows
// across the whole window — the ride-along amendment).
func buildAttentionWidget(viewCtx *view.ViewContext, resp *ocpb.GetOutcomeCompletionSummaryResponse) WidgetData {
	w := WidgetData{
		Ref:      string(home.WidgetAttention),
		Title:    viewCtx.T("home.widget.attention.title"),
		Subtitle: viewCtx.T("home.widget.attention.subtitle"),
	}

	var rows []AttentionRow
	periodText := ""
	var returnedTotal int64
	if resp != nil {
		// Common-period window (M3): every category row is scored at the SAME
		// phase_order (chosen from the All rollup) so the lowest-completion
		// ranking compares like periods; rows with no slice in that period are
		// skipped.
		commonOrder, haveCommon := pickCommonPeriodOrder(resp)
		if haveCommon {
			periodText = fmt.Sprintf("%s %d", viewCtx.T("home.period.label"), commonOrder)
		}
		for _, r := range resp.GetCategoryRows() {
			var slice *ocpb.OutcomeCompletionPeriodSlice
			if haveCommon {
				slice = sliceForOrder(r, commonOrder)
				if slice == nil {
					continue
				}
			}
			if !haveCommon {
				slice = pickPeriodSlice(r)
			}
			if slice == nil || slice.GetExpectedCells() == 0 {
				continue
			}
			rows = append(rows, AttentionRow{
				CategoryID: r.GetCategoryId(),
				Name:       r.GetCategoryName(),
				Ratio:      formatCount(slice.GetRecordedCells()) + " / " + formatCount(slice.GetExpectedCells()),
				RatePct:    formatRate(slice.GetRecordedCells(), slice.GetExpectedCells()),
				rateValue:  float64(slice.GetRecordedCells()) / float64(slice.GetExpectedCells()),
			})
		}
		if all := resp.GetAllRollup(); all != nil {
			if slice := pickPeriodSlice(all); slice != nil {
				returnedTotal = slice.GetApprovalCounts().GetReturned()
			} else {
				returnedTotal = all.GetApprovalCounts().GetReturned()
			}
		}
	}
	// Lowest completion first (stable insertion order for equal rates —
	// the response arrives sort_order-ordered).
	sortAttentionRows(rows)

	if len(rows) == 0 && returnedTotal == 0 {
		w.State = StateEmpty
		w.EmptyText = viewCtx.T("home.widget.attention.empty")
		return w
	}

	w.State = StateFull
	w.Attention = &AttentionData{
		Subtitle:        w.Subtitle,
		PeriodText:      periodText,
		CategoryLabel:   viewCtx.T("home.widget.attention.category"),
		RatioLabel:      viewCtx.T("home.widget.attention.ratio"),
		RateLabel:       viewCtx.T("home.widget.attention.rate"),
		Rows:            rows,
		ExceptionsLabel: viewCtx.T("home.widget.attention.exceptions"),
		ExceptionsValue: formatCount(returnedTotal),
	}
	return w
}

// sortAttentionRows orders by windowed completion rate ascending, stable
// (equal rates keep the response's job_category sort_order).
func sortAttentionRows(rows []AttentionRow) {
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].rateValue < rows[j].rateValue })
}

// buildReviewPipelineWidget (Pulse) renders the review ladder as an ordered
// funnel of forward stages plus the returned count, from the selected row's
// windowed approval counts (the SAME aggregate — no new backend, 20260809 H2).
func buildReviewPipelineWidget(viewCtx *view.ViewContext, row *ocpb.OutcomeCompletionCategoryRow) WidgetData {
	w := WidgetData{
		Ref:      string(home.WidgetReviewPipeline),
		Title:    viewCtx.T("home.widget.pipeline.title"),
		Subtitle: viewCtx.T("home.widget.pipeline.subtitle"),
	}

	var counts *ocpb.OutcomeCompletionApprovalCounts
	periodText := ""
	if slice := pickPeriodSlice(row); slice != nil {
		counts = slice.GetApprovalCounts()
		periodText = fmt.Sprintf("%s %d", viewCtx.T("home.period.label"), slice.GetPhaseOrder())
	} else if row != nil {
		counts = row.GetApprovalCounts()
	}

	total := counts.GetNotStarted() + counts.GetInProgress() + counts.GetForReview() +
		counts.GetVerified() + counts.GetPublished() + counts.GetReturned()
	if counts == nil || total == 0 {
		w.State = StateEmpty
		w.EmptyText = viewCtx.T("home.widget.pipeline.empty")
		return w
	}

	w.State = StateFull
	w.Pipeline = &PipelineData{
		PeriodText: periodText,
		Stages: []CountRow{
			{Key: "not-started", Label: viewCtx.T("home.widget.approval.not_started"), Value: formatCount(counts.GetNotStarted())},
			{Key: "in-progress", Label: viewCtx.T("home.widget.approval.in_progress"), Value: formatCount(counts.GetInProgress())},
			{Key: "for-review", Label: viewCtx.T("home.widget.approval.for_review"), Value: formatCount(counts.GetForReview())},
			{Key: "verified", Label: viewCtx.T("home.widget.approval.verified"), Value: formatCount(counts.GetVerified())},
			{Key: "published", Label: viewCtx.T("home.widget.approval.published"), Value: formatCount(counts.GetPublished())},
		},
		ReturnedLabel: viewCtx.T("home.widget.pipeline.returned_queue"),
		ReturnedValue: formatCount(counts.GetReturned()),
	}
	return w
}

// buildPerformanceWidget (Performance) ranks every category's completion rate
// at the common period (M3) against the app-supplied target (H1). Worst-first,
// mirroring the attention ranking.
func buildPerformanceWidget(viewCtx *view.ViewContext, resp *ocpb.GetOutcomeCompletionSummaryResponse, targetBP int32) WidgetData {
	w := WidgetData{
		Ref:      string(home.WidgetPerformanceTable),
		Title:    viewCtx.T("home.widget.performance.title"),
		Subtitle: viewCtx.T("home.widget.performance.subtitle"),
	}

	var rows []PerformanceRow
	periodText := ""
	if resp != nil {
		commonOrder, haveCommon := pickCommonPeriodOrder(resp)
		if haveCommon {
			periodText = fmt.Sprintf("%s %d", viewCtx.T("home.period.label"), commonOrder)
		}
		for _, r := range resp.GetCategoryRows() {
			var slice *ocpb.OutcomeCompletionPeriodSlice
			if haveCommon {
				slice = sliceForOrder(r, commonOrder)
				if slice == nil {
					continue
				}
			}
			if !haveCommon {
				slice = pickPeriodSlice(r)
			}
			if slice == nil || slice.GetExpectedCells() == 0 {
				continue
			}
			bp := int32(slice.GetRecordedCells() * 10000 / slice.GetExpectedCells())
			rows = append(rows, PerformanceRow{
				CategoryID:  r.GetCategoryId(),
				Name:        r.GetCategoryName(),
				Ratio:       formatCount(slice.GetRecordedCells()) + " / " + formatCount(slice.GetExpectedCells()),
				RatePct:     formatRate(slice.GetRecordedCells(), slice.GetExpectedCells()),
				BelowTarget: targetBP > 0 && bp < targetBP,
				rateBP:      bp,
			})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].rateBP < rows[j].rateBP })

	if len(rows) == 0 {
		w.State = StateEmpty
		w.EmptyText = viewCtx.T("home.widget.performance.empty")
		return w
	}

	w.State = StateFull
	w.Performance = &PerformanceData{
		PeriodText:    periodText,
		TargetLabel:   viewCtx.T("home.widget.performance.target"),
		TargetPct:     formatTargetBP(targetBP),
		CategoryLabel: viewCtx.T("home.widget.performance.category"),
		RatioLabel:    viewCtx.T("home.widget.performance.ratio"),
		RateLabel:     viewCtx.T("home.widget.performance.rate"),
		Rows:          rows,
	}
	return w
}

// formatTargetBP renders an app-supplied target (basis points) as a percent
// string; zero (no declared target) renders "" so the template omits the line.
func formatTargetBP(bp int32) string {
	if bp <= 0 {
		return ""
	}
	return fmt.Sprintf("%.1f%%", float64(bp)/100)
}

// buildQuickLinksWidget filters the declared links by the session permission
// set (empty Permission renders ungated — the sidebar-filter mirror; a nil
// permission set fails closed for gated links).
func buildQuickLinksWidget(viewCtx *view.ViewContext, links []home.QuickLink, perms *types.UserPermissions) WidgetData {
	w := WidgetData{
		Ref:   string(home.WidgetQuickLinks),
		Title: viewCtx.T("home.widget.quick.title"),
	}
	for _, l := range links {
		if l.RouteKey == "" || l.Label == "" {
			continue // fail-safe skip of half-declared links
		}
		if l.Permission != "" && !perms.HasCode(l.Permission) {
			continue
		}
		w.Links = append(w.Links, LinkData{
			Label:      l.Label,
			RouteKey:   l.RouteKey,
			ParamName:  l.ParamName,
			ParamValue: l.ParamValue,
		})
	}
	// quick_links performs no data read, so it has no denied/empty data
	// states — with zero visible links the card renders its title with an
	// empty body (links are individually permission-filtered, never faked).
	w.State = StateFull
	return w
}

// formatRate renders recorded/expected as a percentage ("94.5%"); zero
// expected renders an em-dash-free "0.0%" guard never divides by zero.
func formatRate(recorded, expected int64) string {
	if expected <= 0 {
		return "0.0%"
	}
	return fmt.Sprintf("%.1f%%", float64(recorded)/float64(expected)*100)
}
