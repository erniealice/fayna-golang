package document

import (
	"sort"
	"strconv"
	"strings"
	"time"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplatetaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	phaseoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
)

type projectedJob struct {
	jobID       string
	categoryID  string
	categoryKey string
	templateID  string
	template    string
}

// buildProjectedYearFinalData maps the one-client typed projection onto the
// existing report-card template contract. It reads no extra entities and uses
// the same root builder/blank-seeding path as the legacy document renderer.
func buildProjectedYearFinalData(d *Deps, card *exportpb.ClientReportCardProjection, printedBy, printedAt string, now time.Time) map[string]any {
	if card == nil {
		return buildReportCardData(reportCard{DocumentHeaderName: d.DocumentHeaderName})
	}
	context := card.GetContext()
	groupName := ""
	period := ""
	if context != nil {
		groupName = strings.TrimSpace(context.GetSubscriptionGroupName())
		period = strings.TrimSpace(context.GetPriceScheduleName())
	}
	name, ayFromGroup := groupParts(groupName)
	grade, sectionName := gradeSection(name)
	if period == "" {
		period = ayFromGroup
	}
	groupLabel := sectionName
	if grade != "" {
		groupLabel = grade + " - " + sectionName
	}
	clientName := ""
	if card.GetClient() != nil {
		clientName = strings.TrimSpace(card.GetClient().GetName())
	}
	attributes := make(map[string]any, len(card.GetAttributes()))
	clientReference := ""
	for _, attribute := range card.GetAttributes() {
		if attribute == nil {
			continue
		}
		code := strings.TrimSpace(attribute.GetCode())
		if code == "" {
			continue
		}
		attributes[code] = strings.TrimSpace(attribute.GetValue())
		if code == strings.TrimSpace(d.DocOptions.ClientReferenceAttributeCode) {
			clientReference = strings.TrimSpace(attribute.GetValue())
		}
	}
	historical := context != nil && context.GetHistorical()
	jobCategories, groupLead, leadDisplay := buildProjectedJobCategories(card, strings.TrimSpace(d.DocOptions.GroupCategoryFilter), historical)
	subjects, formation, itemRatings, groupRatingOne, groupRatingTwo := buildProjectedLegacyCardRows(d, card)
	rc := reportCard{
		DocumentHeaderName:    firstNonEmpty(strings.TrimSpace(d.DocumentHeaderName), d.Labels.Landing.Title, "Report Card"),
		SchedulePeriod:        ayFromGroup,
		SchedulePeriodDisplay: period,
		SchedulePeriodSpaced:  strings.Replace(period, "-", " - ", 1),
		ClientName:            clientName,
		GroupLevel:            grade,
		SubscriptionGroupName: sectionName,
		PriceScheduleID:       "",
		GroupLabel:            groupLabel,
		GroupLead:             groupLead,
		LeadStaffDisplay:      leadDisplay,
		ClientReference:       clientReference,
		ClientAttributes:      attributes,
		Subjects:              subjects,
		FormationGroups:       formation,
		ItemRatings:           itemRatings,
		GroupRatingPhase1:     groupRatingOne,
		GroupRatingPhase2:     groupRatingTwo,
		PrintedBy:             printedBy,
		PrintedAt:             printedAt,
		PrintedByName:         printedBy,
		PrintedAtLong:         now.Format("January 2, 2006 03:04 PM"),
		JobIDs:                append([]string(nil), card.GetRenderGateJobIds()...),
		JobCategories:         jobCategories,
	}
	data := buildReportCardData(rc)
	// The operator path keeps its historical contract. The explicit client
	// export exposes the same projected coded cells through the neutral matrix
	// shape and omits the vertical-specific legacy alias.
	delete(data, "conduct_rows")
	data["outcome_sections"] = buildProjectedOutcomeSections(card, historical)
	outcomeCells, outcomeTotals := buildProjectedOutcomeCellIndex(card, historical)
	data["outcome_cells"] = outcomeCells
	data["outcome_totals"] = outcomeTotals
	return data
}

func buildProjectedJobCategories(card *exportpb.ClientReportCardProjection, groupCategoryCode string, historical bool) (map[string]any, string, string) {
	categories := map[string]*jobcategorypb.JobCategory{}
	for _, category := range card.GetJobCategories() {
		if category != nil && strings.TrimSpace(category.GetId()) != "" {
			categories[category.GetId()] = category
		}
	}
	templates := map[string]string{}
	for _, template := range card.GetJobTemplates() {
		if template != nil && strings.TrimSpace(template.GetId()) != "" {
			templates[template.GetId()] = strings.TrimSpace(template.GetName())
		}
	}
	items := make([]projectedJob, 0, len(card.GetJobs()))
	for _, job := range card.GetJobs() {
		if job == nil || strings.TrimSpace(job.GetId()) == "" || !historical && !job.GetActive() {
			continue
		}
		categoryID := strings.TrimSpace(job.GetJobCategoryId())
		if categoryID == "" {
			if template := cardTemplateByID(card, job.GetJobTemplateId()); template != nil {
				categoryID = strings.TrimSpace(template.GetJobCategoryId())
			}
		}
		category := categories[categoryID]
		categoryKey := "uncategorized"
		if category != nil && strings.TrimSpace(category.GetCode()) != "" {
			categoryKey = strings.TrimSpace(category.GetCode())
		}
		items = append(items, projectedJob{
			jobID:       job.GetId(),
			categoryID:  categoryID,
			categoryKey: categoryKey,
			templateID:  job.GetJobTemplateId(),
			template:    firstNonEmpty(templates[job.GetJobTemplateId()], job.GetName()),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		left, right := categories[items[i].categoryID], categories[items[j].categoryID]
		if left != nil && right != nil && left.GetSortOrder() != right.GetSortOrder() {
			return left.GetSortOrder() < right.GetSortOrder()
		}
		if items[i].categoryKey != items[j].categoryKey {
			return items[i].categoryKey < items[j].categoryKey
		}
		if items[i].template != items[j].template {
			return strings.ToLower(items[i].template) < strings.ToLower(items[j].template)
		}
		return items[i].jobID < items[j].jobID
	})

	phaseTemplateByID := map[string]*jobtemplatephasepb.JobTemplatePhase{}
	for _, phase := range card.GetJobTemplatePhases() {
		if phase != nil && strings.TrimSpace(phase.GetId()) != "" && (historical || phase.GetActive()) {
			phaseTemplateByID[phase.GetId()] = phase
		}
	}
	phasesByJob := map[string][]*jobphasepb.JobPhase{}
	for _, phase := range card.GetJobPhases() {
		if phase == nil || strings.TrimSpace(phase.GetId()) == "" {
			continue
		}
		phasesByJob[phase.GetJobId()] = append(phasesByJob[phase.GetJobId()], phase)
	}
	phaseSummaryByID := latestPhaseSummaryByID(card.GetPhaseOutcomeSummaries(), historical)
	finalByJob := latestFinalSummaryByJob(card.GetJobOutcomeSummaries(), historical)
	tasksByPhase := map[string][]*jobtaskpb.JobTask{}
	templateTaskByID := map[string]*jobtemplatetaskpb.JobTemplateTask{}
	for _, task := range card.GetJobTemplateTasks() {
		if task != nil && strings.TrimSpace(task.GetId()) != "" && (historical || task.GetActive()) {
			templateTaskByID[task.GetId()] = task
		}
	}
	for _, task := range card.GetJobTasks() {
		if task != nil && (historical || task.GetActive()) && strings.TrimSpace(task.GetId()) != "" {
			tasksByPhase[task.GetJobPhaseId()] = append(tasksByPhase[task.GetJobPhaseId()], task)
		}
	}
	templateCriteriaByTask := map[string][]*ttcpb.TemplateTaskCriteria{}
	for _, binding := range card.GetTemplateTaskCriteria() {
		if binding != nil && (historical || binding.GetActive()) {
			templateCriteriaByTask[binding.GetJobTemplateTaskId()] = append(templateCriteriaByTask[binding.GetJobTemplateTaskId()], binding)
		}
	}
	for taskID := range templateCriteriaByTask {
		sort.SliceStable(templateCriteriaByTask[taskID], func(i, j int) bool {
			left, right := templateCriteriaByTask[taskID][i], templateCriteriaByTask[taskID][j]
			if left.GetSequenceOrder() != right.GetSequenceOrder() {
				return left.GetSequenceOrder() < right.GetSequenceOrder()
			}
			return left.GetId() < right.GetId()
		})
	}
	criteriaByID := map[string]*criteriapb.OutcomeCriteria{}
	for _, criterion := range card.GetOutcomeCriteria() {
		if criterion != nil && (historical || criterion.GetActive()) {
			criteriaByID[criterion.GetId()] = criterion
		}
	}
	outcomesByKey := map[string]*exportpb.ClientReportCardTaskOutcome{}
	for _, outcome := range card.GetTaskOutcomes() {
		if outcome == nil {
			continue
		}
		key := outcome.GetJobTaskId() + "\x00" + outcome.GetTemplateTaskCriteriaId()
		if current := outcomesByKey[key]; current == nil || outcome.GetRecordedDate() > current.GetRecordedDate() {
			outcomesByKey[key] = outcome
		}
	}

	result := map[string]any{}
	jobsByCategory := map[string][]any{}
	groupLeads := map[string][]string{}
	for _, item := range items {
		jobPhases := append([]*jobphasepb.JobPhase(nil), phasesByJob[item.jobID]...)
		sort.Slice(jobPhases, func(i, j int) bool {
			left, right := phaseTemplateByID[jobPhases[i].GetTemplatePhaseId()], phaseTemplateByID[jobPhases[j].GetTemplatePhaseId()]
			if left != nil && right != nil && left.GetPhaseOrder() != right.GetPhaseOrder() {
				return left.GetPhaseOrder() < right.GetPhaseOrder()
			}
			return jobPhases[i].GetId() < jobPhases[j].GetId()
		})
		for _, phase := range jobPhases {
			sort.SliceStable(tasksByPhase[phase.GetId()], func(i, j int) bool {
				left, right := tasksByPhase[phase.GetId()][i], tasksByPhase[phase.GetId()][j]
				leftOrder, rightOrder := int32(^uint32(0)>>1), int32(^uint32(0)>>1)
				if templateTask := templateTaskByID[left.GetTemplateTaskId()]; templateTask != nil {
					leftOrder = templateTask.GetStepOrder()
				}
				if templateTask := templateTaskByID[right.GetTemplateTaskId()]; templateTask != nil {
					rightOrder = templateTask.GetStepOrder()
				}
				if leftOrder != rightOrder {
					return leftOrder < rightOrder
				}
				if left.GetName() != right.GetName() {
					return left.GetName() < right.GetName()
				}
				return left.GetId() < right.GetId()
			})
		}
		jobTemplatePhases := map[string]any{}
		criteriaRows := map[string]map[string]any{}
		criteriaOrder := make([]string, 0)
		for _, phase := range jobPhases {
			templatePhase := phaseTemplateByID[phase.GetTemplatePhaseId()]
			phaseCode := ""
			if templatePhase != nil {
				phaseCode = strings.TrimSpace(templatePhase.GetCode())
			}
			if phaseCode == "" {
				continue
			}
			phaseData := map[string]any{}
			if summary := phaseSummaryByID[phase.GetId()]; summary != nil {
				phaseData["phase_outcome_summary_scaled_label"] = summary.GetScaledLabel()
			}
			phaseTasks := map[string]any{}
			var phaseTotal float64
			hasPhaseNumericOutcome := false
			for _, task := range tasksByPhase[phase.GetId()] {
				templateTask := templateTaskByID[task.GetTemplateTaskId()]
				if templateTask == nil || !historical && !templateTask.GetActive() {
					continue
				}
				taskCode := strings.TrimSpace(templateTask.GetCode())
				if taskCode == "" {
					taskCode = genericPathKey(templateTask.GetName())
				}
				if taskCode == "" {
					continue
				}
				taskOutcomeRows := map[string]any{}
				for _, binding := range templateCriteriaByTask[templateTask.GetId()] {
					criterion := criteriaByID[binding.GetOutcomeCriteriaId()]
					if criterion == nil {
						continue
					}
					criterionKey := strings.TrimSpace(criterion.GetCode())
					if criterionKey == "" {
						criterionKey = genericPathKey(criterion.GetName())
					}
					if criterionKey == "" {
						continue
					}
					outcome := outcomesByKey[task.GetId()+"\x00"+binding.GetId()]
					outcomeData := map[string]any{}
					if outcome != nil {
						if outcome.NumericValue != nil {
							outcomeData["numeric_value"] = strconv.FormatFloat(outcome.GetNumericValue(), 'f', -1, 64)
							phaseTotal += outcome.GetNumericValue()
							hasPhaseNumericOutcome = true
						}
						outcomeData["scaled_label"] = outcome.GetScaledLabel()
						outcomeData["comment"] = outcome.GetDeterminationNote()
					}
					taskOutcomeRows[criterionKey] = outcomeData
					criterionPhaseRows := criteriaRows[criterion.GetId()]
					if criterionPhaseRows == nil {
						criterionPhaseRows = map[string]any{"outcome_criteria_label_display": criterion.GetName(), "job_template_phases": map[string]any{}}
						criteriaRows[criterion.GetId()] = criterionPhaseRows
						criteriaOrder = append(criteriaOrder, criterion.GetId())
					}
					phaseValues := criterionPhaseRows["job_template_phases"].(map[string]any)
					if outcome != nil && outcome.NumericValue != nil {
						phaseValues[phaseCode] = map[string]any{"task_outcome_numeric_value_max_derived": strconv.FormatFloat(outcome.GetNumericValue(), 'f', -1, 64)}
					}
					if criterion.MaxScore != nil {
						criterionPhaseRows["maximum"] = formatClientReportMaximum(float64(criterion.GetMaxScore()))
					}
				}
				phaseTasks[taskCode] = map[string]any{"task_outcomes": taskOutcomeRows}
			}
			phaseData["job_template_tasks"] = phaseTasks
			if hasPhaseNumericOutcome {
				phaseData["task_outcome_numeric_value_total_derived"] = strconv.FormatFloat(phaseTotal, 'f', -1, 64)
			} else {
				phaseData["task_outcome_numeric_value_total_derived"] = ""
			}
			jobTemplatePhases[phaseCode] = phaseData
		}
		criteriaList := make([]any, 0, len(criteriaOrder))
		for _, criterionID := range criteriaOrder {
			criteriaList = append(criteriaList, criteriaRows[criterionID])
		}
		assignmentNames := strings.Split(projectedJobTeacherNames(card, item.jobID, jobPhases), " / ")
		if len(assignmentNames) == 1 && assignmentNames[0] == "" {
			assignmentNames = nil
		}
		if item.categoryKey == groupCategoryCode {
			groupLeads[item.categoryKey] = append(groupLeads[item.categoryKey], assignmentNames...)
		}
		finalLabel := ""
		if summary := finalByJob[item.jobID]; summary != nil {
			finalLabel = summary.GetScaledLabel()
		}
		jobsByCategory[item.categoryKey] = append(jobsByCategory[item.categoryKey], map[string]any{
			"job_template_name_display":        item.template,
			"staff_line_display":               strings.Join(assignmentNames, " / "),
			"job_outcome_summary_scaled_label": finalLabel,
			"job_template_phases":              jobTemplatePhases,
			"outcome_criteria":                 criteriaList,
		})
	}
	for categoryKey, jobs := range jobsByCategory {
		result[categoryKey] = map[string]any{"jobs": jobs}
	}
	groupLead := ""
	leadDisplay := ""
	if groupCategoryCode != "" {
		groupLeads[groupCategoryCode] = sortedUniqueStrings(groupLeads[groupCategoryCode])
		groupLead = strings.Join(groupLeads[groupCategoryCode], " / ")
		if len(jobsByCategory[groupCategoryCode]) == 1 {
			leadDisplay = groupLead
		}
	}
	return result, groupLead, leadDisplay
}

func cardTemplateByID(card *exportpb.ClientReportCardProjection, templateID string) *jobtemplatepb.JobTemplate {
	for _, template := range card.GetJobTemplates() {
		if template != nil && strings.TrimSpace(template.GetId()) == strings.TrimSpace(templateID) {
			return template
		}
	}
	return nil
}

// buildProjectedLegacyCardRows fills the normalized reportCard fields consumed
// by buildReportCardData. It mirrors the legacy transcript walk from only the
// selected client's typed projection: per-job/phase/criterion maxima come from
// latest task outcomes; stored phase and job summary labels remain authoritative.
func buildProjectedLegacyCardRows(d *Deps, card *exportpb.ClientReportCardProjection) ([]itemRow, []formationGroup, []ratingRow, string, string) {
	if card == nil {
		return nil, nil, nil, "", ""
	}
	historical := card.GetContext().GetHistorical()
	categoryByID := map[string]*jobcategorypb.JobCategory{}
	for _, category := range card.GetJobCategories() {
		if category != nil {
			categoryByID[category.GetId()] = category
		}
	}
	templateByID := map[string]*jobtemplatepb.JobTemplate{}
	for _, template := range card.GetJobTemplates() {
		if template != nil {
			templateByID[template.GetId()] = template
		}
	}
	templatePhaseByID := map[string]*jobtemplatephasepb.JobTemplatePhase{}
	for _, phase := range card.GetJobTemplatePhases() {
		if phase != nil && (historical || phase.GetActive()) {
			templatePhaseByID[phase.GetId()] = phase
		}
	}
	templateTaskByID := map[string]*jobtemplatetaskpb.JobTemplateTask{}
	for _, task := range card.GetJobTemplateTasks() {
		if task != nil && (historical || task.GetActive()) {
			templateTaskByID[task.GetId()] = task
		}
	}
	criteriaByID := map[string]*criteriapb.OutcomeCriteria{}
	for _, criterion := range card.GetOutcomeCriteria() {
		if criterion != nil && (historical || criterion.GetActive()) {
			criteriaByID[criterion.GetId()] = criterion
		}
	}
	linksByTask := map[string][]*ttcpb.TemplateTaskCriteria{}
	for _, link := range card.GetTemplateTaskCriteria() {
		if link != nil && (historical || link.GetActive()) && criteriaByID[link.GetOutcomeCriteriaId()] != nil {
			linksByTask[link.GetJobTemplateTaskId()] = append(linksByTask[link.GetJobTemplateTaskId()], link)
		}
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
		if phase != nil && (historical || phase.GetActive()) && templatePhaseByID[phase.GetTemplatePhaseId()] != nil {
			phasesByJob[phase.GetJobId()] = append(phasesByJob[phase.GetJobId()], phase)
		}
	}
	for jobID := range phasesByJob {
		sort.SliceStable(phasesByJob[jobID], func(i, j int) bool {
			left := templatePhaseByID[phasesByJob[jobID][i].GetTemplatePhaseId()]
			right := templatePhaseByID[phasesByJob[jobID][j].GetTemplatePhaseId()]
			if left.GetPhaseOrder() != right.GetPhaseOrder() {
				return left.GetPhaseOrder() < right.GetPhaseOrder()
			}
			return phasesByJob[jobID][i].GetId() < phasesByJob[jobID][j].GetId()
		})
	}
	tasksByPhase := map[string][]*jobtaskpb.JobTask{}
	for _, task := range card.GetJobTasks() {
		if task != nil && (historical || task.GetActive()) && templateTaskByID[task.GetTemplateTaskId()] != nil {
			tasksByPhase[task.GetJobPhaseId()] = append(tasksByPhase[task.GetJobPhaseId()], task)
		}
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
	phaseSummaries := latestPhaseSummaryByID(card.GetPhaseOutcomeSummaries(), historical)
	finalSummaries := latestFinalSummaryByJob(card.GetJobOutcomeSummaries(), historical)

	academicCode, groupCode := "", ""
	if d != nil {
		academicCode = strings.TrimSpace(d.CategoryFilter)
		groupCode = strings.TrimSpace(d.DocOptions.GroupCategoryFilter)
	}
	jobs := make([]*jobpb.Job, 0, len(card.GetJobs()))
	categoryIDForJob := map[string]string{}
	for _, job := range card.GetJobs() {
		if job == nil || !historical && !job.GetActive() || strings.TrimSpace(job.GetId()) == "" {
			continue
		}
		categoryID := strings.TrimSpace(job.GetJobCategoryId())
		if categoryID == "" {
			if template := templateByID[job.GetJobTemplateId()]; template != nil {
				categoryID = strings.TrimSpace(template.GetJobCategoryId())
			}
		}
		categoryIDForJob[job.GetId()] = categoryID
		jobs = append(jobs, job)
	}
	sort.SliceStable(jobs, func(i, j int) bool {
		leftCategory, rightCategory := categoryByID[categoryIDForJob[jobs[i].GetId()]], categoryByID[categoryIDForJob[jobs[j].GetId()]]
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
		leftName, rightName := jobs[i].GetName(), jobs[j].GetName()
		if template := templateByID[jobs[i].GetJobTemplateId()]; template != nil {
			leftName = firstNonEmpty(leftName, template.GetName())
		}
		if template := templateByID[jobs[j].GetJobTemplateId()]; template != nil {
			rightName = firstNonEmpty(rightName, template.GetName())
		}
		if leftName != rightName {
			return leftName < rightName
		}
		return jobs[i].GetId() < jobs[j].GetId()
	})

	transcripts := map[string]*transcript{}
	for _, job := range jobs {
		tr := &transcript{marks: map[string]map[int32]float64{}, seq: map[string]int32{}}
		for _, phase := range phasesByJob[job.GetId()] {
			templatePhase := templatePhaseByID[phase.GetTemplatePhaseId()]
			order := templatePhase.GetPhaseOrder()
			for _, task := range tasksByPhase[phase.GetId()] {
				templateTask := templateTaskByID[task.GetTemplateTaskId()]
				for _, link := range linksByTask[templateTask.GetId()] {
					outcome := outcomes[task.GetId()+"\x00"+link.GetId()]
					if outcome == nil || outcome.NumericValue == nil {
						continue
					}
					criterionID := link.GetOutcomeCriteriaId()
					if tr.marks[criterionID] == nil {
						tr.marks[criterionID] = map[int32]float64{}
					}
					if current, ok := tr.marks[criterionID][order]; !ok || outcome.GetNumericValue() > current {
						tr.marks[criterionID][order] = outcome.GetNumericValue()
					}
					if current, ok := tr.seq[criterionID]; !ok || link.GetSequenceOrder() < current {
						tr.seq[criterionID] = link.GetSequenceOrder()
					}
				}
			}
		}
		transcripts[job.GetId()] = tr
	}

	var subjects []itemRow
	var formation []formationGroup
	var ratingJobs []*jobpb.Job
	var groupJobs []*jobpb.Job
	academicNames := map[string]bool{}
	for _, job := range jobs {
		category := categoryByID[categoryIDForJob[job.GetId()]]
		if category == nil {
			continue
		}
		isAcademic := academicCode == "" || strings.EqualFold(strings.TrimSpace(category.GetCode()), academicCode)
		if isAcademic {
			name := strings.TrimSpace(job.GetName())
			if template := templateByID[job.GetJobTemplateId()]; template != nil {
				name = firstNonEmpty(name, template.GetName())
			}
			academicNames[strings.ToLower(cleanSubject(name))] = true
			tr := transcripts[job.GetId()]
			crit, hasMarks := tr.yearCriteria()
			row := itemRow{Name: cleanSubject(name), CritA: crit.a, CritB: crit.b, CritC: crit.c, CritD: crit.d, Total: crit.total}
			for _, phase := range phasesByJob[job.GetId()] {
				order := templatePhaseByID[phase.GetTemplatePhaseId()].GetPhaseOrder()
				summary := phaseSummaries[phase.GetId()]
				if summary != nil && order == 1 {
					row.Sem1Band = strings.TrimSpace(summary.GetScaledLabel())
				}
				if summary != nil && order == 2 {
					row.Sem2Band = strings.TrimSpace(summary.GetScaledLabel())
				}
			}
			if summary := finalSummaries[job.GetId()]; summary != nil {
				row.YearFinal = strings.TrimSpace(summary.GetScaledLabel())
			}
			if isNonEnrolledPlaceholder(row, hasMarks) {
				continue
			}
			criterionNames := map[string]string{}
			for criterionID, criterion := range criteriaByID {
				criterionNames[criterionID] = criterion.GetName()
			}
			row.Criteria, row.OrderTotals = tr.criterionRowsByOrder(criterionNames)
			row.Sem1Total, row.Sem2Total = row.OrderTotals[1], row.OrderTotals[2]
			row.ItemTitle = row.Name
			teacherNames := projectedJobTeacherNames(card, job.GetId(), phasesByJob[job.GetId()])
			row.StaffLine = teacherNames
			subjects = append(subjects, row)
		}
		if academicCode != "" && !strings.EqualFold(strings.TrimSpace(category.GetCode()), academicCode) {
			if groupCode != "" && strings.EqualFold(strings.TrimSpace(category.GetCode()), groupCode) {
				groupJobs = append(groupJobs, job)
			} else {
				ratingJobs = append(ratingJobs, job)
			}
			if summary := finalSummaries[job.GetId()]; summary != nil {
				average := strings.TrimSpace(summary.GetScaledLabel())
				if !outcome_summary.IsNonEnrolledCell(outcome_summary.EnrollmentEvidence{HasMarks: true}, average) {
					formation = appendFormationRow(formation, category, firstNonEmpty(templateByID[job.GetJobTemplateId()].GetName(), job.GetName()), average)
				}
			}
		}
	}

	rating := &ratingContext{nameOf: map[string]string{}, pos: map[string]map[int32]string{}, avg: map[string]string{}, groupPos: map[int32]string{}}
	for _, job := range ratingJobs {
		rating.strandJobs = append(rating.strandJobs, job)
		name := strings.TrimSpace(job.GetName())
		if template := templateByID[job.GetJobTemplateId()]; template != nil {
			name = firstNonEmpty(name, template.GetName())
		}
		rating.nameOf[job.GetId()] = cleanSubject(name)
		if summary := finalSummaries[job.GetId()]; summary != nil {
			rating.avg[job.GetId()] = strings.TrimSpace(summary.GetScaledLabel())
		}
		rating.pos[job.GetId()] = map[int32]string{}
		for _, phase := range phasesByJob[job.GetId()] {
			if summary := phaseSummaries[phase.GetId()]; summary != nil {
				rating.pos[job.GetId()][templatePhaseByID[phase.GetTemplatePhaseId()].GetPhaseOrder()] = strings.TrimSpace(summary.GetScaledLabel())
			}
		}
	}
	if len(groupJobs) > 0 {
		groupJob := groupJobs[0]
		for _, phase := range phasesByJob[groupJob.GetId()] {
			if summary := phaseSummaries[phase.GetId()]; summary != nil {
				rating.groupPos[templatePhaseByID[phase.GetTemplatePhaseId()].GetPhaseOrder()] = strings.TrimSpace(summary.GetScaledLabel())
			}
		}
	}
	merged := mergeRotationPairs(rating, academicNames, nil)
	items, groupOne, groupTwo := buildItemRatings(rating, merged)
	return subjects, formation, items, groupOne, groupTwo
}

func appendFormationRow(groups []formationGroup, category *jobcategorypb.JobCategory, name, average string) []formationGroup {
	for i := range groups {
		if groups[i].Title == strings.TrimSpace(category.GetName()) {
			groups[i].Rows = append(groups[i].Rows, formationRow{Subject: cleanSubject(name), Average: average})
			return groups
		}
	}
	return append(groups, formationGroup{Title: strings.TrimSpace(category.GetName()), Rows: []formationRow{{Subject: cleanSubject(name), Average: average}}})
}

func projectedJobTeacherNames(card *exportpb.ClientReportCardProjection, jobID string, phases []*jobphasepb.JobPhase) string {
	phaseIDs := map[string]struct{}{}
	for _, phase := range phases {
		phaseIDs[phase.GetId()] = struct{}{}
	}
	staffNames := map[string]string{}
	for _, staff := range card.GetStaff() {
		if staff != nil && strings.TrimSpace(staff.GetStaffId()) != "" {
			staffNames[staff.GetStaffId()] = strings.TrimSpace(staff.GetDisplayName())
		}
	}
	names := []string{}
	for _, assignment := range card.GetTeacherAssignments() {
		if assignment == nil || assignment.GetJobPhaseId() == "" {
			continue
		}
		if _, ok := phaseIDs[assignment.GetJobPhaseId()]; !ok {
			continue
		}
		if assignment.GetJobId() != "" && strings.TrimSpace(assignment.GetJobId()) != jobID {
			continue
		}
		name := strings.TrimSpace(assignment.GetDisplayName())
		if name == "" {
			name = staffNames[assignment.GetStaffId()]
		}
		if name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(sortedUniqueStrings(names), " / ")
}

func latestPhaseSummaryByID(summaries []*phaseoutcomepb.PhaseOutcomeSummary, historical bool) map[string]*phaseoutcomepb.PhaseOutcomeSummary {
	result := map[string]*phaseoutcomepb.PhaseOutcomeSummary{}
	for _, summary := range summaries {
		if summary == nil || !historical && !summary.GetActive() {
			continue
		}
		if current := result[summary.GetJobPhaseId()]; current == nil || summary.GetDateModified() > current.GetDateModified() || summary.GetDateModified() == current.GetDateModified() && summary.GetId() > current.GetId() {
			result[summary.GetJobPhaseId()] = summary
		}
	}
	return result
}

func latestFinalSummaryByJob(summaries []*jobsumpb.JobOutcomeSummary, historical bool) map[string]*jobsumpb.JobOutcomeSummary {
	result := map[string]*jobsumpb.JobOutcomeSummary{}
	for _, summary := range summaries {
		if summary == nil || !historical && !summary.GetActive() {
			continue
		}
		if current := result[summary.GetJobId()]; current == nil || summary.GetDateModified() > current.GetDateModified() || summary.GetDateModified() == current.GetDateModified() && summary.GetId() > current.GetId() {
			result[summary.GetJobId()] = summary
		}
	}
	return result
}

func genericPathKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var result strings.Builder
	separator := false
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			result.WriteRune(r)
			separator = false
		} else if result.Len() > 0 && !separator {
			result.WriteByte('_')
			separator = true
		}
	}
	return strings.Trim(result.String(), "_")
}

func sortedUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	values = append([]string(nil), values...)
	sort.Strings(values)
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
