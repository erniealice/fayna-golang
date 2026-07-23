// Package form — options.go holds pure-function option builders for the
// job_template_phase drawer form.
// No Deps, no repo imports — pure functions only.
package form

import (
	"context"
	"sort"

	"github.com/erniealice/pyeza-golang/types"

	scoringschemepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_scheme"
)

// BuildScoringSchemeOptions calls the narrow ListScoringSchemes closure and
// returns select options for the Scoring/Grading Scheme picker. A nil
// closure or a failed call yields an empty slice — the field stays optional.
func BuildScoringSchemeOptions(ctx context.Context, listFn func(context.Context, *scoringschemepb.ListScoringSchemesRequest) (*scoringschemepb.ListScoringSchemesResponse, error), selected string) []types.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &scoringschemepb.ListScoringSchemesRequest{})
	if err != nil || resp == nil {
		return nil
	}
	items := resp.GetData()
	opts := make([]types.SelectOption, 0, len(items))
	for _, s := range items {
		opts = append(opts, types.SelectOption{
			Value:    s.GetId(),
			Label:    s.GetName(),
			Selected: s.GetId() == selected,
		})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}
