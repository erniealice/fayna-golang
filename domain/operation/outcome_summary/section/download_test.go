package section

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func drawerResponse() *exportpb.GetSubscriptionGroupOutcomeExportResponse {
	return &exportpb.GetSubscriptionGroupOutcomeExportResponse{
		Success: true,
		Context: &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: "group-1"},
		JobCategories: []*exportpb.JobCategoryOption{
			{JobCategoryId: "cat-b", Code: "academic", Name: "Academic", SortOrder: 2, FinalOutcomeAvailable: true, JobTemplatePhases: []*exportpb.JobTemplatePhaseOption{
				{Code: "mid", Name: "Midterm", SequenceOrder: 2},
				{Code: "ambiguous", Name: "Do not offer", SequenceOrder: 3, Ambiguous: true},
			}},
			{JobCategoryId: "cat-a", Code: "deportment", Name: "Deportment", SortOrder: 1, FinalOutcomeAvailable: false, JobTemplatePhases: []*exportpb.JobTemplatePhaseOption{
				{Code: "q1", Name: "Quarter 1", SequenceOrder: 1},
			}},
		},
	}
}

func drawerDeps(resp *exportpb.GetSubscriptionGroupOutcomeExportResponse) (*DrawerDeps, *int) {
	calls := 0
	return &DrawerDeps{
		Routes:               outcome_summary.Routes{SectionDownloadDrawerURL: "/report-cards/section/{id}/download", SectionExportURL: "/report-cards/section/{id}/export"},
		Labels:               outcome_summary.DefaultLabels(),
		ResolvePrincipalKind: func(context.Context) int32 { return outcome_summary.PrincipalKindOperatorOwner },
		Options: outcome_summary.Options{List: outcome_summary.ListOptions{Entity: outcome_summary.ListEntitySubscriptionGroup}, SectionExport: outcome_summary.SectionExportOptions{
			Enabled:             true,
			DefaultCategoryCode: "academic",
			ProfileByCategoryCode: map[string]bindingpb.RenderProfile{
				"academic": bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1,
			},
		}},
		GetSubscriptionGroupOutcomeExport: func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
			calls++
			if req.GetSubscriptionGroupId() != "group-1" || req.GetJobCategoryId() != "" || req.GetOutcomeSelector() != nil {
				return nil, context.Canceled
			}
			return resp, nil
		},
	}, &calls
}

func TestDownloadDrawer_CategoryRefreshBuildsPhaseOptions(t *testing.T) {
	deps, calls := drawerDeps(drawerResponse())
	viewUnderTest := NewDownloadDrawer(deps)
	request := httptest.NewRequest("GET", "/report-cards/section/group-1/download?job_category_id=cat-b", nil)
	request.SetPathValue("id", "group-1")
	ctx := view.WithUserPermissions(context.Background(), types.NewUserPermissions([]string{"subscription_group_outcome_export:read"}))
	result := viewUnderTest.Handle(ctx, &view.ViewContext{Request: request})
	if result.Error != nil || result.Template != "outcome-summary-section-download-drawer-form" {
		t.Fatalf("result = template %q error %v", result.Template, result.Error)
	}
	if *calls != 1 {
		t.Fatalf("composite calls = %d, want one", *calls)
	}
	data := result.Data.(*DrawerData)
	if len(data.Categories) != 2 || data.Categories[0].Value != "cat-a" || data.Categories[1].Value != "cat-b" {
		t.Fatalf("categories = %+v", data.Categories)
	}
	if !data.Categories[1].Selected || data.Categories[0].Selected {
		t.Fatalf("category selection = %+v", data.Categories)
	}
	if len(data.Periods) != 2 || data.Periods[0].Value != "phase:mid" || data.Periods[1].Value != "final" {
		t.Fatalf("academic periods = %+v", data.Periods)
	}
	if !data.Formats[0].Selected || data.Formats[1].Disabled {
		t.Fatalf("formats = %+v, want selected/enabled CSV and mapped PDF", data.Formats)
	}
	if data.RefreshAction != "/report-cards/section/group-1/download" || data.FormAction != "/report-cards/section/group-1/export" {
		t.Fatalf("actions = %q / %q", data.RefreshAction, data.FormAction)
	}

	// Invalid request category falls back to the trusted configured code, while
	// a missing configured code falls back to the first deterministic category.
	request = httptest.NewRequest("GET", "/report-cards/section/group-1/download?job_category_id=nope", nil)
	request.SetPathValue("id", "group-1")
	deps.Options.SectionExport.DefaultCategoryCode = "missing"
	result = viewUnderTest.Handle(ctx, &view.ViewContext{Request: request})
	data = result.Data.(*DrawerData)
	if len(data.Periods) != 1 || data.Periods[0].Value != "phase:q1" {
		t.Fatalf("first-category fallback periods = %+v", data.Periods)
	}
	if !data.Formats[0].Selected || !data.Formats[1].Disabled {
		t.Fatalf("unmapped category formats = %+v, want disabled PDF", data.Formats)
	}
}

func TestDownloadDrawer_FailsClosed(t *testing.T) {
	request := httptest.NewRequest("GET", "/report-cards/section/group-1/download", nil)
	request.SetPathValue("id", "group-1")
	ctx := view.WithUserPermissions(context.Background(), types.NewUserPermissions([]string{"subscription_group_outcome_export:read"}))
	for name, deps := range map[string]*DrawerDeps{
		"nil deps":         nil,
		"feature disabled": {Routes: outcome_summary.Routes{SectionDownloadDrawerURL: "/drawer", SectionExportURL: "/export"}, Labels: outcome_summary.DefaultLabels(), ResolvePrincipalKind: func(context.Context) int32 { return outcome_summary.PrincipalKindOperatorOwner }},
		"missing closure":  {Routes: outcome_summary.Routes{SectionDownloadDrawerURL: "/drawer", SectionExportURL: "/export"}, Labels: outcome_summary.DefaultLabels(), ResolvePrincipalKind: func(context.Context) int32 { return outcome_summary.PrincipalKindOperatorOwner }, Options: outcome_summary.Options{SectionExport: outcome_summary.SectionExportOptions{Enabled: true}}},
	} {
		t.Run(name, func(t *testing.T) {
			result := NewDownloadDrawer(deps).Handle(ctx, &view.ViewContext{Request: request})
			if name == "nil deps" {
				if result.StatusCode != 403 {
					t.Fatalf("nil deps status = %d, want 403", result.StatusCode)
				}
				return
			}
			if result.Error == nil {
				t.Fatal("expected configuration error")
			}
		})
	}

	forbidden := NewDownloadDrawer(&DrawerDeps{}).Handle(view.WithUserPermissions(context.Background(), types.NewUserPermissions(nil)), &view.ViewContext{Request: request})
	if forbidden.StatusCode != 403 {
		t.Fatalf("forbidden status = %d", forbidden.StatusCode)
	}
}

func TestDownloadDrawer_RequiresExportCapabilityAndAllowedKind(t *testing.T) {
	request := httptest.NewRequest("GET", "/report-cards/section/group-1/download", nil)
	request.SetPathValue("id", "group-1")
	deps, calls := drawerDeps(drawerResponse())
	viewUnderTest := NewDownloadDrawer(deps)
	for _, tc := range []struct {
		name       string
		codes      []string
		kind       int32
		wantStatus int
		wantCalls  int
	}{
		{name: "legacy list only", codes: []string{"job_outcome_summary:list"}, kind: outcome_summary.PrincipalKindOperatorOwner, wantStatus: 403},
		{name: "operator staff", codes: []string{"subscription_group_outcome_export:read"}, kind: outcome_summary.PrincipalKindOperatorStaff, wantStatus: 200, wantCalls: 1},
		{name: "unresolved", codes: []string{"subscription_group_outcome_export:read"}, kind: 0, wantStatus: 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := *calls
			deps.ResolvePrincipalKind = func(context.Context) int32 { return tc.kind }
			ctx := view.WithUserPermissions(context.Background(), types.NewUserPermissions(tc.codes))
			result := viewUnderTest.Handle(ctx, &view.ViewContext{Request: request})
			if result.StatusCode != tc.wantStatus || *calls != before+tc.wantCalls {
				t.Fatalf("status/calls=%d/%d, want %d/%d", result.StatusCode, *calls, tc.wantStatus, before+tc.wantCalls)
			}
		})
	}
}
