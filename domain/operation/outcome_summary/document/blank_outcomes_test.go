package document

import "testing"

func TestBlankDocumentOutcomeValuesKeepsStructureAndRubric(t *testing.T) {
	data := map[string]any{
		"student_name":      "Año, N",
		"client_attributes": map[string]any{"value": "reference-1"},
		"jobs": []any{map[string]any{
			"job_name":      "Art",
			"phase_grade":   "Meeting",
			"phase_total":   "3",
			"phase_maximum": "4",
			"assessments": []any{map[string]any{
				"assessment_name":   "Studio activity",
				"achievement_level": "Proficient",
				"maximum":           "4",
				"comment":           "Careful work",
			}},
		}},
		"outcome_sections": []any{map[string]any{
			"row_name": "Technique",
			"cells":    []any{map[string]any{"task_name": "Studio task", "value": "3"}},
			"total":    "3",
		}},
		"outcome_cells": map[string]any{"art": map[string]any{"value": "3", "numeric_value": "3"}},
		"job_categories": map[string]any{"academic": map[string]any{"jobs": []any{map[string]any{
			"job_template_name_display":        "Art",
			"job_outcome_summary_scaled_label": "Excellent",
			"job_template_phases":              map[string]any{"s1": map[string]any{"task_outcome_numeric_value_total_derived": "3"}},
		}}}},
	}
	blankDocumentOutcomeValues(data)
	job := data["jobs"].([]any)[0].(map[string]any)
	assessment := job["assessments"].([]any)[0].(map[string]any)
	if job["job_name"] != "Art" || job["phase_grade"] != "" || job["phase_total"] != "" || job["phase_maximum"] != "4" {
		t.Fatalf("job structure/results = %#v", job)
	}
	if assessment["assessment_name"] != "Studio activity" || assessment["maximum"] != "4" || assessment["achievement_level"] != "" || assessment["comment"] != "" {
		t.Fatalf("assessment structure/results = %#v", assessment)
	}
	section := data["outcome_sections"].([]any)[0].(map[string]any)
	cell := section["cells"].([]any)[0].(map[string]any)
	if section["row_name"] != "Technique" || section["total"] != "" || cell["task_name"] != "Studio task" || cell["value"] != "" {
		t.Fatalf("outcome section = %#v", section)
	}
	if data["outcome_cells"].(map[string]any)["art"].(map[string]any)["numeric_value"] != "" || data["client_attributes"].(map[string]any)["value"] != "reference-1" {
		t.Fatalf("outcome or identity attributes = %#v", data)
	}
	categoryJob := data["job_categories"].(map[string]any)["academic"].(map[string]any)["jobs"].([]any)[0].(map[string]any)
	if categoryJob["job_template_name_display"] != "Art" || categoryJob["job_outcome_summary_scaled_label"] != "" {
		t.Fatalf("category job = %#v", categoryJob)
	}
}
