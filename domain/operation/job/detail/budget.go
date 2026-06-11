package detail

import (
	"context"
	"log"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplatetaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
)

// loadBudgetTab populates pageData.Budget with the JobTemplate-derived
// phase/task hour plan.
//
// v1 source-of-truth:
//   - read JobTemplate via deps.ReadJobTemplate (to verify it exists)
//   - list JobTemplatePhases by template_id via deps.ListJobTemplatePhasesByTemplate
//   - list JobTemplateTasks by phase_id via deps.ListJobTemplateTasksByPhase
//   - roll up estimated_duration_minutes into hours per phase and grand total
//
// Money is intentionally skipped in v1: resource→PriceProduct→bill_rate
// resolution requires the resource/PriceProduct wiring that is not yet
// exposed in fayna detail deps. The tab shows "Hours per phase" instead.
//
// TODO(Wave3): replace this loader with a JobInputPlan reader. When the
// resource→PriceProduct→bill_rate chain is wired, compute centavos subtotals
// per task and fold into BudgetTask.Rate + BudgetTask.Subtotal.
func loadBudgetTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, _ string, templateID string) {
	if templateID == "" {
		// No template attached — render empty state.
		pageData.Budget = BudgetSnapshot{HasBudget: false}
		return
	}

	if deps.ReadJobTemplate == nil || deps.ListJobTemplatePhasesByTemplate == nil {
		// Deps not wired yet — graceful empty state.
		// TODO(composition): wire ReadJobTemplate + ListJobTemplatePhasesByTemplate
		// in packages/fayna-golang/block/wiring.go wireJobDeps().
		pageData.Budget = BudgetSnapshot{HasBudget: false}
		return
	}

	// Verify the template exists.
	tmplResp, err := deps.ReadJobTemplate(ctx, &jobtemplatepb.ReadJobTemplateRequest{
		Data: &jobtemplatepb.JobTemplate{Id: templateID},
	})
	if err != nil {
		log.Printf("loadBudgetTab: failed to read job template %s: %v", templateID, err)
		pageData.Budget = BudgetSnapshot{HasBudget: false}
		return
	}
	if tmplResp == nil || len(tmplResp.GetData()) == 0 {
		pageData.Budget = BudgetSnapshot{HasBudget: false}
		return
	}

	// List phases for this template.
	phasesResp, err := deps.ListJobTemplatePhasesByTemplate(ctx, &jobtemplatephasepb.ListByJobTemplateRequest{
		JobTemplateId: templateID,
	})
	if err != nil {
		log.Printf("loadBudgetTab: failed to list phases for template %s: %v", templateID, err)
		pageData.Budget = BudgetSnapshot{HasBudget: true}
		return
	}

	var phases []BudgetPhase
	var grandTotalHours float64

	for _, phase := range phasesResp.GetJobTemplatePhases() {
		bp := BudgetPhase{Name: phase.GetName()}

		// List tasks for this phase if dep is wired.
		if deps.ListJobTemplateTasksByPhase != nil {
			tasksResp, err := deps.ListJobTemplateTasksByPhase(ctx, &jobtemplatetaskpb.ListJobTemplateTasksByPhaseRequest{
				JobTemplatePhaseId: phase.GetId(),
			})
			if err != nil {
				log.Printf("loadBudgetTab: failed to list tasks for phase %s: %v", phase.GetId(), err)
			} else {
				for _, task := range tasksResp.GetJobTemplateTasks() {
					hours := 0.0
					if d := task.GetEstimatedDurationMinutes(); d > 0 {
						hours = float64(d) / 60.0
					}
					bp.Tasks = append(bp.Tasks, BudgetTask{
						Name:  task.GetName(),
						Hours: hours,
					})
					bp.PhaseHours += hours
				}
			}
		}

		grandTotalHours += bp.PhaseHours
		phases = append(phases, bp)
	}

	pageData.Budget = BudgetSnapshot{
		Phases:     phases,
		TotalHours: grandTotalHours,
		HasBudget:  true,
	}
}
