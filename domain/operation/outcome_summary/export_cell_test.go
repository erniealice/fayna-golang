package outcome_summary

import (
	"testing"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

func TestExportCellValue_LabelWinsOverScore(t *testing.T) {
	label := "A"
	score := 7.5
	cell := &exportpb.SubscriptionGroupOutcomeCell{
		JobTemplateId:      "job-1",
		ScaledLabel:        &label,
		ScaledScore:        &score,
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasPositiveMark: true},
	}
	if got := ExportCellValue(cell, "—"); got != "A" {
		t.Fatalf("value = %q, want label", got)
	}
}

func TestExportCellValue_ScoreFallback(t *testing.T) {
	score := 8.5
	cell := &exportpb.SubscriptionGroupOutcomeCell{
		JobTemplateId:      "job-1",
		ScaledScore:        &score,
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasPositiveMark: true},
	}
	if got := ExportCellValue(cell, "—"); got != "8.5" {
		t.Fatalf("value = %q, want score fallback", got)
	}
}

func TestExportCellValue_StoredNumericZeroStaysPresent(t *testing.T) {
	score := float64(0)
	cell := &exportpb.SubscriptionGroupOutcomeCell{
		JobTemplateId:      "job-1",
		ScaledScore:        &score,
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasPositiveMark: true},
	}
	if got := ExportCellValue(cell, "—"); got != "0" {
		t.Fatalf("value = %q, want stored zero", got)
	}
}

func TestExportCellValue_NonEnrolledSuppression(t *testing.T) {
	label := "1"
	cell := &exportpb.SubscriptionGroupOutcomeCell{
		JobTemplateId:      "job-1",
		ScaledLabel:        &label,
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasMarks: true},
	}
	if got := ExportCellValue(cell, "—"); got != "—" {
		t.Fatalf("value = %q, want blank text for a non-enrolled cell", got)
	}
}

func TestExportCellValue_MissingJob(t *testing.T) {
	cell := &exportpb.SubscriptionGroupOutcomeCell{
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasPositiveMark: true},
	}
	if got := ExportCellValue(cell, "—"); got != "—" {
		t.Fatalf("value = %q, want blank text for a missing job", got)
	}
}

func TestExportCellValue_MatchesExportCSVPolicy(t *testing.T) {
	label := "0"
	positiveZero := &exportpb.SubscriptionGroupOutcomeCell{
		JobTemplateId:      "job-a",
		ScaledLabel:        &label,
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasMarks: true, HasPositiveMark: true},
	}
	placeholder := &exportpb.SubscriptionGroupOutcomeCell{
		JobTemplateId:      "job-b",
		ScaledLabel:        &label,
		EnrollmentEvidence: &exportpb.EnrollmentEvidence{HasMarks: true},
	}
	if got := ExportCellValue(positiveZero, ""); got != "0" {
		t.Fatalf("positive stored zero = %q, want CSV value 0", got)
	}
	if got := ExportCellValue(placeholder, ""); got != "" {
		t.Fatalf("placeholder = %q, want the empty CSV field", got)
	}
}
