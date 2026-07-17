package document

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

// --- W5 report-card PDF download handler tests ----------------------------
//
// These pin the ?format param contract added in W5: DOCX stays the unchanged
// default, PDF routes through the injected GeneratePDF closure, the
// LibreOffice-absent sentinel maps to 503 (vs 500 for any other error), an
// unknown format is 400, and the auth/IDOR gates fire identically for both
// formats (a foreign section 404s before either closure is called).

const okPermCode = "job_outcome_summary:list"

func sp(s string) *string { return &s }

// stubDocBytes / stubPDFBytes are non-empty sentinels so len>0 passes.
var stubDocBytes = []byte("DOCX-BYTES")
var stubPDFBytes = []byte("%PDF-1.7\nPDF-BYTES")

// fullCardDeps returns Deps whose fetch chain yields exactly one kept subject
// (Mathematics, year-final band "7" so the non-enrolled-placeholder suppressor
// keeps it) for section sec-1 / client stu-1, with the two generator closures
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
	}
}

func reqWithPerms(t *testing.T, target, section, client string, granted bool) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, target, nil)
	r.SetPathValue("id", section)
	r.SetPathValue("client_id", client)
	var perms *types.UserPermissions
	if granted {
		perms = types.NewUserPermissions([]string{okPermCode})
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

func TestDownload_IDOR_ForeignSection_404_BothFormats(t *testing.T) {
	for _, f := range []string{"docx", "pdf"} {
		d := fullCardDeps(
			func([]byte, map[string]any) ([]byte, error) { t.Fatalf("generator must not run for a foreign section"); return nil, nil },
			func([]byte, map[string]any) ([]byte, error) { t.Fatalf("generator must not run for a foreign section"); return nil, nil },
		)
		d.ListSubscriptionGroups = groupsFn() // foreign / missing → no rows
		h := NewDownloadHandler(d)
		w := httptest.NewRecorder()
		h(w, reqWithPerms(t, "/doc?format="+f, "sec-1", "stu-1", true))
		if w.Code != http.StatusNotFound {
			t.Fatalf("format=%q foreign section must 404, got %d", f, w.Code)
		}
	}
}

func TestDownload_DOCX_Unchanged(t *testing.T) {
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { return stubDocBytes, nil },
		func([]byte, map[string]any) ([]byte, error) { t.Fatalf("PDF closure must not run for docx"); return nil, nil },
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
	if !strings.Contains(cd, `filename="report-card-`) || !strings.HasSuffix(cd, `.docx"`) {
		t.Fatalf("docx Content-Disposition unchanged shape expected, got %q", cd)
	}
	if got := w.Body.Bytes(); string(got) != string(stubDocBytes) {
		t.Fatalf("docx body = %q, want the docx bytes", got)
	}
}

func TestDownload_PDF_ContentTypeAndFilename(t *testing.T) {
	h := NewDownloadHandler(fullCardDeps(
		func([]byte, map[string]any) ([]byte, error) { t.Fatalf("DOCX closure must not run for pdf"); return nil, nil },
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
	// LOCKED filename: "Report Card - {Student} - {AY} - {unixMilli}.pdf".
	if !strings.Contains(cd, "filename*=UTF-8''") {
		t.Fatalf("pdf Content-Disposition must carry an RFC-5987 filename*, got %q", cd)
	}
	if !strings.Contains(cd, "Report%20Card%20-%20") {
		t.Fatalf("pdf filename* must be the LOCKED 'Report Card - ...' form, got %q", cd)
	}
	if !strings.Contains(cd, "2025-2026") {
		t.Fatalf("pdf filename must include the AY, got %q", cd)
	}
	if !strings.HasSuffix(cd, ".pdf") && !strings.Contains(cd, ".pdf") {
		t.Fatalf("pdf filename must end .pdf, got %q", cd)
	}
	// ASCII fallback must be present and reference the locked prefix.
	if !strings.Contains(cd, `filename="Report Card - `) {
		t.Fatalf("pdf Content-Disposition must carry the ASCII filename= fallback, got %q", cd)
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

func (stubLibreOfficeUnavailableErr) Error() string                 { return "soffice binary absent on host" }
func (stubLibreOfficeUnavailableErr) LibreOfficeUnavailable() bool  { return true }

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
