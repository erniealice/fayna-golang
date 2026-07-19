package list

import (
	"context"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
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

// --- R9 W-A2 dynamic category-column tests -------------------------------
//
// The landing gains one count column per ACTIVE job_category (sort_order
// order), cells = the typed composite count+eye (pyeza BuildCompositeCell),
// with a two-layer degrade: knob off OR empty/error category read → EXACTLY
// today's static landing. NULL / out-of-corpus subjects fold into the single
// Uncategorized bucket column (§3.0 — never dropped, never duplicated).

// dynamicDeps builds a knob-on landing fixture: one active schedule, one
// active section (g-1, "Grade 10 A"), category corpus via the tab-support
// closure, and summary rows spread across categories.
func dynamicDeps() *ListViewDeps {
	deps := &ListViewDeps{}
	deps.Options.List.Entity = "subscription_group"
	deps.Options.List.ColumnsByField = "job_category"
	deps.Routes.SectionURL = "/report-cards/section/{id}"
	deps.Routes.SectionExportURL = "/report-cards/section/{id}/export"
	deps.Labels.Landing.CellViewAction = "View {category} report cards for {section}"
	deps.Labels.Landing.UncategorizedColumn = "Uncategorized"
	deps.ListPriceSchedules = func(_ context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error) {
		if hasFilters(req.GetFilters().GetFilters()) {
			return &priceschedulepb.ListPriceSchedulesResponse{}, nil
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
			{Id: "g-1", Name: "Grade 10 A", Active: true, PriceScheduleId: strptr("ps-active")},
		}}, nil
	}
	// Summary rows: t1..t3 in cat-a (t1 duplicated across staff — dedup), t4 in
	// cat-b, none in cat-c. Largest per-subject cohort = 28.
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		return &summarypb.ListJobTemplateSummariesResponse{Summaries: []*summarypb.JobTemplateSummary{
			{JobTemplateId: "t1", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a"},
			{JobTemplateId: "t1", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a"}, // second staff row — dedup
			{JobTemplateId: "t2", SubscriptionGroupId: "g-1", JobCount: 27, JobCategoryId: "cat-a"},
			{JobTemplateId: "t3", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a"},
			{JobTemplateId: "t4", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-b"},
		}}, nil
	}
	// Corpus arrives UNSORTED and includes an inactive category — the view must
	// keep ACTIVE only, ordered by sort_order (cat-a=1, cat-b=2, cat-c=3).
	deps.ListJobListTabSupport = func(context.Context) ([]*jobcategorypb.JobCategory, []*jobtemplatepb.JobTemplate, error) {
		return []*jobcategorypb.JobCategory{
				{Id: "cat-c", Name: "Homeroom Deportment", Active: true, SortOrder: i32ptr(3)},
				{Id: "cat-a", Name: "Academic", Active: true, SortOrder: i32ptr(1)},
				{Id: "cat-old", Name: "Retired", Active: false, SortOrder: i32ptr(0)},
				{Id: "cat-b", Name: "Subject Deportment", Active: true, SortOrder: i32ptr(2)},
			}, []*jobtemplatepb.JobTemplate{
				{Id: "t1", Name: "Math", JobCategoryId: strptr("cat-a")},
				{Id: "t2", Name: "Science", JobCategoryId: strptr("cat-a")},
				{Id: "t3", Name: "English", JobCategoryId: strptr("cat-a")},
				{Id: "t4", Name: "Math Conduct", JobCategoryId: strptr("cat-b")},
			}, nil
	}
	return deps
}

// TestLandingDynamicCategoryColumns pins the dynamic column set: headers are
// job_category.name DATA in sort_order order, cells are typed composite
// count+eye cells with URL-query-encoded ?jc= hrefs, collision-proof
// rc-eye-<section>-<category> test ids, and an aria name carrying BOTH the
// category and the section. A zero-count category still renders its column
// (and its eye — the empty tab is still navigable).
func TestLandingDynamicCategoryColumns(t *testing.T) {
	deps := dynamicDeps()
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))

	wantKeys := []string{"section", "students", "jc-cat-a", "jc-cat-b", "jc-cat-c"}
	if len(pd.Table.Columns) != len(wantKeys) {
		t.Fatalf("columns = %d, want %d (%v)", len(pd.Table.Columns), len(wantKeys), pd.Table.Columns)
	}
	for i, k := range wantKeys {
		if pd.Table.Columns[i].Key != k {
			t.Fatalf("column[%d].Key = %q, want %q", i, pd.Table.Columns[i].Key, k)
		}
	}
	for i, want := range []string{"Academic", "Subject Deportment", "Homeroom Deportment"} {
		if got := pd.Table.Columns[2+i].Label; got != want {
			t.Fatalf("category column[%d] label = %q, want %q (job_category.name DATA, sort_order order)", 2+i, got, want)
		}
	}

	if len(pd.Table.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(pd.Table.Rows))
	}
	cells := pd.Table.Rows[0].Cells
	if len(cells) != 5 {
		t.Fatalf("cells = %d, want 5 (no Uncategorized bucket here)", len(cells))
	}
	// Per-category subject counts: cat-a=3 (t1 deduped), cat-b=1, cat-c=0.
	for i, want := range []string{"3", "1", "0"} {
		c := cells[2+i]
		if c.Type != "composite" || c.Value != want {
			t.Fatalf("cell[%d] = {Type:%q Value:%q}, want composite %q", 2+i, c.Type, c.Value, want)
		}
		if c.Composite == nil {
			t.Fatalf("cell[%d].Composite = nil", 2+i)
		}
	}
	eye := cells[2].Composite
	if eye.EyeHref != "/report-cards/section/g-1?jc=cat-a" {
		t.Fatalf("EyeHref = %q, want %q", eye.EyeHref, "/report-cards/section/g-1?jc=cat-a")
	}
	if eye.EyeTestID != "rc-eye-g-1-cat-a" {
		t.Fatalf("EyeTestID = %q, want %q (collision-proof full ids)", eye.EyeTestID, "rc-eye-g-1-cat-a")
	}
	if eye.EyeName != "View Academic report cards for Grade 10 A" {
		t.Fatalf("EyeName = %q, want the lyngua frame with BOTH nouns substituted", eye.EyeName)
	}
	// The zero-count category keeps a working eye (empty tab is navigable).
	if zero := cells[4].Composite; zero.EyeHref != "/report-cards/section/g-1?jc=cat-c" {
		t.Fatalf("zero-count EyeHref = %q, want %q", zero.EyeHref, "/report-cards/section/g-1?jc=cat-c")
	}
}

// TestLandingZeroCategoriesDegradesToStatic pins the two-layer degrade: with
// the knob ON but an EMPTY category read — and equally a FAILED read — the
// landing's columns and rows are DEEP-EQUAL to the knob-off static render
// (today's landing, byte-identical).
func TestLandingZeroCategoriesDegradesToStatic(t *testing.T) {
	baseline := dynamicDeps()
	baseline.Options.List.ColumnsByField = "" // knob OFF — the static reference
	ctx, vc := landingReq("")
	ref := mustPageData(t, NewView(baseline).Handle(ctx, vc))

	empty := dynamicDeps()
	empty.ListJobListTabSupport = func(context.Context) ([]*jobcategorypb.JobCategory, []*jobtemplatepb.JobTemplate, error) {
		return nil, nil, nil
	}
	ctx2, vc2 := landingReq("")
	got := mustPageData(t, NewView(empty).Handle(ctx2, vc2))
	if !reflect.DeepEqual(ref.Table.Columns, got.Table.Columns) {
		t.Fatalf("empty-corpus columns differ from static baseline:\n got %+v\nwant %+v", got.Table.Columns, ref.Table.Columns)
	}
	if !reflect.DeepEqual(ref.Table.Rows, got.Table.Rows) {
		t.Fatalf("empty-corpus rows differ from static baseline:\n got %+v\nwant %+v", got.Table.Rows, ref.Table.Rows)
	}

	failed := dynamicDeps()
	failed.ListJobListTabSupport = func(context.Context) ([]*jobcategorypb.JobCategory, []*jobtemplatepb.JobTemplate, error) {
		return nil, nil, context.DeadlineExceeded
	}
	ctx3, vc3 := landingReq("")
	got3 := mustPageData(t, NewView(failed).Handle(ctx3, vc3))
	if !reflect.DeepEqual(ref.Table.Columns, got3.Table.Columns) || !reflect.DeepEqual(ref.Table.Rows, got3.Table.Rows) {
		t.Fatalf("failed-read landing differs from static baseline (must degrade, not error)")
	}
}

// TestLandingCategoryKnobGatesTheRead pins the config gate + statement budget:
// knob OFF → the tab-support closure is NEVER called (zero extra statements);
// knob ON → called EXACTLY ONCE (the +1 header read of the 11→12 Scenario-A
// budget, §3.6).
func TestLandingCategoryKnobGatesTheRead(t *testing.T) {
	calls := 0
	deps := dynamicDeps()
	inner := deps.ListJobListTabSupport
	deps.ListJobListTabSupport = func(ctx context.Context) ([]*jobcategorypb.JobCategory, []*jobtemplatepb.JobTemplate, error) {
		calls++
		return inner(ctx)
	}
	deps.Options.List.ColumnsByField = ""
	ctx, vc := landingReq("")
	_ = mustPageData(t, NewView(deps).Handle(ctx, vc))
	if calls != 0 {
		t.Fatalf("tab-support calls = %d with the knob OFF, want 0", calls)
	}

	deps.Options.List.ColumnsByField = "job_category"
	ctx2, vc2 := landingReq("")
	_ = mustPageData(t, NewView(deps).Handle(ctx2, vc2))
	if calls != 1 {
		t.Fatalf("tab-support calls = %d with the knob ON, want exactly 1 (+1 statement)", calls)
	}
}

// TestLandingEmptyScopedBandSkipsCategoryRead pins the §3.6 "empty scoped band
// unchanged" budget row: the header read sits AFTER the empty-band guard, so a
// scoped landing with no schedule in band issues NO tab-support read and keeps
// the static column set.
func TestLandingEmptyScopedBandSkipsCategoryRead(t *testing.T) {
	calls := 0
	deps := dynamicDeps()
	deps.ListJobListTabSupport = func(context.Context) ([]*jobcategorypb.JobCategory, []*jobtemplatepb.JobTemplate, error) {
		calls++
		return nil, nil, nil
	}
	ctx, vc := landingReq("past") // fixture has no inactive schedule → empty band
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))
	if calls != 0 {
		t.Fatalf("tab-support calls = %d on the empty scoped band, want 0", calls)
	}
	if len(pd.Table.Columns) != 3 || pd.Table.Columns[2].Key != "subjects" {
		t.Fatalf("empty scoped band columns = %+v, want the static 3-column set", pd.Table.Columns)
	}
}

// TestLandingUncategorizedBucket pins the §3.0 NULL policy: a summary subject
// whose template FK is NULL lands in the single named Uncategorized column as
// a BARE count (no eye — no addressable ?jc= target), never dropped, never
// duplicated into the real categories.
func TestLandingUncategorizedBucket(t *testing.T) {
	deps := dynamicDeps()
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		return &summarypb.ListJobTemplateSummariesResponse{Summaries: []*summarypb.JobTemplateSummary{
			{JobTemplateId: "t1", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a"},
			{JobTemplateId: "t9", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: ""}, // NULL FK
		}}, nil
	}
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))

	last := pd.Table.Columns[len(pd.Table.Columns)-1]
	if last.Key != "jc-uncategorized" || last.Label != "Uncategorized" {
		t.Fatalf("last column = {%q %q}, want the Uncategorized bucket", last.Key, last.Label)
	}
	cells := pd.Table.Rows[0].Cells
	uncat := cells[len(cells)-1]
	if uncat.Type != "composite" || uncat.Value != "1" {
		t.Fatalf("uncat cell = {Type:%q Value:%q}, want composite \"1\"", uncat.Type, uncat.Value)
	}
	if uncat.Composite.EyeHref != "" || uncat.Composite.EyeTestID != "" {
		t.Fatalf("uncat cell carries an eye (%q) — the NULL bucket must render a bare count", uncat.Composite.EyeHref)
	}
	// cat-a still counts only its own subject — no duplication.
	if cells[2].Value != "1" {
		t.Fatalf("cat-a count = %q, want \"1\" (NULL subject must not duplicate into categories)", cells[2].Value)
	}
}

// TestLandingNoUncategorizedColumnWhenFullyCategorized: with every subject in
// an active category and no NULL template stub, NO bucket column renders.
func TestLandingNoUncategorizedColumnWhenFullyCategorized(t *testing.T) {
	deps := dynamicDeps()
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))
	for _, c := range pd.Table.Columns {
		if c.Key == "jc-uncategorized" {
			t.Fatalf("Uncategorized column rendered with a fully-categorized corpus")
		}
	}
}

// TestLandingForeignCategoryFoldsIntoBucket pins the corpus/authority edge
// (§3.0, codex §4 pt 1): a subject whose effective category id is NOT in the
// active corpus (stale / inactive / foreign FK) folds into the Uncategorized
// bucket — it gets NO column of its own, is NOT dropped, and carries NO eye.
func TestLandingForeignCategoryFoldsIntoBucket(t *testing.T) {
	deps := dynamicDeps()
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		return &summarypb.ListJobTemplateSummariesResponse{Summaries: []*summarypb.JobTemplateSummary{
			{JobTemplateId: "t1", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-a"},
			{JobTemplateId: "t8", SubscriptionGroupId: "g-1", JobCount: 28, JobCategoryId: "cat-old"}, // inactive category
		}}, nil
	}
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))

	for _, c := range pd.Table.Columns {
		if c.Key == "jc-cat-old" {
			t.Fatalf("inactive category cat-old got its own column — corpus must stay ACTIVE-only")
		}
	}
	last := pd.Table.Columns[len(pd.Table.Columns)-1]
	if last.Key != "jc-uncategorized" {
		t.Fatalf("last column = %q, want jc-uncategorized (the fold target)", last.Key)
	}
	cells := pd.Table.Rows[0].Cells
	if got := cells[len(cells)-1].Value; got != "1" {
		t.Fatalf("folded bucket count = %q, want \"1\" (never dropped)", got)
	}
}

// TestLandingExportLinkNeutralizesRowID pins the Q-R9-7 export-link fix: the
// landing's per-row download URL pre-seeds an EMPTY "?id=" so the pyeza
// row-action JS's unconditional "&id=<row id>" append is inert server-side
// (Query().Get("id") returns the FIRST — empty — value; the bogus section-id
// row lookup never happens and the whole-section CSV is guaranteed).
func TestLandingExportLinkNeutralizesRowID(t *testing.T) {
	deps := dynamicDeps()
	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))
	acts := pd.Table.Rows[0].Actions
	if len(acts) != 2 {
		t.Fatalf("actions = %d, want 2 (view + download)", len(acts))
	}
	dl := acts[1]
	if dl.Action != "download" {
		t.Fatalf("action[1].Action = %q, want download", dl.Action)
	}
	if want := "/report-cards/section/g-1/export?id="; dl.URL != want {
		t.Fatalf("download URL = %q, want %q (empty id pre-seed neutralizes the JS row-id append)", dl.URL, want)
	}
}

// TestLandingHistoricalCountsBucketByCategory pins the historical (inactive-AY)
// fallback's bucketing (§3.0 + §3.6): counts ride the SAME member/job reads
// (no per-category fan-out); the effective category is the template's CURRENT
// FK via the tab-support stub map, the job's frozen snapshot ONLY for a
// template absent from the active corpus, and "" (no snapshot either) folds
// into the Uncategorized bucket.
func TestLandingHistoricalCountsBucketByCategory(t *testing.T) {
	deps := dynamicDeps()
	// Only an INACTIVE schedule + section exist → the historical path fills.
	deps.ListPriceSchedules = func(_ context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error) {
		if hasFilters(req.GetFilters().GetFilters()) {
			return &priceschedulepb.ListPriceSchedulesResponse{Data: []*priceschedulepb.PriceSchedule{
				{Id: "ps-past", Name: "AY 2024-25", Active: false},
			}}, nil
		}
		return &priceschedulepb.ListPriceSchedulesResponse{}, nil
	}
	deps.ListSubscriptionGroups = func(_ context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		if hasFilters(req.GetFilters().GetFilters()) {
			return &subscriptiongrouppb.ListSubscriptionGroupsResponse{Data: []*subscriptiongrouppb.SubscriptionGroup{
				{Id: "g-past", Name: "Grade 9 A", Active: false, PriceScheduleId: strptr("ps-past")},
			}}, nil
		}
		return &subscriptiongrouppb.ListSubscriptionGroupsResponse{}, nil
	}
	// The active-bound aggregate sees nothing historical.
	deps.ListJobTemplateSummaries = func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		return &summarypb.ListJobTemplateSummariesResponse{}, nil
	}
	deps.ListSubscriptionGroupMembers = func(_ context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
		return &subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse{Data: []*subscriptiongroupmemberpb.SubscriptionGroupMember{
			{SubscriptionGroupId: "g-past", SubscriptionId: "sub-1", ClientId: "c1"},
			{SubscriptionGroupId: "g-past", SubscriptionId: "sub-2", ClientId: "c2"},
		}}, nil
	}
	deps.ListJobs = func(_ context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
		return &jobpb.ListJobsResponse{Data: []*jobpb.Job{
			// t1 is in the ACTIVE stub map → current FK cat-a (snapshot ignored).
			{JobTemplateId: strptr("t1"), JobCategoryId: strptr("cat-b")},
			// t-old is absent from the active corpus → frozen snapshot cat-b.
			{JobTemplateId: strptr("t-old"), JobCategoryId: strptr("cat-b")},
			// t-none: absent + no snapshot → the Uncategorized bucket.
			{JobTemplateId: strptr("t-none")},
		}}, nil
	}

	ctx, vc := landingReq("")
	pd := mustPageData(t, NewView(deps).Handle(ctx, vc))
	if len(pd.Table.Rows) != 1 {
		t.Fatalf("rows = %d, want 1 (the historical section)", len(pd.Table.Rows))
	}
	cells := pd.Table.Rows[0].Cells
	// section, students, cat-a, cat-b, cat-c, uncategorized
	if len(cells) != 6 {
		t.Fatalf("cells = %d, want 6 (3 categories + bucket): %+v", len(cells), pd.Table.Columns)
	}
	if cells[1].Value != "2" {
		t.Fatalf("students = %q, want \"2\" (frozen roster)", cells[1].Value)
	}
	for i, want := range []string{"1", "1", "0", "1"} {
		if got := cells[2+i].Value; got != want {
			t.Fatalf("cell[%d] = %q, want %q (current-FK, snapshot-fallback, zero, uncategorized)", 2+i, got, want)
		}
	}
}

// TestCellAccessibleNameFrames pins the aria-frame contract: both placeholders
// substituted from DATA; a frame missing either noun falls back to "" so the
// typed cell composes its default BOTH-nouns name.
func TestCellAccessibleNameFrames(t *testing.T) {
	if got := cellAccessibleName("View {category} report cards for {section}", "Academic", "7A"); got != "View Academic report cards for 7A" {
		t.Fatalf("substituted frame = %q", got)
	}
	if got := cellAccessibleName("View report cards", "Academic", "7A"); got != "" {
		t.Fatalf("placeholder-less frame = %q, want \"\" (fall back to the typed default)", got)
	}
	if got := cellAccessibleName("View {category}", "Academic", "7A"); got != "" {
		t.Fatalf("single-noun frame = %q, want \"\" (must name BOTH)", got)
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

// i32ptr returns a pointer to v (proto optional int32 fields are *int32).
func i32ptr(v int32) *int32 { return &v }
