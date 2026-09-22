// Package form — options.go holds pure-function option builders for the
// template_task_criteria drawer form. Each builder takes a narrow
// list-closure signature (not the full ModuleDeps).
package form

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/erniealice/pyeza-golang/types"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	scorescalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
	scorescalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
	ttcrdpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria_rating_description"
)

const (
	RatingModeUnspecified            = "RATING_MODE_UNSPECIFIED"
	RatingModeStandard               = "RATING_MODE_STANDARD"
	RatingModeNumericWithDescription = "RATING_MODE_NUMERIC_WITH_DESCRIPTION"
)

// RatingModeLabels carries translated labels for the two authoring choices.
// The form package stays independent from the parent module's label type.
type RatingModeLabels struct {
	Standard               string
	NumericWithDescription string
}

// BuildRatingModeOptions returns the stable authoring choices. UNSPECIFIED is
// deliberately not shown as a user choice: it is the legacy/null wire state,
// while Standard is the explicit equivalent.
func BuildRatingModeOptions(selected string, labels RatingModeLabels) []types.SelectOption {
	if selected == "" || selected == RatingModeUnspecified {
		selected = RatingModeStandard
	}
	if labels.Standard == "" {
		labels.Standard = "Standard"
	}
	if labels.NumericWithDescription == "" {
		labels.NumericWithDescription = "Numeric with description"
	}
	return []types.SelectOption{
		{Value: RatingModeStandard, Label: labels.Standard, Selected: selected == RatingModeStandard},
		{Value: RatingModeNumericWithDescription, Label: labels.NumericWithDescription, Selected: selected == RatingModeNumericWithDescription},
	}
}

// BuildScoreScaleOptions lists active reusable scales. Filtering by numeric
// compatibility remains a server-side invariant; the form is only a picker.
func BuildScoreScaleOptions(ctx context.Context, listFn func(context.Context, *scorescalepb.ListScoreScalesRequest) (*scorescalepb.ListScoreScalesResponse, error), selected string) []types.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &scorescalepb.ListScoreScalesRequest{})
	if err != nil || resp == nil {
		return nil
	}
	opts := make([]types.SelectOption, 0, len(resp.GetData()))
	for _, scale := range resp.GetData() {
		if scale == nil || !scale.GetActive() {
			continue
		}
		label := scale.GetName()
		if label == "" {
			label = scale.GetId()
		}
		if scale.GetInputUnit() != "" {
			label += " (" + scale.GetInputUnit() + ")"
		}
		opts = append(opts, types.SelectOption{Value: scale.GetId(), Label: label, Selected: scale.GetId() == selected})
	}
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// BuildScoreScaleNameMap returns the same active scale set used by the select,
// keyed for the band rows' hierarchy label.
func BuildScoreScaleNameMap(ctx context.Context, listFn func(context.Context, *scorescalepb.ListScoreScalesRequest) (*scorescalepb.ListScoreScalesResponse, error)) map[string]string {
	names := map[string]string{}
	if listFn == nil {
		return names
	}
	resp, err := listFn(ctx, &scorescalepb.ListScoreScalesRequest{})
	if err != nil || resp == nil {
		return names
	}
	for _, scale := range resp.GetData() {
		if scale == nil || !scale.GetActive() {
			continue
		}
		name := scale.GetName()
		if name == "" {
			name = scale.GetId()
		}
		names[scale.GetId()] = name
	}
	return names
}

// BuildRatingDescriptionRows joins reusable active bands with the existing
// binding-specific child rows. All scales are returned so the browser can
// switch the visible band group without another request; POST handling only
// persists rows belonging to the selected scale.
func BuildRatingDescriptionRows(
	ctx context.Context,
	listBands func(context.Context, *scorescalebandpb.ListScoreScaleBandsRequest) (*scorescalebandpb.ListScoreScaleBandsResponse, error),
	listDescriptions func(context.Context, *ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaRequest) (*ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaResponse, error),
	parentID string,
	scaleNames map[string]string,
) []RatingDescriptionRow {
	if listBands == nil {
		return nil
	}
	bandResp, err := listBands(ctx, &scorescalebandpb.ListScoreScaleBandsRequest{})
	if err != nil || bandResp == nil {
		return nil
	}
	descriptionByBand := map[string]*ttcrdpb.TemplateTaskCriteriaRatingDescription{}
	if listDescriptions != nil && parentID != "" {
		if resp, err := listDescriptions(ctx, &ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaRequest{TemplateTaskCriteriaId: parentID}); err == nil && resp != nil {
			for _, row := range resp.GetTemplateTaskCriteriaRatingDescriptions() {
				if row != nil && row.GetActive() {
					descriptionByBand[row.GetScoreScaleBandId()] = row
				}
			}
		}
	}
	rows := make([]RatingDescriptionRow, 0, len(bandResp.GetData()))
	for _, band := range bandResp.GetData() {
		if band == nil || !band.GetActive() {
			continue
		}
		row := RatingDescriptionRow{
			BandID:     band.GetId(),
			ScaleID:    band.GetScoreScaleId(),
			ScaleName:  scaleNames[band.GetScoreScaleId()],
			BandLabel:  band.GetOutputLabel(),
			InputLabel: formatBandInput(band),
		}
		if existing := descriptionByBand[band.GetId()]; existing != nil {
			row.ID = existing.GetId()
			row.Description = existing.GetDescription()
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].ScaleName != rows[j].ScaleName {
			return rows[i].ScaleName < rows[j].ScaleName
		}
		return rows[i].InputLabel < rows[j].InputLabel
	})
	return rows
}

func formatBandInput(band *scorescalebandpb.ScoreScaleBand) string {
	if match := strings.TrimSpace(band.GetInputMatch()); match != "" {
		return "= " + match
	}
	var parts []string
	if band.InputMin != nil {
		parts = append(parts, strconv.FormatFloat(band.GetInputMin(), 'f', -1, 64))
	} else {
		parts = append(parts, "−∞")
	}
	if band.InputMax != nil {
		parts = append(parts, strconv.FormatFloat(band.GetInputMax(), 'f', -1, 64))
	} else {
		parts = append(parts, "+∞")
	}
	return strings.Join(parts, " ≤ score < ")
}

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
