package subscription_group

import (
	"context"
	"strings"
	"testing"

	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
)

// TestBuildRows_CellFormatComposite: the group grid shows the stored composite
// under "{composite}", falls back to the scaled label when no composite is
// stored, and keeps enrollment blanking keyed on the scaled label.
func TestBuildRows_CellFormatComposite(t *testing.T) {
	clients := map[string]client{"stu1": {clientID: "stu1", name: "Doe, Jane"}}
	templateIDs := []string{"tA", "tB", "tP"}
	cellJob := map[string]string{"stu1\x00tA": "jobA", "stu1\x00tB": "jobB", "stu1\x00tP": "jobP"}
	labels := map[string]string{"jobA": "O", "jobB": "7", "jobP": "1"}
	f := func(v float64) *float64 { return &v }
	scores := map[string]*float64{"jobA": f(94), "jobP": f(0)} // jobB: label-only
	ev := map[string]outcome_summary.EnrollmentEvidence{
		"jobA": {HasMarks: true, HasPositiveMark: true},
		"jobB": {HasMarks: true, HasPositiveMark: true},
		"jobP": {HasMarks: true}, // all-zero scaffold → blank
	}
	for _, tt := range []struct {
		format string
		want   []string
	}{
		{"", []string{"O", "7", ""}},
		{"{composite}", []string{"94", "7", ""}},
		{"{composite} / {scaled}", []string{"94 / O", "7", ""}},
	} {
		rows := buildRows(clients, templateIDs, cellJob, labels, scores, tt.format, ev, "sec1", outcome_summary.Routes{}, outcome_summary.Labels{}, false, false)
		for i, want := range tt.want {
			cell := rows[0].Cells[2+i]
			if got := types.CellCSV(cell); got != want {
				t.Errorf("format %q column %d CSV = %q, want %q", tt.format, i, got, want)
			}
			if want != "" && !strings.Contains(string(cell.HTML), ">"+want+"<") {
				t.Errorf("format %q column %d HTML = %q, want text %q", tt.format, i, cell.HTML, want)
			}
		}
	}
}

// TestFetchSummaryLabels_ImportedZeroCompositeIsAbsent: an imported
// (authoritative) final's stored 0 is a placeholder, so no composite is
// returned and the cell falls back to its label; a computed 0 is real.
func TestFetchSummaryLabels_ImportedZeroCompositeIsAbsent(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	lbl := func(s string) *string { return &s }
	deps := &Deps{ListJobOutcomeSummarys: func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
		return &jobsumpb.ListJobOutcomeSummarysResponse{Data: []*jobsumpb.JobOutcomeSummary{
			{JobId: "imported0", Active: true, IsAuthoritative: true, SummaryScore: f(0), ScaledLabel: lbl("100")},
			{JobId: "imported28", Active: true, IsAuthoritative: true, SummaryScore: f(28), ScaledLabel: lbl("7")},
			{JobId: "computed0", Active: true, SummaryScore: f(0), ScaledLabel: lbl("1")},
		}}, nil
	}}
	labels, scores := fetchSummaryLabels(context.Background(), deps, []string{"imported0", "imported28", "computed0"}, outcome_summary.Labels{})
	if _, ok := scores["imported0"]; ok {
		t.Error("imported zero composite must read as absent")
	}
	if s := scores["imported28"]; s == nil || *s != 28 {
		t.Errorf("imported real composite = %v, want 28", s)
	}
	if s := scores["computed0"]; s == nil || *s != 0 {
		t.Errorf("computed zero composite = %v, want present 0", s)
	}
	if got := outcome_summary.ApplyCellFormat("{composite}", labels["imported0"], scores["imported0"]); got != "100" {
		t.Errorf("imported zero cell = %q, want label fallback 100", got)
	}
}
