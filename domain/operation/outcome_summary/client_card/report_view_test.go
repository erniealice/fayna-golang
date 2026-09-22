package client_card

import (
	"context"
	"errors"
	"html/template"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func TestClientReportViewUsesOnlyAggregate(t *testing.T) {
	deps := clientFatalDeps(t)
	category := clientCategory("cat-a", true, clientPhase("P1", "Period A", 1, false))
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return clientOptions("sg-1", category), nil
		}
		return clientMatrix("sg-1", category, []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Alpha"}}, clientRow("client-a", "Ava", "Adams", clientCell("tmpl-a", "A", nil, true, true))), nil
	}

	result := NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", ""), clientViewContext(t, deps.Routes, "sg-1", "client-a", ""))
	if result.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", result.StatusCode)
	}
	page := clientPageData(t, result)
	if page.Table == nil || len(page.Table.Rows) != 1 {
		t.Fatalf("table = %+v, want one subject row", page.Table)
	}
}

func TestClientReportViewForeignClientForbidden(t *testing.T) {
	deps := clientReportDeps()
	category := clientCategory("cat-a", true)
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return clientOptions("sg-1", category), nil
		}
		return clientMatrix("sg-1", category, []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Alpha"}}, nil), nil
	}

	result := NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-foreign", ""), clientViewContext(t, deps.Routes, "sg-1", "client-foreign", ""))
	if result.StatusCode != 403 || result.Template != "forbidden" {
		t.Fatalf("result = template %q/status %d, want forbidden", result.Template, result.StatusCode)
	}
}

func TestClientReportViewMatrixErrorIsForbidden(t *testing.T) {
	deps := clientReportDeps()
	category := clientCategory("cat-a", true)
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return clientOptions("sg-1", category), nil
		}
		return nil, errors.New("aggregate read failed")
	}

	result := NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a"), clientViewContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a"))
	if result.StatusCode != 403 || result.Template != "forbidden" {
		t.Fatalf("result = template %q/status %d, want forbidden", result.Template, result.StatusCode)
	}
	if data, ok := result.Data.(*view.ForbiddenPageData); !ok || data.Perm != "subscription_group_outcome_export:read" {
		t.Fatalf("forbidden data = %#v, want export permission", result.Data)
	}
}

func TestClientReportViewCallCap(t *testing.T) {
	deps := clientReportDeps()
	categories := make([]*exportpb.JobCategoryOption, 0, maxNarrowMatrixCalls+4)
	for index := 0; index < maxNarrowMatrixCalls+4; index++ {
		categories = append(categories, clientCategory("cat-"+string(rune('a'+index)), true))
	}
	matrixCalls := 0
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return clientOptions("sg-1", categories...), nil
		}
		matrixCalls++
		return clientMatrix("sg-1", clientCategory(req.GetJobCategoryId(), true), []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Alpha"}}, nil), nil
	}

	result := NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", ""), clientViewContext(t, deps.Routes, "sg-1", "client-a", ""))
	if result.StatusCode != 403 {
		t.Fatalf("status = %d, want 403 after bounded reads", result.StatusCode)
	}
	if matrixCalls != maxNarrowMatrixCalls {
		t.Fatalf("matrix calls = %d, want cap %d", matrixCalls, maxNarrowMatrixCalls)
	}
}

func TestClientReportViewCallCapTruncatesPhasesNotBlanks(t *testing.T) {
	deps := clientReportDeps()
	phases := make([]*exportpb.JobTemplatePhaseOption, 0, maxNarrowMatrixCalls+1)
	for index := 0; index < maxNarrowMatrixCalls+1; index++ {
		code := "phase-" + string(rune('a'+index))
		phases = append(phases, clientPhase(code, code, int32(index+1), false))
	}
	category := clientCategory("cat-a", true, phases...)
	matrixCalls, phaseCalls := 0, 0
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return clientOptions("sg-1", category), nil
		}
		matrixCalls++
		if code := req.GetJobTemplatePhaseCode(); code != "" {
			phaseCalls++
			return clientMatrix("sg-1", category, clientColumns(), clientRow("client-a", "Ava", "Adams", clientCell("tmpl-a", code, nil, true, true), clientCell("tmpl-b", code, nil, true, true))), nil
		}
		return clientMatrix("sg-1", category, clientColumns(), clientRow("client-a", "Ava", "Adams", clientCell("tmpl-a", "Y-A", nil, true, true), clientCell("tmpl-b", "Y-B", nil, true, true))), nil
	}

	page := clientPageData(t, NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a"), clientViewContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a")))
	if matrixCalls > maxNarrowMatrixCalls {
		t.Fatalf("matrix calls = %d, want at most %d", matrixCalls, maxNarrowMatrixCalls)
	}
	if phaseCalls != maxNarrowMatrixCalls-1 {
		t.Fatalf("phase calls = %d, want %d after final matrix", phaseCalls, maxNarrowMatrixCalls-1)
	}
	wantColumns := 1 + phaseCalls + 1
	if len(page.Table.Columns) != wantColumns {
		t.Fatalf("columns = %d, want %d (subject, fetched phases, final)", len(page.Table.Columns), wantColumns)
	}
	for index := 1; index <= phaseCalls; index++ {
		if page.Table.Rows[0].Cells[index].Value == deps.Labels.SubscriptionGroup.RatingEmpty {
			t.Fatalf("fetched phase cell %d is blank: %+v", index, page.Table.Rows[0].Cells)
		}
	}
	if got := page.Table.Columns[len(page.Table.Columns)-1].Key; got != "year-final" {
		t.Fatalf("last column = %q, want year-final", got)
	}
}

func TestClientReportViewPhaseOrderAndFinalColumn(t *testing.T) {
	deps := clientReportDeps()
	category := clientCategory("cat-a", true,
		clientPhase("P2", "Period B", 2, false),
		clientPhase("P-ambiguous", "Ignored", 3, true),
		clientPhase("P1", "Period A", 1, false),
	)
	var selectors []string
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return clientOptions("sg-1", category), nil
		}
		if req.GetFinalOutcome() {
			selectors = append(selectors, "final")
			return clientMatrix("sg-1", category, clientColumns(), clientRow("client-a", "Ava", "Adams", clientCell("tmpl-a", "Y-A", nil, true, true), clientCell("tmpl-b", "Y-B", nil, true, true))), nil
		}
		code := req.GetJobTemplatePhaseCode()
		selectors = append(selectors, code)
		label := "P1-A"
		if code == "P2" {
			label = "P2-A"
		}
		return clientMatrix("sg-1", category, clientColumns(), clientRow("client-a", "Ava", "Adams", clientCell("tmpl-a", label, nil, true, true), clientCell("tmpl-b", label, nil, true, true))), nil
	}

	page := clientPageData(t, NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a"), clientViewContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a")))
	if got, want := selectors, []string{"final", "P1", "P2"}; !equalStrings(got, want) {
		t.Fatalf("selectors = %v, want %v", got, want)
	}
	keys := make([]string, 0, len(page.Table.Columns))
	for _, column := range page.Table.Columns {
		keys = append(keys, column.Key)
	}
	if got, want := keys, []string{"subject", "phase-P1", "phase-P2", "year-final"}; !equalStrings(got, want) {
		t.Fatalf("column keys = %v, want %v", got, want)
	}
	if page.Table.Rows[0].Cells[1].Value != "P1-A" || page.Table.Rows[0].Cells[2].Value != "P2-A" || page.Table.Rows[0].Cells[3].Value != "Y-A" {
		t.Fatalf("row cells = %+v, want phase order followed by final", page.Table.Rows[0].Cells)
	}
	if page.Table.Columns[3].Label != deps.Labels.Client.YearColumn+" "+deps.Labels.Client.FinalColumn {
		t.Fatalf("final column label = %q, want year/final labels", page.Table.Columns[3].Label)
	}
}

func TestClientReportViewMissingFinalStillRendersPhases(t *testing.T) {
	deps := clientReportDeps()
	category := clientCategory("cat-a", false, clientPhase("P1", "Period A", 1, false))
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return clientOptions("sg-1", category), nil
		}
		if req.GetFinalOutcome() {
			t.Fatal("unavailable final matrix was requested")
		}
		return clientMatrix("sg-1", category, clientColumns(), clientRow("client-a", "Ava", "Adams", clientCell("tmpl-a", "P1-A", nil, true, true), clientCell("tmpl-b", "P1-B", nil, true, true))), nil
	}

	page := clientPageData(t, NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a"), clientViewContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a")))
	if page.Table == nil || page.NotComputed {
		t.Fatalf("page = %+v, want phase-only table", page)
	}
	if page.Table.Rows[0].Cells[1].Value != "P1-A" {
		t.Fatalf("phase value = %q, want P1-A", page.Table.Rows[0].Cells[1].Value)
	}
	if page.Table.Rows[0].Cells[2].Value != deps.Labels.SubscriptionGroup.RatingEmpty {
		t.Fatalf("final value = %q, want rating-empty label %q", page.Table.Rows[0].Cells[2].Value, deps.Labels.SubscriptionGroup.RatingEmpty)
	}
}

func TestClientReportViewSkipsFinalWhenUnavailable(t *testing.T) {
	deps := clientReportDeps()
	category := clientCategory("cat-a", false, clientPhase("P1", "Period A", 1, false))
	selectors := []string{}
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return clientOptions("sg-1", category), nil
		}
		if req.GetFinalOutcome() {
			t.Fatal("final matrix was requested while unavailable")
		}
		selectors = append(selectors, req.GetJobTemplatePhaseCode())
		return clientMatrix("sg-1", category, clientColumns(), clientRow("client-a", "Ava", "Adams", clientCell("tmpl-a", "P1-A", nil, true, true), clientCell("tmpl-b", "P1-B", nil, true, true))), nil
	}

	page := clientPageData(t, NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a"), clientViewContext(t, deps.Routes, "sg-1", "client-a", "jc=cat-a")))
	if got, want := selectors, []string{"P1"}; !equalStrings(got, want) {
		t.Fatalf("selectors = %v, want only the available phase", got)
	}
	if page.Table.Rows[0].Cells[1].Value != "P1-A" || page.Table.Rows[0].Cells[2].Value != deps.Labels.SubscriptionGroup.RatingEmpty {
		t.Fatalf("row cells = %+v, want phase value and blank final", page.Table.Rows[0].Cells)
	}
}

func TestClientLegacyPathUnchangedForReadHolder(t *testing.T) {
	deps := clientReportDeps()
	aggregateCalled := false
	deps.GetSubscriptionGroupOutcomeExport = func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		aggregateCalled = true
		return nil, nil
	}
	deps.ListSubscriptionGroups = func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		return &subscriptiongrouppb.ListSubscriptionGroupsResponse{Data: []*subscriptiongrouppb.SubscriptionGroup{{Id: "sg-1", Name: "Group", Active: true}}}, nil
	}
	deps.ListSubscriptionGroupMembers = func(context.Context, *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
		return &subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse{Data: []*subscriptiongroupmemberpb.SubscriptionGroupMember{{SubscriptionId: "sub-1", ClientId: "client-a", Active: true}}}, nil
	}
	deps.ListJobs = func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
		return &jobpb.ListJobsResponse{}, nil
	}
	ctx := clientContextWithPermissions(t, deps.Routes, "sg-1", "client-a", "", []string{"job_outcome_summary:list", "job_outcome_summary:read"})
	result := NewView(deps).Handle(ctx, clientViewContext(t, deps.Routes, "sg-1", "client-a", ""))
	if result.StatusCode != 200 || result.Template != "outcome-summary-client" {
		t.Fatalf("legacy result = template %q/status %d, want legacy success", result.Template, result.StatusCode)
	}
	if aggregateCalled {
		t.Fatal("narrow aggregate was called for a legacy read holder")
	}
}

func TestClientReportViewUsesRouteAndLabels(t *testing.T) {
	cases := []struct {
		name           string
		groupRoute     string
		clientRoute    string
		clientLabel    string
		groupLabel     string
		groupName      string
		wantGroupHref  string
		wantClientHref string
	}{
		{name: "generic", groupRoute: "/generic/group/{id}", clientRoute: "/generic/group/{id}/client/{client_id}", clientLabel: "Client A", groupLabel: "Group A", groupName: "Generic Group", wantGroupHref: "/generic/group/sg-1", wantClientHref: "/generic/group/sg-1/client/client-a"},
		{name: "education", groupRoute: "/tier/group/{id}", clientRoute: "/tier/group/{id}/client/{client_id}", clientLabel: "Client B", groupLabel: "Group B", groupName: "Tier Group", wantGroupHref: "/tier/group/sg-1", wantClientHref: "/tier/group/sg-1/client/client-a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := clientReportDeps()
			deps.Routes.SubscriptionGroupURL = tc.groupRoute
			deps.Routes.ClientCardURL = tc.clientRoute
			deps.Labels.SubscriptionGroup.Title = tc.groupLabel
			deps.Labels.Client.SubjectColumn = tc.clientLabel
			category := clientCategory("cat-a", true)
			deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
				if req.GetOutcomeSelector() == nil {
					response := clientOptions("sg-1", category)
					response.Context.SubscriptionGroupName = tc.groupName
					return response, nil
				}
				return clientMatrix("sg-1", category, clientColumns(), clientRow("client-a", "Ava", "Adams", clientCell("tmpl-a", "A", nil, true, true), clientCell("tmpl-b", "B", nil, true, true))), nil
			}
			page := clientPageData(t, NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", ""), clientViewContext(t, deps.Routes, "sg-1", "client-a", "")))
			if page.Table.Columns[0].Label != tc.clientLabel || page.PageData.HeaderBreadcrumb != tc.groupLabel {
				t.Fatalf("labels did not flow: first column=%q breadcrumb=%q", page.Table.Columns[0].Label, page.PageData.HeaderBreadcrumb)
			}
			if page.PageData.HeaderBreadcrumbURL != tc.wantGroupHref {
				t.Fatalf("group href = %q, want %q", page.PageData.HeaderBreadcrumbURL, tc.wantGroupHref)
			}
			if page.PageData.HeaderTitle != "Adams, Ava — "+tc.groupName {
				t.Fatalf("header title = %q, want data-composed title", page.PageData.HeaderTitle)
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
		{name: "aggregate error", response: clientOptions("sg-1", clientCategory("cat-a", true)), err: errors.New("aggregate read failed")},
	}
	var wantTemplate, wantPerm string
	var wantStatus int
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := clientReportDeps()
			deps.GetSubscriptionGroupOutcomeExport = func(context.Context, *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
				return tc.response, tc.err
			}
			result := NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", ""), clientViewContext(t, deps.Routes, "sg-1", "client-a", ""))
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
	category := clientCategory(malicious, true)
	deps := clientReportDeps()
	deps.GetSubscriptionGroupOutcomeExport = func(_ context.Context, req *exportpb.GetSubscriptionGroupOutcomeExportRequest) (*exportpb.GetSubscriptionGroupOutcomeExportResponse, error) {
		if req.GetOutcomeSelector() == nil {
			return clientOptions("sg-1", category), nil
		}
		if req.GetJobCategoryId() != malicious {
			t.Fatalf("category id = %q, want query value", req.GetJobCategoryId())
		}
		return clientMatrix("sg-1", category, clientColumns(), &exportpb.SubscriptionGroupOutcomeClientRow{
			ClientId:   "client-a",
			ClientName: malicious,
			Cells:      []*exportpb.SubscriptionGroupOutcomeCell{clientCell("tmpl-a", "A", nil, true, true), clientCell("tmpl-b", "B", nil, true, true)},
		}), nil
	}

	query := url.Values{"jc": []string{malicious}}.Encode()
	page := clientPageData(t, NewView(deps).Handle(clientContext(t, deps.Routes, "sg-1", "client-a", query), clientViewContext(t, deps.Routes, "sg-1", "client-a", query)))
	if got := assertClientReportPlainString(t, "header title", page.PageData.HeaderTitle); got != malicious+" — Group" {
		t.Fatalf("header title = %q, want plain client name", got)
	}
	if strings.Contains(page.PageData.HeaderBreadcrumbURL, malicious) {
		t.Fatalf("raw client value leaked into breadcrumb href: %q", page.PageData.HeaderBreadcrumbURL)
	}
	for _, row := range page.Table.Rows {
		for _, cell := range row.Cells {
			if strings.Contains(cell.Href, malicious) {
				t.Fatalf("raw client/category value leaked into cell href: %q", cell.Href)
			}
		}
	}
}

func assertClientReportPlainString(t *testing.T, label string, value any) string {
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

func clientReportDeps() *Deps {
	deps := &Deps{Routes: outcome_summary.DefaultRoutes(), Labels: outcome_summary.DefaultLabels()}
	deps.Options.List.Entity = outcome_summary.ListEntitySubscriptionGroup
	deps.Options.SubscriptionGroupExport.Enabled = true
	deps.ResolvePrincipalKind = func(context.Context) int32 { return outcome_summary.PrincipalKindStaff }
	return deps
}

func clientFatalDeps(t *testing.T) *Deps {
	t.Helper()
	deps := clientReportDeps()
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
	deps.ListPhaseOutcomeSummarysByJob = func(context.Context, *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error) {
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
	return deps
}

func clientViewContext(t *testing.T, routes outcome_summary.Routes, subscriptionGroupID, clientID, query string) *view.ViewContext {
	t.Helper()
	requestURL := route.ResolveURL(routes.ClientCardURL, "id", subscriptionGroupID, "client_id", clientID)
	if query != "" {
		requestURL += "?" + query
	}
	request := httptest.NewRequest("GET", requestURL, nil)
	request.SetPathValue("id", subscriptionGroupID)
	request.SetPathValue("client_id", clientID)
	return &view.ViewContext{Request: request, CurrentPath: request.URL.Path, CacheVersion: "test"}
}

func clientContext(t *testing.T, routes outcome_summary.Routes, subscriptionGroupID, clientID, query string) context.Context {
	return clientContextWithPermissions(t, routes, subscriptionGroupID, clientID, query, []string{
		"job_outcome_summary:list",
		"subscription_group_outcome_export:read",
	})
}

func clientContextWithPermissions(t *testing.T, routes outcome_summary.Routes, subscriptionGroupID, clientID, query string, permissions []string) context.Context {
	t.Helper()
	vc := clientViewContext(t, routes, subscriptionGroupID, clientID, query)
	return view.WithUserPermissions(vc.Request.Context(), types.NewUserPermissions(permissions))
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

func clientCategory(id string, final bool, phases ...*exportpb.JobTemplatePhaseOption) *exportpb.JobCategoryOption {
	return &exportpb.JobCategoryOption{JobCategoryId: id, Name: id, FinalOutcomeAvailable: final, JobTemplatePhases: phases}
}

func clientPhase(code, name string, order int32, ambiguous bool) *exportpb.JobTemplatePhaseOption {
	return &exportpb.JobTemplatePhaseOption{Code: code, Name: name, SequenceOrder: order, Ambiguous: ambiguous}
}

func clientColumns() []*exportpb.JobTemplateColumn {
	return []*exportpb.JobTemplateColumn{{JobTemplateId: "tmpl-a", DisplayName: "Alpha"}, {JobTemplateId: "tmpl-b", DisplayName: "Beta"}}
}

func clientOptions(subscriptionGroupID string, categories ...*exportpb.JobCategoryOption) *exportpb.GetSubscriptionGroupOutcomeExportResponse {
	return &exportpb.GetSubscriptionGroupOutcomeExportResponse{
		Context:       &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: subscriptionGroupID, SubscriptionGroupName: "Group"},
		JobCategories: categories,
		Success:       true,
	}
}

func clientMatrix(subscriptionGroupID string, category *exportpb.JobCategoryOption, columns []*exportpb.JobTemplateColumn, rows ...*exportpb.SubscriptionGroupOutcomeClientRow) *exportpb.GetSubscriptionGroupOutcomeExportResponse {
	return &exportpb.GetSubscriptionGroupOutcomeExportResponse{
		Context:            &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: subscriptionGroupID, SubscriptionGroupName: "Group"},
		JobCategories:      []*exportpb.JobCategoryOption{category},
		JobTemplateColumns: columns,
		ClientRows:         rows,
		Success:            true,
	}
}

func clientRow(clientID, firstName, lastName string, cells ...*exportpb.SubscriptionGroupOutcomeCell) *exportpb.SubscriptionGroupOutcomeClientRow {
	return &exportpb.SubscriptionGroupOutcomeClientRow{ClientId: clientID, ClientFirstName: firstName, ClientLastName: lastName, Cells: cells}
}

func clientCell(templateID, label string, score *float64, hasMarks, hasPositiveMark bool) *exportpb.SubscriptionGroupOutcomeCell {
	cell := &exportpb.SubscriptionGroupOutcomeCell{
		JobTemplateId:      templateID,
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasMarks: hasMarks, HasPositiveMark: hasPositiveMark},
	}
	if label != "" {
		cell.ScaledLabel = &label
	}
	if score != nil {
		cell.ScaledScore = score
	}
	return cell
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
