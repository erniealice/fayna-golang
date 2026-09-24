package document

import (
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
)

func TestClientPhaseCriteriaLetterPrefix(t *testing.T) {
	card := clientPhaseProjectionFixture()
	bare, err := buildClientPhaseReportData(&Deps{}, card, testProgressReportPhaseCode, "", "")
	if err != nil {
		t.Fatal(err)
	}
	labels := outcome_summary.Labels{ClientDocument: outcome_summary.ClientDocumentLabels{
		CriterionPrefix: "Objective {letter}: ",
		NotAssessed:     "Not yet assessed",
	}}
	data, err := buildClientPhaseReportData(&Deps{Labels: labels}, card, testProgressReportPhaseCode, "", "")
	if err != nil {
		t.Fatal(err)
	}
	bareJobs, jobs := bare["jobs"].([]any), data["jobs"].([]any)
	if len(jobs) == 0 || len(jobs) != len(bareJobs) {
		t.Fatalf("jobs = %d, bare jobs = %d; want equal and non-zero", len(jobs), len(bareJobs))
	}
	checked, placeholders := 0, 0
	for j := range jobs {
		got := jobs[j].(map[string]any)["assessments"].([]any)
		want := bareJobs[j].(map[string]any)["assessments"].([]any)
		for i := range got {
			a, b := got[i].(map[string]any), want[i].(map[string]any)
			// Letters restart per activity (position inside the activity), so
			// only the shape is checked here; the first criterion is always A.
			name := a["assessment_name"].(string)
			bareName := b["assessment_name"].(string)
			if len(name) < len("Objective A: ") || !strings.HasPrefix(name, "Objective ") || name[10] < 'A' || name[10] > 'Z' || name[11:] != ": "+bareName {
				t.Fatalf("job %d criterion %d name = %q, want \"Objective <letter>: %s\"", j, i, name, bareName)
			}
			if i == 0 && name[10] != 'A' {
				t.Fatalf("first criterion letter = %q, want A", name[10])
			}
			if strings.Contains(a["assessment_name"].(string), "\n") {
				t.Fatalf("document criterion name must never contain a line break: %q", a["assessment_name"])
			}
			if b["achievement_level"] == "" && b["comment"] == "" {
				if a["comment"] != "Not yet assessed" {
					t.Fatalf("unassessed criterion comment = %q, want Not yet assessed", a["comment"])
				}
				placeholders++
			} else if a["comment"] != b["comment"] {
				t.Fatalf("assessed criterion comment changed: %q vs %q", a["comment"], b["comment"])
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("fixture produced no assessments")
	}
	t.Logf("checked %d criteria, %d placeholders", checked, placeholders)
}
