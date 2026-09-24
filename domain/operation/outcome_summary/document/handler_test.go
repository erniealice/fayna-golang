package document

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

// --- W5 report-card PDF download handler tests ----------------------------
//
// These pin the ?format param contract added in W5: DOCX stays the unchanged
// default, PDF routes through the injected GeneratePDF closure, the
// LibreOffice-absent sentinel maps to 503 (vs 500 for any other error), an
// unknown format is 400, and the auth/IDOR gates fire identically for both
// formats (a foreign group 404s before either closure is called).

var okPermCodes = []string{"job_outcome_summary:list", "job_outcome_summary:read"}

func sp(s string) *string { return &s }

// stubDocBytes / stubPDFBytes are non-empty sentinels so len>0 passes.
var stubDocBytes = []byte("DOCX-BYTES")
var stubPDFBytes = []byte("%PDF-1.7\nPDF-BYTES")

// fullCardDeps returns Deps whose fetch chain yields exactly one kept subject
// (Mathematics, year-final band "7" so the non-enrolled-placeholder suppressor
// keeps it) for group sec-1 / client stu-1, with the two generator closures
// injected by the caller.
func fullCardDeps(gen, pdf func([]byte, map[string]any) ([]byte, error)) *Deps {
	return &Deps{
		Labels:       outcome_summary.Labels{},
		CommonLabels: pyeza.CommonLabels{},
		GenerateDoc:  gen,
		GeneratePDF:  pdf,
		ListSubscriptionGroups: groupsFn(&subscriptiongrouppb.SubscriptionGroup{
			Id: "sec-1", Active: true, Name: "Grade 10 Gold (AY 2025-2026)",
		}),
		ListSubscriptionGroupMembers: membersFn(&subscriptiongroupmemberpb.SubscriptionGroupMember{
			ClientId: "stu-1", SubscriptionId: "sub-1", Active: true,
		}),
		ListJobs: func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
			return &jobpb.ListJobsResponse{Data: []*jobpb.Job{{
				Id: "job-1", JobTemplateId: sp("tmpl-1"), OriginId: sp("sub-1"), Active: true,
				OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION,
			}}}, nil
		},
		ListJobTemplates: func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error) {
			return &jobtemplatepb.ListJobTemplatesResponse{Data: []*jobtemplatepb.JobTemplate{{Id: "tmpl-1", Name: "Mathematics"}}}, nil
		},
		ListJobOutcomeSummarys: func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
			return &jobsumpb.ListJobOutcomeSummarysResponse{Data: []*jobsumpb.JobOutcomeSummary{{JobId: "job-1", Active: true, ScaledLabel: sp("7")}}}, nil
		},
		ListClients: func(context.Context, *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error) {
			return &clientpb.ListClientsResponse{Data: []*clientpb.Client{{Id: "stu-1", LastName: sp("Dela Cruz"), FirstName: sp("Juan")}}}, nil
		},
		// D5 render-gate reads (required, wired): empty phase set → the gate proves
		// the card safe (never-workflowed) and renders. The gate now FAILS CLOSED if
		// these are unwired (codex §B4), so the happy-path test must supply them.
		ListJobPhases: func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
			return &jobphasepb.ListJobPhasesResponse{Success: true}, nil
		},
		ListJobTasks: func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
			return &jobtaskpb.ListJobTasksResponse{Success: true}, nil
		},
		ListTaskOutcomes: func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
			return &taskoutcomepb.ListTaskOutcomesResponse{Success: true}, nil
		},
	}
}

func reqWithPerms(t *testing.T, target, group, client string, granted bool) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, target, nil)
	r.SetPathValue("id", group)
	r.SetPathValue("client_id", client)
	var perms *types.UserPermissions
	if granted {
		perms = types.NewUserPermissions(okPermCodes)
	} else {
		perms = types.NewEmptyUserPermissions()
	}
	return r.WithContext(view.WithUserPermissions(r.Context(), perms))
}

func TestDownload_InvalidFormat_400(t *testing.T) {
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
		func([]byte, map[string]any) ([]byte, error) { return stubPDFBytes, nil },
	))
	w := httptest.NewRecorder()
	h(w, reqWithPerms(t, "/doc?format=xml", "sec-1", "stu-1", true))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid format must 400, got %d", w.Code)
	}
}

func TestDownload_PDFNotWired_503(t *testing.T) {
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
		nil, // GeneratePDF unwired
	))
	w := httptest.NewRecorder()
	h(w, reqWithPerms(t, "/doc?format=pdf", "sec-1", "stu-1", true))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("pdf with GeneratePDF nil must 503, got %d", w.Code)
	}
}

func TestDownload_LegacyUnpublishedSheetRendersBlankPDF(t *testing.T) {
	var rendered map[string]any
	deps := fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
		func(_ []byte, data map[string]any) ([]byte, error) {
			rendered = data
			return stubPDFBytes, nil
		},
	)
	deps.ListJobPhases = func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
		p := phase("phase-1", jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW, "")
		p.JobId = "job-1"
		return &jobphasepb.ListJobPhasesResponse{Data: []*jobphasepb.JobPhase{p}, Success: true}, nil
	}
	deps.ListJobTasks = func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
		return &jobtaskpb.ListJobTasksResponse{Data: []*jobtaskpb.JobTask{{Id: "task-1", JobPhaseId: "phase-1", Active: true}}, Success: true}, nil
	}
	deps.ListTaskOutcomes = func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
		return &taskoutcomepb.ListTaskOutcomesResponse{Data: []*taskoutcomepb.TaskOutcome{{Id: "outcome-1", JobTaskId: "task-1", Active: true}}, Success: true}, nil
	}
	w := httptest.NewRecorder()
	NewDownloadHandler(deps)(w, reqWithPerms(t, "/doc?format=pdf", "sec-1", "stu-1", true))
	if w.Code != http.StatusOK || rendered == nil {
		t.Fatalf("status=%d body=%s, want blank PDF", w.Code, w.Body.String())
	}
	if rendered["student_name"] != "Dela Cruz, Juan" {
		t.Fatalf("student identity = %#v", rendered["student_name"])
	}
	subjects := rendered["subjects"].([]any)
	if len(subjects) != 1 {
		t.Fatalf("subjects = %#v", subjects)
	}
	subject := subjects[0].(map[string]any)
	if subject["subject_name"] != "Mathematics" || subject["myp_overall"] != "" {
		t.Fatalf("subject label/outcome = %#v", subject)
	}
}

func TestDownload_Forbidden_BothFormats(t *testing.T) {
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
		func([]byte, map[string]any) ([]byte, error) { return stubPDFBytes, nil },
	))
	for _, f := range []string{"", "docx", "pdf"} {
		w := httptest.NewRecorder()
		h(w, reqWithPerms(t, "/doc?format="+f, "sec-1", "stu-1", false))
		if w.Code != http.StatusForbidden {
			t.Fatalf("format=%q without perms must 403, got %d", f, w.Code)
		}
	}
}

func TestDownload_IDOR_ForeignGroup_404_BothFormats(t *testing.T) {
	for _, f := range []string{"docx", "pdf"} {
		d := fullCardDeps(
			func([]byte, map[string]any) ([]byte, error) {
				t.Fatalf("generator must not run for a foreign group")
				return nil, nil
			},
			func([]byte, map[string]any) ([]byte, error) {
				t.Fatalf("generator must not run for a foreign group")
				return nil, nil
			},
		)
		d.ListSubscriptionGroups = groupsFn() // foreign / missing → no rows
		h := NewDownloadHandler(d)
		w := httptest.NewRecorder()
		h(w, reqWithPerms(t, "/doc?format="+f, "sec-1", "stu-1", true))
		if w.Code != http.StatusNotFound {
			t.Fatalf("format=%q foreign group must 404, got %d", f, w.Code)
		}
	}
}

func TestDownload_DOCX_Unchanged(t *testing.T) {
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
		func([]byte, map[string]any) ([]byte, error) {
			t.Fatalf("PDF closure must not run for docx")
			return nil, nil
		},
	))
	// Omitted format defaults to docx.
	w := httptest.NewRecorder()
	h(w, reqWithPerms(t, "/doc", "sec-1", "stu-1", true))
	if w.Code != http.StatusOK {
		t.Fatalf("docx must 200, got %d (%s)", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != docxContentType {
		t.Fatalf("docx content-type = %q, want %q", ct, docxContentType)
	}
	cd := w.Header().Get("Content-Disposition")
	if !operatorFilenamePattern("docx").MatchString(cd) {
		t.Fatalf("docx Content-Disposition = %q, want {name}-stu-1-{unix}.docx", cd)
	}
	if got := w.Body.Bytes(); string(got) != string(stubDocBytes) {
		t.Fatalf("docx body = %q, want the docx bytes", got)
	}
}

func TestDownload_PDF_ContentTypeAndFilename(t *testing.T) {
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) {
			t.Fatalf("DOCX closure must not run for pdf")
			return nil, nil
		},
		func([]byte, map[string]any) ([]byte, error) { return stubPDFBytes, nil },
	))
	w := httptest.NewRecorder()
	h(w, reqWithPerms(t, "/doc?format=pdf", "sec-1", "stu-1", true))
	if w.Code != http.StatusOK {
		t.Fatalf("pdf must 200, got %d (%s)", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != pdfContentType {
		t.Fatalf("pdf content-type = %q, want %q", ct, pdfContentType)
	}
	if ns := w.Header().Get("X-Content-Type-Options"); ns != "nosniff" {
		t.Fatalf("pdf must set nosniff, got %q", ns)
	}
	cd := w.Header().Get("Content-Disposition")
	if !operatorFilenamePattern("pdf").MatchString(cd) {
		t.Fatalf("pdf Content-Disposition = %q, want {name}-stu-1-{unix}.pdf", cd)
	}
	if !strings.Contains(cd, "filename*=UTF-8''") {
		t.Fatalf("pdf Content-Disposition must carry an RFC-5987 filename*, got %q", cd)
	}
	if got := w.Body.Bytes(); string(got) != string(stubPDFBytes) {
		t.Fatalf("pdf body mismatch")
	}
}

func TestDownload_PDF_LibreOfficeUnavailable_503(t *testing.T) {
	// FALLBACK PATH: the fycha sentinel's stable substring ("LibreOffice is not
	// installed"). fayna must not import fycha, so the documented string fallback
	// must still classify this as 503.
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
		func([]byte, map[string]any) ([]byte, error) {
			return nil, fmt.Errorf("PDF conversion unavailable: LibreOffice is not installed (see https://www.libreoffice.org/download/)")
		},
	))
	w := httptest.NewRecorder()
	h(w, reqWithPerms(t, "/doc?format=pdf", "sec-1", "stu-1", true))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("LibreOffice-absent (string fallback) must 503, got %d (%s)", w.Code, w.Body.String())
	}
}

// stubLibreOfficeUnavailableErr mirrors the STRUCTURAL contract fycha's
// document.ErrLibreOfficeUnavailable exposes: a method LibreOfficeUnavailable()
// bool. Crucially its message does NOT contain the "LibreOffice is not installed"
// substring — so a test that passes ONLY because of the string fallback would fail
// here. This is what proves the preferred interface path works independently.
type stubLibreOfficeUnavailableErr struct{}

func (stubLibreOfficeUnavailableErr) Error() string                { return "soffice binary absent on host" }
func (stubLibreOfficeUnavailableErr) LibreOfficeUnavailable() bool { return true }

func TestDownload_PDF_LibreOfficeUnavailable_InterfacePath_503(t *testing.T) {
	// PREFERRED PATH: a typed error implementing LibreOfficeUnavailable() bool, whose
	// MESSAGE lacks the magic substring. Also wrapped once with %w to prove the
	// handler walks the chain via errors.As.
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
		func([]byte, map[string]any) ([]byte, error) {
			return nil, fmt.Errorf("render pipeline: %w", stubLibreOfficeUnavailableErr{})
		},
	))
	w := httptest.NewRecorder()
	h(w, reqWithPerms(t, "/doc?format=pdf", "sec-1", "stu-1", true))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("LibreOffice-absent (structural interface) must 503, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestDownload_PDF_GenericError_500(t *testing.T) {
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
		func([]byte, map[string]any) ([]byte, error) {
			return nil, fmt.Errorf("template processing exploded")
		},
	))
	w := httptest.NewRecorder()
	h(w, reqWithPerms(t, "/doc?format=pdf", "sec-1", "stu-1", true))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("generic pdf error must 500, got %d", w.Code)
	}
}

func operatorFilenamePattern(ext string) *regexp.Regexp {
	return regexp.MustCompile(`filename="[a-z0-9-]+-stu-1-[0-9]{10}\.` + ext + `"`)
}
