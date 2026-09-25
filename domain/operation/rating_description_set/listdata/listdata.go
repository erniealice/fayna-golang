// Package listdata holds the fayna-side DTO for the rating_description_set
// LIST page's enriched rows, sourced from a single espyna page-data use case
// authorized ONCE under rating_description_set:list
// (codex-review-impl3.out.md finding #1 — the list page must not separately
// call ordinary ListScoreScales / ListRatingDescriptionSetEntries /
// ListRatingDescriptionSetProductPlans, each gated by an unrelated
// permission the prescribed roles (per copya.md's locked grant matrix,
// Section Template Manager most narrowly) do not all hold). Declared in
// this subpackage (not the parent rating_description_set package) for the
// same import-cycle-avoidance reason as the sibling form.FormPageData DTOs
// established by fix2-views.
package listdata

import (
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
)

// Row is one enriched rating_description_set list row.
type Row struct {
	Set            *setpb.RatingDescriptionSet
	ScoreScaleName string
	EntryCount     int32
	LinkCount      int32
}

// PageData carries every row the list page needs.
type PageData struct {
	Rows []Row
}
