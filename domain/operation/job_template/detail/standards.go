package detail

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	templatetaskcriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	templateTaskCriteria "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria"
)

// loadStandardsTab populates PageData.StandardsTable with TemplateTaskCriteria
// pinnings: which OutcomeCriteria are pinned to which JobTemplateTask and phase.
//
// Strategy: walk phases → tasks → criteria (three nested calls). All three deps
// must be non-nil; otherwise an empty-state panel is shown.
//
// The "+ Add Standard" CTA and per-row edit/remove actions are wired to the
// template_task_criteria module's real routes (CriteriaRoutes), permission-
// gated on template_task_criteria:create / template_task_criteria:update /
// template_task_criteria:delete —
// fail-closed inside the tab body per the child-list roster pattern (an
// empty/no-CTA table, never view.Forbidden — a tab-swap target must stay a
// partial).
func loadStandardsTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, templateID string) {
	perms := view.GetUserPermissions(ctx)

	if deps.CriteriaRoutes.AddURL != "" && perms.Can("template_task_criteria", "create") {
		pageData.StandardsAddURL = deps.CriteriaRoutes.AddURL + "?job_template_id=" + templateID + "&return_table=jt-standards-table"
	}

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

	// 3. Build table rows. Per-row edit/remove actions are permission-gated —
	// fail-closed (no action rendered when unpermitted or the route is
	// unwired).
	criteriaLabels := deps.CriteriaLabels
	if criteriaLabels.Actions.Edit == "" || criteriaLabels.Actions.Delete == "" {
		criteriaLabels = templateTaskCriteria.DefaultLabels()
	}
	canEdit := deps.CriteriaRoutes.EditURL != "" && perms.Can("template_task_criteria", "update")
	canDelete := deps.CriteriaRoutes.DeleteURL != "" && perms.Can("template_task_criteria", "delete")
	hasActions := (canEdit || canDelete) && len(allRows) > 0
	rows := make([]types.TableRow, 0, len(allRows))
	for _, r := range allRows {
		criteriaName := ""
		if oc := r.criteria.GetOutcomeCriteria(); oc != nil {
			criteriaName = oc.GetName()
		}
		if criteriaName == "" {
			criteriaName = r.criteria.GetOutcomeCriteriaId()
		}
		var actions []types.TableAction
		if canEdit {
			editURL := route.ResolveURL(deps.CriteriaRoutes.EditURL, "id", r.criteria.GetId())
			editURL += "?" + url.Values{
				"job_template_id": {templateID},
			}.Encode()
			actions = append(actions, types.TableAction{
				Type:        "edit",
				Label:       criteriaLabels.Actions.Edit,
				HxGet:       editURL,
				HxTarget:    "#sheetContent",
				HxSwap:      "innerHTML",
				OnClick:     "lf.Sheet.open()",
				DrawerTitle: criteriaLabels.Actions.Edit,
			})
		}
		if canDelete {
			actions = append(actions, types.TableAction{
				Type:     "delete",
				Action:   "delete",
				Label:    criteriaLabels.Actions.Delete,
				URL:      deps.CriteriaRoutes.DeleteURL + "?return_table=jt-standards-table",
				ItemName: criteriaName,
			})
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
			Actions: actions,
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
		ShowActions: hasActions,
		ShowSort:    false,
		ShowColumns: false,
		ShowDensity: false,
		ShowEntries: false,
		RefreshURL:  route.ResolveURL(deps.Routes.TabActionURL, "id", templateID, "tab", "standards"),
	}
	types.ApplyColumnStyles(pageData.StandardsTable.Columns, pageData.StandardsTable.Rows)
}
