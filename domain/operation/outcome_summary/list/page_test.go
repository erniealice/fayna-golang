package list

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// TestListContentTemplateDispatch (T9) pins the boosted-nav dispatch invariant:
// the ViewAdapter derives the content-partial name as {result.Template}-content,
// so PageData.ContentTemplate MUST equal result.Template + "-content". A drift
// here reproduces the "vanishing tabs" bug (the full-page and boosted-nav render
// paths desync). templates_golden pins the template NAMES; nothing pinned the
// dispatch relationship until now.
func TestListContentTemplateDispatch(t *testing.T) {
	v := NewView(&ListViewDeps{
		// Non-nil so renderFlat (the zero-Options path) does not nil-deref; an
		// empty response renders the flat empty table.
		ListJobOutcomeSummarys: func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
			return &jobsumpb.ListJobOutcomeSummarysResponse{}, nil
		},
	})

	req := httptest.NewRequest("GET", "/outcomes/summaries", nil)
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions([]string{"job_outcome_summary:list"}))
	res := v.Handle(ctx, &view.ViewContext{Request: req, CurrentPath: "/outcomes/summaries", CacheVersion: "test"})

	if res.Template != "outcome-summary-list" {
		t.Fatalf("Template = %q, want %q", res.Template, "outcome-summary-list")
	}
	pd, ok := res.Data.(*PageData)
	if !ok {
		t.Fatalf("Data type = %T, want *PageData", res.Data)
	}
	if want := res.Template + "-content"; pd.ContentTemplate != want {
		t.Fatalf("ContentTemplate = %q, want %q (boosted-nav dispatch invariant)", pd.ContentTemplate, want)
	}
}

// --- scope-boundary guard (empty-band) tests -----------------------------
//
// These cover the landing (view-1) dispatch, which the flat-dispatch test above
// does not touch. The invariant under test: a SCOPED landing (current|past)
// whose activeness band has NO schedules must render ZERO section rows — it must
// never fall open to the unfiltered group set. The UNSCOPED landing (scope "")
// stays unfiltered.

// TestLandingScopedPastEmptyBandRendersZeroRows: /report-cards/list/past in a
// workspace with NO inactive schedule renders an empty landing (no rows, no
// tabs) — NOT the open AY's sections — and performs none of the group/count
// reads.
func TestLandingScopedPastEmptyBandRendersZeroRows(t *testing.T) {
	var groupsCalled, summariesCalled bool
	deps := &ListViewDeps{}
	deps.Options.List.Entity = "subscription_group"
	deps.ListPriceSchedules = func(_ context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error) {
		if hasFilters(req.GetFilters().GetFilters()) {
			// inactive (past) band: empty.
			return &priceschedulepb.ListPriceSchedulesResponse{}, nil
		}
		// active (current) band: one open schedule.
		return &priceschedulepb.ListPriceSchedulesResponse{Data: []*priceschedulepb.PriceSchedule{
			{Id: "ps-active", Name: "AY 2025-26", Active: true},
		}}, nil
	}
	deps.ListSubscriptionGroups = func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		groupsCalled = true
		return &subscriptiongrouppb.ListSubscriptionGroupsResponse{}, nil
	}
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		summariesCalled = true
		return &summarypb.ListJobTemplateSummariesResponse{}, nil
	}

	ctx, vc := landingReq("past")
	res := NewView(deps).Handle(ctx, vc)
	pd := mustPageData(t, res)

	if len(pd.Table.Rows) != 0 {
		t.Fatalf("Table.Rows = %d, want 0 (scope=past must not fall open to the active AY)", len(pd.Table.Rows))
	}
	if len(pd.TabItems) != 0 {
		t.Fatalf("TabItems = %d, want 0 (no tabstrip for an empty band)", len(pd.TabItems))
	}
	if !pd.Landing {
		t.Fatalf("Landing = false, want true (empty landing, not the flat list)")
	}
	if pd.ActiveSubNav != "report-cards-past" {
		t.Fatalf("ActiveSubNav = %q, want %q", pd.ActiveSubNav, "report-cards-past")
	}
	if want := res.Template + "-content"; pd.ContentTemplate != want {
		t.Fatalf("ContentTemplate = %q, want %q (boosted-nav dispatch invariant)", pd.ContentTemplate, want)
	}
	if groupsCalled {
		t.Fatalf("ListSubscriptionGroups called; the group read must be skipped for an empty scoped band")
	}
	if summariesCalled {
		t.Fatalf("ListJobTemplateSummaries called; the count read must be skipped for an empty scoped band")
	}
}

// TestLandingScopedCurrentEmptyBandRendersZeroRows: /report-cards/list/current
// in a workspace with NO active schedule renders an empty landing and skips the
// group/count reads (the symmetric case).
func TestLandingScopedCurrentEmptyBandRendersZeroRows(t *testing.T) {
	var groupsCalled, summariesCalled bool
	deps := &ListViewDeps{}
	deps.Options.List.Entity = "subscription_group"
	deps.ListPriceSchedules = func(_ context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error) {
		if hasFilters(req.GetFilters().GetFilters()) {
			// inactive (past) band: one closed schedule.
			return &priceschedulepb.ListPriceSchedulesResponse{Data: []*priceschedulepb.PriceSchedule{
				{Id: "ps-past", Name: "AY 2024-25", Active: false},
			}}, nil
		}
		// active (current) band: empty.
		return &priceschedulepb.ListPriceSchedulesResponse{}, nil
	}
	deps.ListSubscriptionGroups = func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		groupsCalled = true
		return &subscriptiongrouppb.ListSubscriptionGroupsResponse{}, nil
	}
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		summariesCalled = true
		return &summarypb.ListJobTemplateSummariesResponse{}, nil
	}

	ctx, vc := landingReq("current")
	res := NewView(deps).Handle(ctx, vc)
	pd := mustPageData(t, res)

	if len(pd.Table.Rows) != 0 {
		t.Fatalf("Table.Rows = %d, want 0 (scope=current must not fall open to the past AY)", len(pd.Table.Rows))
	}
	if len(pd.TabItems) != 0 {
		t.Fatalf("TabItems = %d, want 0 (no tabstrip for an empty band)", len(pd.TabItems))
	}
	if pd.ActiveSubNav != "report-cards-current" {
		t.Fatalf("ActiveSubNav = %q, want %q", pd.ActiveSubNav, "report-cards-current")
	}
	if groupsCalled {
		t.Fatalf("ListSubscriptionGroups called; the group read must be skipped for an empty scoped band")
	}
	if summariesCalled {
		t.Fatalf("ListJobTemplateSummaries called; the count read must be skipped for an empty scoped band")
	}
}

// TestLandingScopedPastRendersOnlyInactiveBand: /report-cards/list/past with an
// inactive schedule present renders ONLY that band's sections (existing
// behavior) — the active AY's section must not leak in.
func TestLandingScopedPastRendersOnlyInactiveBand(t *testing.T) {
	deps := &ListViewDeps{}
	deps.Options.List.Entity = "subscription_group"
	deps.ListPriceSchedules = func(_ context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error) {
		if hasFilters(req.GetFilters().GetFilters()) {
			return &priceschedulepb.ListPriceSchedulesResponse{Data: []*priceschedulepb.PriceSchedule{
				{Id: "ps-past", Name: "AY 2024-25", Active: false},
			}}, nil
		}
		return &priceschedulepb.ListPriceSchedulesResponse{Data: []*priceschedulepb.PriceSchedule{
			{Id: "ps-active", Name: "AY 2025-26", Active: true},
		}}, nil
	}
	deps.ListSubscriptionGroups = func(_ context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		if hasFilters(req.GetFilters().GetFilters()) {
			// inactive groups belong to the past schedule.
			return &subscriptiongrouppb.ListSubscriptionGroupsResponse{Data: []*subscriptiongrouppb.SubscriptionGroup{
				{Id: "g-past-1", Name: "Grade 9 A", Active: false, PriceScheduleId: strptr("ps-past")},
				{Id: "g-past-2", Name: "Grade 9 B", Active: false, PriceScheduleId: strptr("ps-past")},
			}}, nil
		}
		// active group belongs to the current schedule — must NOT render under past.
		return &subscriptiongrouppb.ListSubscriptionGroupsResponse{Data: []*subscriptiongrouppb.SubscriptionGroup{
			{Id: "g-active-1", Name: "Grade 10 A", Active: true, PriceScheduleId: strptr("ps-active")},
		}}, nil
	}

	ctx, vc := landingReq("past")
	res := NewView(deps).Handle(ctx, vc)
	pd := mustPageData(t, res)

	if len(pd.TabItems) != 1 {
		t.Fatalf("TabItems = %d, want 1 (only the past band's tab)", len(pd.TabItems))
	}
	if len(pd.Table.Rows) != 2 {
		t.Fatalf("Table.Rows = %d, want 2 (only the inactive schedule's sections)", len(pd.Table.Rows))
	}
	for _, r := range pd.Table.Rows {
		if r.ID == "g-active-1" {
			t.Fatalf("active-AY section g-active-1 rendered under scope=past; band boundary leaked")
		}
	}
	if pd.ActiveSubNav != "report-cards-past" {
		t.Fatalf("ActiveSubNav = %q, want %q", pd.ActiveSubNav, "report-cards-past")
	}
}

// TestLandingUnscopedUnfilteredBackcompat: scope "" (the /report-cards ListURL
// landing) is unaffected by the guard — even with NO inactive schedule it stays
// unfiltered and renders the default (active) schedule's sections, exactly as
// before. This is the backward-compatibility contract.
func TestLandingUnscopedUnfilteredBackcompat(t *testing.T) {
	deps := &ListViewDeps{}
	deps.Options.List.Entity = "subscription_group"
	deps.ListPriceSchedules = func(_ context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error) {
		if hasFilters(req.GetFilters().GetFilters()) {
			return &priceschedulepb.ListPriceSchedulesResponse{}, nil // no inactive schedule
		}
		return &priceschedulepb.ListPriceSchedulesResponse{Data: []*priceschedulepb.PriceSchedule{
			{Id: "ps-active", Name: "AY 2025-26", Active: true},
		}}, nil
	}
	deps.ListSubscriptionGroups = func(_ context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		if hasFilters(req.GetFilters().GetFilters()) {
			return &subscriptiongrouppb.ListSubscriptionGroupsResponse{}, nil
		}
		return &subscriptiongrouppb.ListSubscriptionGroupsResponse{Data: []*subscriptiongrouppb.SubscriptionGroup{
			{Id: "g-active-1", Name: "Grade 10 A", Active: true, PriceScheduleId: strptr("ps-active")},
		}}, nil
	}

	ctx, vc := landingReq("")
	res := NewView(deps).Handle(ctx, vc)
	pd := mustPageData(t, res)

	if len(pd.TabItems) != 1 {
		t.Fatalf("TabItems = %d, want 1 (unscoped landing stays unfiltered)", len(pd.TabItems))
	}
	if len(pd.Table.Rows) != 1 {
		t.Fatalf("Table.Rows = %d, want 1 (the active schedule's section renders)", len(pd.Table.Rows))
	}
	if pd.ActiveSubNav != "report-cards" {
		t.Fatalf("ActiveSubNav = %q, want %q (base report-cards row)", pd.ActiveSubNav, "report-cards")
	}
}

// --- test helpers --------------------------------------------------------

// landingReq builds a permission-granted GET at the scoped landing route,
// populating the {scope} path value the router would set (Go 1.22+
// SetPathValue). An empty scope leaves the path value unset (the unscoped
// landing).
func landingReq(scope string) (context.Context, *view.ViewContext) {
	req := httptest.NewRequest("GET", "/report-cards/list/"+scope, nil)
	if scope != "" {
		req.SetPathValue("scope", scope)
	}
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions([]string{"job_outcome_summary:list"}))
	return ctx, &view.ViewContext{Request: req, CurrentPath: req.URL.Path, CacheVersion: "test"}
}

// mustPageData asserts the landing template + PageData shape and returns it.
func mustPageData(t *testing.T, res view.ViewResult) *PageData {
	t.Helper()
	if res.Template != "outcome-summary-list" {
		t.Fatalf("Template = %q, want %q", res.Template, "outcome-summary-list")
	}
	pd, ok := res.Data.(*PageData)
	if !ok {
		t.Fatalf("Data type = %T, want *PageData", res.Data)
	}
	return pd
}

// hasFilters reports whether a list request carried the explicit active=false
// filter (the inactive-band read) — the listAll* merge issues the default
// (active) call with no filters and the inactive call with one.
func hasFilters[T any](filters []T) bool { return len(filters) > 0 }

// strptr returns a pointer to s (proto oneof string fields are *string).
func strptr(s string) *string { return &s }
