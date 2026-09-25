// Package listdata holds the fayna-side DTO for the rating_description_set_
// product_plan (AY setup / "Descriptor Assignments") LIST page's enriched
// data, sourced from a single espyna page-data use case authorized ONCE
// under rating_description_set_product_plan:list
// (codex-review-impl3.out.md finding #1 — the assignment list must not
// separately call ordinary ListProductPlans / ListRatingDescriptionSets,
// each gated by a permission the prescribed roles do not all hold — and
// finding #3 — offering names must be scoped through BOTH parent
// product.workspace_id AND plan.workspace_id, not a generic list).
// Declared in this subpackage for the same import-cycle-avoidance reason as
// the sibling form.FormPageData DTOs established by fix2-views.
package listdata

import (
	linkpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_product_plan"
)

// ScheduleOption is one academic-year (price_schedule) selector entry.
type ScheduleOption struct {
	ID            string
	Name          string
	Active        bool
	DateTimeStart int64
}

// SetInfo is the display name/version of the rating_description_set a link
// points at, at ANY status (a link may point at a DEPRECATED set).
type SetInfo struct {
	Name    string
	Version int32
}

// LinkRow is one enriched assignment-list row.
type LinkRow struct {
	Link            *linkpb.RatingDescriptionSetProductPlan
	ProductPlanName string
	Set             SetInfo
}

// PageData carries every AY option, the selected/default one, and that AY's
// enriched link rows.
type PageData struct {
	Schedules          []ScheduleOption
	SelectedScheduleID string
	Links              []LinkRow
}
