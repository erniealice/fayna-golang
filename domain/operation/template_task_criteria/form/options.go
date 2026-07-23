// Package form — options.go holds pure-function option builders for the
// template_task_criteria drawer form. Each builder takes a narrow
// list-closure signature (not the full ModuleDeps).
package form

import (
	"context"
	"fmt"
	"sort"

	"github.com/erniealice/pyeza-golang/types"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
)

// BuildOutcomeCriteriaOptions calls the narrow ListOutcomeCriterias closure
// and returns select options for the Outcome Criteria picker. A nil closure
// or a failed call yields an empty slice — the template falls back to the
// raw-id text input.
func BuildOutcomeCriteriaOptions(ctx context.Context, listFn func(context.Context, *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error), selected string) []types.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &criteriapb.ListOutcomeCriteriasRequest{})
	if err != nil || resp == nil {
		return nil
	}
	items := resp.GetData()
	opts := make([]types.SelectOption, 0, len(items))
	for _, c := range items {
		opts = append(opts, types.SelectOption{
			Value:    c.GetId(),
			Label:    c.GetName(),
			Selected: c.GetId() == selected,
		})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// BuildTemplateTaskOptions walks a job_template's phases -> tasks (via the
// two narrow list closures) and returns select options for the Job Template
// Task picker, scoped to ONE template. Used when the drawer is opened from a
// job_template detail Standards tab (?job_template_id=) — the Standards tab
// aggregates criteria across every task in the template, so there is no
// single task to pre-lock; the picker is scoped to the template instead
// (Context-discriminator "Template" per ui-drawer-form-template-anatomy).
// A nil closure, a failed call, or no phases yields an empty slice — the
// template falls back to the raw-id text input.
func BuildTemplateTaskOptions(ctx context.Context,
	listPhases func(context.Context, *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error),
	listTasks func(context.Context, *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error),
	templateID string, selected string,
) []types.SelectOption {
	if listPhases == nil || listTasks == nil || templateID == "" {
		return nil
	}
	phaseResp, err := listPhases(ctx, &jobtemplatephasepb.ListByJobTemplateRequest{JobTemplateId: templateID})
	if err != nil || phaseResp == nil {
		return nil
	}
	var opts []types.SelectOption
	for _, p := range phaseResp.GetJobTemplatePhases() {
		taskResp, err := listTasks(ctx, &jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest{JobTemplatePhaseId: p.GetId()})
		if err != nil || taskResp == nil {
			continue
		}
		for _, t := range taskResp.GetJobTemplateTasks() {
			opts = append(opts, types.SelectOption{
				Value:    t.GetId(),
				Label:    fmt.Sprintf("%s — %s", p.GetName(), t.GetName()),
				Selected: t.GetId() == selected,
			})
		}
	}
	return opts
}
