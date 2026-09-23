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
	deps.Routes.SubscriptionGroupDownloadDrawerURL = "/action/report-cards/group/{id}/download"
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
	if !strings.HasSuffix(page.DownloadDrawerURL, "?mode=fixed&job_category_id=cat-final") {
		t.Fatalf("download drawer URL = %q, want validated active category", page.DownloadDrawerURL)
	}
	if page.Table == nil || page.Table.PrimaryAction == nil {
		t.Fatal("computed report table has no primary download action")
	}
	if action := page.Table.PrimaryAction; action.ActionURL != page.DownloadDrawerURL || action.TestID != "rc-subscription-group-download-open" || action.Label != deps.Labels.SubscriptionGroupExport.DownloadAction || action.SheetTitle != deps.Labels.SubscriptionGroupExport.DrawerTitle {
		t.Fatalf("table primary download action = %+v, want the active category drawer", action)
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

	ctx := reportContextWithPermissions(t, deps.Routes, "sg-1", "", []string{"job_outcome_summary:list", "subscription_group_outcome_export:read", "attribute:list", "client_attribute:list"})
	page := reportPageData(t, NewView(deps).Handle(ctx, reportViewContext(t, deps.Routes, "sg-1", "")))
	if len(page.Table.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(page.Table.Rows))
	}
	if got := page.Table.Rows[0].Cells[0].Value; got != "1 Adams, Ann" {
		t.Fatalf("row 0 client = %q, want 1 Adams, Ann", got)
	}
	if got := page.Table.Rows[1].Cells[0].Value; got != "2 Baker, Bob" {
		t.Fatalf("row 1 client = %q, want 2 Baker, Bob", got)
	}
	wantHref := route.ResolveURL(deps.Routes.ClientCardURL, "id", "sg-1", "client_id", "client-a")
	if page.Table.Rows[0].Cells[0].Href != wantHref {
		t.Fatalf("client href = %q, want %q", page.Table.Rows[0].Cells[0].Href, wantHref)
	}
	if got := page.Table.Rows[0].Cells[2].Value; got != deps.Labels.SubscriptionGroup.RatingEmpty {
		t.Fatalf("suppressed value = %q, want rating-empty label %q", got, deps.Labels.SubscriptionGroup.RatingEmpty)
	}
	if got := page.Table.Rows[1].Cells[2].Value; got != "8.5" {
		t.Fatalf("score fallback = %q, want 8.5", got)
	}
	if page.Table.Rows[0].Cells[0].Type != "link" {
		t.Fatalf("client cell type = %q, want link", page.Table.Rows[0].Cells[0].Type)
	}
	// Cells[1] is the frozen per-row actions cell (view + download); it is
	// deliberately excluded from the "no per-cell link" matrix assertion below.
	for _, row := range page.Table.Rows {
		for _, cell := range row.Cells[2:] {
			if cell.Type == "link" || cell.Href != "" {
				t.Fatalf("matrix cell has a per-cell link: %+v", cell)
			}
		}
	}
}

// reportSingleRowMatrix builds a minimal explicitMatrix (one template column,
// one client row) for direct buildReportTable unit tests below — mirroring
// how page_test.go's buildRows tests call the static builder directly,
// bypassing the ctx/permission plumbing that decides WHICH gating booleans
// buildReportTable is called with (that decision is exercised end-to-end by
// the reportViewDeps()-based tests above/below).
func reportSingleRowMatrix(clientID, firstName, lastName string) *explicitMatrix {
	return &explicitMatrix{
		columns: []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Alpha"}},
		rows: []explicitRow{{
			clientID:  clientID,
			firstName: firstName,
			lastName:  lastName,
			cells: map[string]*exportpb.SubscriptionGroupOutcomeCell{
				"tmpl-a": reportCell("tmpl-a", "A", nil, true, true),
			},
		}},
	}
}

// TestBuildReportTable_ColumnOrderClientActionsTemplates pins AC-ROW-01's
// column order (client -> frozen actions -> template columns) and the
// freeze2 width/class contract the tabbed grid must match to align with the
// static grid's CSS (page.go buildColumns/buildGroupTable).
func TestBuildReportTable_ColumnOrderClientActionsTemplates(t *testing.T) {
	deps := reportViewDeps()
	deps.Routes.ClientDownloadDrawerURL = "/action/report-cards/section/{id}/student/{client_id}/download"
	matrix := reportSingleRowMatrix("client-1", "Ann", "Adams")

	table := buildReportTable(deps, matrix, "group-1", true, false)

	if len(table.Columns) != 3 {
		t.Fatalf("columns = %d, want client + actions + one template column", len(table.Columns))
	}
	if table.Columns[0].Key != "client" || table.Columns[0].Width != "14rem" {
		t.Fatalf("column 0 = %+v, want client column at 14rem", table.Columns[0])
	}
	if table.Columns[1].Key != actionsColumnKey || table.Columns[1].Label != "" || table.Columns[1].Width != "5rem" {
		t.Fatalf("column 1 = %+v, want blank frozen %q column at 5rem", table.Columns[1], actionsColumnKey)
	}
	if table.Columns[2].Key != "tmpl-tmpl-a" {
		t.Fatalf("column 2 = %+v, want the template column last", table.Columns[2])
	}
	if table.TableClass != "data-table-freeze2" {
		t.Fatalf("table class = %q, want data-table-freeze2", table.TableClass)
	}
	if table.ShowActions {
		t.Fatal("ShowActions must stay false: the actions cell is the frozen second column, not a trailing built-in actions column")
	}
	if len(table.Rows) != 1 || len(table.Rows[0].Cells) != 3 {
		t.Fatalf("row cells = %+v, want client + actions + one template cell", table.Rows)
	}
	if table.Rows[0].Cells[1].Type != "html" {
		t.Fatalf("actions cell type = %q, want html", table.Rows[0].Cells[1].Type)
	}

	// Cross-function contract with export.go: the actions column (and only
	// that column) is skipped by key, never by position.
	skip := exportSkipColumns(table.Columns)
	if skip[0] || skip[2] {
		t.Fatalf("exportSkipColumns skipped a data column: %+v", skip)
	}
	if !skip[1] {
		t.Fatalf("exportSkipColumns must skip the actions column: %+v", skip)
	}
}

// TestBuildReportTable_ExplicitExportOpensSameDrawerAsClientCard covers
// AC-ROW-01: the tabbed grid's per-row action must sit right after the
// client-name column and open the SAME Period/Format drawer as the
// client-card header, using the shared rowActionsCell helper (byte-identical
// to the static grid's buildRows output for the same inputs).
func TestBuildReportTable_ExplicitExportOpensSameDrawerAsClientCard(t *testing.T) {
	deps := reportViewDeps()
	deps.Routes.ClientDownloadDrawerURL = "/action/report-cards/section/{id}/student/{client_id}/download"
	deps.Labels.Client = outcome_summary.PeriodLabels{ViewAction: "View", DownloadAction: "Download"}
	deps.Labels.ClientDocumentDownload = outcome_summary.ClientDocumentDownloadLabels{DrawerTitle: "Download Progress Report"}
	matrix := reportSingleRowMatrix("client-1", "Ann", "Adams")

	table := buildReportTable(deps, matrix, "group-1", true, false)

	cell := string(table.Rows[0].Cells[1].HTML)
	wantDrawerURL := route.ResolveURL(deps.Routes.ClientDownloadDrawerURL, "id", "group-1", "client_id", "client-1")
	for _, want := range []string{
		"rc-view-client-1", "rc-download-client-1",
		`hx-get="` + wantDrawerURL + `"`,
		`hx-target="#sheetContent"`, `data-lf-action="sheet-open"`,
		`data-lf-sheet-title="Download Progress Report"`,
	} {
		if !strings.Contains(cell, want) {
			t.Errorf("actions cell missing %q: %s", want, cell)
		}
	}
	if strings.Contains(cell, "/document?") || strings.Contains(cell, " download>") {
		t.Errorf("explicit-export row action must open the drawer, not directly download: %s", cell)
	}
}

// TestBuildReportTable_LegacyFallbackKeepsPeriodlessFullCardDownload covers
// the legacy-detail-only branch of the shared rowActionsCell helper (reached
// in production only via the static grid, since narrowReportViewEnabled
// requires !CanLegacyDetail before routing here — exercised directly so the
// tabbed builder's wiring of the SAME helper/branch is pinned too).
func TestBuildReportTable_LegacyFallbackKeepsPeriodlessFullCardDownload(t *testing.T) {
	deps := reportViewDeps()
	matrix := reportSingleRowMatrix("client-1", "Ann", "Adams")

	table := buildReportTable(deps, matrix, "group-1", false, true)

	cell := string(table.Rows[0].Cells[1].HTML)
	wantDocumentURL := route.ResolveURL(deps.Routes.ClientDocumentURL, "id", "group-1", "client_id", "client-1") + "?format=pdf"
	for _, want := range []string{
		"rc-view-client-1", "rc-download-client-1",
		`href="` + wantDocumentURL + `"`,
		`hx-boost="false"`, " download>",
	} {
		if !strings.Contains(cell, want) {
			t.Errorf("legacy fallback action missing %q: %s", want, cell)
		}
	}
	if strings.Contains(cell, "data-lf-action=\"sheet-open\"") {
		t.Errorf("legacy fallback action must stay direct, not drawer-based: %s", cell)
	}
}

// TestBuildReportTable_NoDownloadPermissionKeepsViewOnly covers the neither
// branch: the view-client-card action always renders, but no download
// control appears without either capability.
func TestBuildReportTable_NoDownloadPermissionKeepsViewOnly(t *testing.T) {
	deps := reportViewDeps()
	matrix := reportSingleRowMatrix("client-1", "Ann", "Adams")

	table := buildReportTable(deps, matrix, "group-1", false, false)

	cell := string(table.Rows[0].Cells[1].HTML)
	if !strings.Contains(cell, "rc-view-client-1") {
		t.Fatalf("view action missing without download capability: %s", cell)
	}
	if strings.Contains(cell, "rc-download-") {
		t.Fatalf("no download capability must omit the download control entirely: %s", cell)
	}
}

func TestReportViewRequiresBothAttributeReadGrantsBeforeBandReads(t *testing.T) {
	for _, missing := range []string{"attribute:list", "client_attribute:list"} {
		t.Run(missing, func(t *testing.T) {
			deps := reportViewDeps()
			deps.Options.Row.GroupByField = "client_attributes.gender"
			deps.Options.SubscriptionGroupExport.GroupByAttributeModule = "client"
			attributeCalls, matrixCalls := 0, 0
			deps.ListAttributes = func(context.Context, *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error) {
				attributeCalls++
				return nil, nil
			}
			deps.ListClientAttributes = func(context.Context, *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error) {
				attributeCalls++
				return nil, nil
			}
			category := reportCategory("cat-a", "Academic", true)
			deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
				if req.GetOutcomeSelector() == nil {
					return reportOptions("sg-1", category), nil
				}
				matrixCalls++
				return reportMatrix("sg-1", category, []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "A"}}, nil), nil
			}
			permissions := []string{"job_outcome_summary:list", "subscription_group_outcome_export:read", "client_attribute:list"}
			if missing != "attribute:list" {
				permissions = append(permissions, "attribute:list")
			}
			if missing == "client_attribute:list" {
				permissions = []string{"job_outcome_summary:list", "subscription_group_outcome_export:read", "attribute:list"}
			}
			ctx := reportContextWithPermissions(t, deps.Routes, "sg-1", "", permissions)
			result := NewView(deps).Handle(ctx, reportViewContext(t, deps.Routes, "sg-1", ""))
			if result.StatusCode != 403 || result.Template != "forbidden" {
				t.Fatalf("result = template %q/status %d, want forbidden", result.Template, result.StatusCode)
			}
			if attributeCalls != 0 || matrixCalls != 0 {
				t.Fatalf("attribute/matrix reads = %d/%d, want zero", attributeCalls, matrixCalls)
			}
		})
	}
}

func TestReportViewUsesConfiguredGenderBandsAndNumbersAfterOrdering(t *testing.T) {
	deps := reportViewDeps()
	deps.Options.Row.GroupByField = "client_attributes.gender"
	deps.Options.Row.GroupValueOrder = []string{"Male", "Female"}
	deps.Options.Row.SortField = "last_name"
	deps.Options.SubscriptionGroupExport.GroupByAttributeModule = "client"
	deps.ListAttributes = func(context.Context, *commonpb.ListAttributesRequest) (*commonpb.ListAttributesResponse, error) {
		return &commonpb.ListAttributesResponse{Success: true, Data: []*commonpb.Attribute{{Id: "gender-id", Code: "gender", Module: "client", Active: true}}}, nil
	}
	deps.ListClientAttributes = func(_ context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error) {
		var rows []*clientattributepb.ClientAttribute
		for _, filter := range req.GetFilters().GetFilters() {
			if clients := filter.GetListFilter(); clients != nil {
				for _, id := range clients.GetValues() {
					if id == "male-a" || id == "male-z" {
						rows = append(rows, &clientattributepb.ClientAttribute{ClientId: id, AttributeId: "gender-id", Value: "Male", Active: true})
					} else if id == "female" {
						rows = append(rows, &clientattributepb.ClientAttribute{ClientId: id, AttributeId: "gender-id", Value: "Female", Active: true})
					}
				}
			}
		}
		return &clientattributepb.ListClientAttributesResponse{Success: true, Data: rows}, nil
	}
	category := reportCategory("cat-final", "Academic", true)
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return reportOptions("sg-1", category), nil
		}
		return reportMatrix("sg-1", category, []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Alpha"}}, []*exportpb.SubscriptionGroupOutcomeClientRow{
			reportClientRow("male-z", "[12] Zed Display", "", "Zulu", reportCell("tmpl-a", "Z", nil, true, true)),
			reportClientRow("female", "Beta Display", "Beta", "Beta", reportCell("tmpl-a", "F", nil, true, true)),
			reportClientRow("missing", "Missing Display", "Missing", "Missing", reportCell("tmpl-a", "M", nil, true, true)),
			reportClientRow("male-a", "Alpha Display", "Alpha", "Alpha", reportCell("tmpl-a", "A", nil, true, true)),
		}), nil
	}

	ctx := reportContextWithPermissions(t, deps.Routes, "sg-1", "", []string{"job_outcome_summary:list", "subscription_group_outcome_export:read", "attribute:list", "client_attribute:list"})
	page := reportPageData(t, NewView(deps).Handle(ctx, reportViewContext(t, deps.Routes, "sg-1", "")))
	if len(page.Table.Groups) != 3 {
		t.Fatalf("groups = %+v, want Male, Female, and missing-value bands", page.Table.Groups)
	}
	want := []struct{ id, title, name string }{
		{"rc-band-male", "Male", "1 Alpha, Alpha"},
		{"rc-band-female", "Female", "3 Beta, Beta"},
		{"rc-band-none", "—", "4 Missing, Missing"},
	}
	if got := page.Table.Groups[0].Rows[1].Cells[0].Value; got != "2 Zed Display" {
		t.Fatalf("second Male row = %q, want number 2 after last-name ordering", got)
	}
	for i, expected := range want {
		group := page.Table.Groups[i]
		if group.ID != expected.id || group.Title != expected.title || len(group.Rows) == 0 {
			t.Fatalf("group %d = %+v, want id/title %q/%q", i, group, expected.id, expected.title)
		}
		if got := group.Rows[0].Cells[0].Value; got != expected.name {
			t.Fatalf("group %s first row = %q, want %q", group.ID, got, expected.name)
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
