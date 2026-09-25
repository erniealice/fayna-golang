// Package form holds the template-facing data shape and option builders for
// the rating description set entry drawer.
package form

import (
	"context"
	"sort"
	"strings"

	pyeza "github.com/erniealice/pyeza-golang/types"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	scalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// PickerOption is one criterion <select> entry.
type PickerOption struct {
	ID   string
	Name string
}

// BandPickerOption is one level (score_scale_band) row/entry. See the
// espyna-side BandFormPickerOption doc for why BandRole is carried through
// unfiltered (shared by both the entry drawer's Level picker and the set
// detail page's matrix rows).
type BandPickerOption struct {
	ID            string
	Name          string
	ScoreScaleID  string
	BandRole      string
	SequenceOrder int32
}

// EntrySnapshot is the pre-fill shape for an existing entry, returned by the
// Drawer page-data variant's Edit path — loaded through a parent-update-
// authorized read (rating_description_set:update), never the entry's own
// :read gate (codex-review-impl4.out.md "Update-only drawer" finding).
type EntrySnapshot struct {
	ID                     string
	RatingDescriptionSetID string
	OutcomeCriteriaID      string
	ScoreScaleBandID       string
	Description            string
}

// FormPageData carries every picker option the entry drawer (and the set
// detail page's matrix, which reuses this same page data) needs, sourced
// from a page-data use case authorized under rating_description_set:read
// instead of separate outcome_criteria:list / score_scale_band:list grants
// (codex-review-impl2.out.md finding #6). Declared in this subpackage (not
// the parent rating_description_set_entry package) for the same
// import-cycle-avoidance reason as rating_description_set/form.FormPageData.
//
// Entry and ScoreScaleID are populated ONLY by the Drawer variant
// (deps.GetFormPageData, :update-gated) when called with an entry/set id —
// the read-only detail-matrix variant never sets them (its FormPageData is
// always the workspace-wide criteria/bands picker source, unchanged).
type FormPageData struct {
	Criteria []PickerOption
	Bands    []BandPickerOption
	// ScoreScaleID is the parent set's score_scale_id, resolved server-side
	// under the SAME :update authorization — the Drawer's Level picker no
	// longer needs a separate rating_description_set:read-gated
	// ReadRatingDescriptionSet call to learn it.
	ScoreScaleID string
	// Entry is nil for Add, or when an Edit's entry id does not resolve to a
	// row (absence — the caller treats a nil Entry as Not Found, not a
	// technical failure).
	Entry *EntrySnapshot
}

// BuildCriteriaOptionsFromPageData converts FormPageData.Criteria into
// select options — the page-data-sourced counterpart to BuildCriteriaOptions
// below (finding #6: preferred over the ordinary ListOutcomeCriterias call).
func BuildCriteriaOptionsFromPageData(data *FormPageData, selected string) []pyeza.SelectOption {
	if data == nil {
		return nil
	}
	opts := make([]pyeza.SelectOption, 0, len(data.Criteria))
	for _, c := range data.Criteria {
		opts = append(opts, pyeza.SelectOption{Value: c.ID, Label: c.Name, Selected: c.ID == selected})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// BuildBandOptionsFromPageData filters FormPageData.Bands to one
// score_scale, excludes band_role == "no_description" (Q20 — level 0 is
// never a valid entry target) and sorts by sequence_order — the
// page-data-sourced counterpart to BuildBandOptions below.
func BuildBandOptionsFromPageData(data *FormPageData, scoreScaleID, selected string) []pyeza.SelectOption {
	if data == nil || scoreScaleID == "" {
		return nil
	}
	var bands []BandPickerOption
	for _, b := range data.Bands {
		if b.ScoreScaleID != scoreScaleID || strings.EqualFold(b.BandRole, "no_description") {
			continue
		}
		bands = append(bands, b)
	}
	sort.SliceStable(bands, func(i, j int) bool { return bands[i].SequenceOrder < bands[j].SequenceOrder })
	opts := make([]pyeza.SelectOption, 0, len(bands))
	for _, b := range bands {
		opts = append(opts, pyeza.SelectOption{Value: b.ID, Label: b.Name, Selected: b.ID == selected})
	}
	return opts
}

// Data is the template-facing data shape for the entry Add/Edit drawer.
type Data struct {
	FormAction             string
	WorkspaceID            string // injected by ViewAdapter for action_workspace_guard
	IsEdit                 bool
	ID                     string
	RatingDescriptionSetID string
	OutcomeCriteriaID      string
	ScoreScaleBandID       string
	Description            string
	CriteriaOptions        []pyeza.SelectOption
	BandOptions            []pyeza.SelectOption
	Labels                 any
	CommonLabels           any
}

// BuildCriteriaOptions calls the narrow ListOutcomeCriterias closure and
// returns select options for the Criterion picker. A nil closure or a
// failed call yields an empty slice — the template falls back to the raw-id
// text input (established fallback convention).
func BuildCriteriaOptions(ctx context.Context, listFn func(context.Context, *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error), selected string) []pyeza.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &criteriapb.ListOutcomeCriteriasRequest{})
	if err != nil || resp == nil {
		return nil
	}
	opts := make([]pyeza.SelectOption, 0, len(resp.GetData()))
	for _, c := range resp.GetData() {
		if c == nil {
			continue
		}
		opts = append(opts, pyeza.SelectOption{Value: c.GetId(), Label: c.GetName(), Selected: c.GetId() == selected})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// BuildBandOptions lists the active bands of one score_scale (the parent
// set's scale), excluding band_role == "no_description" (Q20 — level 0 is
// never a valid entry target), sorted by sequence_order.
func BuildBandOptions(ctx context.Context, listFn func(context.Context, *scalebandpb.ListScoreScaleBandsRequest) (*scalebandpb.ListScoreScaleBandsResponse, error), scoreScaleID, selected string) []pyeza.SelectOption {
	if listFn == nil || scoreScaleID == "" {
		return nil
	}
	resp, err := listFn(ctx, &scalebandpb.ListScoreScaleBandsRequest{})
	if err != nil || resp == nil {
		return nil
	}
	type band struct {
		id, label string
		seq       int32
	}
	var bands []band
	for _, b := range resp.GetData() {
		if b == nil || !b.GetActive() || b.GetScoreScaleId() != scoreScaleID {
			continue
		}
		if strings.EqualFold(b.GetBandRole(), "no_description") {
			continue
		}
		bands = append(bands, band{id: b.GetId(), label: b.GetOutputLabel(), seq: b.GetSequenceOrder()})
	}
	sort.SliceStable(bands, func(i, j int) bool { return bands[i].seq < bands[j].seq })
	opts := make([]pyeza.SelectOption, 0, len(bands))
	for _, b := range bands {
		opts = append(opts, pyeza.SelectOption{Value: b.id, Label: b.label, Selected: b.id == selected})
	}
	return opts
}
