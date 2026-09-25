// Package form holds the template-facing data shape and option builders for
// the rating description set Add drawer.
package form

import (
	"context"
	"sort"

	pyeza "github.com/erniealice/pyeza-golang/types"

	scalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
)

// ScalePickerOption is one score_scale <select> entry for the Add-set
// drawer.
type ScalePickerOption struct {
	ID   string
	Name string
}

// FormPageData carries every picker option the rating_description_set
// drawers need, sourced from a page-data use case authorized under
// rating_description_set:create instead of a separate score_scale:list
// grant (codex-review-impl2.out.md finding #6 — live-confirmed: "nobody
// holds score_scale:list ... Add-set Scale picker empty"). Declared in this
// subpackage (not the parent rating_description_set package) so both the
// parent's deps.go and this package's BuildScaleOptionsFromPageData can
// reference it without an import cycle (actions.go, in the parent package,
// already imports this subpackage).
type FormPageData struct {
	ScoreScales []ScalePickerOption
}

// BuildScaleOptionsFromPageData converts FormPageData.ScoreScales into
// select options — the page-data-sourced counterpart to BuildScaleOptions
// below (finding #6: preferred over the ordinary ListScoreScales call).
func BuildScaleOptionsFromPageData(data *FormPageData, selected string) []pyeza.SelectOption {
	if data == nil {
		return nil
	}
	opts := make([]pyeza.SelectOption, 0, len(data.ScoreScales))
	for _, s := range data.ScoreScales {
		label := s.Name
		if label == "" {
			label = s.ID
		}
		opts = append(opts, pyeza.SelectOption{Value: s.ID, Label: label, Selected: s.ID == selected})
	}
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// Data is the template-facing data shape for the Add drawer. MVP scope only
// implements Add (Create) — no Edit (TODO, see sequence log): the set header
// (name/code/scale) is DRAFT-only anyway, and CreateRatingDescriptionSet
// always creates a fresh DRAFT (version 1).
type Data struct {
	FormAction    string
	WorkspaceID   string // injected by ViewAdapter for action_workspace_guard
	Name          string
	Code          string
	ScoreScaleID  string
	SourceRef     string
	ScaleOptions  []pyeza.SelectOption
	Labels        any
	CommonLabels  any
}

// BuildScaleOptions lists active reusable score scales. A nil closure or a
// failed call yields an empty slice — the template falls back to the raw-id
// text input (established fallback convention, e.g.
// template_task_criteria/form/options.go BuildScoreScaleOptions).
func BuildScaleOptions(ctx context.Context, listFn func(context.Context, *scalepb.ListScoreScalesRequest) (*scalepb.ListScoreScalesResponse, error), selected string) []pyeza.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &scalepb.ListScoreScalesRequest{})
	if err != nil || resp == nil {
		return nil
	}
	opts := make([]pyeza.SelectOption, 0, len(resp.GetData()))
	for _, scale := range resp.GetData() {
		if scale == nil || !scale.GetActive() {
			continue
		}
		label := scale.GetName()
		if label == "" {
			label = scale.GetId()
		}
		opts = append(opts, pyeza.SelectOption{Value: scale.GetId(), Label: label, Selected: scale.GetId() == selected})
	}
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}
