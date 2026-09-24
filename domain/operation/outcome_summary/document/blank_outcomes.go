package document

import (
	"strings"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	phaseoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
	"google.golang.org/protobuf/proto"
)

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

// redactBlockedClientPhaseScores returns a copy of the client projection in
// which every result recorded on a blocked (unpublished) sheet is reduced to
// its teacher note. Owner decision 2026-09-24 for the phase (Progress Report)
// document: notes print regardless of sheet status or viewer role; scores,
// levels, totals and grades stay blank until the sheet is published.
//
// Redaction happens on the projection, before any mapping, so every derived
// value (subject levels, phase and progress-to-date totals, fixed cells and
// repeated outcome sections) is blank by construction, per sheet: a published
// Term 1 renders its scores even while Term 2 or Term 3 is still in progress.
// Fields are allowlisted, so a new score field added to either message stays
// redacted until it is explicitly admitted here.
func redactBlockedClientPhaseScores(card *exportpb.ClientReportCardProjection, gate clientSheetGate) *exportpb.ClientReportCardProjection {
	if card == nil || !gate.anyBlocked() {
		return card
	}
	out := proto.Clone(card).(*exportpb.ClientReportCardProjection)

	phases := make(map[string]*jobphasepb.JobPhase, len(out.GetJobPhases()))
	for _, phase := range out.GetJobPhases() {
		if phase != nil {
			phases[strings.TrimSpace(phase.GetId())] = phase
		}
	}
	taskPhase := make(map[string]string, len(out.GetJobTasks()))
	for _, task := range out.GetJobTasks() {
		if task != nil {
			taskPhase[strings.TrimSpace(task.GetId())] = strings.TrimSpace(task.GetJobPhaseId())
		}
	}
	// A phase the projection does not describe is treated as blocked.
	blocked := func(jobPhaseID string) bool {
		return gate.phaseBlocked(phases[strings.TrimSpace(jobPhaseID)])
	}

	for i, outcome := range out.GetTaskOutcomes() {
		if outcome == nil || !blocked(taskPhase[strings.TrimSpace(outcome.GetJobTaskId())]) {
			continue
		}
		out.TaskOutcomes[i] = &exportpb.ClientReportCardTaskOutcome{
			JobTaskId:              outcome.GetJobTaskId(),
			TemplateTaskCriteriaId: outcome.GetTemplateTaskCriteriaId(),
			DeterminationNote:      outcome.DeterminationNote,
			RecordedDate:           outcome.RecordedDate,
		}
	}
	for i, summary := range out.GetPhaseOutcomeSummaries() {
		if summary == nil || !blocked(summary.GetJobPhaseId()) {
			continue
		}
		out.PhaseOutcomeSummaries[i] = &phaseoutcomepb.PhaseOutcomeSummary{
			Id:           summary.GetId(),
			JobPhaseId:   summary.GetJobPhaseId(),
			JobId:        summary.GetJobId(),
			Narrative:    summary.Narrative,
			Active:       summary.GetActive(),
			DateCreated:  summary.DateCreated,
			DateModified: summary.DateModified,
			WorkspaceId:  summary.WorkspaceId,
		}
	}
	return out
}
