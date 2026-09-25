package outcome_summary

import (
	"testing"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func reportCell(label *string, composite *float64, ev *exportpb.EnrollmentEvidence) *exportpb.SubscriptionGroupOutcomeCell {
	return &exportpb.SubscriptionGroupOutcomeCell{JobTemplateId: "job-1", ScaledLabel: label, SummaryScore: composite, EnrollmentEvidence: ev}
}

func strp(s string) *string   { return &s }
func f64p(v float64) *float64 { return &v }
func enrolled() *exportpb.EnrollmentEvidence {
	return &exportpb.EnrollmentEvidence{HasMarks: true, HasPositiveMark: true}
}

func TestFormatReportCell_Templates(t *testing.T) {
	cell := reportCell(strp("O"), f64p(100), enrolled())
	for _, tt := range []struct{ format, want string }{
		{"", "O"},
		{"{scaled}", "O"},
		{"{composite}", "100"},
		{"{composite} / {scaled}", "100 / O"},
		{"{composite} ({scaled})", "100 (O)"},
	} {
		if got := FormatReportCell(cell, "—", tt.format); got != tt.want {
			t.Errorf("format %q = %q, want %q", tt.format, got, tt.want)
		}
	}
	// Verbatim composite formatting: no rounding, no trailing zeros; 0 is present.
	if got := FormatReportCell(reportCell(strp("7"), f64p(27.5), enrolled()), "—", "{composite}"); got != "27.5" {
		t.Errorf("fractional composite = %q", got)
	}
	if got := FormatReportCell(reportCell(strp("1"), f64p(0), enrolled()), "—", "{composite}"); got != "0" {
		t.Errorf("stored zero composite = %q, want 0", got)
	}
}

func TestFormatReportCell_MissingTokenCollapses(t *testing.T) {
	labelOnly := reportCell(strp("O"), nil, enrolled())
	if got := FormatReportCell(labelOnly, "—", "{composite}"); got != "O" {
		t.Errorf("score-less row under {composite} = %q, want label fallback", got)
	}
	if got := FormatReportCell(labelOnly, "—", "{composite} / {scaled}"); got != "O" {
		t.Errorf("score-less row under both = %q, want %q (no dangling separator)", got, "O")
	}
	scoreOnly := reportCell(nil, f64p(28), enrolled())
	if got := FormatReportCell(scoreOnly, "—", "{composite} / {scaled}"); got != "28" {
		t.Errorf("label-less enrolled row = %q, want 28", got)
	}
	if got := FormatReportCell(reportCell(nil, nil, enrolled()), "—", "{composite} / {scaled}"); got != "—" {
		t.Errorf("both missing = %q, want blank", got)
	}
}

func TestFormatReportCell_NonEnrolledStillBlank(t *testing.T) {
	// Untaken-elective scaffold: all-zero marks, floored "1" year-final.
	scaffold := reportCell(strp("1"), f64p(0), &exportpb.EnrollmentEvidence{HasMarks: true})
	for _, format := range []string{"{scaled}", "{composite}", "{composite} / {scaled}"} {
		if got := FormatReportCell(scaffold, "—", format); got != "—" {
			t.Errorf("format %q non-enrolled = %q, want blank", format, got)
		}
	}
	if got := FormatReportCell(nil, "—", "{composite}"); got != "—" {
		t.Errorf("nil cell = %q, want blank", got)
	}
	noJob := &exportpb.SubscriptionGroupOutcomeCell{SummaryScore: f64p(50), EnrollmentEvidence: enrolled()}
	if got := FormatReportCell(noJob, "—", "{composite}"); got != "—" {
		t.Errorf("missing job cell = %q, want blank", got)
	}
}

func TestCellFormat_InvalidFallsBackToScaled(t *testing.T) {
	for _, bad := range []string{"{total}", "static text", "{composite} {grade}", "{composite"} {
		o := Options{SubscriptionGroupExport: SubscriptionGroupExportOptions{CellFormat: bad}}
		if got := o.ReportCellFormat(); got != CellFormatScaled {
			t.Errorf("invalid %q → %q, want %q", bad, got, CellFormatScaled)
		}
	}
	for _, good := range []string{"{composite}", "{scaled}", "{composite} / {scaled}"} {
		o := Options{SubscriptionGroupExport: SubscriptionGroupExportOptions{CellFormat: good}}
		if got := o.ReportCellFormat(); got != good {
			t.Errorf("valid %q → %q", good, got)
		}
	}
	if got := (Options{}).ReportCellFormat(); got != CellFormatScaled {
		t.Errorf("zero options → %q, want default", got)
	}
}
