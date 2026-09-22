package action

// rating_description.go — the binding-level rating description → outcome
// narrative bridge (owner decision 2026-09-21).
//
// A template_task_criteria binding in RATING_MODE_NUMERIC_WITH_DESCRIPTION
// carries one description per scale band. When a grader records a numeric value
// for such a cell, the matching description is SNAPSHOTTED into the outcome's
// determination_note — the one canonical narrative field the per-cell drawer
// already edits — so the grid cell stays a bare number and the wording travels
// with the recorded grade (a later edit of the criterion description never
// rewrites history).
//
// The description set is read from the same server-derived MINE matrix that
// authorizes the write; the POST body never supplies it. The typed value
// columns are untouched: text_value is never written here.

import (
	"math"
	"strconv"
	"strings"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// ratingMatcher is one resolved band: an exact numeric match or a half-open
// [min, max) range, plus the binding's wording for it.
type ratingMatcher struct {
	kind        enums.ScaleKind
	min, max    *float64
	match       *float64
	description string
}

// ratingDescriptions is the description set for one column. A nil/empty value
// means the binding is not in numeric-with-description mode (or carries no
// usable rows) and describe always returns "".
type ratingDescriptions []ratingMatcher

// ratingDescriptionsFromColumn projects a matrix column's descriptions. Only a
// binding explicitly in NUMERIC_WITH_DESCRIPTION mode yields matchers; blank
// wording and unsupported scale kinds are dropped (never guessed).
func ratingDescriptionsFromColumn(col *matrixpb.CriterionColumn) ratingDescriptions {
	if col == nil || col.GetRatingMode() != enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION {
		return nil
	}
	out := make(ratingDescriptions, 0, len(col.GetRatingDescriptions()))
	for _, d := range col.GetRatingDescriptions() {
		text := strings.TrimSpace(d.GetDescription())
		if d == nil || text == "" {
			continue
		}
		m := ratingMatcher{kind: d.GetScaleKind(), min: d.InputMin, max: d.InputMax, description: text}
		switch d.GetScaleKind() {
		case enums.ScaleKind_SCALE_KIND_EXACT_MAP:
			raw := strings.TrimSpace(d.GetInputMatch())
			v, err := strconv.ParseFloat(raw, 64)
			if raw == "" || err != nil {
				continue
			}
			m.match = &v
		case enums.ScaleKind_SCALE_KIND_RANGE_MAP:
		default:
			continue
		}
		out = append(out, m)
	}
	return out
}

// describe returns the description for v: an exact match wins, then the first
// half-open range containing v. "" when nothing matches.
func (r ratingDescriptions) describe(v float64) string {
	for _, m := range r {
		if m.kind == enums.ScaleKind_SCALE_KIND_EXACT_MAP && m.match != nil && math.Abs(*m.match-v) < 1e-9 {
			return m.description
		}
	}
	for _, m := range r {
		if m.kind != enums.ScaleKind_SCALE_KIND_RANGE_MAP {
			continue
		}
		if (m.min == nil || v >= *m.min) && (m.max == nil || v < *m.max) {
			return m.description
		}
	}
	return ""
}

// enabled reports whether the column can produce descriptions at all.
func (r ratingDescriptions) enabled() bool { return len(r) > 0 }
