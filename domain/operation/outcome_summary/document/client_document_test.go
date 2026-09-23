package document

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	espynaports "github.com/erniealice/espyna-golang/ports"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
	"google.golang.org/protobuf/proto"

	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	phaseoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func TestDownload_ExplicitProgressReport_UsesExportPermissionAndPhaseBinding(t *testing.T) {
	for _, format := range []string{"docx", "pdf"} {
		t.Run(format, func(t *testing.T) {
			card := clientPhaseProjectionFixture()
			card.Context.PlanId = strptr("plan-1")
			card.Context.PriceScheduleId = strptr("schedule-1")
			card.ClientSubscriptionIds = []string{"subscription-1"}
			allowClientProjectionRender(card, "group-1")
			response := &exportpb.GetSubscriptionGroupClientReportCardResponse{Success: true, ReportCard: card}
			projectionCalls, resolverCalls, renderCalls := 0, 0, 0
			deps := &Deps{
				Options:              outcome_summary.Options{Document: outcome_summary.DocumentOptions{ClientAttributeCodes: []string{"gender"}}},
				ResolvePrincipalKind: func(context.Context) int32 { return outcome_summary.PrincipalKindStaff },
				DocumentHeaderName:   "Sample Organization",
				GetSubscriptionGroupClientReportCard: func(_ context.Context, req *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
					projectionCalls++
					if req.GetSubscriptionGroupId() != "group-1" || req.GetClientId() != "client-1" || len(req.GetClientAttributeCodes()) != 1 || req.GetClientAttributeCodes()[0] != "gender" {
						t.Fatalf("projection request = %+v", req)
					}
					return response, nil
				},
				ResolveTemplateBytes: func(_ context.Context, scheduleID, phaseCode string) ([]byte, error) {
					resolverCalls++
					if scheduleID != "schedule-1" || phaseCode != "progress_report" {
						t.Fatalf("phase template scope = schedule %q phase %q", scheduleID, phaseCode)
					}
					return []byte("phase-template"), nil
				},
				ListJobPhases: func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
					return &jobphasepb.ListJobPhasesResponse{Success: true}, nil
				},
			}
			generate := func(template []byte, data map[string]any) ([]byte, error) {
				renderCalls++
				if string(template) != "phase-template" {
					t.Errorf("template = %q", template)
				}
				if data["phase_name"] != "Progress Report" || data["student_name"] != "Año, N" {
					t.Errorf("phase data markers = phase %v student %v", data["phase_name"], data["student_name"])
				}
				if jobs, ok := data["jobs"].([]any); !ok || len(jobs) != 4 {
					t.Errorf("rendered jobs = %#v, want all four available categories", data["jobs"])
				}
				if format == "pdf" {
					return stubPDFBytes, nil
				}
				return stubDocBytes, nil
			}
			deps.GenerateDoc = generate
			deps.GeneratePDF = generate

			handler := NewDownloadHandler(deps)
			r := httptest.NewRequest(http.MethodGet, "/document?period=progress_report&format="+format, nil)
			r.SetPathValue("id", "group-1")
			r.SetPathValue("client_id", "client-1")
			r = r.WithContext(view.WithUserPermissions(r.Context(), types.NewUserPermissions([]string{"subscription_group_outcome_export:read"})))
			w := httptest.NewRecorder()
			handler(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
			}
			wantType := docxContentType
			if format == "pdf" {
				wantType = pdfContentType
			}
			if got := w.Header().Get("Content-Type"); got != wantType {
				t.Errorf("content-type = %q, want %q", got, wantType)
			}
			if disposition := w.Header().Get("Content-Disposition"); !strings.Contains(disposition, "progress-report-grade-10") || !strings.Contains(disposition, "."+format) {
				t.Errorf("content-disposition = %q", disposition)
			}
			if projectionCalls != 1 || resolverCalls != 1 || renderCalls != 1 {
				t.Errorf("calls projection=%d resolver=%d render=%d, want one each", projectionCalls, resolverCalls, renderCalls)
			}
		})
	}
}

func TestDownload_ExplicitUnpublishedSheetRendersStructureWithBlankOutcomes(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.ClientSubscriptionIds = []string{"subscription-1"}
	allowClientProjectionRender(card, "group-1")
	card.RenderGateSheets[2].AllPublished = false
	card.RenderGateSheets[2].AnyWorkflowEntered = true
	card.RenderGateSheets[2].HasData = true
	if blocked, err := clientProjectionRenderStatus(card, "group-1"); err != nil || !blocked {
		t.Fatalf("gate = blocked %v, err %v; want proven unpublished sheet", blocked, err)
	}
	var rendered map[string]any
	deps := &Deps{
		ResolvePrincipalKind: func(context.Context) int32 { return outcome_summary.PrincipalKindStaff },
		GetSubscriptionGroupClientReportCard: func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
			return &exportpb.GetSubscriptionGroupClientReportCardResponse{Success: true, ReportCard: card}, nil
		},
		ResolveTemplateBytes: func(context.Context, string, string) ([]byte, error) { return []byte("template"), nil },
		GeneratePDF: func(_ []byte, data map[string]any) ([]byte, error) {
			rendered = data
			return stubPDFBytes, nil
		},
	}
	r := httptest.NewRequest(http.MethodGet, "/document?period=progress_report&format=pdf", nil)
	r.SetPathValue("id", "group-1")
	r.SetPathValue("client_id", "client-1")
	r = r.WithContext(view.WithUserPermissions(r.Context(), types.NewUserPermissions([]string{"subscription_group_outcome_export:read"})))
	w := httptest.NewRecorder()
	NewDownloadHandler(deps)(w, r)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != pdfContentType || rendered == nil {
		t.Fatalf("download status=%d type=%q body=%s", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}
	if rendered["student_name"] != "Año, N" || rendered["phase_name"] != "Progress Report" {
		t.Fatalf("document identity/period = %#v", rendered)
	}
	jobs, ok := rendered["jobs"].([]any)
	if !ok || len(jobs) == 0 {
		t.Fatalf("jobs = %#v, want complete structure", rendered["jobs"])
	}
	for _, raw := range jobs {
		job := raw.(map[string]any)
		if job["phase_grade"] != "" || job["phase_total"] != "" || job["progress_to_date_total"] != "" {
			t.Fatalf("unpublished outcome value leaked: %#v", job)
		}
		for _, rawAssessment := range job["assessments"].([]any) {
			assessment := rawAssessment.(map[string]any)
			if assessment["achievement_level"] != "" || assessment["comment"] != "" {
				t.Fatalf("unpublished assessment leaked: %#v", assessment)
			}
		}
	}
	sections := rendered["outcome_sections"].([]any)
	if len(sections) == 0 {
		t.Fatalf("outcome sections = %#v, want activity/task structure", sections)
	}
}

func TestClientProjectionRenderGateUsesDistinctTemplateAndSingletonSheets(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.JobPhases = append(card.JobPhases, &jobphasepb.JobPhase{Id: "phase-singleton", JobId: "job-art", Active: true})
	card.RenderGateAppliedSubscriptionGroupId = "group-1"
	card.RenderGateSheets = []*exportpb.ClientReportCardRenderGateSheet{
		{JobTemplatePhaseId: ptr("template-phase-bio"), AppliedSubscriptionGroupId: "group-1", TargetCount: 3, AllPublished: true, HasData: true},
		{JobTemplatePhaseId: ptr("template-phase-chem"), AppliedSubscriptionGroupId: "group-1", TargetCount: 2, AllPublished: true, HasData: true},
		{JobTemplatePhaseId: ptr("template-phase-art"), AppliedSubscriptionGroupId: "group-1", TargetCount: 7, AllPublished: false, AnyWorkflowEntered: true, HasData: true},
		{JobPhaseId: ptr("phase-singleton"), AppliedSubscriptionGroupId: "group-1", TargetCount: 1, AllPublished: true, HasData: true},
	}
	blocked, err := clientProjectionRenderStatus(card, "group-1")
	if err != nil {
		t.Fatalf("clientProjectionRenderStatus() error = %v", err)
	}
	if !blocked {
		t.Fatal("clientProjectionRenderStatus() = false, want blocked for entered, unpublished, data-bearing group sheet")
	}
}

func TestClientProjectionRenderGateFailsClosedOnIncompleteOrMismatchedCoverage(t *testing.T) {
	base := clientPhaseProjectionFixture()
	allowClientProjectionRender(base, "group-1")
	tests := []struct {
		name   string
		mutate func(*exportpb.ClientReportCardProjection)
	}{
		{name: "projection group echo", mutate: func(card *exportpb.ClientReportCardProjection) {
			card.RenderGateAppliedSubscriptionGroupId = "group-other"
		}},
		{name: "missing sheet", mutate: func(card *exportpb.ClientReportCardProjection) {
			card.RenderGateSheets = card.RenderGateSheets[:len(card.RenderGateSheets)-1]
		}},
		{name: "duplicate sheet", mutate: func(card *exportpb.ClientReportCardProjection) {
			card.RenderGateSheets = append(card.RenderGateSheets, card.RenderGateSheets[0])
		}},
		{name: "foreign row group", mutate: func(card *exportpb.ClientReportCardProjection) {
			card.RenderGateSheets[0].AppliedSubscriptionGroupId = "group-other"
		}},
		{name: "short template coverage", mutate: func(card *exportpb.ClientReportCardProjection) { card.RenderGateSheets[2].TargetCount = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := proto.Clone(base).(*exportpb.ClientReportCardProjection)
			tt.mutate(card)
			if _, err := clientProjectionRenderStatus(card, "group-1"); err == nil {
				t.Fatal("clientProjectionRenderStatus() error = nil, want fail-closed coverage error")
			}
		})
	}
}

func allowClientProjectionRender(card *exportpb.ClientReportCardProjection, groupID string) {
	card.RenderGateAppliedSubscriptionGroupId = groupID
	card.RenderGateSheets = []*exportpb.ClientReportCardRenderGateSheet{
		{JobTemplatePhaseId: ptr("template-phase-bio"), AppliedSubscriptionGroupId: groupID, TargetCount: 1, AllPublished: true, HasData: true},
		{JobTemplatePhaseId: ptr("template-phase-chem"), AppliedSubscriptionGroupId: groupID, TargetCount: 1, AllPublished: true, HasData: true},
		{JobTemplatePhaseId: ptr("template-phase-art"), AppliedSubscriptionGroupId: groupID, TargetCount: 2, AllPublished: true, HasData: true},
	}
}

func TestDownload_ExplicitClientPeriodRequiresExportPermissionBeforeProjection(t *testing.T) {
	for _, period := range []string{"year_final", "progress_report"} {
		t.Run(period, func(t *testing.T) {
			projectionCalls := 0
			deps := &Deps{
				ResolvePrincipalKind: func(context.Context) int32 { return outcome_summary.PrincipalKindStaff },
				GetSubscriptionGroupClientReportCard: func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
					projectionCalls++
					return nil, nil
				},
			}
			r := httptest.NewRequest(http.MethodGet, "/document?period="+period, nil)
			r.SetPathValue("id", "group-1")
			r.SetPathValue("client_id", "client-1")
			r = r.WithContext(view.WithUserPermissions(r.Context(), types.NewUserPermissions([]string{"job_outcome_summary:list", "job_outcome_summary:read"})))
			w := httptest.NewRecorder()
			NewDownloadHandler(deps)(w, r)
			if w.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403", w.Code)
			}
			if projectionCalls != 0 {
				t.Fatalf("projection calls = %d, want zero before export permission check", projectionCalls)
			}
		})
	}
}

func TestDownload_ExplicitProjectionErrorMapsTypedNotFoundAndDependencyFailure(t *testing.T) {
	tests := []struct {
		name       string
		projection func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error)
		wantStatus int
	}{
		{
			name: "typed non-enumerating not found",
			projection: func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
				return nil, fmt.Errorf("query: %w", espynaports.ErrClientReportNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "dependency failure",
			projection: func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
				return nil, errors.New("database unavailable")
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "incomplete response",
			projection: func(context.Context, *exportpb.GetSubscriptionGroupClientReportCardRequest) (*exportpb.GetSubscriptionGroupClientReportCardResponse, error) {
				return &exportpb.GetSubscriptionGroupClientReportCardResponse{}, nil
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := &Deps{
				Options: outcome_summary.Options{
					SubscriptionGroupExport: outcome_summary.SubscriptionGroupExportOptions{
						Enabled:            true,
						WholeReportProfile: bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1,
					},
				},
				ResolvePrincipalKind:                 func(context.Context) int32 { return outcome_summary.PrincipalKindStaff },
				GetSubscriptionGroupClientReportCard: tt.projection,
				GenerateDoc:                          func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
			}
			r := httptest.NewRequest(http.MethodGet, "/document?period=s1", nil)
			r.SetPathValue("id", "group-1")
			r.SetPathValue("client_id", "client-1")
			r = r.WithContext(view.WithUserPermissions(r.Context(), types.NewUserPermissions([]string{"subscription_group_outcome_export:read"})))
			w := httptest.NewRecorder()
			NewDownloadHandler(deps)(w, r)
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d body=%s, want %d", w.Code, w.Body.String(), tt.wantStatus)
			}
		})
	}
}

func TestClientProjectionPhaseCatalogIncludesAllActiveClientTemplateCodes(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.JobTemplatePhases = append(card.JobTemplatePhases,
		&jobtemplatephasepb.JobTemplatePhase{Id: "template-phase-art-s1", JobTemplateId: "template-art", Name: "Semester 1", Code: strptr("s1"), PhaseOrder: 1, Active: true},
		&jobtemplatephasepb.JobTemplatePhase{Id: "template-phase-chem-s1", JobTemplateId: "template-chem", Name: "Semester One", Code: strptr("s1"), PhaseOrder: 1, Active: true},
		&jobtemplatephasepb.JobTemplatePhase{Id: "template-phase-unused", JobTemplateId: "not-a-client-template", Name: "Unused", Code: strptr("unused"), PhaseOrder: 0, Active: true},
		&jobtemplatephasepb.JobTemplatePhase{Id: "template-phase-inactive", JobTemplateId: "template-art", Name: "Inactive", Code: strptr("inactive"), PhaseOrder: 0, Active: false},
	)
	card.JobPhases = append(card.JobPhases,
		&jobphasepb.JobPhase{Id: "phase-art-s1", JobId: "job-art", TemplatePhaseId: strptr("template-phase-art-s1"), Active: true},
		&jobphasepb.JobPhase{Id: "phase-chem-s1", JobId: "job-chem", TemplatePhaseId: strptr("template-phase-chem-s1"), Active: true},
	)
	got := clientProjectionPhaseCatalog(card)
	want := []outcome_summary.ClientReportPhase{
		{Code: "progress_report", Name: "Progress Report", Order: 0},
		{Code: "s1", Name: "Semester 1", Order: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("clientProjectionPhaseCatalog() = %#v, want %#v", got, want)
	}
	if !clientProjectionHasPhase(card, "S1") || clientProjectionHasPhase(card, "inactive") || clientProjectionHasPhase(card, "unused") {
		t.Fatalf("phase availability did not follow active client template phases: %#v", got)
	}
}

func TestBuildProjectedYearFinalDataMapsCategoryJobsAndStoredSummary(t *testing.T) {
	card := clientPhaseProjectionFixture()
	categoryA := card.JobCategories[1]
	categoryA.Code = strptr("category_a")
	categoryB := card.JobCategories[0]
	categoryB.Code = strptr("category_b")
	card.Jobs[0].JobCategoryId = strptr("cat-academic")
	card.Jobs[1].JobCategoryId = strptr("cat-other")
	card.Jobs[2].JobCategoryId = strptr("cat-academic")
	card.Jobs[3].JobCategoryId = strptr("cat-academic")
	for _, phase := range card.JobTemplatePhases {
		if phase.GetId() == "template-phase-chem" {
			phase.PhaseOrder = 1
		}
	}
	card.JobOutcomeSummaries = []*jobsumpb.JobOutcomeSummary{
		{JobId: "job-art", Active: true, ScaledLabel: strptr("Excellent")},
		{JobId: "job-chem", Active: true, ScaledLabel: strptr("95")},
	}
	card.PhaseOutcomeSummaries = append(card.PhaseOutcomeSummaries, &phaseoutcomepb.PhaseOutcomeSummary{Id: "summary-chem", JobPhaseId: "phase-chem", Active: true, ScaledLabel: strptr("95")})
	data := buildProjectedYearFinalData(&Deps{DocumentHeaderName: "Sample Organization", CategoryFilter: "category_a", DocOptions: outcome_summary.DocumentOptions{GroupCategoryFilter: "category_b"}}, card, "Teacher A", "2026-09-23", testTime())
	subjects, ok := data["subjects"].([]any)
	if !ok || len(subjects) != 1 {
		t.Fatalf("subjects = %#v, want only the enrolled Art subject", data["subjects"])
	}
	if subject := subjects[0].(map[string]any); subject["subject_name"] != "Art" || subject["myp_overall"] != "Excellent" {
		t.Errorf("subject row = %#v, want source subject and stored final label", subject)
	}
	formations, ok := data["formation_groups"].([]any)
	if !ok || len(formations) != 1 {
		t.Fatalf("formation_groups = %#v, want projected non-academic category", data["formation_groups"])
	}
	formationRows := formations[0].(map[string]any)["rows"].([]any)
	if len(formationRows) != 1 || formationRows[0].(map[string]any)["row_average"] != "95" {
		t.Errorf("formation rows = %#v, want stored selected-client summary", formationRows)
	}
	if data["group_conduct_sem1"] != "95" {
		t.Errorf("group phase rating = %#v, want projected phase summary", data["group_conduct_sem1"])
	}
	if _, hasLegacyVerticalKey := data["conduct_rows"]; hasLegacyVerticalKey {
		t.Errorf("explicit client Year Final exposes legacy conduct_rows key")
	}
	if sections, ok := data["outcome_sections"].([]any); !ok || len(sections) != 1 {
		t.Errorf("outcome_sections = %#v, want generic projected section", data["outcome_sections"])
	}
	categories, ok := data["job_categories"].(map[string]any)
	if !ok {
		t.Fatalf("job_categories = %#v", data["job_categories"])
	}
	category, ok := categories["category_a"].(map[string]any)
	if !ok {
		t.Fatalf("category_a = %#v", categories["category_a"])
	}
	jobs, ok := category["jobs"].([]any)
	if !ok || len(jobs) != 3 {
		t.Fatalf("category_a jobs = %#v, want three", category["jobs"])
	}
	first := jobs[0].(map[string]any)
	if first["job_template_name_display"] != "Art" || first["job_outcome_summary_scaled_label"] != "Excellent" {
		t.Errorf("first category job = %#v", first)
	}
	phases, ok := first["job_template_phases"].(map[string]any)
	if !ok {
		t.Fatalf("job phases = %#v", first["job_template_phases"])
	}
	progress, ok := phases["progress_report"].(map[string]any)
	if !ok || progress["phase_outcome_summary_scaled_label"] != "Meeting" {
		t.Errorf("progress report phase = %#v", phases["progress_report"])
	}
	criteria, ok := first["outcome_criteria"].([]any)
	if !ok {
		t.Fatalf("outcome_criteria = %#v", first["outcome_criteria"])
	}
	criterionNames := make([]string, 0, len(criteria))
	for _, criterion := range criteria {
		criterionNames = append(criterionNames, criterion.(map[string]any)["outcome_criteria_label_display"].(string))
	}
	if want := []string{"Planning", "Technique", "Reflection"}; !reflect.DeepEqual(criterionNames, want) {
		t.Errorf("criterion order = %v, want binding sequence order %v", criterionNames, want)
	}
}

func TestBuildProjectedYearFinalDataKeepsHistoricalInactiveProjectionRows(t *testing.T) {
	card := clientPhaseProjectionFixture()
	card.Context.Historical = true
	card.JobOutcomeSummaries = []*jobsumpb.JobOutcomeSummary{{JobId: "job-art", Active: false, ScaledLabel: strptr("85")}}
	for _, job := range card.Jobs {
		job.Active = false
	}
	for _, template := range card.JobTemplates {
		template.Active = false
	}
	for _, phase := range card.JobPhases {
		phase.Active = false
	}
	for _, phase := range card.JobTemplatePhases {
		phase.Active = false
	}
	for _, task := range card.JobTasks {
		task.Active = false
	}
	for _, task := range card.JobTemplateTasks {
		task.Active = false
	}
	for _, link := range card.TemplateTaskCriteria {
		link.Active = false
	}
	for _, criterion := range card.OutcomeCriteria {
		criterion.Active = false
	}
	for _, summary := range card.PhaseOutcomeSummaries {
		summary.Active = false
	}
	data := buildProjectedYearFinalData(&Deps{CategoryFilter: "academic", DocumentHeaderName: "Sample Organization"}, card, "Teacher A", "2026-09-23", testTime())
	subjects, ok := data["subjects"].([]any)
	if !ok || len(subjects) != 1 {
		t.Fatalf("historical subjects = %#v, want inactive enrolled subject", data["subjects"])
	}
	if subjects[0].(map[string]any)["myp_overall"] != "85" {
		t.Errorf("historical final = %#v, want inactive historical summary", subjects[0])
	}
}

func testTime() time.Time { return time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC) }
