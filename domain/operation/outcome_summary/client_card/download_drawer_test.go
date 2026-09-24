package client_card

import (
	"bytes"
	"context"
	"html/template"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func TestClientDownloadDrawerTemplate_RendersFetchDownloadFeedback(t *testing.T) {
	labels := outcome_summary.DefaultLabels().ClientDocumentDownload
	data := &DownloadDrawerData{
		FormURL: "/report-cards/section/group-1/student/client-1/document",
		Labels:  labels,
		CommonLabels: map[string]any{
			"Buttons": map[string]string{"Cancel": "Cancel"},
		},
	}
	tmpl, err := template.ParseFiles("../templates/client-card-download-drawer.html")
	if err != nil {
		t.Fatalf("parse client download drawer template: %v", err)
	}
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "outcome-summary-client-download-drawer-form", data); err != nil {
		t.Fatalf("render client download drawer template: %v", err)
	}

	for _, want := range []string{
		`data-lf-download-form`,
		`data-lf-download-error-fallback="` + labels.DownloadErrorFallback + `"`,
		`data-lf-busy-label="` + labels.DownloadingAction + `"`,
		`data-testid="rc-client-download-notice" hidden>`,
		`role="status" aria-live="polite"`,
		labels.DownloadNotice,
		`data-testid="rc-client-download-error" hidden></div>`,
		`role="alert"`,
	} {
		if !strings.Contains(rendered.String(), want) {
			t.Errorf("rendered client download drawer is missing %q:\n%s", want, rendered.String())
		}
	}
}

func downloadDrawerDeps(response *exportpb.GetSubscriptionGroupClientReportCardResponse, calls *int) *DrawerDeps {
	return &DrawerDeps{
		Routes: outcome_summary.Routes{
			ClientDownloadDrawerURL: "/action/report-cards/section/{id}/student/{client_id}/download",
			ClientDocumentURL:       "/report-cards/section/{id}/student/{client_id}/document",
		},
		Labels:               outcome_summary.DefaultLabels(),
		ResolvePrincipalKind: func(context.Context) int32 { return outcome_summary.PrincipalKindStaff },
		Options: outcome_summary.Options{SubscriptionGroupExport: outcome_summary.SubscriptionGroupExportOptions{
			Enabled:            true,
			WholeReportProfile: bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1,
		}},
		ClientAttributeCodes: []string{"gender"},
		GetSubscriptionGroupClientReportCard: func(_ context.Context, req *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
			*calls++
			if req.GetSubscriptionGroupId() != "group-1" || req.GetClientId() != "client-1" || len(req.GetClientAttributeCodes()) != 1 || req.GetClientAttributeCodes()[0] != "gender" {
				return nil, context.Canceled
			}
			return response, nil
		},
	}
}

func TestClientDownloadDrawer_UsesScopedProjectionAndOffersAvailablePeriods(t *testing.T) {
	calls := 0
	s1, s2, progress, inactive, unreferenced := "s1", "s2", "progress_report", "s3", "unused"
	phaseS1, phaseS2, phaseProgress := "template-phase-s1", "template-phase-s2", "template-phase-progress"
	phaseS1Duplicate, phaseInactive, phaseUnreferenced := "template-phase-s1-copy", "template-phase-inactive", "template-phase-unused"
	response := &exportpb.GetSubscriptionGroupClientReportCardResponse{
		Success: true,
		ReportCard: &exportpb.ClientReportCardProjection{
			Context:               &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: "group-1"},
			Client:                &exportpb.ClientReportCardClient{ClientId: "client-1"},
			ClientSubscriptionIds: []string{"subscription-1"},
			Jobs:                  []*jobpb.Job{{Id: "job-1"}, {Id: "job-2"}},
			JobTemplatePhases: []*jobtemplatephasepb.JobTemplatePhase{
				{Id: phaseProgress, Active: true, Code: &progress, Name: "Progress Report", PhaseOrder: 3},
				{Id: phaseS2, Active: true, Code: &s2, Name: "Term 2", PhaseOrder: 2},
				{Id: phaseS1Duplicate, Active: true, Code: &s1, Name: "Term 1 duplicate", PhaseOrder: 1},
				{Id: phaseInactive, Active: false, Code: &inactive, Name: "Term 3", PhaseOrder: 4},
				{Id: phaseUnreferenced, Active: true, Code: &unreferenced, Name: "Unused", PhaseOrder: 0},
				{Id: phaseS1, Active: true, Code: &s1, Name: "Term 1", PhaseOrder: 1},
			},
			JobPhases: []*jobphasepb.JobPhase{
				{JobId: "job-1", Active: true, TemplatePhaseId: &phaseS1},
				{JobId: "job-1", Active: true, TemplatePhaseId: &phaseS2},
				{JobId: "job-1", Active: true, TemplatePhaseId: &phaseProgress},
				{JobId: "job-2", Active: true, TemplatePhaseId: &phaseS1Duplicate},
				{JobId: "job-1", Active: false, TemplatePhaseId: &phaseInactive},
				{JobId: "foreign-job", Active: true, TemplatePhaseId: &phaseUnreferenced},
			},
			JobOutcomeSummaries: []*jobsumpb.JobOutcomeSummary{{Id: "summary-1", JobId: "job-1", Active: true}},
		},
	}
	request := httptest.NewRequest("GET", "/action/report-cards/section/group-1/student/client-1/download", nil)
	request.SetPathValue("id", "group-1")
	request.SetPathValue("client_id", "client-1")
	ctx := view.WithUserPermissions(context.Background(), types.NewUserPermissions([]string{"subscription_group_outcome_export:read"}))
	result := NewDownloadDrawer(downloadDrawerDeps(response, &calls)).Handle(ctx, &view.ViewContext{Request: request})
	if result.Error != nil || result.Template != "outcome-summary-client-download-drawer-form" {
		t.Fatalf("result = template %q status %d error %v", result.Template, result.StatusCode, result.Error)
	}
	if calls != 1 {
		t.Fatalf("typed scoped projection calls = %d, want 1", calls)
	}
	data := result.Data.(*DownloadDrawerData)
	if data.FormURL != "/report-cards/section/group-1/student/client-1/document" {
		t.Fatalf("form URL = %q", data.FormURL)
	}
	if len(data.Periods) != 4 {
		t.Fatalf("period options = %+v, want s1, s2, progress_report, year_final", data.Periods)
	}
	wantValues := []string{"s1", "s2", "progress_report", clientReportYearFinalPeriod}
	wantLabels := []string{"Term 1", "Term 2", "Progress Report", "Year Final"}
	for index, option := range data.Periods {
		if option.Value != wantValues[index] || option.Label != wantLabels[index] {
			t.Errorf("period option %d = %+v, want value=%q label=%q", index, option, wantValues[index], wantLabels[index])
		}
		if option.Selected != (index == 0) {
			t.Errorf("period option %d selected=%v, want %v", index, option.Selected, index == 0)
		}
	}
	if len(data.Formats) != 2 || data.Formats[0].Value != "pdf" || !data.Formats[0].Selected || data.Formats[1].Value != "docx" || data.Formats[1].Selected {
		t.Fatalf("format options = %+v", data.Formats)
	}
}

func TestDownloadDrawerData_FormURLUsesWorkspaceRewriteSuffix(t *testing.T) {
	field, ok := reflect.TypeOf(DownloadDrawerData{}).FieldByName("FormURL")
	if !ok || !strings.HasSuffix(field.Name, "URL") {
		t.Fatalf("download form field must end in URL for workspace rewriting")
	}
}

func TestClientReportPeriodOptions_SelectsYearFinalWhenNoActiveReferencedPhase(t *testing.T) {
	labels := outcome_summary.DefaultLabels()
	projection := &exportpb.ClientReportCardProjection{
		Jobs:                []*jobpb.Job{{Id: "job-1"}},
		JobTemplatePhases:   []*jobtemplatephasepb.JobTemplatePhase{{Id: "phase-1", Active: true, Code: stringPtr("s1"), Name: "Term 1"}},
		JobPhases:           []*jobphasepb.JobPhase{{JobId: "job-1", Active: false, TemplatePhaseId: stringPtr("phase-1")}},
		JobOutcomeSummaries: []*jobsumpb.JobOutcomeSummary{{Id: "summary-1", JobId: "job-1", Active: true}},
	}
	options := clientReportPeriodOptions(labels, projection)
	if len(options) != 1 || options[0].Value != clientReportYearFinalPeriod || !options[0].Selected {
		t.Fatalf("period options = %+v, want selected Year Final only", options)
	}
}

func TestClientReportPeriodOptions_OffersYearFinalWithoutPublishedSummary(t *testing.T) {
	projection := &exportpb.ClientReportCardProjection{Jobs: []*jobpb.Job{{Id: "job-1"}}}
	options := clientReportPeriodOptions(outcome_summary.DefaultLabels(), projection)
	if len(options) != 1 || options[0].Value != clientReportYearFinalPeriod || !options[0].Selected {
		t.Fatalf("period options = %+v, want selected Year Final with projected job and no summary", options)
	}
}

func stringPtr(value string) *string { return &value }

func TestClientDownloadDrawer_DeniesWithoutExplicitPermissionBeforeQuery(t *testing.T) {
	calls := 0
	request := httptest.NewRequest("GET", "/action/report-cards/section/group-1/student/client-1/download", nil)
	request.SetPathValue("id", "group-1")
	request.SetPathValue("client_id", "client-1")
	ctx := view.WithUserPermissions(context.Background(), types.NewUserPermissions([]string{"job_outcome_summary:read"}))
	result := NewDownloadDrawer(downloadDrawerDeps(nil, &calls)).Handle(ctx, &view.ViewContext{Request: request})
	if result.StatusCode != 403 {
		t.Fatalf("status = %d, want 403", result.StatusCode)
	}
	if calls != 0 {
		t.Fatalf("projection query calls = %d, want zero before explicit permission", calls)
	}
}

func TestClientDownloadDrawer_RejectsMismatchedProjectionContext(t *testing.T) {
	calls := 0
	request := httptest.NewRequest("GET", "/action/report-cards/section/group-1/student/client-1/download", nil)
	request.SetPathValue("id", "group-1")
	request.SetPathValue("client_id", "client-1")
	response := &exportpb.GetSubscriptionGroupClientReportCardResponse{Success: true, ReportCard: &exportpb.ClientReportCardProjection{
		Context: &exportpb.SubscriptionGroupOutcomeExportContext{SubscriptionGroupId: "other-group"},
		Client:  &exportpb.ClientReportCardClient{ClientId: "client-1"},
	}}
	ctx := view.WithUserPermissions(context.Background(), types.NewUserPermissions([]string{"subscription_group_outcome_export:read"}))
	result := NewDownloadDrawer(downloadDrawerDeps(response, &calls)).Handle(ctx, &view.ViewContext{Request: request})
	if result.StatusCode != 404 {
		t.Fatalf("status = %d, want 404 for mismatched group", result.StatusCode)
	}
}
