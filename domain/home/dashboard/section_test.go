package dashboard

// section_test.go — routed-section behavior (docs/plan/20260809-persona-home-
// routed-sections): the /home redirect resolver, per-section roster resolution,
// the two new widget framings (review_pipeline, performance_table), the common-
// period ranking rule (codex M3), and the new widgets' deny/unavailable degrade
// (codex H2/AB1). Reuses testResponse/slice/withPerms/widgetByRef from
// page_test.go (same package).

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/erniealice/pyeza-golang/view"

	home "github.com/erniealice/fayna-golang/domain/home"

	ocpb "github.com/erniealice/esqyma/pkg/schema/v1/service/dashboard/outcome_completion"
)

const ocRead = "outcome_completion:read"

// sectionOptions is a sections-based Options fixture (the 20260809 shape): all
// four routed sections on one variant, overview alone carrying tabs.
func sectionOptions() home.Options {
	sections := []home.Section{
		{Key: home.SectionOverview, HasTabs: true, Widgets: []home.WidgetRef{
			home.WidgetCompletionSummary, home.WidgetApprovalProgress, home.WidgetQuickLinks,
		}},
		{Key: home.SectionPulse, Widgets: []home.WidgetRef{home.WidgetReviewPipeline}},
		{Key: home.SectionPerformance, Widgets: []home.WidgetRef{home.WidgetPerformanceTable}},
		{Key: home.SectionAttention, Widgets: []home.WidgetRef{home.WidgetAttention}},
	}
	roster := home.Variant{Sections: sections}
	return home.Options{
		Variants:                     map[home.Slot]home.Variant{home.SlotStaff: roster},
		DefaultSlot:                  home.SlotStaff,
		Tab:                          home.TabOptions{GroupByField: home.TabEntityJobCategory},
		PerformanceTargetBasisPoints: 8000,
	}
}

func secViewCtx(url string) *view.ViewContext {
	vc := testViewCtx(url)
	extra := map[string]string{
		"home.section.overview.title":         "Overview",
		"home.section.pulse.title":            "Operations Pulse",
		"home.section.performance.title":      "Performance",
		"home.section.attention.title":        "Needs Attention",
		"home.widget.pipeline.title":          "Review pipeline",
		"home.widget.pipeline.empty":          "No work in review yet",
		"home.widget.pipeline.returned_queue": "Returned for revision",
		"home.widget.performance.title":       "Completion vs target",
		"home.widget.performance.target":      "Target",
		"home.widget.performance.category":    "Category",
		"home.widget.performance.ratio":       "Recorded / expected",
		"home.widget.performance.rate":        "Completion",
		"home.widget.performance.empty":       "No categories to show yet",
		"home.widget.attention.category":      "Category",
		"home.widget.attention.ratio":         "Recorded / expected",
		"home.widget.attention.rate":          "Completion",
		"home.widget.approval.not_started":    "Not started",
		"home.widget.approval.in_progress":    "In progress",
		"home.widget.approval.for_review":     "For review",
		"home.widget.approval.verified":       "Verified",
		"home.widget.approval.published":      "Published",
	}
	for k, v := range extra {
		vc.Messages[k] = v
	}
	return vc
}

func sectionModule(resp *ocpb.GetOutcomeCompletionSummaryResponse, summaryErr error, overviewURL string) *Module {
	return NewModule(&Deps{
		Routes:  home.DefaultRoutes(),
		Options: sectionOptions(),
		GetCompletionSummary: func(context.Context) (*ocpb.GetOutcomeCompletionSummaryResponse, error) {
			return resp, summaryErr
		},
		ResolvePrincipalKind: func(context.Context) int32 { return home.SlotStaff },
		ResolveOverviewURL:   func(context.Context) string { return overviewURL },
		LogDeny:              func(context.Context, string, string) {},
	})
}

// rotationCommonPeriodResponse creates a response where the All rollup selects
// phase_order=2 (period 2), but one category lacks that slice entirely.
func rotationCommonPeriodResponse() *ocpb.GetOutcomeCompletionSummaryResponse {
	catA := &ocpb.OutcomeCompletionCategoryRow{
		CategoryId:    "cat-a",
		CategoryName:  "Alpha",
		ExpectedCells: 70,
		RecordedCells: 60,
		PeriodSlices: []*ocpb.OutcomeCompletionPeriodSlice{
			slice(1, 20, 20, nil),
			slice(3, 50, 40, nil),
		},
	}
	catB := &ocpb.OutcomeCompletionCategoryRow{
		CategoryId:    "cat-b",
		CategoryName:  "Beta",
		ExpectedCells: 90,
		RecordedCells: 70,
		PeriodSlices: []*ocpb.OutcomeCompletionPeriodSlice{
			slice(1, 20, 20, nil),
			slice(2, 20, 8, &ocpb.OutcomeCompletionApprovalCounts{NotStarted: 2}),
		},
	}

	return &ocpb.GetOutcomeCompletionSummaryResponse{
		Success:      true,
		CategoryRows: []*ocpb.OutcomeCompletionCategoryRow{catA, catB},
		AllRollup: &ocpb.OutcomeCompletionCategoryRow{
			ExpectedCells: 160,
			RecordedCells: 140,
			ApprovalCounts: &ocpb.OutcomeCompletionApprovalCounts{
				NotStarted: 4,
				Returned:   7,
			},
			PeriodSlices: []*ocpb.OutcomeCompletionPeriodSlice{
				slice(1, 100, 100, nil),
				slice(2, 40, 20, &ocpb.OutcomeCompletionApprovalCounts{NotStarted: 2, Returned: 7}),
			},
		},
	}
}

// A1 — /home redirects to the resolver-provided (workspace-qualified) overview
// URL, and falls back to the bare Routes.OverviewURL when unwired (codex M5).
func TestRedirectToOverviewUsesResolver(t *testing.T) {
	m := sectionModule(testResponse(), nil, "/w/mmis/home/overview")
	res := m.redirectToOverview().Handle(context.Background(), secViewCtx("/home"))
	if res.Redirect != "/w/mmis/home/overview" {
		t.Fatalf("redirect target: want workspace-qualified /w/mmis/home/overview, got %q", res.Redirect)
	}
	if res.StatusCode != http.StatusFound {
		t.Errorf("redirect status must be 302, got %d", res.StatusCode)
	}

	// Nil resolver → bare Routes.OverviewURL.
	bare := NewModule(&Deps{Routes: home.DefaultRoutes(), Options: sectionOptions(),
		ResolvePrincipalKind: func(context.Context) int32 { return home.SlotStaff }})
	res = bare.redirectToOverview().Handle(context.Background(), secViewCtx("/home"))
	if res.Redirect != "/home/overview" {
		t.Fatalf("nil-resolver fallback: want /home/overview, got %q", res.Redirect)
	}
}

// A2/A5 — each section resolves its own roster + section key; only overview has tabs.
func TestBuildSectionDataPerSection(t *testing.T) {
	m := sectionModule(testResponse(), nil, "/home/overview")
	ctx := withPerms(ocRead)
	cases := []struct {
		key           home.SectionKey
		wantTabs      bool
		wantWidgetRef home.WidgetRef
	}{
		{home.SectionOverview, true, home.WidgetCompletionSummary},
		{home.SectionPulse, false, home.WidgetReviewPipeline},
		{home.SectionPerformance, false, home.WidgetPerformanceTable},
		{home.SectionAttention, false, home.WidgetAttention},
	}
	for _, c := range cases {
		d := m.buildSectionData(ctx, secViewCtx("/home/"+string(c.key)), c.key)
		if d.Section != string(c.key) {
			t.Errorf("%s: data.Section = %q", c.key, d.Section)
		}
		if d.TabsEnabled != c.wantTabs {
			t.Errorf("%s: TabsEnabled = %v, want %v", c.key, d.TabsEnabled, c.wantTabs)
		}
		widgetByRef(t, d, c.wantWidgetRef) // must be present
	}
}

// A6 — pulse assembles the review pipeline (5 forward stages + returned count).
func TestPulsePipelineAssembly(t *testing.T) {
	m := sectionModule(testResponse(), nil, "/home/overview")
	d := m.buildSectionData(withPerms(ocRead), secViewCtx("/home/pulse"), home.SectionPulse)
	w := widgetByRef(t, d, home.WidgetReviewPipeline)
	if w.State != StateFull || w.Pipeline == nil {
		t.Fatalf("pulse: want full pipeline, got state=%q pipeline=%v", w.State, w.Pipeline)
	}
	if len(w.Pipeline.Stages) != 5 {
		t.Errorf("pipeline must have 5 forward stages, got %d", len(w.Pipeline.Stages))
	}
	// All-rollup period 2 has Returned:1 (testResponse).
	if w.Pipeline.ReturnedValue != "1" {
		t.Errorf("returned count: want 1, got %q", w.Pipeline.ReturnedValue)
	}
}

// A7 + M3 — performance ranks worst-first at the COMMON period and flags rows
// below the app target. At common period 2: cat-a = 0/40 (0%), cat-b = 20/25
// (80%). Worst-first ⇒ cat-a then cat-b. Target 8000bp (80%): cat-a below.
func TestPerformanceTableAssembly(t *testing.T) {
	m := sectionModule(testResponse(), nil, "/home/overview")
	d := m.buildSectionData(withPerms(ocRead), secViewCtx("/home/performance"), home.SectionPerformance)
	w := widgetByRef(t, d, home.WidgetPerformanceTable)
	if w.State != StateFull || w.Performance == nil {
		t.Fatalf("performance: want full table, got state=%q", w.State)
	}
	if len(w.Performance.Rows) != 2 {
		t.Fatalf("want 2 category rows, got %d", len(w.Performance.Rows))
	}
	if w.Performance.Rows[0].CategoryID != "cat-a" || w.Performance.Rows[1].CategoryID != "cat-b" {
		t.Errorf("ranking must be worst-first (cat-a, cat-b), got %s, %s",
			w.Performance.Rows[0].CategoryID, w.Performance.Rows[1].CategoryID)
	}
	if !w.Performance.Rows[0].BelowTarget {
		t.Error("cat-a (0%%) must be flagged below the 80%% target")
	}
	if w.Performance.Rows[1].BelowTarget {
		t.Error("cat-b (80%%) must NOT be flagged below the 80%% target")
	}
	if w.Performance.TargetPct == "" {
		t.Error("target line must render (app declared 8000bp)")
	}
}

// M3 invariance — when the common period is selected, both attention and
// performance must use that same phase_order and skip categories without it.
func TestAttentionAndPerformanceUseCommonPeriodWithoutMixing(t *testing.T) {
	m := sectionModule(rotationCommonPeriodResponse(), nil, "/home/overview")
	ctx := withPerms(ocRead)

	perf := m.buildSectionData(ctx, secViewCtx("/home/performance"), home.SectionPerformance)
	wPerf := widgetByRef(t, perf, home.WidgetPerformanceTable)
	if wPerf.State != StateFull || wPerf.Performance == nil {
		t.Fatalf("performance: want full table, got state=%q", wPerf.State)
	}
	if wPerf.Performance.PeriodText != "Period 2" {
		t.Fatalf("performance: common period should be 2, got %q", wPerf.Performance.PeriodText)
	}
	if len(wPerf.Performance.Rows) != 1 {
		t.Fatalf("performance: want only the category with period 2, got %d rows", len(wPerf.Performance.Rows))
	}
	if wPerf.Performance.PeriodText != "Period 2" {
		t.Fatalf("performance: common period should be 2, got %q", wPerf.Performance.PeriodText)
	}
	if wPerf.Performance.Rows[0].CategoryID != "cat-b" {
		t.Errorf("performance: expect only cat-b at period 2, got %s", wPerf.Performance.Rows[0].CategoryID)
	}
	if wPerf.Performance.Rows[0].Ratio != "8 / 20" {
		t.Errorf("performance: cat-b should use period 2 slice (8/20), got %q", wPerf.Performance.Rows[0].Ratio)
	}
	if wPerf.Performance.CategoryLabel != "Category" {
		t.Fatalf("performance: category header label must be explicit, got %q", wPerf.Performance.CategoryLabel)
	}
	if wPerf.Performance.RatioLabel != "Recorded / expected" {
		t.Fatalf("performance: ratio header label must be explicit, got %q", wPerf.Performance.RatioLabel)
	}
	if wPerf.Performance.RateLabel != "Completion" {
		t.Fatalf("performance: rate header label must be explicit, got %q", wPerf.Performance.RateLabel)
	}

	att := m.buildSectionData(ctx, secViewCtx("/home/attention"), home.SectionAttention)
	wAtt := widgetByRef(t, att, home.WidgetAttention)
	if wAtt.State != StateFull || wAtt.Attention == nil {
		t.Fatalf("attention: want full table, got state=%q", wAtt.State)
	}
	if wAtt.Attention.PeriodText != "Period 2" {
		t.Fatalf("attention: common period should be 2, got %q", wAtt.Attention.PeriodText)
	}
	if len(wAtt.Attention.Rows) != 1 {
		t.Fatalf("attention: want only the category with period 2, got %d rows", len(wAtt.Attention.Rows))
	}
	if wAtt.Attention.Rows[0].CategoryID != "cat-b" {
		t.Errorf("attention: expect only cat-b at period 2, got %s", wAtt.Attention.Rows[0].CategoryID)
	}
	if wAtt.Attention.Rows[0].Ratio != "8 / 20" {
		t.Errorf("attention: cat-b should use period 2 slice (8/20), got %q", wAtt.Attention.Rows[0].Ratio)
	}
	if wAtt.Attention.CategoryLabel != "Category" {
		t.Fatalf("attention: category header label must be explicit, got %q", wAtt.Attention.CategoryLabel)
	}
	if wAtt.Attention.RatioLabel != "Recorded / expected" {
		t.Fatalf("attention: ratio header label must be explicit, got %q", wAtt.Attention.RatioLabel)
	}
	if wAtt.Attention.RateLabel != "Completion" {
		t.Fatalf("attention: rate header label must be explicit, got %q", wAtt.Attention.RateLabel)
	}
}

// AB1 / H2 — the new widgets fire the gated read + degrade: nil perms → denied;
// a data error → unavailable (never a blank/ungated section).
func TestNewWidgetsDenyAndUnavailable(t *testing.T) {
	// Deny: no perms in ctx.
	m := sectionModule(testResponse(), nil, "/home/overview")
	for _, c := range []struct {
		key home.SectionKey
		ref home.WidgetRef
	}{{home.SectionPulse, home.WidgetReviewPipeline}, {home.SectionPerformance, home.WidgetPerformanceTable}} {
		d := m.buildSectionData(context.Background(), secViewCtx("/home/"+string(c.key)), c.key)
		w := widgetByRef(t, d, c.ref)
		if w.State != StateDenied {
			t.Errorf("%s under nil perms: want denied, got %q", c.ref, w.State)
		}
	}
	// Unavailable: read errors for a non-authz reason.
	mErr := sectionModule(nil, errors.New("pq: connection refused"), "/home/overview")
	d := mErr.buildSectionData(withPerms(ocRead), secViewCtx("/home/pulse"), home.SectionPulse)
	w := widgetByRef(t, d, home.WidgetReviewPipeline)
	if w.State != StateUnavailable {
		t.Errorf("pulse on read error: want unavailable, got %q", w.State)
	}
}
