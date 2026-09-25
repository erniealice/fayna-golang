package document

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix/criterionlabel"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	categorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplatetaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	phaseoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	templatecriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	ratingdescriptionpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria_rating_description"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

type selectedClientPhaseJob struct {
	job      *jobpb.Job
	template *jobtemplatepb.JobTemplate
	category *categorypb.JobCategory
	phase    *jobphasepb.JobPhase
	tp       *jobtemplatephasepb.JobTemplatePhase
}

// buildClientPhaseReportData maps a typed, single-client projection to the
// generic subscription_group_client_phase_outcome_report_v1 manifest contract.
// It deliberately accepts no vertical labels or category identifiers: labels
// and category names come from the projection, while map keys are manifest keys.
func buildClientPhaseReportData(d *Deps, card *exportpb.ClientReportCardProjection, phaseCode string, printedBy string, printedAt string) (map[string]any, error) {
	if card == nil || card.Context == nil || strings.TrimSpace(card.Context.GetSubscriptionGroupId()) == "" {
		return nil, fmt.Errorf("client phase report projection has no subscription-group context")
	}
	if card.Client == nil || strings.TrimSpace(card.Client.GetClientId()) == "" {
		return nil, fmt.Errorf("client phase report projection has no client identity")
	}
	phaseCode = strings.TrimSpace(phaseCode)
	if phaseCode == "" || strings.EqualFold(phaseCode, clientDocumentPeriodYearFinal) {
		return nil, fmt.Errorf("unsupported client report phase")
	}
	phaseCatalog := clientProjectionPhaseCatalog(card)
	var selectedPhaseName string
	for _, phase := range phaseCatalog {
		if strings.EqualFold(phase.Code, phaseCode) {
			selectedPhaseName = phase.Name
			break
		}
	}
	if selectedPhaseName == "" {
		return nil, fmt.Errorf("client has no active job with the requested report phase")
	}
	clientID := strings.TrimSpace(card.Client.GetClientId())
	historical := card.GetContext().GetHistorical()

	templates := make(map[string]*jobtemplatepb.JobTemplate, len(card.JobTemplates))
	for _, template := range card.JobTemplates {
		if template == nil || strings.TrimSpace(template.GetId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains a malformed job template")
		}
		if _, exists := templates[template.GetId()]; exists {
			return nil, fmt.Errorf("client phase report projection contains duplicate job templates")
		}
		templates[template.GetId()] = template
	}

	categories := make(map[string]*categorypb.JobCategory, len(card.JobCategories))
	for _, category := range card.JobCategories {
		if category == nil || strings.TrimSpace(category.GetId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains a malformed job category")
		}
		if _, exists := categories[category.GetId()]; exists {
			return nil, fmt.Errorf("client phase report projection contains duplicate job categories")
		}
		categories[category.GetId()] = category
	}

	templatePhases := make(map[string]*jobtemplatephasepb.JobTemplatePhase, len(card.JobTemplatePhases))
	phaseByTemplate := make(map[string][]*jobtemplatephasepb.JobTemplatePhase)
	for _, phase := range card.JobTemplatePhases {
		if phase == nil || strings.TrimSpace(phase.GetId()) == "" || strings.TrimSpace(phase.GetJobTemplateId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains a malformed template phase")
		}
		if _, exists := templatePhases[phase.GetId()]; exists {
			return nil, fmt.Errorf("client phase report projection contains duplicate template phases")
		}
		templatePhases[phase.GetId()] = phase
		phaseByTemplate[phase.GetJobTemplateId()] = append(phaseByTemplate[phase.GetJobTemplateId()], phase)
	}
	variantLabels := make(map[string]string, len(card.GetProductVariants()))
	for _, variant := range card.GetProductVariants() {
		if variant != nil && strings.TrimSpace(variant.GetId()) != "" {
			variantLabels[variant.GetId()] = strings.TrimSpace(variant.GetSku())
		}
	}

	criteriaByTemplateTask := make(map[string][]*templatecriteriapb.TemplateTaskCriteria)
	criteriaByID := make(map[string]*criteriapb.OutcomeCriteria)
	for _, criterion := range card.OutcomeCriteria {
		if criterion == nil || strings.TrimSpace(criterion.GetId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains malformed outcome criteria")
		}
		if _, exists := criteriaByID[criterion.GetId()]; exists {
			return nil, fmt.Errorf("client phase report projection contains duplicate outcome criteria")
		}
		criteriaByID[criterion.GetId()] = criterion
	}
	for _, link := range card.TemplateTaskCriteria {
		if link == nil || strings.TrimSpace(link.GetId()) == "" || strings.TrimSpace(link.GetJobTemplateTaskId()) == "" || strings.TrimSpace(link.GetOutcomeCriteriaId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains malformed template-task criteria")
		}
		if !historical && !link.GetActive() {
			continue
		}
		if _, exists := criteriaByID[link.GetOutcomeCriteriaId()]; !exists {
			if historical {
				continue
			}
			return nil, fmt.Errorf("client phase report projection references missing outcome criteria")
		}
		criteriaByTemplateTask[link.GetJobTemplateTaskId()] = append(criteriaByTemplateTask[link.GetJobTemplateTaskId()], link)
	}

	templateTasksByPhase := make(map[string][]*jobtemplatetaskpb.JobTemplateTask)
	templateTasksByID := make(map[string]*jobtemplatetaskpb.JobTemplateTask, len(card.JobTemplateTasks))
	for _, task := range card.JobTemplateTasks {
		if task == nil || strings.TrimSpace(task.GetId()) == "" || strings.TrimSpace(task.GetJobTemplatePhaseId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains a malformed template task")
		}
		if historical || task.GetActive() {
			templateTasksByPhase[task.GetJobTemplatePhaseId()] = append(templateTasksByPhase[task.GetJobTemplatePhaseId()], task)
		}
		templateTasksByID[task.GetId()] = task
	}
	for phaseID := range templateTasksByPhase {
		sort.SliceStable(templateTasksByPhase[phaseID], func(i, j int) bool {
			a, b := templateTasksByPhase[phaseID][i], templateTasksByPhase[phaseID][j]
			if a.GetStepOrder() != b.GetStepOrder() {
				return a.GetStepOrder() < b.GetStepOrder()
			}
			if a.GetName() != b.GetName() {
				return a.GetName() < b.GetName()
			}
			return a.GetId() < b.GetId()
		})
	}
	for taskID := range criteriaByTemplateTask {
		sort.SliceStable(criteriaByTemplateTask[taskID], func(i, j int) bool {
			a, b := criteriaByTemplateTask[taskID][i], criteriaByTemplateTask[taskID][j]
			if a.GetSequenceOrder() != b.GetSequenceOrder() {
				return a.GetSequenceOrder() < b.GetSequenceOrder()
			}
			return a.GetId() < b.GetId()
		})
	}
	ratingDescriptionsByLink := make(map[string][]*ratingdescriptionpb.TemplateTaskCriteriaRatingDescription)
	for _, description := range card.GetRatingDescriptions() {
		if description == nil || strings.TrimSpace(description.GetTemplateTaskCriteriaId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains malformed rating descriptions")
		}
		if historical || description.GetActive() {
			ratingDescriptionsByLink[description.GetTemplateTaskCriteriaId()] = append(ratingDescriptionsByLink[description.GetTemplateTaskCriteriaId()], description)
		}
	}
	for linkID := range ratingDescriptionsByLink {
		sort.SliceStable(ratingDescriptionsByLink[linkID], func(i, j int) bool {
			left, right := ratingDescriptionsByLink[linkID][i], ratingDescriptionsByLink[linkID][j]
			if left.GetSequenceOrder() != right.GetSequenceOrder() {
				return left.GetSequenceOrder() < right.GetSequenceOrder()
			}
			return left.GetId() < right.GetId()
		})
	}

	jobPhasesByID := make(map[string]*jobphasepb.JobPhase, len(card.JobPhases))
	jobPhasesByJob := make(map[string][]*jobphasepb.JobPhase)
	for _, phase := range card.JobPhases {
		if phase == nil || strings.TrimSpace(phase.GetId()) == "" || strings.TrimSpace(phase.GetJobId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains a malformed job phase")
		}
		if _, exists := jobPhasesByID[phase.GetId()]; exists {
			return nil, fmt.Errorf("client phase report projection contains duplicate job phases")
		}
		jobPhasesByID[phase.GetId()] = phase
		jobPhasesByJob[phase.GetJobId()] = append(jobPhasesByJob[phase.GetJobId()], phase)
	}

	phaseSummaries := make(map[string]*phaseoutcomepb.PhaseOutcomeSummary)
	for _, summary := range card.PhaseOutcomeSummaries {
		if summary == nil || strings.TrimSpace(summary.GetJobPhaseId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains a malformed phase summary")
		}
		if historical || summary.GetActive() {
			if current := phaseSummaries[summary.GetJobPhaseId()]; current == nil || summary.GetDateModified() > current.GetDateModified() || summary.GetDateModified() == current.GetDateModified() && summary.GetId() > current.GetId() {
				phaseSummaries[summary.GetJobPhaseId()] = summary
			}
		}
	}

	jobTasksByJobPhase := make(map[string][]*jobtaskpb.JobTask)
	jobTasksByID := make(map[string]*jobtaskpb.JobTask)
	for _, task := range card.JobTasks {
		if task == nil || strings.TrimSpace(task.GetId()) == "" || strings.TrimSpace(task.GetJobPhaseId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains a malformed job task")
		}
		if _, exists := jobTasksByID[task.GetId()]; exists {
			return nil, fmt.Errorf("client phase report projection contains duplicate job tasks")
		}
		jobTasksByID[task.GetId()] = task
		if historical || task.GetActive() {
			jobTasksByJobPhase[task.GetJobPhaseId()] = append(jobTasksByJobPhase[task.GetJobPhaseId()], task)
		}
	}
	for phaseID := range jobTasksByJobPhase {
		sort.SliceStable(jobTasksByJobPhase[phaseID], func(i, j int) bool {
			a, b := jobTasksByJobPhase[phaseID][i], jobTasksByJobPhase[phaseID][j]
			if a.GetStepOrder() != b.GetStepOrder() {
				return a.GetStepOrder() < b.GetStepOrder()
			}
			if a.GetName() != b.GetName() {
				return a.GetName() < b.GetName()
			}
			return a.GetId() < b.GetId()
		})
	}

	taskOutcomes := make(map[string]*exportpb.ClientReportCardTaskOutcome)
	for _, outcome := range card.TaskOutcomes {
		if outcome == nil || strings.TrimSpace(outcome.GetJobTaskId()) == "" || strings.TrimSpace(outcome.GetTemplateTaskCriteriaId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains a malformed task outcome")
		}
		key := outcome.GetJobTaskId() + "\x00" + outcome.GetTemplateTaskCriteriaId()
		if current := taskOutcomes[key]; current == nil || outcome.GetRecordedDate() > current.GetRecordedDate() {
			taskOutcomes[key] = outcome
		}
	}
	for _, outcome := range card.TaskOutcomes {
		if outcome != nil {
			if _, ok := jobTasksByID[outcome.GetJobTaskId()]; !ok {
				return nil, fmt.Errorf("client phase report projection outcome references missing job task")
			}
		}
	}

	selected := make([]selectedClientPhaseJob, 0, len(card.Jobs))
	seenJobs := make(map[string]struct{}, len(card.Jobs))
	for _, job := range card.Jobs {
		if job == nil || strings.TrimSpace(job.GetId()) == "" {
			return nil, fmt.Errorf("client phase report projection contains a malformed job")
		}
		if _, exists := seenJobs[job.GetId()]; exists {
			return nil, fmt.Errorf("client phase report projection contains duplicate jobs")
		}
		seenJobs[job.GetId()] = struct{}{}
		if job.GetClientId() != "" && strings.TrimSpace(job.GetClientId()) != clientID {
			return nil, fmt.Errorf("client phase report projection contains a job for a different client")
		}
		if !historical && !job.GetActive() {
			continue
		}
		templateID := strings.TrimSpace(job.GetJobTemplateId())
		template := templates[templateID]
		if template == nil || (!historical && !template.GetActive()) {
			return nil, fmt.Errorf("client phase report projection job has no active template")
		}
		categoryID := strings.TrimSpace(template.GetJobCategoryId())
		if categoryID == "" {
			categoryID = strings.TrimSpace(job.GetJobCategoryId())
		}
		category := categories[categoryID]
		if category == nil {
			return nil, fmt.Errorf("client phase report projection job has no category")
		}
		var selectedTemplatePhase *jobtemplatephasepb.JobTemplatePhase
		for _, phase := range phaseByTemplate[templateID] {
			if (historical || phase.GetActive()) && strings.EqualFold(strings.TrimSpace(phase.GetCode()), phaseCode) {
				if selectedTemplatePhase != nil {
					return nil, fmt.Errorf("client phase report projection has an ambiguous active template phase")
				}
				selectedTemplatePhase = phase
			}
		}
		if selectedTemplatePhase == nil {
			continue
		}
		var selectedJobPhase *jobphasepb.JobPhase
		for _, phase := range jobPhasesByJob[job.GetId()] {
			if (historical || phase.GetActive()) && strings.TrimSpace(phase.GetTemplatePhaseId()) == selectedTemplatePhase.GetId() {
				if selectedJobPhase != nil {
					return nil, fmt.Errorf("client phase report projection has duplicate active job phases")
				}
				selectedJobPhase = phase
			}
		}
		if selectedJobPhase == nil {
			continue
		}
		selected = append(selected, selectedClientPhaseJob{job: job, template: template, category: category, phase: selectedJobPhase, tp: selectedTemplatePhase})
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("client has no active job with the requested report phase")
	}
	sort.SliceStable(selected, func(i, j int) bool {
		a, b := selected[i], selected[j]
		if categoryOrder(a.category) != categoryOrder(b.category) {
			return categoryOrder(a.category) < categoryOrder(b.category)
		}
		if a.template.GetName() != b.template.GetName() {
			return a.template.GetName() < b.template.GetName()
		}
		return a.job.GetId() < b.job.GetId()
	})

	jobs := make([]any, 0, len(selected))
	phaseName := selectedPhaseName
	outcomeSections := buildProjectedOutcomeSections(card, historical)
	outcomeCells, outcomeTotals, categoryCells, categoryTotals := buildProjectedOutcomeCellIndex(card, historical)
	clientReferenceCode := ""
	groupCategoryCode := ""
	if d != nil {
		clientReferenceCode = strings.TrimSpace(d.DocOptions.ClientReferenceAttributeCode)
		groupCategoryCode = strings.TrimSpace(d.DocOptions.GroupCategoryFilter)
	}
	clientReference := ""
	if clientReferenceCode != "" {
		for _, attribute := range card.GetAttributes() {
			if attribute != nil && strings.EqualFold(strings.TrimSpace(attribute.GetCode()), clientReferenceCode) {
				clientReference = strings.TrimSpace(attribute.GetValue())
				break
			}
		}
	}
	criterionPrefix, notAssessed, maxStaffNames := "", "", 0
	if d != nil {
		criterionPrefix = d.Labels.ClientDocument.CriterionPrefix
		notAssessed = strings.TrimSpace(d.Labels.ClientDocument.NotAssessed)
		maxStaffNames = d.DocOptions.MaxStaffNames
	}
	loopCategories := map[string]bool{}
	if d != nil {
		for _, code := range d.DocOptions.PhaseJobCategoryCodes {
			if code = strings.TrimSpace(code); code != "" {
				loopCategories[strings.ToLower(code)] = true
			}
		}
	}
	groupLeads := make([]string, 0)
	for _, entry := range selected {
		if groupCategoryCode != "" && strings.EqualFold(strings.TrimSpace(entry.category.GetCode()), groupCategoryCode) {
			groupLeads = append(groupLeads, strings.Split(staffNames(card, entry.job.GetId(), entry.phase.GetId()), ", ")...)
		}
		if len(loopCategories) > 0 && !loopCategories[strings.ToLower(strings.TrimSpace(entry.category.GetCode()))] {
			continue
		}
		phase := entry.phase
		summary := phaseSummaries[phase.GetId()]
		jobTaskByTemplateTask := make(map[string]*jobtaskpb.JobTask)
		for _, task := range jobTasksByJobPhase[phase.GetId()] {
			if templateTaskID := strings.TrimSpace(task.GetTemplateTaskId()); templateTaskID != "" {
				jobTaskByTemplateTask[templateTaskID] = task
			}
		}
		assessments := make([]any, 0)
		var phaseTotal, phaseMaximum float64
		hasPhaseNumericOutcome := false
		for _, templateTask := range templateTasksByPhase[entry.tp.GetId()] {
			jobTask := jobTaskByTemplateTask[templateTask.GetId()]
			for criterionIndex, link := range criteriaByTemplateTask[templateTask.GetId()] {
				criterion := criteriaByID[link.GetOutcomeCriteriaId()]
				assessment := map[string]any{
					// Letter by position in the activity (owner 2026-09-24); never
					// a line break inside the document.
					"assessment_name":    criterionlabel.Label(strings.TrimSpace(criterion.GetName()), criterionIndex, criterionPrefix, ""),
					"achievement_level":  "",
					"maximum":            "",
					"assessment_maximum": "",
					"comment":            "",
				}
				if criterion.MaxScore != nil {
					assessment["maximum"] = formatClientReportMaximum(float64(criterion.GetMaxScore()))
					assessment["assessment_maximum"] = assessment["maximum"]
					phaseMaximum += float64(criterion.GetMaxScore())
				}
				descriptions := make([]any, 0, len(ratingDescriptionsByLink[link.GetId()]))
				for _, description := range ratingDescriptionsByLink[link.GetId()] {
					descriptor := map[string]any{
						"rating_label": "",
						"minimum":      "",
						"maximum":      "",
						"output_value": "",
						"description":  strings.TrimSpace(description.GetDescription()),
					}
					if band := description.GetScoreScaleBand(); band != nil {
						descriptor["rating_label"] = strings.TrimSpace(band.GetOutputLabel())
						if band.InputMin != nil {
							descriptor["minimum"] = strconv.FormatFloat(band.GetInputMin(), 'f', -1, 64)
						}
						if band.InputMax != nil {
							descriptor["maximum"] = strconv.FormatFloat(band.GetInputMax(), 'f', -1, 64)
						}
						if band.OutputValue != nil {
							descriptor["output_value"] = strconv.FormatFloat(band.GetOutputValue(), 'f', -1, 64)
						}
					}
					descriptions = append(descriptions, descriptor)
				}
				assessment["rating_descriptions"] = descriptions
				if jobTask != nil {
					if outcome := taskOutcomes[jobTask.GetId()+"\x00"+link.GetId()]; outcome != nil {
						assessment["achievement_level"] = taskOutcomeMark(outcome)
						assessment["comment"] = strings.TrimSpace(outcome.GetDeterminationNote())
						if outcome.NumericValue != nil {
							phaseTotal += outcome.GetNumericValue()
							hasPhaseNumericOutcome = true
						}
					}
				}
				// An empty determination note reads "Not yet assessed", whether or
				// not a level was recorded (owner 2026-09-25).
				if notAssessed != "" && assessment["comment"] == "" {
					assessment["comment"] = notAssessed
				}
				assessments = append(assessments, assessment)
			}
		}
		phaseGrade, phaseComment := "", ""
		if summary != nil {
			if summary.ScaledLabel != nil {
				phaseGrade = strings.TrimSpace(summary.GetScaledLabel())
			} else if summary.ScaledScore != nil {
				phaseGrade = strconv.FormatFloat(summary.GetScaledScore(), 'f', -1, 64)
			}
			phaseComment = strings.TrimSpace(summary.GetNarrative())
		}
		progressTotal, progressMaximum, hasProgressNumericOutcome := projectedProgressToDate(entry, jobPhasesByJob, jobTasksByJobPhase, templatePhases, templateTasksByID, criteriaByTemplateTask, criteriaByID, taskOutcomes, historical)
		jobName := composeJobVariantName(
			firstNonEmpty(strings.TrimSpace(entry.job.GetName()), strings.TrimSpace(entry.template.GetName())),
			variantLabels[strings.TrimSpace(entry.tp.GetOutputProductVariantId())],
		)
		jobs = append(jobs, map[string]any{
			"job_name":                 jobName,
			"job_category_name":        strings.TrimSpace(entry.category.GetName()),
			"staff_name":               capNames(staffNames(card, entry.job.GetId(), phase.GetId()), maxStaffNames),
			"phase_grade":              phaseGrade,
			"phase_comment":            phaseComment,
			"phase_total":              "",
			"phase_maximum":            formatClientReportMaximum(phaseMaximum),
			"progress_to_date_total":   "",
			"progress_to_date_maximum": formatClientReportMaximum(progressMaximum),
			"assessments":              assessments,
		})
		if hasPhaseNumericOutcome {
			jobs[len(jobs)-1].(map[string]any)["phase_total"] = formatClientReportMaximum(phaseTotal)
		}
		if hasProgressNumericOutcome {
			jobs[len(jobs)-1].(map[string]any)["progress_to_date_total"] = formatClientReportMaximum(progressTotal)
		}
	}

	studentName := strings.TrimSpace(card.Client.GetName())
	if studentName == "" {
		studentName = strings.TrimSpace(strings.TrimSpace(card.Client.GetLastName()) + ", " + strings.TrimSpace(card.Client.GetFirstName()))
		studentName = strings.TrimSuffix(studentName, ",")
	}
	groupName := strings.TrimSpace(card.Context.GetSubscriptionGroupName())
	grade, sectionName := gradeSection(groupName)
	if sectionName == "" {
		sectionName = groupName
	}
	sectionName = trimTrailingQualifier(sectionName)
	adviser := strings.Join(sortedUniqueStrings(groupLeads), " / ")
	headerName := ""
	if d != nil {
		headerName = d.DocumentHeaderName
	}
	academicYear := strings.TrimSpace(card.Context.GetPriceScheduleName())
	planLabel := ""
	if d != nil {
		if code := strings.TrimSpace(d.DocOptions.PlanLabelAttributeCode); code != "" {
			for _, attribute := range card.GetPlanAttributes() {
				if attribute != nil && strings.EqualFold(strings.TrimSpace(attribute.GetCode()), code) {
					planLabel = strings.TrimSpace(attribute.GetValue())
					break
				}
			}
		}
	}
	// Per-job copies of the identity so a heading repeated on every job's page
	// can print it (loop items do not fall back to root values).
	for _, raw := range jobs {
		job := raw.(map[string]any)
		job["page_student_name"] = studentName
		job["page_grade_level"] = grade
		job["page_section_name"] = sectionName
		job["page_academic_year"] = academicYear
		job["page_client_reference"] = clientReference
		job["page_adviser"] = adviser
		job["page_plan_label"] = planLabel
	}
	data := map[string]any{
		"school_name":      strings.TrimSpace(headerName),
		"academic_year":    academicYear,
		"student_name":     studentName,
		"grade_level":      grade,
		"section_name":     sectionName,
		"client_reference": clientReference,
		"adviser":          adviser,
		"printed_by":       strings.TrimSpace(printedBy),
		"printed_at":       strings.TrimSpace(printedAt),
		"phase_name":       phaseName,
		"jobs":             jobs,
		"outcome_sections": outcomeSections,
		"outcome_cells":    outcomeCells,
		"outcome_totals":   outcomeTotals,
		"category_cells":   categoryCells,
		"category_totals":  categoryTotals,
	}
	summaryJobCategory, summaryTaskCategory := "", ""
	if d != nil {
		summaryJobCategory, summaryTaskCategory = d.DocOptions.PeriodSummaryJobCategoryCode, d.DocOptions.PeriodSummaryTaskCategoryCode
	}
	for key, value := range buildPeriodSummaries(card, summaryJobCategory, summaryTaskCategory, historical) {
		data[key] = value
	}
	return data, nil
}

// buildProjectedOutcomeCellIndex exposes exact code-keyed outcome paths for
// static template cells. It uses only identifiers present in the typed client
// projection and keeps category, job-template, criterion, phase, and task
// identity in each path so unrelated projected jobs cannot collide silently.
// Path-sensitive code segments are reversibly escaped before they become map
// keys; template token paths use the same escaping convention. If distinct
// source records map to the same full code tuple, that tuple (and its aggregate
// total) is omitted as ambiguous. The repeated outcome_sections representation
// remains lossless.
//
// It also returns a template-independent family keyed by category, criterion
// and activity code only (category_cells / category_totals), so one document
// template serves every grade's template of a category (e.g. the same month
// activities across per-grade attendance templates). The same duplicate and
// ambiguity rules apply at that grain: two different jobs of one category
// yielding the same category/criterion/activity path omit the cell and total.
func buildProjectedOutcomeCellIndex(card *exportpb.ClientReportCardProjection, historical bool) (map[string]any, map[string]any, map[string]any, map[string]any) {
	cellsRoot := map[string]any{}
	totalsRoot := map[string]any{}
	categoryCellsRoot := map[string]any{}
	categoryTotalsRoot := map[string]any{}
	if card == nil || card.GetClient() == nil || strings.TrimSpace(card.GetClient().GetClientId()) == "" {
		return cellsRoot, totalsRoot, categoryCellsRoot, categoryTotalsRoot
	}
	clientID := strings.TrimSpace(card.GetClient().GetClientId())
	categories := map[string]*categorypb.JobCategory{}
	for _, category := range card.GetJobCategories() {
		if category != nil && strings.TrimSpace(category.GetId()) != "" {
			categories[category.GetId()] = category
		}
	}
	templates := map[string]*jobtemplatepb.JobTemplate{}
	for _, template := range card.GetJobTemplates() {
		if template != nil && strings.TrimSpace(template.GetId()) != "" {
			templates[template.GetId()] = template
		}
	}
	phases := map[string]*jobtemplatephasepb.JobTemplatePhase{}
	for _, phase := range card.GetJobTemplatePhases() {
		if phase != nil && strings.TrimSpace(phase.GetId()) != "" && (historical || phase.GetActive()) {
			phases[phase.GetId()] = phase
		}
	}
	tasks := map[string]*jobtemplatetaskpb.JobTemplateTask{}
	for _, task := range card.GetJobTemplateTasks() {
		if task != nil && strings.TrimSpace(task.GetId()) != "" && (historical || task.GetActive()) {
			tasks[task.GetId()] = task
		}
	}
	criteria := map[string]*criteriapb.OutcomeCriteria{}
	for _, criterion := range card.GetOutcomeCriteria() {
		if criterion != nil && strings.TrimSpace(criterion.GetId()) != "" && (historical || criterion.GetActive()) {
			criteria[criterion.GetId()] = criterion
		}
	}
	linksByTask := map[string][]*templatecriteriapb.TemplateTaskCriteria{}
	for _, link := range card.GetTemplateTaskCriteria() {
		if link != nil && strings.TrimSpace(link.GetId()) != "" && (historical || link.GetActive()) && criteria[link.GetOutcomeCriteriaId()] != nil {
			linksByTask[link.GetJobTemplateTaskId()] = append(linksByTask[link.GetJobTemplateTaskId()], link)
		}
	}
	jobTasks := map[string]*jobtaskpb.JobTask{}
	for _, task := range card.GetJobTasks() {
		if task != nil && strings.TrimSpace(task.GetId()) != "" && (historical || task.GetActive()) {
			jobTasks[task.GetId()] = task
		}
	}
	jobPhasesByJob := map[string][]*jobphasepb.JobPhase{}
	for _, phase := range card.GetJobPhases() {
		if phase != nil && strings.TrimSpace(phase.GetId()) != "" && strings.TrimSpace(phase.GetJobId()) != "" && (historical || phase.GetActive()) {
			jobPhasesByJob[phase.GetJobId()] = append(jobPhasesByJob[phase.GetJobId()], phase)
		}
	}
	latestOutcomes := map[string]*exportpb.ClientReportCardTaskOutcome{}
	for _, outcome := range card.GetTaskOutcomes() {
		if outcome == nil || jobTasks[outcome.GetJobTaskId()] == nil || strings.TrimSpace(outcome.GetTemplateTaskCriteriaId()) == "" {
			continue
		}
		key := outcome.GetJobTaskId() + "\x00" + outcome.GetTemplateTaskCriteriaId()
		current := latestOutcomes[key]
		if current == nil || outcome.GetRecordedDate() > current.GetRecordedDate() {
			latestOutcomes[key] = outcome
		}
	}
	type outcomeCellCandidate struct {
		path      []string
		owner     string
		value     string
		numeric   string
		numericOK bool
	}
	type outcomeTotalCandidate struct {
		owner   string
		value   float64
		hasData bool
	}
	candidates := map[string]outcomeCellCandidate{}
	ambiguousCells := map[string]bool{}
	totalCandidates := map[string]outcomeTotalCandidate{}
	ambiguousTotals := map[string]bool{}
	seenContributions := map[string]bool{}
	categoryCandidates := map[string]outcomeCellCandidate{}
	ambiguousCategoryCells := map[string]bool{}
	categoryTotalCandidates := map[string]outcomeTotalCandidate{}
	ambiguousCategoryTotals := map[string]bool{}
	seenCategoryContributions := map[string]bool{}
	for _, job := range card.GetJobs() {
		if job == nil || strings.TrimSpace(job.GetId()) == "" || !historical && !job.GetActive() {
			continue
		}
		if job.GetClientId() != "" && strings.TrimSpace(job.GetClientId()) != clientID {
			continue
		}
		template := templates[job.GetJobTemplateId()]
		if template == nil {
			continue
		}
		categoryID := strings.TrimSpace(job.GetJobCategoryId())
		if categoryID == "" {
			categoryID = strings.TrimSpace(template.GetJobCategoryId())
		}
		category := categories[categoryID]
		if category == nil {
			continue
		}
		categoryCode := strings.TrimSpace(category.GetCode())
		templateCode := strings.TrimSpace(template.GetTemplateCode())
		if !clientReportPathCode(categoryCode) || !clientReportPathCode(templateCode) {
			continue
		}
		jobID := strings.TrimSpace(job.GetId())
		for _, jobPhase := range jobPhasesByJob[jobID] {
			if jobPhase == nil {
				continue
			}
			templatePhase := phases[jobPhase.GetTemplatePhaseId()]
			if templatePhase == nil || strings.TrimSpace(templatePhase.GetJobTemplateId()) != strings.TrimSpace(template.GetId()) {
				continue
			}
			phaseCode := strings.TrimSpace(templatePhase.GetCode())
			if !clientReportPathCode(phaseCode) {
				continue
			}
			for _, jobTask := range card.GetJobTasks() {
				if jobTask == nil || jobTask.GetJobPhaseId() != jobPhase.GetId() || !historical && !jobTask.GetActive() {
					continue
				}
				templateTask := tasks[jobTask.GetTemplateTaskId()]
				if templateTask == nil || strings.TrimSpace(templateTask.GetJobTemplatePhaseId()) != strings.TrimSpace(templatePhase.GetId()) {
					continue
				}
				taskCode := strings.TrimSpace(templateTask.GetCode())
				if !clientReportPathCode(taskCode) {
					continue
				}
				for _, link := range linksByTask[templateTask.GetId()] {
					criterion := criteria[link.GetOutcomeCriteriaId()]
					criterionCode := ""
					if criterion != nil {
						criterionCode = strings.TrimSpace(criterion.GetCode())
					}
					if !clientReportPathCode(criterionCode) {
						continue
					}
					outcome := latestOutcomes[jobTask.GetId()+"\x00"+link.GetId()]
					if outcome == nil {
						continue
					}
					path := []string{
						clientReportPathSegment(categoryCode),
						clientReportPathSegment(templateCode),
						clientReportPathSegment(criterionCode),
						clientReportPathSegment(phaseCode),
						clientReportPathSegment(taskCode),
					}
					key := strings.Join(path, "\x00")
					owner := jobID + "\x00" + jobTask.GetId() + "\x00" + link.GetId()
					cell := outcomeCellCandidate{path: path, owner: owner, value: taskOutcomeMark(outcome)}
					if outcome.NumericValue != nil {
						cell.numeric = strconv.FormatFloat(outcome.GetNumericValue(), 'f', -1, 64)
						cell.numericOK = true
					}
					categoryPath := []string{path[0], path[2], path[4]}
					categoryKey := strings.Join(categoryPath, "\x00")
					categoryCell := outcomeCellCandidate{path: categoryPath, owner: owner, value: cell.value, numeric: cell.numeric, numericOK: cell.numericOK}
					if existing, exists := categoryCandidates[categoryKey]; exists && existing.owner != owner {
						ambiguousCategoryCells[categoryKey] = true
						ambiguousCategoryTotals[path[0]+"\x00"+path[2]] = true
					} else {
						categoryCandidates[categoryKey] = categoryCell
					}
					categoryTotalKey := path[0] + "\x00" + path[2]
					if !seenCategoryContributions[categoryTotalKey+"\x00"+owner] {
						seenCategoryContributions[categoryTotalKey+"\x00"+owner] = true
						categoryTotal := categoryTotalCandidates[categoryTotalKey]
						if categoryTotal.owner != "" && categoryTotal.owner != jobID {
							ambiguousCategoryTotals[categoryTotalKey] = true
						} else {
							categoryTotal.owner = jobID
							if outcome.NumericValue != nil {
								categoryTotal.value += outcome.GetNumericValue()
								categoryTotal.hasData = true
							}
							categoryTotalCandidates[categoryTotalKey] = categoryTotal
						}
					}
					duplicateCellOwner := false
					if existing, exists := candidates[key]; exists && existing.owner != owner {
						duplicateCellOwner = true
						if existing.value != cell.value || existing.numericOK != cell.numericOK || existing.numeric != cell.numeric {
							ambiguousCells[key] = true
						}
					} else {
						candidates[key] = cell
					}
					totalPath := []string{
						clientReportPathSegment(categoryCode),
						clientReportPathSegment(templateCode),
						clientReportPathSegment(criterionCode),
					}
					totalKey := strings.Join(totalPath, "\x00")
					if duplicateCellOwner {
						ambiguousTotals[totalKey] = true
					}
					contributionKey := totalKey + "\x00" + owner
					if seenContributions[contributionKey] {
						continue
					}
					seenContributions[contributionKey] = true
					totalOwner := jobID
					total := totalCandidates[totalKey]
					if total.owner != "" && total.owner != totalOwner {
						ambiguousTotals[totalKey] = true
					} else {
						total.owner = totalOwner
						if outcome.NumericValue != nil {
							total.value += outcome.GetNumericValue()
							total.hasData = true
						}
						totalCandidates[totalKey] = total
					}
				}
			}
		}
	}
	for key, candidate := range candidates {
		if ambiguousCells[key] {
			continue
		}
		leaf := map[string]any{"value": candidate.value}
		if candidate.numericOK {
			leaf["numeric_value"] = candidate.numeric
		}
		setClientReportCodeLeaf(cellsRoot, candidate.path, leaf)
	}
	for key, total := range totalCandidates {
		if !total.hasData || ambiguousTotals[key] {
			continue
		}
		setClientReportCodeScalar(totalsRoot, append(strings.Split(key, "\x00"), "numeric_value"), formatClientReportMaximum(total.value))
	}
	for key, candidate := range categoryCandidates {
		if ambiguousCategoryCells[key] {
			continue
		}
		leaf := map[string]any{"value": candidate.value}
		if candidate.numericOK {
			leaf["numeric_value"] = candidate.numeric
		}
		setClientReportCodeLeaf(categoryCellsRoot, candidate.path, leaf)
	}
	for key, total := range categoryTotalCandidates {
		if !total.hasData || ambiguousCategoryTotals[key] {
			continue
		}
		setClientReportCodeScalar(categoryTotalsRoot, append(strings.Split(key, "\x00"), "numeric_value"), formatClientReportMaximum(total.value))
	}
	return cellsRoot, totalsRoot, categoryCellsRoot, categoryTotalsRoot
}

func clientReportPathCode(code string) bool {
	return strings.TrimSpace(code) != ""
}

// clientReportPathSegment escapes the path delimiter and its escape marker,
// while preserving common code values as-is for readable template tokens.
func clientReportPathSegment(code string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(code), "~", "~0"), ".", "~1")
}

func setClientReportCodeLeaf(root map[string]any, path []string, leaf map[string]any) {
	current := root
	for _, segment := range path {
		next, ok := current[segment].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[segment] = next
		}
		current = next
	}
	for key, value := range leaf {
		current[key] = value
	}
}

func setClientReportCodeScalar(root map[string]any, path []string, value any) {
	current := root
	for _, segment := range path[:len(path)-1] {
		next, ok := current[segment].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[segment] = next
		}
		current = next
	}
	current[path[len(path)-1]] = value
}

// buildProjectedOutcomeSections produces a generic category/criterion
// matrix from the selected client's projected evidence. Cells span every
// eligible projected job phase, independently of the requested phase used by
// the primary jobs[] table. Period identifiers, names, and order come from
// job_template_phase; row and band identifiers/names come from the
// projected category and outcome criterion. Each cell carries its template
// task code/name as well as its phase code/name, preserving configured
// phase/task pairs without assigning calendar meaning in generic Fayna.
//
// Missing outcomes create blank placeholder cells so configured task and
// criterion structure remains visible. A numeric total is emitted only when
// at least one numeric outcome exists; recorded zero remains "0".
func buildProjectedOutcomeSections(card *exportpb.ClientReportCardProjection, historical bool) []any {
	if card == nil {
		return nil
	}
	categories := map[string]*categorypb.JobCategory{}
	for _, category := range card.GetJobCategories() {
		if category != nil && strings.TrimSpace(category.GetId()) != "" {
			categories[category.GetId()] = category
		}
	}
	templates := map[string]*jobtemplatepb.JobTemplate{}
	for _, template := range card.GetJobTemplates() {
		if template != nil && strings.TrimSpace(template.GetId()) != "" {
			templates[template.GetId()] = template
		}
	}
	templatePhases := map[string]*jobtemplatephasepb.JobTemplatePhase{}
	for _, phase := range card.GetJobTemplatePhases() {
		if phase != nil && strings.TrimSpace(phase.GetId()) != "" && (historical || phase.GetActive()) {
			templatePhases[phase.GetId()] = phase
		}
	}
	templateTasks := map[string]*jobtemplatetaskpb.JobTemplateTask{}
	for _, task := range card.GetJobTemplateTasks() {
		if task != nil && strings.TrimSpace(task.GetId()) != "" && (historical || task.GetActive()) {
			templateTasks[task.GetId()] = task
		}
	}
	criteria := map[string]*criteriapb.OutcomeCriteria{}
	for _, criterion := range card.GetOutcomeCriteria() {
		if criterion != nil && strings.TrimSpace(criterion.GetId()) != "" && (historical || criterion.GetActive()) {
			criteria[criterion.GetId()] = criterion
		}
	}
	linksByTask := map[string][]*templatecriteriapb.TemplateTaskCriteria{}
	for _, link := range card.GetTemplateTaskCriteria() {
		if link == nil || !historical && !link.GetActive() || criteria[link.GetOutcomeCriteriaId()] == nil {
			continue
		}
		linksByTask[link.GetJobTemplateTaskId()] = append(linksByTask[link.GetJobTemplateTaskId()], link)
	}
	for taskID := range linksByTask {
		sort.SliceStable(linksByTask[taskID], func(i, j int) bool {
			left, right := linksByTask[taskID][i], linksByTask[taskID][j]
			if left.GetSequenceOrder() != right.GetSequenceOrder() {
				return left.GetSequenceOrder() < right.GetSequenceOrder()
			}
			return left.GetId() < right.GetId()
		})
	}
	phasesByJob := map[string][]*jobphasepb.JobPhase{}
	for _, phase := range card.GetJobPhases() {
		if phase != nil && strings.TrimSpace(phase.GetId()) != "" && (historical || phase.GetActive()) && templatePhases[phase.GetTemplatePhaseId()] != nil {
			phasesByJob[phase.GetJobId()] = append(phasesByJob[phase.GetJobId()], phase)
		}
	}
	for jobID := range phasesByJob {
		sort.SliceStable(phasesByJob[jobID], func(i, j int) bool {
			left := templatePhases[phasesByJob[jobID][i].GetTemplatePhaseId()]
			right := templatePhases[phasesByJob[jobID][j].GetTemplatePhaseId()]
			if left.GetPhaseOrder() != right.GetPhaseOrder() {
				return left.GetPhaseOrder() < right.GetPhaseOrder()
			}
			if left.GetCode() != right.GetCode() {
				return left.GetCode() < right.GetCode()
			}
			return phasesByJob[jobID][i].GetId() < phasesByJob[jobID][j].GetId()
		})
	}
	tasksByPhase := map[string][]*jobtaskpb.JobTask{}
	for _, task := range card.GetJobTasks() {
		if task != nil && strings.TrimSpace(task.GetId()) != "" && (historical || task.GetActive()) && templateTasks[task.GetTemplateTaskId()] != nil {
			tasksByPhase[task.GetJobPhaseId()] = append(tasksByPhase[task.GetJobPhaseId()], task)
		}
	}
	for phaseID := range tasksByPhase {
		sort.SliceStable(tasksByPhase[phaseID], func(i, j int) bool {
			left, right := tasksByPhase[phaseID][i], tasksByPhase[phaseID][j]
			leftTemplate, rightTemplate := templateTasks[left.GetTemplateTaskId()], templateTasks[right.GetTemplateTaskId()]
			if leftTemplate.GetStepOrder() != rightTemplate.GetStepOrder() {
				return leftTemplate.GetStepOrder() < rightTemplate.GetStepOrder()
			}
			if left.GetName() != right.GetName() {
				return left.GetName() < right.GetName()
			}
			return left.GetId() < right.GetId()
		})
	}
	outcomes := map[string]*exportpb.ClientReportCardTaskOutcome{}
	for _, outcome := range card.GetTaskOutcomes() {
		if outcome == nil {
			continue
		}
		key := outcome.GetJobTaskId() + "\x00" + outcome.GetTemplateTaskCriteriaId()
		if current := outcomes[key]; current == nil || outcome.GetRecordedDate() > current.GetRecordedDate() {
			outcomes[key] = outcome
		}
	}

	jobsByID := map[string]*jobpb.Job{}
	for _, job := range card.GetJobs() {
		if job != nil && strings.TrimSpace(job.GetId()) != "" && (historical || job.GetActive()) {
			jobsByID[job.GetId()] = job
		}
	}
	jobIDs := make([]string, 0, len(jobsByID))
	for jobID := range jobsByID {
		jobIDs = append(jobIDs, jobID)
	}
	sort.SliceStable(jobIDs, func(i, j int) bool {
		left, right := jobsByID[jobIDs[i]], jobsByID[jobIDs[j]]
		leftCategory, rightCategory := categories[left.GetJobCategoryId()], categories[right.GetJobCategoryId()]
		leftOrder, rightOrder := int32(^uint32(0)>>1), int32(^uint32(0)>>1)
		if leftCategory != nil && leftCategory.SortOrder != nil {
			leftOrder = leftCategory.GetSortOrder()
		}
		if rightCategory != nil && rightCategory.SortOrder != nil {
			rightOrder = rightCategory.GetSortOrder()
		}
		if leftOrder != rightOrder {
			return leftOrder < rightOrder
		}
		leftName, rightName := left.GetName(), right.GetName()
		if leftTemplate := templates[left.GetJobTemplateId()]; leftTemplate != nil {
			leftName = firstNonEmpty(leftName, leftTemplate.GetName())
		}
		if rightTemplate := templates[right.GetJobTemplateId()]; rightTemplate != nil {
			rightName = firstNonEmpty(rightName, rightTemplate.GetName())
		}
		if leftName != rightName {
			return leftName < rightName
		}
		return jobIDs[i] < jobIDs[j]
	})

	type sectionState struct {
		category  *categorypb.JobCategory
		rows      []any
		rowsByKey map[string]map[string]any
	}
	sectionsByCategory := map[string]*sectionState{}
	categoryIDs := make([]string, 0)
	for _, jobID := range jobIDs {
		job := jobsByID[jobID]
		categoryID := strings.TrimSpace(job.GetJobCategoryId())
		if categoryID == "" {
			if template := templates[job.GetJobTemplateId()]; template != nil {
				categoryID = strings.TrimSpace(template.GetJobCategoryId())
			}
		}
		category := categories[categoryID]
		if category == nil {
			continue
		}
		state := sectionsByCategory[categoryID]
		if state == nil {
			state = &sectionState{category: category, rowsByKey: map[string]map[string]any{}}
			sectionsByCategory[categoryID] = state
			categoryIDs = append(categoryIDs, categoryID)
		}
		phaseIDs := phasesByJob[jobID]
		template := templates[job.GetJobTemplateId()]
		for _, phase := range phaseIDs {
			tasks := tasksByPhase[phase.GetId()]
			for _, task := range tasks {
				templateTask := templateTasks[task.GetTemplateTaskId()]
				for _, link := range linksByTask[templateTask.GetId()] {
					criterion := criteria[link.GetOutcomeCriteriaId()]
					outcome := outcomes[task.GetId()+"\x00"+link.GetId()]
					templatePhase := templatePhases[phase.GetTemplatePhaseId()]
					taskCode := strings.TrimSpace(templateTask.GetCode())
					if taskCode == "" {
						taskCode = genericPathKey(firstNonEmpty(templateTask.GetName(), task.GetName()))
					}
					criterionCode := strings.TrimSpace(criterion.GetCode())
					if criterionCode == "" {
						criterionCode = genericPathKey(criterion.GetName())
					}
					rowKey := jobID + "\x00" + criterion.GetId()
					row := state.rowsByKey[rowKey]
					if row == nil {
						row = map[string]any{
							"job_name": firstNonEmpty(strings.TrimSpace(job.GetName()), func() string {
								if template != nil {
									return strings.TrimSpace(template.GetName())
								}
								return ""
							}()),
							"row_code": criterionCode,
							"row_name": strings.TrimSpace(criterion.GetName()),
							"cells":    []any{},
						}
						state.rowsByKey[rowKey] = row
						state.rows = append(state.rows, row)
					}
					cells := row["cells"].([]any)
					cells = append(cells, map[string]any{
						"period_code":  strings.TrimSpace(templatePhase.GetCode()),
						"period_name":  strings.TrimSpace(templatePhase.GetName()),
						"period_order": templatePhase.GetPhaseOrder(),
						"task_code":    taskCode,
						"task_name":    strings.TrimSpace(firstNonEmpty(templateTask.GetName(), task.GetName())),
						"value":        taskOutcomeMark(outcome),
					})
					row["cells"] = cells
					if outcome != nil && outcome.NumericValue != nil {
						previous, _ := row["total"].(string)
						var total float64
						if previous != "" {
							total, _ = strconv.ParseFloat(previous, 64)
						}
						row["total"] = formatClientReportMaximum(total + outcome.GetNumericValue())
					}
				}
			}
		}
	}
	sort.SliceStable(categoryIDs, func(i, j int) bool {
		left, right := sectionsByCategory[categoryIDs[i]].category, sectionsByCategory[categoryIDs[j]].category
		leftOrder, rightOrder := int32(^uint32(0)>>1), int32(^uint32(0)>>1)
		if left.SortOrder != nil {
			leftOrder = left.GetSortOrder()
		}
		if right.SortOrder != nil {
			rightOrder = right.GetSortOrder()
		}
		if leftOrder != rightOrder {
			return leftOrder < rightOrder
		}
		if left.GetName() != right.GetName() {
			return left.GetName() < right.GetName()
		}
		return categoryIDs[i] < categoryIDs[j]
	})
	sections := make([]any, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		state := sectionsByCategory[categoryID]
		if len(state.rows) == 0 {
			continue
		}
		sections = append(sections, map[string]any{
			"section_code": firstNonEmpty(strings.TrimSpace(state.category.GetCode()), genericPathKey(state.category.GetName())),
			"section_name": strings.TrimSpace(state.category.GetName()),
			"rows":         state.rows,
		})
	}
	return sections
}

func projectedProgressToDate(
	selected selectedClientPhaseJob,
	jobPhasesByJob map[string][]*jobphasepb.JobPhase,
	jobTasksByPhase map[string][]*jobtaskpb.JobTask,
	templatePhases map[string]*jobtemplatephasepb.JobTemplatePhase,
	templateTasks map[string]*jobtemplatetaskpb.JobTemplateTask,
	criteriaByTemplateTask map[string][]*templatecriteriapb.TemplateTaskCriteria,
	criteriaByID map[string]*criteriapb.OutcomeCriteria,
	taskOutcomes map[string]*exportpb.ClientReportCardTaskOutcome,
	historical bool,
) (float64, float64, bool) {
	var total, maximum float64
	hasNumericOutcome := false
	for _, phase := range jobPhasesByJob[selected.job.GetId()] {
		if !historical && !phase.GetActive() {
			continue
		}
		templatePhase := templatePhases[phase.GetTemplatePhaseId()]
		if templatePhase == nil || templatePhase.GetPhaseOrder() > selected.tp.GetPhaseOrder() {
			continue
		}
		for _, task := range jobTasksByPhase[phase.GetId()] {
			templateTask := templateTasks[task.GetTemplateTaskId()]
			if templateTask == nil || !historical && !templateTask.GetActive() {
				continue
			}
			for _, link := range criteriaByTemplateTask[templateTask.GetId()] {
				criterion := criteriaByID[link.GetOutcomeCriteriaId()]
				if criterion == nil || criterion.MaxScore == nil {
					continue
				}
				maximum += float64(criterion.GetMaxScore())
				if outcome := taskOutcomes[task.GetId()+"\x00"+link.GetId()]; outcome != nil && outcome.NumericValue != nil {
					total += outcome.GetNumericValue()
					hasNumericOutcome = true
				}
			}
		}
	}
	return total, maximum, hasNumericOutcome
}

func categoryOrder(category *categorypb.JobCategory) int32 {
	if category.SortOrder == nil {
		return int32(^uint32(0) >> 1)
	}
	return category.GetSortOrder()
}

func taskOutcomeMark(outcome *exportpb.ClientReportCardTaskOutcome) string {
	if outcome == nil {
		return ""
	}
	if outcome.ScaledLabel != nil && strings.TrimSpace(outcome.GetScaledLabel()) != "" {
		return strings.TrimSpace(outcome.GetScaledLabel())
	}
	if outcome.NumericValue != nil {
		return strconv.FormatFloat(outcome.GetNumericValue(), 'f', -1, 64)
	}
	return ""
}

func staffNames(card *exportpb.ClientReportCardProjection, jobID, jobPhaseID string) string {
	jobID = strings.TrimSpace(jobID)
	jobPhaseID = strings.TrimSpace(jobPhaseID)
	byID := make(map[string]string, len(card.Staff))
	for _, staff := range card.Staff {
		if staff != nil && strings.TrimSpace(staff.GetStaffId()) != "" {
			byID[staff.GetStaffId()] = strings.TrimSpace(staff.GetDisplayName())
		}
	}
	names := make([]string, 0, 2)
	seen := make(map[string]struct{})
	for _, assignment := range card.TeacherAssignments {
		if assignment == nil || strings.TrimSpace(jobPhaseID) == "" || strings.TrimSpace(assignment.GetJobPhaseId()) != jobPhaseID {
			continue
		}
		if assignment.GetJobId() != "" && strings.TrimSpace(assignment.GetJobId()) != jobID {
			continue
		}
		name := strings.TrimSpace(assignment.GetDisplayName())
		if name == "" {
			name = byID[assignment.GetStaffId()]
		}
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// composeJobVariantName names a job whose selected period is pinned to a
// product variant as "Arts (Music)", matching the grade-sheet period header
// (composePhaseLabel). A blank variant, or a name that already carries it,
// leaves the name unchanged.
func composeJobVariantName(name, variant string) string {
	name, variant = strings.TrimSpace(name), strings.TrimSpace(variant)
	if name == "" || variant == "" || strings.Contains(strings.ToLower(name), strings.ToLower(variant)) {
		return name
	}
	return name + " (" + variant + ")"
}

// capNames keeps the first max names of a ", "-joined list (max <= 0 = all).
func capNames(joined string, max int) string {
	if max <= 0 || joined == "" {
		return joined
	}
	names := strings.Split(joined, ", ")
	if len(names) <= max {
		return joined
	}
	return strings.Join(names[:max], ", ")
}

func formatClientReportMaximum(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// trimTrailingQualifier drops one trailing parenthetical qualifier from a
// display name ("Palladium (AY 2026-27)" → "Palladium"); the qualifier repeats
// information the document already prints (the academic year).
func trimTrailingQualifier(name string) string {
	name = strings.TrimSpace(name)
	if strings.HasSuffix(name, ")") {
		if open := strings.LastIndex(name, " ("); open > 0 {
			return strings.TrimSpace(name[:open])
		}
	}
	return name
}
