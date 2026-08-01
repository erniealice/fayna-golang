package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	home "github.com/erniealice/fayna-golang/domain/home"

	ocpb "github.com/erniealice/esqyma/pkg/schema/v1/service/dashboard/outcome_completion"
)

// ---------------------------------------------------------------------------
// Harness
// ---------------------------------------------------------------------------

func testOptions() home.Options {
	return home.Options{
		Variants: map[home.Slot]home.Variant{
			home.SlotOperatorStaff: {Widgets: []home.WidgetRef{
				home.WidgetCompletionSummary, home.WidgetApprovalProgress,
				home.WidgetAttention, home.WidgetQuickLinks,
			}},
			home.SlotStaff: {Widgets: []home.WidgetRef{
				home.WidgetCompletionSummary, home.WidgetApprovalProgress, home.WidgetQuickLinks,
			}},
		},
		DefaultSlot: home.SlotStaff,
		Tab:         home.TabOptions{GroupByField: home.TabEntityJobCategory},
	}
}

func testViewCtx(url string) *view.ViewContext {
	return &view.ViewContext{
		Request:      httptest.NewRequest("GET", url, nil),
		CurrentPath:  "/home",
		CacheVersion: "test",
		Messages: map[string]string{
			"home.title":                          "Daily Overview",
			"home.tab.all":                        "All",
			"home.period.label":                   "Period",
			"home.widget.completion.title":        "Outcome completion",
			"home.widget.completion.empty":        "Nothing assigned yet",
			"home.widget.completion.zero":         "No outcomes recorded yet for this period",
			"home.widget.completion.scope_self":   "Assigned to you",
			"home.widget.approval.title":          "Review progress",
			"home.widget.attention.title":         "Needs attention",
			"home.widget.attention.empty":         "Nothing needs attention",
			"home.widget.quick.title":             "Quick links",
			"home.widget.denied":                  "You don't have access to this information",
			"home.widget.denied_hint":             "Ask an administrator if you need it",
			"home.widget.unavailable":             "This information can't be loaded right now",
			"home.widget.unavailable_hint":        "Try refreshing the page; contact an administrator if it keeps happening",
		},
	}
}

func withPerms(codes ...string) context.Context {
	return view.WithUserPermissions(context.Background(), types.NewUserPermissions(codes))
}

// slice builds a period slice.
func slice(order int32, expected, recorded int64, counts *ocpb.OutcomeCompletionApprovalCounts) *ocpb.OutcomeCompletionPeriodSlice {
	if counts == nil {
		counts = &ocpb.OutcomeCompletionApprovalCounts{}
	}
	return &ocpb.OutcomeCompletionPeriodSlice{
		PhaseOrder: order, ExpectedCells: expected, RecordedCells: recorded, ApprovalCounts: counts,
	}
}

func testResponse() *ocpb.GetOutcomeCompletionSummaryResponse {
	catA := &ocpb.OutcomeCompletionCategoryRow{
		CategoryId: "cat-a", CategoryName: "Alpha",
		ExpectedCells: 100, RecordedCells: 40,
		ApprovalCounts: &ocpb.OutcomeCompletionApprovalCounts{InProgress: 4},
		PeriodSlices: []*ocpb.OutcomeCompletionPeriodSlice{
			slice(1, 60, 60, nil),
			slice(2, 40, 0, &ocpb.OutcomeCompletionApprovalCounts{NotStarted: 3, Returned: 1}),
		},
	}
	catB := &ocpb.OutcomeCompletionCategoryRow{
		CategoryId: "cat-b", CategoryName: "Beta",
		ExpectedCells: 50, RecordedCells: 45,
		ApprovalCounts: &ocpb.OutcomeCompletionApprovalCounts{ForReview: 2},
		PeriodSlices: []*ocpb.OutcomeCompletionPeriodSlice{
			slice(1, 25, 25, nil),
			slice(2, 25, 20, &ocpb.OutcomeCompletionApprovalCounts{ForReview: 2}),
		},
	}
	return &ocpb.GetOutcomeCompletionSummaryResponse{
		Success:      true,
		CategoryRows: []*ocpb.OutcomeCompletionCategoryRow{catA, catB},
		AllRollup: &ocpb.OutcomeCompletionCategoryRow{
			ExpectedCells: 150, RecordedCells: 85,
			ApprovalCounts: &ocpb.OutcomeCompletionApprovalCounts{InProgress: 4, ForReview: 2},
			PeriodSlices: []*ocpb.OutcomeCompletionPeriodSlice{
				slice(1, 85, 85, nil),
				slice(2, 65, 20, &ocpb.OutcomeCompletionApprovalCounts{NotStarted: 3, ForReview: 2, Returned: 1}),
			},
		},
	}
}

func newTestModule(resp *ocpb.GetOutcomeCompletionSummaryResponse, kind int32, denies *[]string) *Module {
	return NewModule(&Deps{
		Routes:  home.DefaultRoutes(),
		Options: testOptions(),
		GetCompletionSummary: func(context.Context) (*ocpb.GetOutcomeCompletionSummaryResponse, error) {
			return resp, nil
		},
		ResolvePrincipalKind: func(context.Context) int32 { return kind },
		LogDeny: func(_ context.Context, code, widget string) {
			if denies != nil {
				*denies = append(*denies, code+"|"+widget)
			}
		},
	})
}

func widgetByRef(t *testing.T, d *PageData, ref home.WidgetRef) WidgetData {
	t.Helper()
	for _, w := range d.Widgets {
		if w.Ref == string(ref) {
			return w
		}
	}
	t.Fatalf("widget %s missing from roster (T-9: widgets must never silently vanish)", ref)
	return WidgetData{}
}

// ---------------------------------------------------------------------------
// Q4 / T-9 degrade
// ---------------------------------------------------------------------------

func TestNilPermissionsRenderDeniedCardsAndLogAndSkipRead(t *testing.T) {
	var denies []string
	called := false
	m := NewModule(&Deps{
		Routes:  home.DefaultRoutes(),
		Options: testOptions(),
		GetCompletionSummary: func(context.Context) (*ocpb.GetOutcomeCompletionSummaryResponse, error) {
			called = true
			return testResponse(), nil
		},
		ResolvePrincipalKind: func(context.Context) int32 { return home.SlotStaff },
		LogDeny: func(_ context.Context, code, widget string) {
			denies = append(denies, code+"|"+widget)
		},
	})
	// NO permission set in ctx at all — nil is fail-closed.
	d := m.buildData(context.Background(), testViewCtx("/home"))

	for _, ref := range []home.WidgetRef{home.WidgetCompletionSummary, home.WidgetApprovalProgress} {
		w := widgetByRef(t, d, ref)
		if w.State != StateDenied {
			t.Errorf("%s: want denied state under nil perms, got %q", ref, w.State)
		}
		if w.DeniedText == "" || w.Title == "" {
			t.Errorf("%s: denied card must carry a visible reason + title", ref)
		}
	}
	if called {
		t.Error("aggregate read must NOT be issued when every bundle is denied")
	}
	if len(denies) == 0 {
		t.Error("view-level denies must be logged (AUTHZ_RBAC_DENY closure)")
	}
	// The tabstrip shares the widget bundle — no data, no tabs.
	if d.TabsEnabled {
		t.Error("tabstrip must not render when the shared bundle is denied")
	}
	// quick_links has no bundle: it must still render (not denied).
	if w := widgetByRef(t, d, home.WidgetQuickLinks); w.State == StateDenied {
		t.Error("quick_links must not be bundle-denied (no data read)")
	}
}

// TestSummaryDenyRendersDeniedCardsAndLogs — a read failure CLASSIFIED as an
// authorization deny (ErrSummaryDenied, wrapped by the composition closure)
// renders the designed DENIED cards and emits the AUTHZ_RBAC_DENY closure —
// NEVER the benign "nothing assigned" empty state (skeptic F2).
func TestSummaryDenyRendersDeniedCardsAndLogs(t *testing.T) {
	var denies []string
	m := NewModule(&Deps{
		Routes:  home.DefaultRoutes(),
		Options: testOptions(),
		GetCompletionSummary: func(context.Context) (*ocpb.GetOutcomeCompletionSummaryResponse, error) {
			return nil, fmt.Errorf("%w: use-case gate refused", ErrSummaryDenied)
		},
		ResolvePrincipalKind: func(context.Context) int32 { return home.SlotStaff },
		LogDeny: func(_ context.Context, code, widget string) {
			denies = append(denies, code+"|"+widget)
		},
	})
	d := m.buildData(withPerms("outcome_completion:read"), testViewCtx("/home"))

	for _, ref := range []home.WidgetRef{home.WidgetCompletionSummary, home.WidgetApprovalProgress} {
		w := widgetByRef(t, d, ref)
		if w.State != StateDenied {
			t.Errorf("%s: use-case deny must render the DENIED card, got %q", ref, w.State)
		}
		if w.DeniedText == "" || w.Title == "" {
			t.Errorf("%s: denied card must carry a visible reason + title", ref)
		}
	}
	if len(denies) != 2 {
		t.Errorf("use-case deny must emit the AUTHZ_RBAC_DENY closure per data widget, got %v", denies)
	}
	// quick_links has no bundle — unaffected by the deny.
	if w := widgetByRef(t, d, home.WidgetQuickLinks); w.State != StateFull {
		t.Errorf("quick_links must be unaffected by a summary deny, got %q", w.State)
	}
	if d.TabsEnabled {
		t.Error("no response → no tabstrip")
	}
}

// TestSummaryErrorRendersUnavailableNotEmpty — a NON-authorization read
// failure (DB down, query error) renders the labelled UNAVAILABLE state:
// distinct from denied AND from the legitimate "nothing assigned" empty
// state, with no deny log line (it is not a deny).
func TestSummaryErrorRendersUnavailableNotEmpty(t *testing.T) {
	var denies []string
	m := NewModule(&Deps{
		Routes:  home.DefaultRoutes(),
		Options: testOptions(),
		GetCompletionSummary: func(context.Context) (*ocpb.GetOutcomeCompletionSummaryResponse, error) {
			return nil, errors.New("pq: connection refused")
		},
		ResolvePrincipalKind: func(context.Context) int32 { return home.SlotStaff },
		LogDeny: func(_ context.Context, code, widget string) {
			denies = append(denies, code+"|"+widget)
		},
	})
	d := m.buildData(withPerms("outcome_completion:read"), testViewCtx("/home"))

	for _, ref := range []home.WidgetRef{home.WidgetCompletionSummary, home.WidgetApprovalProgress} {
		w := widgetByRef(t, d, ref)
		if w.State != StateUnavailable {
			t.Errorf("%s: data error must render UNAVAILABLE, got %q", ref, w.State)
		}
		if w.UnavailableText == "" || w.Title == "" {
			t.Errorf("%s: unavailable card must carry a visible reason + title", ref)
		}
	}
	if len(denies) != 0 {
		t.Errorf("a data error is NOT a deny — no AUTHZ_RBAC_DENY closure, got %v", denies)
	}
	if w := widgetByRef(t, d, home.WidgetQuickLinks); w.State != StateFull {
		t.Errorf("quick_links must be unaffected by a data error, got %q", w.State)
	}
}

// ---------------------------------------------------------------------------
// Full render (staff persona)
// ---------------------------------------------------------------------------

func TestStaffFullRenderWindowsToEarliestUnrecordedPeriod(t *testing.T) {
	m := newTestModule(testResponse(), home.SlotStaff, nil)
	ctx := withPerms("outcome_completion:read")
	d := m.buildData(ctx, testViewCtx("/home"))

	if d.Slot != home.SlotStaff {
		t.Fatalf("slot: want 7, got %d", d.Slot)
	}
	// Staff roster: 3 widgets, no attention.
	if len(d.Widgets) != 3 {
		t.Fatalf("staff roster must have 3 widgets, got %d", len(d.Widgets))
	}
	for _, w := range d.Widgets {
		if w.Ref == string(home.WidgetAttention) {
			t.Error("staff roster must not include the attention widget (Q1 lock)")
		}
	}

	// Completion (All rollup): period 1 is fully recorded, so the A2 window
	// must pick period 2 (earliest with unrecorded work) and SHOW it.
	w := widgetByRef(t, d, home.WidgetCompletionSummary)
	if w.State != StateFull {
		t.Fatalf("completion: want full state, got %q", w.State)
	}
	c := w.Completion
	if c.PeriodText != "Period 2" {
		t.Errorf("A2 window must be visible: want 'Period 2', got %q", c.PeriodText)
	}
	if c.Expected != "65" || c.Recorded != "20" {
		t.Errorf("windowed counts: want 20/65, got %s/%s", c.Recorded, c.Expected)
	}
	if c.ScopeText != "Assigned to you" {
		t.Errorf("staff scope line: got %q", c.ScopeText)
	}
	if c.ZeroText != "" {
		t.Errorf("recorded > 0 must not be the zero state")
	}

	// Approval: windowed counts incl. the ride-along buckets.
	a := widgetByRef(t, d, home.WidgetApprovalProgress)
	if a.State != StateFull {
		t.Fatalf("approval: want full, got %q", a.State)
	}
	byKey := map[string]string{}
	for _, r := range a.Approval.Rows {
		byKey[r.Key] = r.Value
	}
	if byKey["not-started"] != "3" || byKey["returned"] != "1" || byKey["for-review"] != "2" {
		t.Errorf("approval ride-along buckets wrong: %v", byKey)
	}

	// Tabs: All + 2 categories, §A-6 testids, All active.
	if !d.TabsEnabled || len(d.Tabs) != 3 {
		t.Fatalf("tabs: want enabled with 3 items, got enabled=%v n=%d", d.TabsEnabled, len(d.Tabs))
	}
	if d.Tabs[0].ID != "home-tab-all" || !d.Tabs[0].Active {
		t.Errorf("first tab must be the active All tab, got %+v", d.Tabs[0])
	}
	if d.Tabs[1].ID != "home-tab-cat-a" || d.Tabs[1].Query != "?t=cat-a" {
		t.Errorf("category tab shape wrong: %+v", d.Tabs[1])
	}
}

func TestTabSelectionValidatesAgainstLiveCategories(t *testing.T) {
	m := newTestModule(testResponse(), home.SlotOperatorStaff, nil)
	ctx := withPerms("outcome_completion:read")

	// Valid selection narrows the row.
	d := m.buildData(ctx, testViewCtx("/home/content?t=cat-b"))
	if d.ActiveTab != "cat-b" {
		t.Fatalf("valid t must select the category, got %q", d.ActiveTab)
	}
	w := widgetByRef(t, d, home.WidgetCompletionSummary)
	if w.Completion.Expected != "25" || w.Completion.Recorded != "20" {
		t.Errorf("cat-b window: want 20/25, got %s/%s", w.Completion.Recorded, w.Completion.Expected)
	}

	// Unknown/tampered t fails safe to All.
	d = m.buildData(ctx, testViewCtx("/home/content?t=not-a-category"))
	if d.ActiveTab != "all" {
		t.Errorf("unknown t must fall back to All, got %q", d.ActiveTab)
	}
}

func TestAdminAttentionRanksLowestFirstWithExceptionsRow(t *testing.T) {
	m := newTestModule(testResponse(), home.SlotOperatorStaff, nil)
	ctx := withPerms("outcome_completion:read")
	d := m.buildData(ctx, testViewCtx("/home"))

	w := widgetByRef(t, d, home.WidgetAttention)
	if w.State != StateFull {
		t.Fatalf("attention: want full, got %q", w.State)
	}
	rows := w.Attention.Rows
	// cat-a windows to period 2 (0/40 = 0%), cat-b to period 2 (20/25 = 80%).
	if len(rows) != 2 || rows[0].CategoryID != "cat-a" || rows[1].CategoryID != "cat-b" {
		t.Fatalf("attention must rank lowest completion first, got %+v", rows)
	}
	if w.Attention.ExceptionsValue != "1" {
		t.Errorf("operational-exceptions (returned) row: want 1, got %q", w.Attention.ExceptionsValue)
	}
}

// ---------------------------------------------------------------------------
// Empty persona + zero-state distinction (Q3)
// ---------------------------------------------------------------------------

func TestEmptyPersonaRendersEmptyStatesNotErrors(t *testing.T) {
	empty := &ocpb.GetOutcomeCompletionSummaryResponse{
		Success:   true,
		AllRollup: &ocpb.OutcomeCompletionCategoryRow{ApprovalCounts: &ocpb.OutcomeCompletionApprovalCounts{}},
	}
	m := newTestModule(empty, home.SlotStaff, nil)
	d := m.buildData(withPerms("outcome_completion:read"), testViewCtx("/home"))

	c := widgetByRef(t, d, home.WidgetCompletionSummary)
	if c.State != StateEmpty || c.EmptyText == "" {
		t.Errorf("zero expected cells must render the labelled EMPTY state, got %q", c.State)
	}
	a := widgetByRef(t, d, home.WidgetApprovalProgress)
	if a.State != StateEmpty {
		t.Errorf("all-zero approval counts must render the empty state, got %q", a.State)
	}
}

func TestFreshPeriodZeroStateIsDistinctFromEmpty(t *testing.T) {
	resp := &ocpb.GetOutcomeCompletionSummaryResponse{
		Success: true,
		AllRollup: &ocpb.OutcomeCompletionCategoryRow{
			ExpectedCells: 4238, RecordedCells: 0,
			ApprovalCounts: &ocpb.OutcomeCompletionApprovalCounts{NotStarted: 90},
			PeriodSlices: []*ocpb.OutcomeCompletionPeriodSlice{
				slice(1, 4238, 0, &ocpb.OutcomeCompletionApprovalCounts{NotStarted: 90}),
			},
		},
	}
	m := newTestModule(resp, home.SlotStaff, nil)
	d := m.buildData(withPerms("outcome_completion:read"), testViewCtx("/home"))

	w := widgetByRef(t, d, home.WidgetCompletionSummary)
	if w.State != StateFull {
		t.Fatalf("expected>0 recorded=0 is the legitimate ZERO state (full card), got %q", w.State)
	}
	if w.Completion.ZeroText == "" {
		t.Error("fresh-period zero must carry the zero-state copy (Q3 empty-vs-zero distinction)")
	}
	if w.Completion.RatePct != "0.0%" {
		t.Errorf("zero rate must render 0.0%%, got %q", w.Completion.RatePct)
	}
}

// ---------------------------------------------------------------------------
// Quick links permission filter
// ---------------------------------------------------------------------------

func TestQuickLinksFilterByPermission(t *testing.T) {
	opts := testOptions()
	opts.QuickLinks = []home.QuickLink{
		{RouteKey: "job.list", ParamName: "status", ParamValue: "active", Label: "Courses", Permission: "job:list"},
		{RouteKey: "outcome_summary.list", Label: "Reports", Permission: "job_outcome_summary:list"},
		{RouteKey: "event.calendar", Label: "Schedule"}, // ungated
		{RouteKey: "", Label: "broken"},                 // half-declared → skipped
	}
	m := NewModule(&Deps{
		Routes:               home.DefaultRoutes(),
		Options:              opts,
		ResolvePrincipalKind: func(context.Context) int32 { return home.SlotStaff },
	})
	d := m.buildData(withPerms("job:list"), testViewCtx("/home"))

	w := widgetByRef(t, d, home.WidgetQuickLinks)
	var labels []string
	for _, l := range w.Links {
		labels = append(labels, l.Label)
	}
	got := strings.Join(labels, ",")
	if got != "Courses,Schedule" {
		t.Errorf("quick links must keep held + ungated links only, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Fail-safety of the read closure
// ---------------------------------------------------------------------------

func TestNilReadClosureDegradesToEmptyStates(t *testing.T) {
	m := NewModule(&Deps{
		Routes:               home.DefaultRoutes(),
		Options:              testOptions(),
		ResolvePrincipalKind: func(context.Context) int32 { return home.SlotStaff },
	})
	d := m.buildData(withPerms("outcome_completion:read"), testViewCtx("/home"))
	w := widgetByRef(t, d, home.WidgetCompletionSummary)
	if w.State != StateEmpty {
		t.Errorf("nil read closure must degrade to the empty state, got %q", w.State)
	}
	if d.TabsEnabled {
		t.Error("no response → no tabstrip")
	}
}
