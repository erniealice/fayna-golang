package document

// blankDocumentOutcomeValues removes recorded results from a document that is
// structurally renderable but contains an unpublished sheet. The render gate
// still has to prove the sheet membership before callers use this fallback.
// Template, task, activity, criterion and rubric labels remain available.
func blankDocumentOutcomeValues(data map[string]any) {
	blankDocumentOutcomeNode(data, false)
}

var documentOutcomeValueKeys = map[string]bool{
	"crit_a": true, "crit_b": true, "crit_c": true, "crit_d": true,
	"crit_sem1": true, "crit_sem2": true, "criteria_total": true,
	"sem1_band": true, "sem2_band": true, "myp_overall": true,
	"sem1_total": true, "sem2_total": true,
	"conduct_sem1": true, "conduct_sem2": true,
	"group_conduct_sem1": true, "group_conduct_sem2": true,
	"row_average":                              true,
	"job_outcome_summary_scaled_label":         true,
	"phase_outcome_summary_scaled_label":       true,
	"task_outcome_numeric_value_max_derived":   true,
	"task_outcome_numeric_value_total_derived": true,
	"numeric_value":                            true, "numeric_value_total_derived": true,
	"scaled_label":      true,
	"achievement_level": true, "comment": true,
	"phase_grade": true, "phase_comment": true,
	"phase_total": true, "progress_to_date_total": true,
	"total": true,
}

func blankDocumentOutcomeNode(node any, inOutcomeCells bool) {
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if key == "client_attributes" {
				continue
			}
			if documentOutcomeValueKeys[key] || (inOutcomeCells && key == "value") {
				value[key] = ""
				continue
			}
			blankDocumentOutcomeNode(child, inOutcomeCells || key == "outcome_cells" || key == "outcome_sections")
		}
	case []any:
		for _, child := range value {
			blankDocumentOutcomeNode(child, inOutcomeCells)
		}
	}
}
