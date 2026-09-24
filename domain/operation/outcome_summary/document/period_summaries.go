package document

import (
	"sort"
	"strconv"
	"strings"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	phaseoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

// periodSummaryCount is the number of period columns the summary tables carry
// (phase order 1..N of the job's template phases).
const periodSummaryCount = 3

// buildPeriodSummaries fills the optional period-summary tokens of the phase
// document from stored, already-computed values only:
//
//   - summary_jobs: one row per job of opts.PeriodSummaryJobCategoryCode; each
//     period cell is that job phase's phase_outcome_summary.summary_score (the
//     grade computation's composite, e.g. the rounded average of the recorded
//     activities);
//   - summary_tasks: one row per activity (template task code) of the single job
//     of opts.PeriodSummaryTaskCategoryCode, one recorded value per period;
//   - summary_average_period_N / summary_transmuted_period_N: that job's phase
//     summary score and scaled label per period.
//
// Unpublished sheets reach here already redacted (scores stripped, notes kept),
// so their cells are blank. Nothing is recomputed here: the grade sheet and the
// document read the same stored summary.
func buildPeriodSummaries(card *exportpb.ClientReportCardProjection, jobCategoryCode, taskCategoryCode string, historical bool) map[string]any {
	out := map[string]any{"summary_jobs": []any{}, "summary_tasks": []any{}}
	for n := 1; n <= periodSummaryCount; n++ {
		out["summary_average_period_"+strconv.Itoa(n)] = ""
		out["summary_transmuted_period_"+strconv.Itoa(n)] = ""
	}
	if card == nil {
		return out
	}
	clientID := strings.TrimSpace(card.GetClient().GetClientId())
	categoryCodeByID := map[string]string{}
	for _, category := range card.GetJobCategories() {
		if category != nil {
			categoryCodeByID[category.GetId()] = strings.ToLower(strings.TrimSpace(category.GetCode()))
		}
	}
	templateCategory := map[string]string{}
	templateName := map[string]string{}
	for _, template := range card.GetJobTemplates() {
		if template != nil {
			templateCategory[template.GetId()] = template.GetJobCategoryId()
			templateName[template.GetId()] = template.GetName()
		}
	}
	templatePhases := map[string]*jobtemplatephasepb.JobTemplatePhase{}
	for _, phase := range card.GetJobTemplatePhases() {
		if phase != nil && (historical || phase.GetActive()) {
			templatePhases[phase.GetId()] = phase
		}
	}
	phasesByJob := map[string][]*jobphasepb.JobPhase{}
	for _, phase := range card.GetJobPhases() {
		if phase != nil && (historical || phase.GetActive()) {
			phasesByJob[phase.GetJobId()] = append(phasesByJob[phase.GetJobId()], phase)
		}
	}
	summaries := map[string]*phaseoutcomepb.PhaseOutcomeSummary{}
	for _, summary := range card.GetPhaseOutcomeSummaries() {
		if summary == nil || (!historical && !summary.GetActive()) {
			continue
		}
		current := summaries[summary.GetJobPhaseId()]
		if current == nil || summary.GetDateModified() > current.GetDateModified() ||
			summary.GetDateModified() == current.GetDateModified() && summary.GetId() > current.GetId() {
			summaries[summary.GetJobPhaseId()] = summary
		}
	}
	// period index (1..N) of a job phase, from its template phase order.
	periodOf := func(phase *jobphasepb.JobPhase) int {
		tp := templatePhases[strings.TrimSpace(phase.GetTemplatePhaseId())]
		if tp == nil {
			return 0
		}
		order := int(tp.GetPhaseOrder())
		if order < 1 || order > periodSummaryCount {
			return 0
		}
		return order
	}
	jobsOf := func(code string) []*jobpb.Job {
		var jobs []*jobpb.Job
		code = strings.ToLower(strings.TrimSpace(code))
		if code == "" {
			return nil
		}
		for _, job := range card.GetJobs() {
			if job == nil || (!historical && !job.GetActive()) {
				continue
			}
			if job.GetClientId() != "" && strings.TrimSpace(job.GetClientId()) != clientID {
				continue
			}
			// Template category first, job category as fallback — the same
			// precedence as the phase builder (same-origin jobs may carry
			// another category on the job row).
			categoryID := strings.TrimSpace(templateCategory[job.GetJobTemplateId()])
			if categoryID == "" {
				categoryID = strings.TrimSpace(job.GetJobCategoryId())
			}
			if categoryCodeByID[categoryID] == code {
				jobs = append(jobs, job)
			}
		}
		sort.SliceStable(jobs, func(i, j int) bool {
			ni := firstNonEmpty(jobs[i].GetName(), templateName[jobs[i].GetJobTemplateId()])
			nj := firstNonEmpty(jobs[j].GetName(), templateName[jobs[j].GetJobTemplateId()])
			if li, lj := strings.ToLower(ni), strings.ToLower(nj); li != lj {
				return li < lj
			}
			return jobs[i].GetId() < jobs[j].GetId()
		})
		return jobs
	}

	rows := []any{}
	for _, job := range jobsOf(jobCategoryCode) {
		// The projection's job rows may carry no name; the template name is the
		// same fallback the phase builder uses.
		row := map[string]any{"summary_job_name": trimTrailingQualifier(firstNonEmpty(job.GetName(), templateName[job.GetJobTemplateId()]))}
		for n := 1; n <= periodSummaryCount; n++ {
			row["summary_period_"+strconv.Itoa(n)] = ""
		}
		for _, phase := range phasesByJob[job.GetId()] {
			if n := periodOf(phase); n > 0 {
				if summary := summaries[phase.GetId()]; summary != nil && summary.SummaryScore != nil {
					row["summary_period_"+strconv.Itoa(n)] = formatClientReportMaximum(summary.GetSummaryScore())
				}
			}
		}
		rows = append(rows, row)
	}
	out["summary_jobs"] = rows

	taskJobs := jobsOf(taskCategoryCode)
	if len(taskJobs) != 1 {
		return out // none, or ambiguous: leave the activity table empty
	}
	job := taskJobs[0]
	// Activities of this job's own template phases only (codes may repeat in
	// other templates of the projection).
	jobTemplatePhase := map[string]bool{}
	for _, phase := range phasesByJob[job.GetId()] {
		jobTemplatePhase[strings.TrimSpace(phase.GetTemplatePhaseId())] = true
	}
	templateTasks := map[string]string{} // template task id -> code
	taskNames := map[string]string{}     // code -> name
	taskOrder := map[string]int32{}      // code -> step order (first period seen)
	for _, task := range card.GetJobTemplateTasks() {
		if task == nil || (!historical && !task.GetActive()) || !jobTemplatePhase[strings.TrimSpace(task.GetJobTemplatePhaseId())] {
			continue
		}
		code := strings.TrimSpace(task.GetCode())
		if code == "" {
			continue
		}
		templateTasks[task.GetId()] = code
		if _, seen := taskNames[code]; !seen {
			taskNames[code] = strings.TrimSpace(task.GetName())
			taskOrder[code] = task.GetStepOrder()
		}
	}
	values := map[string]map[int]string{} // code -> period -> value
	for _, phase := range phasesByJob[job.GetId()] {
		n := periodOf(phase)
		if n == 0 {
			continue
		}
		if summary := summaries[phase.GetId()]; summary != nil {
			if summary.SummaryScore != nil {
				out["summary_average_period_"+strconv.Itoa(n)] = formatClientReportMaximum(summary.GetSummaryScore())
			}
			out["summary_transmuted_period_"+strconv.Itoa(n)] = strings.TrimSpace(summary.GetScaledLabel())
		}
		for _, task := range card.GetJobTasks() {
			if task == nil || task.GetJobPhaseId() != phase.GetId() || (!historical && !task.GetActive()) {
				continue
			}
			code := templateTasks[task.GetTemplateTaskId()]
			if code == "" {
				continue
			}
			for _, outcome := range card.GetTaskOutcomes() {
				if outcome == nil || outcome.GetJobTaskId() != task.GetId() || outcome.NumericValue == nil {
					continue
				}
				if values[code] == nil {
					values[code] = map[int]string{}
				}
				values[code][n] = formatClientReportMaximum(outcome.GetNumericValue())
				break
			}
		}
	}
	codes := make([]string, 0, len(taskNames))
	for code := range taskNames {
		codes = append(codes, code)
	}
	sort.SliceStable(codes, func(i, j int) bool {
		if taskOrder[codes[i]] != taskOrder[codes[j]] {
			return taskOrder[codes[i]] < taskOrder[codes[j]]
		}
		return codes[i] < codes[j]
	})
	taskRows := []any{}
	for _, code := range codes {
		row := map[string]any{"summary_task_name": taskNames[code]}
		for n := 1; n <= periodSummaryCount; n++ {
			row["summary_task_period_"+strconv.Itoa(n)] = values[code][n]
		}
		taskRows = append(taskRows, row)
	}
	out["summary_tasks"] = taskRows
	return out
}
