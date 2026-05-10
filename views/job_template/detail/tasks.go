package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
)

// loadTasksTab populates PageData.TasksTable with a denormalised list of
// JobTemplateTask rows across all phases for the given template.
//
// Strategy: fetch phases via the existing ListPhasesByJobTemplate dep, then
// call ListTasksByPhase once per phase and flatten the results into a single
// table. Both deps must be non-nil; otherwise an empty-state is shown.
//
// Each row includes Edit and Delete row-level actions when TaskRoutes are wired.
// A top-level "+ Add Task" CTA (TasksAddURL on PageData) is also populated so the
// template can render a header button. For v1, operators select the target phase
// manually in the Add drawer.
func loadTasksTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, templateID string) {
	if deps.ListPhasesByJobTemplate == nil || deps.ListTasksByPhase == nil {
		// Either dep unwired — render empty state (table stays nil).
		return
	}

	// Wire Add CTA — top-level; operator selects phase inside the drawer for v1.
	if deps.TaskRoutes.AddURL != "" {
		pageData.TasksAddURL = deps.TaskRoutes.AddURL + "?job_template_phase_id="
	}

	// 1. Collect phases.
	phaseResp, err := deps.ListPhasesByJobTemplate(ctx, &jobtemplatephasepb.ListByJobTemplateRequest{
		JobTemplateId: templateID,
	})
	if err != nil {
		log.Printf("loadTasksTab: failed to list phases for template %s: %v", templateID, err)
		return
	}
	phases := phaseResp.GetJobTemplatePhases()

	// 2. Collect all tasks across phases.
	var allTasks []*jobtemplateTaskpb.JobTemplateTask
	phaseNameByID := make(map[string]string, len(phases))
	for _, p := range phases {
		phaseNameByID[p.GetId()] = p.GetName()
		taskResp, err := deps.ListTasksByPhase(ctx, &jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest{
			JobTemplatePhaseId: p.GetId(),
		})
		if err != nil {
			log.Printf("loadTasksTab: failed to list tasks for phase %s: %v", p.GetId(), err)
			continue
		}
		allTasks = append(allTasks, taskResp.GetJobTemplateTasks()...)
	}

	// 3. Build table rows.
	hasActions := deps.TaskRoutes.EditURL != "" || deps.TaskRoutes.DeleteURL != ""
	rows := make([]types.TableRow, 0, len(allTasks))
	for _, t := range allTasks {
		phaseName := phaseNameByID[t.GetJobTemplatePhaseId()]
		taskID := t.GetId()
		var actions []types.TableAction
		if deps.TaskRoutes.EditURL != "" {
			editURL := route.ResolveURL(deps.TaskRoutes.EditURL, "id", taskID)
			actions = append(actions, types.TableAction{
				Type:        "edit",
				Label:       "Edit Task",
				HxGet:       editURL,
				HxTarget:    "#sheetContent",
				HxSwap:      "innerHTML",
				OnClick:     "lf.Sheet.open()",
				DrawerTitle: "Edit Task",
			})
		}
		if deps.TaskRoutes.DeleteURL != "" {
			actions = append(actions, types.TableAction{
				Type:     "delete",
				Label:    "Delete Task",
				URL:      deps.TaskRoutes.DeleteURL,
				ItemName: t.GetName(),
			})
		}
		estDur := int32(0)
		if t.EstimatedDurationMinutes != nil {
			estDur = *t.EstimatedDurationMinutes
		}
		rows = append(rows, types.TableRow{
			ID: taskID,
			Cells: []types.TableCell{
				{Type: "text", Value: phaseName},
				{Type: "text", Value: fmt.Sprintf("%d", t.GetStepOrder())},
				{Type: "text", Value: t.GetName()},
				{Type: "text", Value: fmt.Sprintf("%d", estDur)},
			},
			Actions: actions,
		})
	}

	pageData.TasksTable = &types.TableConfig{
		ID: "jt-tasks-table",
		Columns: []types.TableColumn{
			{Key: "phase", Label: "Phase"},
			{Key: "step", Label: "#", WidthClass: "col-sm"},
			{Key: "name", Label: deps.Labels.Columns.Name},
			{Key: "duration_min", Label: "Est. Duration (min)", WidthClass: "col-md"},
		},
		Rows:        rows,
		Labels:      deps.TableLabels,
		ShowSearch:  false,
		ShowActions: hasActions,
		ShowSort:    false,
		ShowColumns: false,
		ShowDensity: false,
		ShowEntries: false,
	}
	types.ApplyColumnStyles(pageData.TasksTable.Columns, pageData.TasksTable.Rows)
}
