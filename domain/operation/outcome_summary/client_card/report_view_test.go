package client_card

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	espynaports "github.com/erniealice/espyna-golang/ports"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func TestClientProjectionGroupByPreservesCompleteOutcomeSet(t *testing.T) {
	projection := fourCategoryProjection()
	deps := clientProjectionDeps(projection)
	grouped := clientProjectionPageData(t, deps, "")
	if grouped.Table == nil || len(grouped.Table.Rows) != 0 || len(grouped.Table.Groups) != 4 {
		t.Fatalf("configured default rows/groups = %d/%d, want 0/4", len(grouped.Table.Rows), len(grouped.Table.Groups))
	}
	if got, want := groupTitles(grouped.Table.Groups), []string{"Category A", "Category B", "Category C", "Category D"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("category group order = %v, want %v", got, want)
	}
	flatDeps := clientProjectionDeps(projection)
	flatDeps.Options.ClientCard.Row.GroupByField = ""
	flat := clientProjectionPageData(t, flatDeps, "group_by=job_category")
	if flat.Table == nil || len(flat.Table.Rows) != 4 || len(flat.Table.Groups) != 0 {
		t.Fatalf("flat mode table rows/groups = %d/%d, want 4/0", len(flat.Table.Rows), len(flat.Table.Groups))
	}
	if len(flat.Table.Columns) != 5 {
		t.Fatalf("flat mode columns = %d, want subject + three template phases + Year Final", len(flat.Table.Columns))
	}
	if flat.Table.Rows[0].Cells[1].Value == deps.Labels.SubscriptionGroup.RatingEmpty || flat.Table.Rows[0].Cells[4].Value != "YF category_a" {
		t.Fatalf("projected outcomes were not rendered: %+v", flat.Table.Rows[0].Cells)
	}

	if got, want := normalizeRows(grouped.Table.Groups), normalizeTableRows(flat.Table.Rows); !reflect.DeepEqual(got, want) {
		t.Fatalf("grouping changed row identity or values\ngrouped=%+v\nflat=%+v", got, want)
	}
	if grouped.Table.ToolbarPrefixTemplate != "" || grouped.Table.ShowSearch {
		t.Fatalf("unexpected table controls: toolbar=%q search=%t", grouped.Table.ToolbarPrefixTemplate, grouped.Table.ShowSearch)
	}
	if grouped.Table.PrimaryAction == nil || grouped.Table.PrimaryAction.ActionURL != "/download/sg-1/client-a" {
		t.Fatalf("client download action = %+v", grouped.Table.PrimaryAction)
	}
}

func TestClientProjectionQueryCannotOverrideConfiguredGrouping(t *testing.T) {
	deps := clientProjectionDeps(fourCategoryProjection())
	reads := 0
	deps.GetSubscriptionGroupClientReportCard = func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
		reads++
		return &exportpb.GetSubscriptionGroupClientReportCardResponse{Success: true, ReportCard: fourCategoryProjection()}, nil
	}
	result := NewView(deps).Handle(clientProjectionContext(t, deps.Routes, "sg-1", "client-a", "group_by=student"), clientProjectionViewContext(t, deps.Routes, "sg-1", "client-a", "group_by=student"))
	if result.StatusCode != 200 {
		t.Fatalf("status = %d, want configured view", result.StatusCode)
	}
	if reads != 1 {
		t.Fatalf("projection reads = %d, want one read", reads)
	}
}

func TestClientProjectionUsesOneExactGroupClientRead(t *testing.T) {
	deps := clientProjectionDeps(fourCategoryProjection())
	reads := 0
	deps.GetSubscriptionGroupClientReportCard = func(_ context.Context, request *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
		reads++
		if request.GetSubscriptionGroupId() != "sg-1" || request.GetClientId() != "client-a" {
			t.Fatalf("projection request scope = %q/%q", request.GetSubscriptionGroupId(), request.GetClientId())
		}
		return &exportpb.GetSubscriptionGroupClientReportCardResponse{Success: true, ReportCard: fourCategoryProjection()}, nil
	}
	result := NewView(deps).Handle(clientProjectionContext(t, deps.Routes, "sg-1", "client-a", ""), clientProjectionViewContext(t, deps.Routes, "sg-1", "client-a", ""))
	if result.StatusCode != 200 || reads != 1 {
		t.Fatalf("result status/reads = %d/%d, want 200/1", result.StatusCode, reads)
	}
}

func TestClientProjectionMapsNotFoundAndDependencyFailuresWithoutUsingForbidden(t *testing.T) {
	foreign := fourCategoryProjection()
	foreign.Context.SubscriptionGroupId = "another-group"
	cases := []struct {
		name     string
		response *exportpb.GetSubscriptionGroupClientReportCardResponse
		err      error
		want     int
	}{
		{name: "typed not found", err: espynaports.ErrClientReportNotFound, want: http.StatusNotFound},
		{name: "projection dependency failure", err: errors.New("query unavailable"), want: http.StatusServiceUnavailable},
		{name: "nil response", want: http.StatusServiceUnavailable},
		{name: "unsuccessful response", response: &exportpb.GetSubscriptionGroupClientReportCardResponse{}, want: http.StatusServiceUnavailable},
		{name: "successful empty projection", response: &exportpb.GetSubscriptionGroupClientReportCardResponse{Success: true}, want: http.StatusNotFound},
		{name: "foreign projection", response: &exportpb.GetSubscriptionGroupClientReportCardResponse{Success: true, ReportCard: foreign}, want: http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := clientProjectionDeps(fourCategoryProjection())
			deps.GetSubscriptionGroupClientReportCard = func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
				return tc.response, tc.err
			}
			vc := clientProjectionViewContext(t, deps.Routes, "sg-1", "client-a", "")
			result := NewView(deps).Handle(clientProjectionContext(t, deps.Routes, "sg-1", "client-a", ""), vc)
			if result.StatusCode != tc.want {
				t.Fatalf("status = %d, want %d (error %v)", result.StatusCode, tc.want, result.Error)
			}
		})
	}
}

func clientProjectionDeps(projection *exportpb.ClientReportCardProjection) *Deps {
	deps := &Deps{Routes: outcome_summary.DefaultRoutes(), Labels: outcome_summary.DefaultLabels()}
	deps.Routes.ClientDownloadDrawerURL = "/download/{id}/{client_id}"
	deps.Routes.ClientDocumentURL = "/document/{id}/{client_id}"
	deps.ClientDocumentMounted = true
	deps.Options.List.Entity = outcome_summary.ListEntitySubscriptionGroup
	deps.Options.SubscriptionGroupExport.Enabled = true
	deps.Options.ClientCard.Row.GroupByField = outcome_summary.ListColumnsJobCategory
	deps.Options.ClientCard.IncludeAllCategories = true
	deps.ResolvePrincipalKind = func(context.Context) int32 { return outcome_summary.PrincipalKindStaff }
	deps.GetSubscriptionGroupClientReportCard = func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
		return &exportpb.GetSubscriptionGroupClientReportCardResponse{Success: true, ReportCard: projection}, nil
	}
	return deps
}

func clientProjectionPageData(t *testing.T, deps *Deps, query string) *PageData {
	t.Helper()
	return clientPageData(t, NewView(deps).Handle(
		clientProjectionContext(t, deps.Routes, "sg-1", "client-a", query),
		clientProjectionViewContext(t, deps.Routes, "sg-1", "client-a", query),
	))
}

func clientProjectionViewContext(t *testing.T, routes outcome_summary.Routes, groupID, clientID, query string) *view.ViewContext {
	t.Helper()
	requestURL := route.ResolveURL(routes.ClientCardURL, "id", groupID, "client_id", clientID)
	if query != "" {
		requestURL += "?" + query
	}
	request := httptest.NewRequest("GET", requestURL, nil)
	request.SetPathValue("id", groupID)
	request.SetPathValue("client_id", clientID)
	return &view.ViewContext{Request: request, CurrentPath: request.URL.Path, CacheVersion: "test"}
}

func clientProjectionContext(t *testing.T, routes outcome_summary.Routes, groupID, clientID, query string) context.Context {
	t.Helper()
	vc := clientProjectionViewContext(t, routes, groupID, clientID, query)
	return view.WithUserPermissions(vc.Request.Context(), types.NewUserPermissions([]string{"subscription_group_outcome_export:read"}))
}

func fourCategoryProjection() *exportpb.ClientReportCardProjection {
	categories := []*jobcategorypb.JobCategory{
		{Id: "cat-c", Name: "Category C", Code: strPtr("category_c"), SortOrder: int32Ptr(30)},
		{Id: "cat-d", Name: "Category D", Code: strPtr("category_d"), SortOrder: int32Ptr(40)},
		{Id: "cat-b", Name: "Category B", Code: strPtr("category_b"), SortOrder: int32Ptr(20)},
		{Id: "cat-a", Name: "Category A", Code: strPtr("category_a"), SortOrder: int32Ptr(10)},
	}
	jobs := make([]*jobpb.Job, 0, 4)
	templates := make([]*jobtemplatepb.JobTemplate, 0, 4)
	phases := make([]*jobphasepb.JobPhase, 0, 12)
	phaseSummaries := make([]*phasesumpb.PhaseOutcomeSummary, 0, 12)
	yearSummaries := make([]*jobsumpb.JobOutcomeSummary, 0, 4)
	for _, category := range categories {
		jobID := "job-" + category.GetCode()
		templateID := "template-" + category.GetCode()
		jobs = append(jobs, &jobpb.Job{Id: jobID, JobTemplateId: strPtr(templateID)})
		templates = append(templates, &jobtemplatepb.JobTemplate{Id: templateID, Name: category.GetName(), JobCategoryId: strPtr(category.GetId())})
		for phaseIndex, code := range []string{"progress_report", "term_1", "term_2"} {
			templatePhaseID := "tp-" + category.GetCode() + "-" + code
			phaseID := "phase-" + category.GetCode() + "-" + code
			phases = append(phases, &jobphasepb.JobPhase{Id: phaseID, JobId: jobID, Name: code, PhaseOrder: int32(phaseIndex + 1), TemplatePhaseId: strPtr(templatePhaseID)})
			phaseSummaries = append(phaseSummaries, &phasesumpb.PhaseOutcomeSummary{Id: "summary-" + phaseID, JobPhaseId: phaseID, ScaledLabel: strPtr(code + " " + category.GetCode()), Active: true})
		}
		yearSummaries = append(yearSummaries, &jobsumpb.JobOutcomeSummary{Id: "year-" + jobID, JobId: jobID, ScaledLabel: strPtr("YF " + category.GetCode()), Active: true})
	}
	templatePhases := make([]*jobtemplatephasepb.JobTemplatePhase, 0, len(phases))
	for _, category := range categories {
		for phaseIndex, code := range []string{"progress_report", "term_1", "term_2"} {
			templatePhases = append(templatePhases, &jobtemplatephasepb.JobTemplatePhase{Id: "tp-" + category.GetCode() + "-" + code, Name: code, Code: strPtr(code), PhaseOrder: int32(phaseIndex + 1), Active: true})
		}
	}
	return &exportpb.ClientReportCardProjection{
		Context:               &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: "sg-1", SubscriptionGroupName: "Group A"},
		Client:                &exportpb.ClientReportCardClient{ClientId: "client-a", Name: "Learner A"},
		ClientSubscriptionIds: []string{"subscription-a"},
		Jobs:                  jobs, JobTemplates: templates, JobCategories: categories,
		JobPhases: phases, JobTemplatePhases: templatePhases,
		PhaseOutcomeSummaries: phaseSummaries, JobOutcomeSummaries: yearSummaries,
	}
}

func normalizeRows(groups []types.TableRowGroup) map[string][]types.TableCell {
	out := make(map[string][]types.TableCell)
	for _, group := range groups {
		for _, row := range group.Rows {
			out[row.ID] = row.Cells
		}
	}
	return out
}

func normalizeTableRows(rows []types.TableRow) map[string][]types.TableCell {
	out := make(map[string][]types.TableCell)
	for _, row := range rows {
		out[row.ID] = row.Cells
	}
	return out
}

func groupTitles(groups []types.TableRowGroup) []string {
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		out = append(out, group.Title)
	}
	return out
}

func clientPageData(t *testing.T, result view.ViewResult) *PageData {
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

func strPtr(value string) *string { return &value }
func int32Ptr(value int32) *int32 { return &value }
