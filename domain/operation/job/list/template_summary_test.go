package list

import (
	"context"
	"errors"
	"testing"

	"github.com/erniealice/espyna-golang/shared/tableparams"
	job "github.com/erniealice/fayna-golang/domain/operation/job"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

func TestDeliverySummaryPropagatesQueryError(t *testing.T) {
	t.Parallel()

	queryErr := errors.New("summary service unavailable")
	deps := &ListViewDeps{
		Labels: job.DefaultLabels(),
		ListJobTemplateSummaries: func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
			return nil, queryErr
		},
	}

	t.Run("flat path", func(t *testing.T) {
		t.Parallel()
		if _, _, err := buildDeliverySummaryTable(context.Background(), deps, "active", summaryParams(), "", false, nil); !errors.Is(err, queryErr) {
			t.Fatalf("buildDeliverySummaryTable() err = %v, want %v", err, queryErr)
		}
	})

	t.Run("tabbed path", func(t *testing.T) {
		t.Parallel()
		if _, _, err := buildDeliverySummaryTable(context.Background(), deps, "active", summaryParams(), "", true, nil); !errors.Is(err, queryErr) {
			t.Fatalf("buildDeliverySummaryTableTabbed() err = %v, want %v", err, queryErr)
		}
	})
}

func TestBuildDeliverySummaryTable_mapsTemplateSummaryRows(t *testing.T) {
	t.Parallel()

	deps := &ListViewDeps{
		Labels: job.DefaultLabels(),
		ListJobTemplateSummaries: func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
			return &summarypb.ListJobTemplateSummariesResponse{
				Summaries: []*summarypb.JobTemplateSummary{
					{
						JobTemplateId:         "tmpl-a",
						SubscriptionGroupId:   "grp-a",
						JobTemplateName:       "Math",
						SubscriptionGroupName: "Grade 9",
						Deliverers:            []*summarypb.Deliverer{{StaffName: "Zed"}},
						JobCount:              4,
						PriceScheduleName:     "AY 2024",
					},
					{
						JobTemplateId:         "tmpl-b",
						SubscriptionGroupId:   "grp-b",
						JobTemplateName:       "Science",
						SubscriptionGroupName: "Grade 10",
						Deliverers:            []*summarypb.Deliverer{{StaffName: "Zulu"}, {StaffName: "Amy"}},
						JobCount:              12,
						PriceScheduleName:     "AY 2026",
					},
					{
						JobTemplateId:         "tmpl-c",
						SubscriptionGroupId:   "grp-c",
						JobTemplateName:       "History",
						SubscriptionGroupName: "Grade 11",
						Deliverers:            []*summarypb.Deliverer{{StaffName: "Sam"}},
						JobCount:              7,
						PriceScheduleName:     "AY 2025",
					},
				},
			}, nil
		},
	}

	table, _, err := buildDeliverySummaryTable(context.Background(), deps, "active", summaryParams(), "", false, nil)
	if err != nil {
		t.Fatalf("buildDeliverySummaryTable() err = %v", err)
	}

	if len(table.Rows) != 3 {
		t.Fatalf("table row count = %d, want 3", len(table.Rows))
	}
	var scienceRow types.TableRow
	for _, r := range table.Rows {
		if r.Cells[0].Value == "Science" {
			scienceRow = r
			break
		}
	}
	if scienceRow.Cells == nil {
		t.Fatal(`missing "Science" row`)
	}
	if got, want := scienceRow.Cells[2].Value, "Amy, Zulu"; got != want {
		t.Fatalf("Science row deliverer = %q, want %q", got, want)
	}
	if got, want := scienceRow.DataAttrs["schedule"], "AY 2026"; got != want {
		t.Fatalf("Science row schedule = %q, want %q", got, want)
	}
}

func summaryParams() tableparams.TableQueryParams {
	return tableparams.TableQueryParams{Page: 1, PageSize: 25, SortColumn: "group", SortDir: "asc"}
}

func TestBuildDeliverySummaryTable_forwardsServerPageContract(t *testing.T) {
	t.Parallel()
	currentPage, totalPages := int32(3), int32(4)
	params := tableparams.TableQueryParams{Page: 3, PageSize: 250, Search: "math", SortColumn: "name", SortDir: "desc"}
	var gotReq *summarypb.ListJobTemplateSummariesRequest
	deps := &ListViewDeps{
		Labels: job.DefaultLabels(),
		Routes: job.DefaultRoutes(),
		ListJobTemplateSummaries: func(_ context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
			gotReq = req
			return &summarypb.ListJobTemplateSummariesResponse{
				Summaries:         []*summarypb.JobTemplateSummary{{JobTemplateId: "fallback", JobTemplateName: "Conduct", JobCount: 99, TemplateGrainFallback: true}},
				Pagination:        &commonpb.PaginationResponse{TotalItems: 321, CurrentPage: &currentPage, TotalPages: &totalPages},
				JobCategoryCounts: []*summarypb.JobCategorySummaryCount{{JobCategoryId: "cat-empty-page", SummaryCount: 77}},
			}, nil
		},
	}

	table, counts, err := buildDeliverySummaryTable(context.Background(), deps, "completed", params, "cat/a b", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if gotReq.GetStatus() != "JOB_STATUS_COMPLETED" || gotReq.GetJobCategoryId() != "cat/a b" || !gotReq.GetIncludeTemplateFallback() {
		t.Fatalf("forwarded request = %+v", gotReq)
	}
	if got, want := gotReq.GetPagination().GetOffset().GetPage(), int32(3); got != want {
		t.Fatalf("page = %d, want %d", got, want)
	}
	if got, want := gotReq.GetPagination().GetLimit(), int32(maxTemplateSummaryPageSize); got != want {
		t.Fatalf("limit = %d, want %d", got, want)
	}
	if got := gotReq.GetSearch().GetQuery(); got != "math" {
		t.Fatalf("search = %q", got)
	}
	if got := gotReq.GetSort().GetFields()[0].GetField(); got != "name" {
		t.Fatalf("sort field = %q", got)
	}
	if got := len(gotReq.GetSort().GetFields()); got != 1 {
		t.Fatalf("summary sort fields = %d, want only the requested field; composite tie-breakers are adapter-owned", got)
	}
	if got := counts["cat-empty-page"]; got != 77 {
		t.Fatalf("response count = %d, want 77", got)
	}
	if got := table.Rows[0].Cells[3].Value; got != "" {
		t.Fatalf("fallback item cell = %q, want blank", got)
	}
	sp := table.ServerPagination
	if sp.CurrentPage != 3 || sp.TotalRows != 321 || sp.TotalPages != 4 || sp.PageSize != maxTemplateSummaryPageSize || sp.SearchQuery != "math" || sp.SortColumn != "name" || sp.SortDirection != "desc" {
		t.Fatalf("server pagination = %+v", sp)
	}
	// Pagination/refresh MUST target the table-only endpoint (Routes.TableURL),
	// NOT the full-page ListURL. Pointing it at ListURL makes the JS full-card
	// swap re-render the whole page shell and nest it inside the table-card.
	if got, want := sp.PaginationURL, "/action/job/table/completed?jc=cat%2Fa+b"; got != want {
		t.Fatalf("pagination URL = %q, want %q (table-only endpoint, not full-page ListURL)", got, want)
	}
}

func TestTemplateSummaryDownloadAction_usesGroupDrawerAndExactContract(t *testing.T) {
	t.Parallel()

	deps := drawerSummaryDeps()
	deps.MatrixDownloadDrawerURL = "/action/outcome-matrix/download/{id}"
	deps.MatrixGroupDownloadDrawerURL = "/action/outcome-matrix/download/{id}/subscription-group/{group_id}"
	deps.MatrixDownloadDrawerTitle = "Export outcomes"
	deps.Options.RowLink.ScopeByField = job.RowLinkEntitySubscriptionGroup

	table := buildDrawerSummaryTable(t, deps, types.NewUserPermissions([]string{"task_outcome:read"}))
	if got, want := len(table.Rows[0].Actions), 2; got != want {
		t.Fatalf("action count = %d, want %d", got, want)
	}
	action := table.Rows[0].Actions[1]
	if action.Type != "download" || action.Action != "outcome-matrix-download" {
		t.Fatalf("download action identity = %#v", action)
	}
	if action.HxGet != "/action/outcome-matrix/download/tmpl-a/subscription-group/grp-a" {
		t.Fatalf("HxGet = %q", action.HxGet)
	}
	if action.HxTarget != "#sheetContent" || action.HxSwap != "innerHTML" {
		t.Fatalf("HTMX contract = %#v", action)
	}
	if action.DrawerTitle != "Export outcomes" || action.Label != "Export outcomes" {
		t.Fatalf("drawer labels = %#v", action)
	}
	if action.TestID != "job-outcome-download-tmpl-a-grp-a" {
		t.Fatalf("TestID = %q", action.TestID)
	}
	if action.Disabled {
		t.Fatal("permitted download action is disabled")
	}
}

func TestTemplateSummaryDownloadAction_fallsBackAndFailsClosed(t *testing.T) {
	t.Parallel()

	t.Run("template route and common download label", func(t *testing.T) {
		deps := drawerSummaryDeps()
		deps.MatrixDownloadDrawerURL = "/action/outcome-matrix/download/{id}"
		deps.CommonLabels.Actions.Download = "Download"
		deps.CommonLabels.Errors.PermissionDenied = "Permission denied"

		action := buildDrawerSummaryTable(t, deps, types.NewUserPermissions([]string{"task_outcome:read"})).Rows[0].Actions[1]
		if action.HxGet != "/action/outcome-matrix/download/tmpl-a" {
			t.Fatalf("fallback HxGet = %q", action.HxGet)
		}
		if action.Label != "Download" || action.DrawerTitle != "Download" {
			t.Fatalf("fallback labels = %#v", action)
		}
		if action.TestID != "job-outcome-download-tmpl-a-grp-a" {
			t.Fatalf("fallback TestID = %q", action.TestID)
		}
	})

	t.Run("nil permissions and missing permission disable", func(t *testing.T) {
		deps := drawerSummaryDeps()
		deps.MatrixDownloadDrawerURL = "/action/outcome-matrix/download/{id}"
		deps.CommonLabels.Errors.PermissionDenied = "Permission denied"
		for _, perms := range []*types.UserPermissions{nil, types.NewUserPermissions([]string{"job:list"})} {
			action := buildDrawerSummaryTable(t, deps, perms).Rows[0].Actions[1]
			if !action.Disabled || action.DisabledTooltip != "Permission denied" {
				t.Fatalf("disabled action = %#v", action)
			}
		}
	})

	t.Run("no drawer route adds no action", func(t *testing.T) {
		deps := drawerSummaryDeps()
		if got := len(buildDrawerSummaryTable(t, deps, types.NewUserPermissions([]string{"task_outcome:read"})).Rows[0].Actions); got != 1 {
			t.Fatalf("action count = %d, want view action only", got)
		}
	})
}

func drawerSummaryDeps() *ListViewDeps {
	return &ListViewDeps{
		Labels:       job.DefaultLabels(),
		CommonLabels: pyeza.CommonLabels{},
		ListJobTemplateSummaries: func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
			return &summarypb.ListJobTemplateSummariesResponse{
				Summaries: []*summarypb.JobTemplateSummary{{
					JobTemplateId:         "tmpl-a",
					SubscriptionGroupId:   "grp-a",
					JobTemplateName:       "Math",
					SubscriptionGroupName: "Grade 9",
				}},
			}, nil
		},
	}
}

func buildDrawerSummaryTable(t *testing.T, deps *ListViewDeps, perms *types.UserPermissions) *types.TableConfig {
	t.Helper()
	table, _, err := buildDeliverySummaryTable(context.Background(), deps, "active", summaryParams(), "", false, perms)
	if err != nil {
		t.Fatalf("buildDeliverySummaryTable() err = %v", err)
	}
	return table
}
