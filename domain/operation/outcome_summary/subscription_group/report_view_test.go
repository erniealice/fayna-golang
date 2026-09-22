package subscription_group

import (
	"context"
	"errors"
	"html/template"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	workspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/workspace_user"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	subscriptiongroupworkspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_workspace_user"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func TestSubscriptionGroupReportViewUsesOnlyAggregate(t *testing.T) {
	deps := reportFatalDeps(t)
	deps.Options.List.Entity = outcome_summary.ListEntitySubscriptionGroup
	deps.Options.SubscriptionGroupExport.Enabled = true
	deps.ResolvePrincipalKind = func(context.Context) int32 { return outcome_summary.PrincipalKindStaff }
	calls := 0
	category := reportCategory("cat-a", "Alpha", true)
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		calls++
		if req.GetOutcomeSelector() == nil {
			return reportOptions("sg-1", category), nil
		}
		return reportMatrix("sg-1", category, []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Alpha"}}, []*exportpb.SubscriptionGroupOutcomeClientRow{
			reportClientRow("client-a", "", "Ava", "Adams", reportCell("tmpl-a", "A", nil, true, true)),
		}), nil
	}

	result := NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", ""), reportViewContext(t, deps.Routes, "sg-1", ""))
	if result.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", result.StatusCode)
	}
	if calls != 2 {
		t.Fatalf("aggregate calls = %d, want options plus one matrix", calls)
	}
	page, ok := result.Data.(*PageData)
	if !ok || page.Table == nil {
		t.Fatalf("result data = %T, want populated *PageData", result.Data)
	}
}

func TestSubscriptionGroupReportViewUnassignedIsForbidden(t *testing.T) {
	deps := reportViewDeps()
	deps.GetSubscriptionGroupOutcomeExport = func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		return &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: true}, nil
	}
	result := NewView(deps).Handle(reportContext(t, deps.Routes, "sg-foreign", ""), reportViewContext(t, deps.Routes, "sg-foreign", ""))
	assertReportForbidden(t, result)
}

func TestSubscriptionGroupReportViewCategoryTabsAndDefault(t *testing.T) {
	deps := reportViewDeps()
	categories := []*exportpb.JobCategoryOption{
		reportCategory("cat-empty", "Empty", false),
		reportCategory("cat-final", "Final", true),
	}
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return reportOptions("sg-1", categories...), nil
		}
		if req.GetJobCategoryId() != "cat-final" {
			t.Fatalf("matrix category = %q, want cat-final", req.GetJobCategoryId())
		}
		return reportMatrix("sg-1", categories[1], []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Alpha"}}, nil), nil
	}

	result := NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", "jc=unknown"), reportViewContext(t, deps.Routes, "sg-1", "jc=unknown"))
	page := reportPageData(t, result)
	if page.ActiveTab != groupTabKey("cat-final") {
		t.Fatalf("active tab = %q, want final category", page.ActiveTab)
	}
	if page.TabsAria != deps.Labels.SubscriptionGroup.CategoryTabsAriaLabel {
		t.Fatalf("tabs aria = %q, want %q", page.TabsAria, deps.Labels.SubscriptionGroup.CategoryTabsAriaLabel)
	}
	if len(page.TabItems) != 2 || page.TabItems[0].Label != "Empty" || page.TabItems[1].Label != "Final" {
		t.Fatalf("tabs = %+v, want one tab per aggregate category", page.TabItems)
	}
	wantHref := groupCategoryURL(route.ResolveURL(deps.Routes.SubscriptionGroupURL, "id", "sg-1"), "cat-final")
	if page.TabItems[1].Href != wantHref {
		t.Fatalf("final tab href = %q, want %q", page.TabItems[1].Href, wantHref)
	}
}

func TestSubscriptionGroupReportViewRowsSortedAndLinked(t *testing.T) {
	deps := reportViewDeps()
	category := reportCategory("cat-final", "Final", true)
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return reportOptions("sg-1", category), nil
		}
		return reportMatrix("sg-1", category,
			[]*exportpb.JobTemplateColumn{
				{JobTemplateId: "tmpl-a", DisplayName: "Alpha"},
				{JobTemplateId: "tmpl-b", DisplayName: "Beta"},
			},
			[]*exportpb.SubscriptionGroupOutcomeClientRow{
				reportClientRow("client-b", "B Display", "Bob", "Baker", reportCell("tmpl-b", "B", nil, true, true), reportCell("tmpl-a", "", floatptr(8.5), true, true)),
				reportClientRow("client-a", "A Display", "Ann", "Adams", reportCell("tmpl-a", "1", nil, true, false), reportCell("tmpl-b", "A", nil, true, true)),
			}), nil
	}

	page := reportPageData(t, NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", ""), reportViewContext(t, deps.Routes, "sg-1", "")))
	if len(page.Table.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(page.Table.Rows))
	}
	if got := page.Table.Rows[0].Cells[0].Value; got != "Adams, Ann" {
		t.Fatalf("row 0 client = %q, want Adams, Ann", got)
	}
	if got := page.Table.Rows[1].Cells[0].Value; got != "Baker, Bob" {
		t.Fatalf("row 1 client = %q, want Baker, Bob", got)
	}
	wantHref := route.ResolveURL(deps.Routes.ClientCardURL, "id", "sg-1", "client_id", "client-a")
	if page.Table.Rows[0].Cells[0].Href != wantHref {
		t.Fatalf("client href = %q, want %q", page.Table.Rows[0].Cells[0].Href, wantHref)
	}
	if got := page.Table.Rows[0].Cells[1].Value; got != deps.Labels.SubscriptionGroup.RatingEmpty {
		t.Fatalf("suppressed value = %q, want rating-empty label %q", got, deps.Labels.SubscriptionGroup.RatingEmpty)
	}
	if got := page.Table.Rows[1].Cells[1].Value; got != "8.5" {
		t.Fatalf("score fallback = %q, want 8.5", got)
	}
	if page.Table.Rows[0].Cells[0].Type != "link" {
		t.Fatalf("client cell type = %q, want link", page.Table.Rows[0].Cells[0].Type)
	}
	for _, row := range page.Table.Rows {
		for _, cell := range row.Cells[1:] {
			if cell.Type == "link" || cell.Href != "" {
				t.Fatalf("matrix cell has a per-cell link: %+v", cell)
			}
		}
	}
}

func TestSubscriptionGroupReportViewUsesRouteAndLabels(t *testing.T) {
	cases := []struct {
		name           string
		listRoute      string
		groupRoute     string
		clientRoute    string
		groupTitle     string
		clientLabel    string
		tabsLabel      string
		wantListHref   string
		wantGroupHref  string
		wantClientHref string
	}{
		{name: "generic", listRoute: "/generic/reports", groupRoute: "/generic/group/{id}", clientRoute: "/generic/group/{id}/client/{client_id}", groupTitle: "Group outcomes", clientLabel: "Client", tabsLabel: "Categories", wantListHref: "/generic/reports", wantGroupHref: "/generic/group/sg-1", wantClientHref: "/generic/group/sg-1/client/client-a"},
		{name: "education", listRoute: "/tier/report-cards", groupRoute: "/tier/group/{id}", clientRoute: "/tier/group/{id}/client/{client_id}", groupTitle: "Report Cards", clientLabel: "Learner", tabsLabel: "Periods", wantListHref: "/tier/report-cards", wantGroupHref: "/tier/group/sg-1", wantClientHref: "/tier/group/sg-1/client/client-a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := reportViewDeps()
			deps.Routes.ListURL = tc.listRoute
			deps.Routes.SubscriptionGroupURL = tc.groupRoute
			deps.Routes.ClientCardURL = tc.clientRoute
			deps.Labels.SubscriptionGroup.Title = tc.groupTitle
			deps.Labels.SubscriptionGroup.ClientColumn = tc.clientLabel
			deps.Labels.SubscriptionGroup.CategoryTabsAriaLabel = tc.tabsLabel
			category := reportCategory("cat-a", "Period A", true)
			deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
				if req.GetOutcomeSelector() == nil {
					return reportOptions("sg-1", category), nil
				}
				return reportMatrix("sg-1", category, []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Item A"}}, []*exportpb.SubscriptionGroupOutcomeClientRow{
					reportClientRow("client-a", "", "Ava", "Adams", reportCell("tmpl-a", "A", nil, true, true)),
				}), nil
			}

			viewContext := reportViewContext(t, deps.Routes, "sg-1", "")
			if got := route.ResolveURL(deps.Routes.ClientCardURL, "id", "sg-1", "client_id", "client-a"); got != tc.wantClientHref {
				t.Fatalf("client href = %q, want %q", got, tc.wantClientHref)
			}
			page := reportPageData(t, NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", ""), viewContext))
			if page.PageData.Title != tc.groupTitle || page.PageData.HeaderBreadcrumb != tc.groupTitle {
				t.Fatalf("group labels did not flow: title=%q breadcrumb=%q", page.PageData.Title, page.PageData.HeaderBreadcrumb)
			}
			if page.PageData.HeaderBreadcrumbURL != tc.wantListHref {
				t.Fatalf("list href = %q, want %q", page.PageData.HeaderBreadcrumbURL, tc.wantListHref)
			}
			if page.TabsAria != tc.tabsLabel || page.Table.Columns[0].Label != tc.clientLabel {
				t.Fatalf("labels did not flow: tabs=%q client-column=%q", page.TabsAria, page.Table.Columns[0].Label)
			}
			if page.Table.Rows[0].Cells[0].Href != tc.wantClientHref {
				t.Fatalf("row client href = %q, want %q", page.Table.Rows[0].Cells[0].Href, tc.wantClientHref)
			}
			wantGroupHref := tc.wantGroupHref + "?jc=cat-a"
			if page.TabItems[0].Href != wantGroupHref {
				t.Fatalf("category href = %q, want %q", page.TabItems[0].Href, wantGroupHref)
			}
		})
	}
}

func TestReportViewForbiddenResponsesAreIndistinguishable(t *testing.T) {
	cases := []struct {
		name     string
		response *exportpb.GetSubscriptionGroupOutcomeExportResponse
		err      error
	}{
		{name: "nil aggregate context", response: &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: true}},
		{name: "different group context", response: &exportpb.GetSubscriptionGroupOutcomeExportResponse{Success: true, Context: &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: "sg-other"}}},
		{name: "aggregate error", response: reportOptions("sg-1", reportCategory("cat-a", "Alpha", true)), err: errors.New("aggregate read failed")},
	}
	var wantTemplate, wantPerm string
	var wantStatus int
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := reportViewDeps()
			deps.GetSubscriptionGroupOutcomeExport = func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
				return tc.response, tc.err
			}
			result := NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", ""), reportViewContext(t, deps.Routes, "sg-1", ""))
			data, ok := result.Data.(*view.ForbiddenPageData)
			if !ok {
				t.Fatalf("data = %T, want forbidden data", result.Data)
			}
			if result.Template != "forbidden" || result.StatusCode != 403 || data.Perm != "subscription_group_outcome_export:read" {
				t.Fatalf("result = template %q/status %d/perm %q, want export forbidden", result.Template, result.StatusCode, data.Perm)
			}
			if wantTemplate == "" {
				wantTemplate, wantStatus, wantPerm = result.Template, result.StatusCode, data.Perm
				return
			}
			if result.Template != wantTemplate || result.StatusCode != wantStatus || data.Perm != wantPerm {
				t.Fatalf("result = template %q/status %d/perm %q, want %q/%d/%q", result.Template, result.StatusCode, data.Perm, wantTemplate, wantStatus, wantPerm)
			}
		})
	}
}

func TestReportViewNamesAndJcAreNotMarkedSafe(t *testing.T) {
	malicious := `<script>"'&`
	category := reportCategory(malicious, malicious, true)
	deps := reportViewDeps()
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return reportOptions("sg-1", category), nil
		}
		if req.GetJobCategoryId() != malicious {
			t.Fatalf("category id = %q, want query value", req.GetJobCategoryId())
		}
		return reportMatrix("sg-1", category, []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Item"}}, []*exportpb.SubscriptionGroupOutcomeClientRow{
			reportClientRow("client-a", malicious, "", "", reportCell("tmpl-a", "A", nil, true, true)),
		}), nil
	}

	query := url.Values{"jc": []string{malicious}}.Encode()
	page := reportPageData(t, NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", query), reportViewContext(t, deps.Routes, "sg-1", query)))
	if got := assertReportPlainString(t, "category label", page.TabItems[0].Label); got != malicious {
		t.Fatalf("category label = %q, want plain category name", got)
	}
	if got := assertReportPlainString(t, "client name", page.Table.Rows[0].Cells[0].Value); !strings.HasSuffix(got, malicious) {
		t.Fatalf("client name = %q, want plain client name", got)
	}
	for _, tab := range page.TabItems {
		if strings.Contains(tab.Href, malicious) {
			t.Fatalf("raw category value leaked into tab href: %q", tab.Href)
		}
		parsed, err := url.Parse(tab.Href)
		if err != nil || parsed.Query().Get("jc") != malicious {
			t.Fatalf("tab href = %q, want encoded jc value", tab.Href)
		}
	}
	if strings.Contains(page.Table.Rows[0].Cells[0].Href, malicious) {
		t.Fatalf("raw client value leaked into client href: %q", page.Table.Rows[0].Cells[0].Href)
	}
}

func assertReportPlainString(t *testing.T, label string, value any) string {
	t.Helper()
	if _, safe := value.(template.HTML); safe {
		t.Fatalf("%s is template.HTML", label)
	}
	typeOfValue := reflect.TypeOf(value)
	if typeOfValue == nil || typeOfValue.Kind() != reflect.String {
		t.Fatalf("%s type = %v, want plain string", label, typeOfValue)
	}
	return reflect.ValueOf(value).String()
}

func TestSubscriptionGroupReportViewNoFinalIsNotComputed(t *testing.T) {
	deps := reportViewDeps()
	category := reportCategory("cat-pending", "Pending", false)
	calls := 0
	deps.GetSubscriptionGroupOutcomeExport = func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		calls++
		return reportOptions("sg-1", category), nil
	}

	page := reportPageData(t, NewView(deps).Handle(reportContext(t, deps.Routes, "sg-1", ""), reportViewContext(t, deps.Routes, "sg-1", "")))
	if calls != 1 {
		t.Fatalf("aggregate calls = %d, want options only for unavailable final", calls)
	}
	if !page.NotComputed || page.Table != nil {
		t.Fatalf("not-computed page = %+v, want banner without table", page)
	}
	if page.Banner != deps.Labels.SubscriptionGroup.NotComputedBanner {
		t.Fatalf("banner = %q, want %q", page.Banner, deps.Labels.SubscriptionGroup.NotComputedBanner)
	}
}

func TestSubscriptionGroupLegacyPathUnchangedForReadHolder(t *testing.T) {
	deps := reportViewDeps()
	aggregateCalled := false
	deps.GetSubscriptionGroupOutcomeExport = func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		aggregateCalled = true
		return nil, nil
	}
	deps.ListSubscriptionGroups = func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		return &subscriptiongrouppb.ListSubscriptionGroupsResponse{Data: []*subscriptiongrouppb.SubscriptionGroup{{Id: "sg-1", Name: "Group", Active: true}}}, nil
	}
	ctx := reportContextWithPermissions(t, deps.Routes, "sg-1", "", []string{"job_outcome_summary:list", "job_outcome_summary:read"})
	result := NewView(deps).Handle(ctx, reportViewContext(t, deps.Routes, "sg-1", ""))
	if result.StatusCode != 200 || result.Template != "outcome-summary-subscription-group" {
		t.Fatalf("legacy result = template %q/status %d, want legacy success", result.Template, result.StatusCode)
	}
	if aggregateCalled {
		t.Fatal("narrow aggregate was called for a legacy read holder")
	}
}

func TestNarrowPagesHaveNoVerticalLiterals(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Dir(filename)
	paths := []string{
		filepath.Join(root, "report_view.go"),
		filepath.Join(root, "..", "client_card", "report_view.go"),
	}
	stringLiteral := regexp.MustCompile(`(?s)"(?:\\.|[^"\\])*"`)
	verticalLiteral := regexp.MustCompile(`(?i)\b(student|section|teacher|grade|class)\b`)
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read narrow view source %s: %v", path, err)
		}
		for _, literal := range stringLiteral.FindAll(raw, -1) {
			value, err := strconv.Unquote(string(literal))
			if err == nil && verticalLiteral.MatchString(value) {
				t.Errorf("vertical literal in %s: %s", path, literal)
			}
		}
	}
}

func reportViewDeps() *Deps {
	deps := &Deps{
		Routes: outcome_summary.DefaultRoutes(),
		Labels: outcome_summary.DefaultLabels(),
	}
	deps.Options.List.Entity = outcome_summary.ListEntitySubscriptionGroup
	deps.Options.SubscriptionGroupExport.Enabled = true
	deps.ResolvePrincipalKind = func(context.Context) int32 { return outcome_summary.PrincipalKindStaff }
	return deps
}

func reportFatalDeps(t *testing.T) *Deps {
	t.Helper()
	deps := reportViewDeps()
	unexpected := func() { t.Helper(); t.Fatal("unexpected dependency call") }
	deps.ListJobCategories = func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListSubscriptionGroups = func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListSubscriptionGroupMembers = func(context.Context, *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListJobs = func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListJobTemplates = func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListClients = func(context.Context, *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListJobOutcomeSummarys = func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListClientAttributes = func(context.Context, *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ResolveAttributeIDByCode = func(context.Context, string) (string, error) {
		unexpected()
		return "", nil
	}
	deps.ListAttributes = func(context.Context, *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ResolveSubscriptionGroupDocumentTemplate = func(context.Context, *exportpb.ResolveSubscriptionGroupOutcomeDocumentForRenderRequest) (*outcome_summary.ResolvedSubscriptionGroupDocumentTemplate, error) {
		unexpected()
		return nil, nil
	}
	deps.GeneratePDF = func([]byte, map[string]any) ([]byte, error) {
		unexpected()
		return nil, nil
	}
	deps.ListJobPhases = func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListJobTasks = func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListTaskOutcomes = func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListSubscriptionGroupWorkspaceUsers = func(context.Context, *subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersRequest) (*subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersResponse, error) {
		unexpected()
		return nil, nil
	}
	deps.ListWorkspaceUsers = func(context.Context, *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error) {
		unexpected()
		return nil, nil
	}
	return deps
}

func reportViewContext(t *testing.T, routes outcome_summary.Routes, subscriptionGroupID, query string) *view.ViewContext {
	t.Helper()
	requestURL := route.ResolveURL(routes.SubscriptionGroupURL, "id", subscriptionGroupID)
	if query != "" {
		requestURL += "?" + query
	}
	request := httptest.NewRequest("GET", requestURL, nil)
	request.SetPathValue("id", subscriptionGroupID)
	return &view.ViewContext{Request: request, CurrentPath: request.URL.Path, CacheVersion: "test"}
}

func reportContext(t *testing.T, routes outcome_summary.Routes, subscriptionGroupID, query string) context.Context {
	return reportContextWithPermissions(t, routes, subscriptionGroupID, query, []string{
		"job_outcome_summary:list",
		"subscription_group_outcome_export:read",
	})
}

func reportContextWithPermissions(t *testing.T, routes outcome_summary.Routes, subscriptionGroupID, query string, permissions []string) context.Context {
	t.Helper()
	vc := reportViewContext(t, routes, subscriptionGroupID, query)
	return view.WithUserPermissions(vc.Request.Context(), types.NewUserPermissions(permissions))
}

func reportPageData(t *testing.T, result view.ViewResult) *PageData {
	t.Helper()
	if result.StatusCode != 200 {
		t.Fatalf("status = %d, want 200; data=%T", result.StatusCode, result.Data)
	}
	page, ok := result.Data.(*PageData)
	if !ok {
		t.Fatalf("data = %T, want *PageData", result.Data)
	}
	return page
}

func assertReportForbidden(t *testing.T, result view.ViewResult) {
	t.Helper()
	if result.StatusCode != 403 || result.Template != "forbidden" {
		t.Fatalf("result = template %q/status %d, want forbidden", result.Template, result.StatusCode)
	}
}

func reportOptions(subscriptionGroupID string, categories ...*exportpb.JobCategoryOption) *exportpb.GetSubscriptionGroupOutcomeExportResponse {
	return &exportpb.GetSubscriptionGroupOutcomeExportResponse{
		Context:       &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: subscriptionGroupID, SubscriptionGroupName: "Group"},
		JobCategories: categories,
		Success:       true,
	}
}

func reportMatrix(subscriptionGroupID string, category *exportpb.JobCategoryOption, columns []*exportpb.JobTemplateColumn, rows []*exportpb.SubscriptionGroupOutcomeClientRow) *exportpb.GetSubscriptionGroupOutcomeExportResponse {
	return &exportpb.GetSubscriptionGroupOutcomeExportResponse{
		Context:            &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: subscriptionGroupID, SubscriptionGroupName: "Group"},
		JobCategories:      []*exportpb.JobCategoryOption{category},
		JobTemplateColumns: columns,
		ClientRows:         rows,
		Success:            true,
	}
}

func reportCategory(id, name string, final bool) *exportpb.JobCategoryOption {
	return &exportpb.JobCategoryOption{JobCategoryId: id, Name: name, FinalOutcomeAvailable: final}
}

func reportClientRow(clientID, displayName, firstName, lastName string, cells ...*exportpb.SubscriptionGroupOutcomeCell) *exportpb.SubscriptionGroupOutcomeClientRow {
	return &exportpb.SubscriptionGroupOutcomeClientRow{ClientId: clientID, ClientName: displayName, ClientFirstName: firstName, ClientLastName: lastName, Cells: cells}
}

func reportCell(templateID, label string, score *float64, hasMarks, hasPositiveMark bool) *exportpb.SubscriptionGroupOutcomeCell {
	cell := &exportpb.SubscriptionGroupOutcomeCell{
		JobTemplateId: templateID,
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{
			HasMarks:        hasMarks,
			HasPositiveMark: hasPositiveMark,
		},
	}
	if label != "" {
		cell.ScaledLabel = &label
	}
	cell.ScaledScore = score
	return cell
}

func floatptr(value float64) *float64 { return &value }
