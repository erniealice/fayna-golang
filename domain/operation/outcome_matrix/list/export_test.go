package list

// export_test.go — the 20260720 export-drawer additions to the CSV handler:
// period→hide-token mapping + validation, the roster composite phase-column
// derivation, and the composite CSV shape + csvSafe application.

import (
	"context"
	"encoding/csv"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	outcome_matrix "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// codedPhases builds a 2-phase response (s1/s2) with codes populated.
func codedPhases() []*matrixpb.PhaseColumn {
	return []*matrixpb.PhaseColumn{
		{JobTemplatePhaseId: "p1", Label: "Semester 1", SequenceOrder: 1, Code: "s1"},
		{JobTemplatePhaseId: "p2", Label: "Semester 2", SequenceOrder: 2, Code: "s2"},
	}
}

func TestPeriodKnown(t *testing.T) {
	phases := codedPhases()
	cases := []struct {
		period string
		want   bool
	}{
		{"", true},      // all
		{"final", true}, // reserved composite
		{"s1", true},    // phase code
		{"s2", true},    // phase code
		{"s3", false},   // unknown code → 400
		{"garbage", false},
		{"p1", false}, // a phase ID is NOT a period token (codes only)
	}
	for _, c := range cases {
		if got := periodKnown(c.period, phases); got != c.want {
			t.Errorf("periodKnown(%q) = %v, want %v", c.period, got, c.want)
		}
	}
	// Zero-phase template: only "" and "final" are known.
	if !periodKnown("", nil) || !periodKnown("final", nil) {
		t.Error("zero-phase: all/final must be known")
	}
	if periodKnown("s1", nil) {
		t.Error("zero-phase: s1 must be unknown")
	}
}

func TestPeriodHideTokens(t *testing.T) {
	phases := codedPhases()

	// s1 keeps p1, hides p2 (the OTHER phase).
	if got := periodHideTokens("s1", phases); len(got) != 1 || got[0] != "p2" {
		t.Errorf("periodHideTokens(s1) = %v, want [p2]", got)
	}
	// s2 keeps p2, hides p1.
	if got := periodHideTokens("s2", phases); len(got) != 1 || got[0] != "p1" {
		t.Errorf("periodHideTokens(s2) = %v, want [p1]", got)
	}
	// "" (all) and "final" never prune.
	if got := periodHideTokens("", phases); got != nil {
		t.Errorf("periodHideTokens(all) = %v, want nil", got)
	}
	if got := periodHideTokens("final", phases); got != nil {
		t.Errorf("periodHideTokens(final) = %v, want nil", got)
	}
	// A codeless phase is hidden by any semester selection (it is never the target).
	mixed := append(codedPhases(), &matrixpb.PhaseColumn{JobTemplatePhaseId: "p3", Label: "Extra", SequenceOrder: 3})
	got := periodHideTokens("s1", mixed)
	if len(got) != 2 {
		t.Fatalf("periodHideTokens(s1) over 3 phases = %v, want 2 hidden", got)
	}
}

func TestRosterPhaseColumns(t *testing.T) {
	rows := []*matrixpb.OutcomeSummaryRosterRow{
		{
			ClientId: "c1",
			Phases: []*matrixpb.OutcomeSummaryPhaseEntry{
				{JobTemplatePhaseId: "p2", Label: "Semester 2", SequenceOrder: 2},
				{JobTemplatePhaseId: "p1", Label: "Semester 1", SequenceOrder: 1},
			},
		},
		{
			ClientId: "c2",
			Phases: []*matrixpb.OutcomeSummaryPhaseEntry{
				{JobTemplatePhaseId: "p1", Label: "Semester 1", SequenceOrder: 1},
			},
		},
	}
	cols := rosterPhaseColumns(rows)
	if len(cols) != 2 {
		t.Fatalf("want 2 union'd phase cols, got %d", len(cols))
	}
	// sequence order: p1 (seq 1) then p2 (seq 2).
	if cols[0].id != "p1" || cols[1].id != "p2" {
		t.Fatalf("phase columns not in sequence order: %+v", cols)
	}
}

func TestWriteFinalCompositeCSV(t *testing.T) {
	roster := &matrixpb.GetOutcomeSummaryRosterResponse{
		Rows: []*matrixpb.OutcomeSummaryRosterRow{
			{
				ClientId:    "client-1",
				ClientLabel: "client-1",
				Phases: []*matrixpb.OutcomeSummaryPhaseEntry{
					{JobTemplatePhaseId: "p1", Code: "s1", Label: "Semester 1", SequenceOrder: 1, ScaledLabel: "A"},
					{JobTemplatePhaseId: "p2", Code: "s2", Label: "Semester 2", SequenceOrder: 2, ScaledLabel: "=cmd"},
				},
				YearFinalLabel:           "A",
				YearFinalIsAuthoritative: true,
			},
			{
				ClientId:    "client-2",
				ClientLabel: "client-2",
				Phases: []*matrixpb.OutcomeSummaryPhaseEntry{
					{JobTemplatePhaseId: "p1", Code: "s1", Label: "Semester 1", SequenceOrder: 1, ScaledLabel: "B"},
					{JobTemplatePhaseId: "p2", Code: "s2", Label: "Semester 2", SequenceOrder: 2, ScaledLabel: "C"},
				},
				YearFinalLabel: "B",
			},
		},
	}
	deps := &PageViewDeps{
		Labels: outcome_matrix.DefaultLabels(),
		GetOutcomeSummaryRoster: func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
			return roster, nil
		},
	}

	rec := httptest.NewRecorder()
	writeFinalCompositeCSV(context.Background(), rec, deps, "Arts", "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "-final.csv") {
		t.Errorf("Content-Disposition missing -final suffix: %q", cd)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing nosniff header")
	}

	recs, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(recs) != 3 {
		t.Fatalf("want header + 2 rows, got %d records", len(recs))
	}
	// Header: Client · Semester 1 Final · Semester 2 Final · Final.
	wantHeader := []string{"Client", "Semester 1 Final", "Semester 2 Final", "Final"}
	for i, h := range wantHeader {
		if recs[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, recs[0][i], h)
		}
	}
	// Row 1: label falls back to the (short) id (no ListClients wired); the s2
	// cell "=cmd" is csvSafe-guarded with a leading tab; year final "A".
	if recs[1][0] != "client-1" {
		t.Errorf("row1 label = %q, want client-1", recs[1][0])
	}
	if recs[1][1] != "A" {
		t.Errorf("row1 s1 = %q, want A", recs[1][1])
	}
	if recs[1][2] != "\t=cmd" {
		t.Errorf("row1 s2 not csvSafe-guarded: %q", recs[1][2])
	}
	if recs[1][3] != "A" {
		t.Errorf("row1 final = %q, want A", recs[1][3])
	}
	// Row 2.
	if recs[2][0] != "client-2" || recs[2][1] != "B" || recs[2][2] != "C" || recs[2][3] != "B" {
		t.Errorf("row2 = %v, want [client-2 B C B]", recs[2])
	}
}

func TestWriteFinalCompositeCSV_ZeroRows404(t *testing.T) {
	deps := &PageViewDeps{
		Labels: outcome_matrix.DefaultLabels(),
		GetOutcomeSummaryRoster: func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
			return &matrixpb.GetOutcomeSummaryRosterResponse{Success: true}, nil
		},
	}
	rec := httptest.NewRecorder()
	writeFinalCompositeCSV(context.Background(), rec, deps, "Arts", "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if rec.Code != 404 {
		t.Fatalf("empty roster: status = %d, want 404 (never an empty CSV)", rec.Code)
	}

	// Nil closure → 404 (no composite source), never a 500.
	rec2 := httptest.NewRecorder()
	writeFinalCompositeCSV(context.Background(), rec2, &PageViewDeps{Labels: outcome_matrix.DefaultLabels()}, "Arts", "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if rec2.Code != 404 {
		t.Fatalf("nil roster closure: status = %d, want 404", rec2.Code)
	}
}

// TestWriteFinalCompositeCSV_ScopePassthrough pins Finding 1's fix at the fayna
// boundary: the resolved scope is threaded UNCHANGED into the roster request, so a
// MINE export carries MINE (never silently widened to the full workspace roster).
// It also pins the fail-closed leg: when the adapter returns zero rows for a
// MINE-scoped non-staff caller, the composite path 404s (never an empty CSV).
func TestWriteFinalCompositeCSV_ScopePassthrough(t *testing.T) {
	// (1) MINE request carries MINE scope through to the roster read.
	var gotScope matrixpb.OutcomeMatrixScope = matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_UNSPECIFIED
	deps := &PageViewDeps{
		Labels: outcome_matrix.DefaultLabels(),
		GetOutcomeSummaryRoster: func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
			gotScope = req.GetScope()
			return &matrixpb.GetOutcomeSummaryRosterResponse{
				Rows: []*matrixpb.OutcomeSummaryRosterRow{{ClientId: "c1", ClientLabel: "c1"}},
			}, nil
		},
	}
	rec := httptest.NewRecorder()
	writeFinalCompositeCSV(context.Background(), rec, deps, "Arts", "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE)
	if gotScope != matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE {
		t.Fatalf("roster request scope = %v, want MINE (scope must pass through unchanged)", gotScope)
	}

	// (2) ALL request carries ALL scope through.
	writeFinalCompositeCSV(context.Background(), httptest.NewRecorder(), deps, "Arts", "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if gotScope != matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL {
		t.Fatalf("roster request scope = %v, want ALL", gotScope)
	}

	// (3) MINE non-staff → adapter fails closed to zero rows → 404 at the view.
	closed := &PageViewDeps{
		Labels: outcome_matrix.DefaultLabels(),
		GetOutcomeSummaryRoster: func(ctx context.Context, req *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
			if req.GetScope() == matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE {
				return &matrixpb.GetOutcomeSummaryRosterResponse{Success: true}, nil // fail-closed
			}
			return &matrixpb.GetOutcomeSummaryRosterResponse{
				Rows: []*matrixpb.OutcomeSummaryRosterRow{{ClientId: "c1", ClientLabel: "c1"}},
			}, nil
		},
	}
	rec3 := httptest.NewRecorder()
	writeFinalCompositeCSV(context.Background(), rec3, closed, "Arts", "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE)
	if rec3.Code != 404 {
		t.Fatalf("MINE non-staff (zero rows): status = %d, want 404", rec3.Code)
	}
}

// --- format=pdf branch (P5) --------------------------------------------------

// pdfMatrixResp is the minimal GetOutcomeMatrix response the pdf branch reads
// (job_template_name fallback for the header/filename).
func pdfMatrixResp() *matrixpb.GetOutcomeMatrixResponse {
	return &matrixpb.GetOutcomeMatrixResponse{JobTemplateName: "Arts"}
}

// pdfRoster is a 1-student roster (s1/s2 + year final) for the happy path.
func pdfRoster() *matrixpb.GetOutcomeSummaryRosterResponse {
	return &matrixpb.GetOutcomeSummaryRosterResponse{
		Rows: []*matrixpb.OutcomeSummaryRosterRow{{
			ClientId:    "c1",
			ClientLabel: "c1",
			Phases: []*matrixpb.OutcomeSummaryPhaseEntry{
				{JobTemplatePhaseId: "p1", Label: "Semester 1", SequenceOrder: 1, ScaledLabel: "6"},
				{JobTemplatePhaseId: "p2", Label: "Semester 2", SequenceOrder: 2, ScaledLabel: "7"},
			},
			YearFinalLabel: "7",
		}},
	}
}

// libreOfficeAbsentErr is a stub satisfying the structural fycha sentinel contract
// (LibreOfficeUnavailable() bool) so the handler maps it to 503, proving the
// errors.As idiom without importing fycha.
type libreOfficeAbsentErr struct{}

func (libreOfficeAbsentErr) Error() string                { return "soffice not found" }
func (libreOfficeAbsentErr) LibreOfficeUnavailable() bool { return true }

// TestGradeSheetPDF_FailLoud_NoTemplate proves the Q1 fail-loud contract: a nil
// resolver, a resolver miss (nil bytes), and a nil GeneratePDF all 503 with the
// lyngua NoTemplateError body — never an embedded fallback, never a 200.
func TestGradeSheetPDF_FailLoud_NoTemplate(t *testing.T) {
	labels := outcome_matrix.DefaultLabels()
	wantBody := labels.Export.NoTemplateError

	// (a) GeneratePDF wired, but ResolveSheetTemplateBytes is nil → 503.
	deps := &PageViewDeps{
		Labels:      labels,
		GeneratePDF: func(_ []byte, _ map[string]any) ([]byte, error) { return []byte("pdf"), nil },
	}
	rec := httptest.NewRecorder()
	writeGradeSheetPDF(context.Background(), rec, deps, pdfMatrixResp(), "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if rec.Code != 503 || !strings.Contains(rec.Body.String(), wantBody) {
		t.Fatalf("nil resolver: status=%d body=%q, want 503 + %q", rec.Code, rec.Body.String(), wantBody)
	}

	// (b) Resolver present but returns nil bytes (miss) → 503 fail-loud.
	deps2 := &PageViewDeps{
		Labels:                    labels,
		GeneratePDF:               func(_ []byte, _ map[string]any) ([]byte, error) { return []byte("pdf"), nil },
		ResolveSheetTemplateBytes: func(_ context.Context, _, _ string) ([]byte, error) { return nil, nil },
	}
	rec2 := httptest.NewRecorder()
	writeGradeSheetPDF(context.Background(), rec2, deps2, pdfMatrixResp(), "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if rec2.Code != 503 || !strings.Contains(rec2.Body.String(), wantBody) {
		t.Fatalf("resolver miss: status=%d body=%q, want 503 + %q", rec2.Code, rec2.Body.String(), wantBody)
	}

	// (c) Nil GeneratePDF (render not configured) → 503.
	deps3 := &PageViewDeps{Labels: labels}
	rec3 := httptest.NewRecorder()
	writeGradeSheetPDF(context.Background(), rec3, deps3, pdfMatrixResp(), "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if rec3.Code != 503 || !strings.Contains(rec3.Body.String(), wantBody) {
		t.Fatalf("nil GeneratePDF: status=%d body=%q, want 503 + %q", rec3.Code, rec3.Body.String(), wantBody)
	}
}

// TestGradeSheetPDF_SofficeUnavailable proves the soffice-503 mapping: a wired
// GeneratePDF returning the structural LibreOffice-unavailable sentinel is a 503
// (not a 500), and a generic error is a 500.
func TestGradeSheetPDF_SofficeUnavailable(t *testing.T) {
	labels := outcome_matrix.DefaultLabels()
	base := func(gen func([]byte, map[string]any) ([]byte, error)) *PageViewDeps {
		return &PageViewDeps{
			Labels:                    labels,
			GeneratePDF:               gen,
			ResolveSheetTemplateBytes: func(_ context.Context, _, _ string) ([]byte, error) { return []byte("TPL"), nil },
			GetOutcomeSummaryRoster: func(_ context.Context, _ *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
				return pdfRoster(), nil
			},
		}
	}

	// Structural sentinel → 503.
	rec := httptest.NewRecorder()
	writeGradeSheetPDF(context.Background(), rec, base(func(_ []byte, _ map[string]any) ([]byte, error) {
		return nil, libreOfficeAbsentErr{}
	}), pdfMatrixResp(), "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if rec.Code != 503 {
		t.Fatalf("soffice sentinel: status = %d, want 503", rec.Code)
	}

	// Substring fallback (typed wrapper dropped) → 503.
	rec2 := httptest.NewRecorder()
	writeGradeSheetPDF(context.Background(), rec2, base(func(_ []byte, _ map[string]any) ([]byte, error) {
		return nil, errors.New("conversion failed: LibreOffice is not installed")
	}), pdfMatrixResp(), "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if rec2.Code != 503 {
		t.Fatalf("soffice substring: status = %d, want 503", rec2.Code)
	}

	// Generic error → 500 (not misclassified as unavailable).
	rec3 := httptest.NewRecorder()
	writeGradeSheetPDF(context.Background(), rec3, base(func(_ []byte, _ map[string]any) ([]byte, error) {
		return nil, errors.New("template parse error")
	}), pdfMatrixResp(), "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if rec3.Code != 500 {
		t.Fatalf("generic error: status = %d, want 500", rec3.Code)
	}
}

// TestGradeSheetPDF_HappyPath proves a wired pipeline streams the PDF with the
// nosniff header + a .pdf attachment disposition, and that the builder passed the
// template bytes + a students loop through to GeneratePDF.
func TestGradeSheetPDF_HappyPath(t *testing.T) {
	var gotTpl []byte
	var gotData map[string]any
	deps := &PageViewDeps{
		Labels:                    outcome_matrix.DefaultLabels(),
		ResolveSheetTemplateBytes: func(_ context.Context, _, _ string) ([]byte, error) { return []byte("TPL"), nil },
		GetOutcomeSummaryRoster: func(_ context.Context, _ *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
			return pdfRoster(), nil
		},
		GeneratePDF: func(tpl []byte, data map[string]any) ([]byte, error) {
			gotTpl, gotData = tpl, data
			return []byte("%PDF-1.7 fake"), nil
		},
	}
	rec := httptest.NewRecorder()
	writeGradeSheetPDF(context.Background(), rec, deps, pdfMatrixResp(), "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing nosniff header")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type = %q, want application/pdf", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, ".pdf") || !strings.Contains(cd, "attachment") {
		t.Errorf("Content-Disposition = %q, want attachment .pdf", cd)
	}
	if string(gotTpl) != "TPL" {
		t.Errorf("template bytes not threaded to GeneratePDF: %q", gotTpl)
	}
	students, ok := gotData["students"].([]any)
	if !ok || len(students) != 1 {
		t.Fatalf("students loop = %v, want 1 item", gotData["students"])
	}
}

// TestGradeSheetPDF_ZeroRoster404 proves a zero-row roster (foreign/empty template
// or MINE non-staff) 404s — never an empty PDF (composite-CSV IDOR parity).
func TestGradeSheetPDF_ZeroRoster404(t *testing.T) {
	deps := &PageViewDeps{
		Labels:                    outcome_matrix.DefaultLabels(),
		GeneratePDF:               func(_ []byte, _ map[string]any) ([]byte, error) { return []byte("pdf"), nil },
		ResolveSheetTemplateBytes: func(_ context.Context, _, _ string) ([]byte, error) { return []byte("TPL"), nil },
		GetOutcomeSummaryRoster: func(_ context.Context, _ *matrixpb.GetOutcomeSummaryRosterRequest) (*matrixpb.GetOutcomeSummaryRosterResponse, error) {
			return &matrixpb.GetOutcomeSummaryRosterResponse{Success: true}, nil
		},
	}
	rec := httptest.NewRecorder()
	writeGradeSheetPDF(context.Background(), rec, deps, pdfMatrixResp(), "tmpl-1", matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL)
	if rec.Code != 404 {
		t.Fatalf("zero roster: status = %d, want 404", rec.Code)
	}
}
