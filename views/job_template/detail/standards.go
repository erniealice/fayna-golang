package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/pyeza-golang/types"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	templatetaskcriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
)

// loadStandardsTab populates PageData.StandardsTable with TemplateTaskCriteria
// pinnings: which OutcomeCriteria are pinned to which JobTemplateTask and phase.
//
// Strategy: walk phases → tasks → criteria (three nested calls). All three deps
// must be non-nil; otherwise an empty-state panel is shown.
//
// TODO(P6.template-children): wire ListCriteriaByTask when the
// TemplateTaskCriteria view module lands. The "Add Standard" CTA URL
// (/app/template-task-criteria/add?template_id={id}) does not exist yet.
func loadStandardsTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, templateID string) {
	if deps.ListPhasesByJobTemplate == nil || deps.ListTasksByPhase == nil || deps.ListCriteriaByTask == nil {
		// One or more deps unwired — render empty state (table stays nil).
		return
	}

	// 1. Fetch phases.
	phaseResp, err := deps.ListPhasesByJobTemplate(ctx, &jobtemplatephasepb.ListByJobTemplateRequest{
		JobTemplateId: templateID,
	})
	if err != nil {
		log.Printf("loadStandardsTab: failed to list phases for template %s: %v", templateID, err)
		return
	}
	phases := phaseResp.GetJobTemplatePhases()

	type criteriaRow struct {
		phaseName string
		taskName  string
		taskStep  int32
		seqOrder  int32
		criteria  *templatetaskcriteriapb.TemplateTaskCriteria
	}

	var allRows []criteriaRow

	// 2. Walk phases → tasks → criteria.
	for _, p := range phases {
		taskResp, err := deps.ListTasksByPhase(ctx, &jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest{
			JobTemplatePhaseId: p.GetId(),
		})
		if err != nil {
			log.Printf("loadStandardsTab: failed to list tasks for phase %s: %v", p.GetId(), err)
			continue
		}
		for _, t := range taskResp.GetJobTemplateTasks() {
			critResp, err := deps.ListCriteriaByTask(ctx, &templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest{
				JobTemplateTaskId: t.GetId(),
			})
			if err != nil {
				log.Printf("loadStandardsTab: failed to list criteria for task %s: %v", t.GetId(), err)
				continue
			}
			for _, c := range critResp.GetTemplateTaskCriterias() {
				allRows = append(allRows, criteriaRow{
					phaseName: p.GetName(),
					taskName:  t.GetName(),
					taskStep:  t.GetStepOrder(),
					seqOrder:  c.GetSequenceOrder(),
					criteria:  c,
				})
			}
		}
	}

	// 3. Build table rows.
	rows := make([]types.TableRow, 0, len(allRows))
	for _, r := range allRows {
		criteriaName := ""
		if oc := r.criteria.GetOutcomeCriteria(); oc != nil {
			criteriaName = oc.GetName()
		}
		if criteriaName == "" {
			criteriaName = r.criteria.GetOutcomeCriteriaId()
		}
		rows = append(rows, types.TableRow{
			ID: r.criteria.GetId(),
			Cells: []types.TableCell{
				{Type: "text", Value: r.phaseName},
				{Type: "text", Value: r.taskName},
				{Type: "text", Value: fmt.Sprintf("%d", r.taskStep)},
				{Type: "text", Value: criteriaName},
				{Type: "text", Value: fmt.Sprintf("%d", r.seqOrder)},
			},
		})
	}

	pageData.StandardsTable = &types.TableConfig{
		ID: "jt-standards-table",
		Columns: []types.TableColumn{
			{Key: "phase", Label: "Phase"},
			{Key: "task", Label: "Task"},
			{Key: "step", Label: "#", WidthClass: "col-sm"},
			{Key: "criteria", Label: "Outcome Criteria"},
			{Key: "seq", Label: "Seq", WidthClass: "col-sm"},
		},
		Rows:        rows,
		Labels:      deps.TableLabels,
		ShowSearch:  false,
		ShowActions: false,
		ShowSort:    false,
		ShowColumns: false,
		ShowDensity: false,
		ShowEntries: false,
	}
	types.ApplyColumnStyles(pageData.StandardsTable.Columns, pageData.StandardsTable.Rows)
}
