// Package form holds the template-facing data shape and option builders for
// the rating description set product plan (AY setup) Relink drawer.
package form

import (
	"context"
	"sort"

	pyeza "github.com/erniealice/pyeza-golang/types"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	productplanpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product_plan"
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
)

// PickerOption is one offering/set <select> entry.
type PickerOption struct {
	ID   string
	Name string
}

// SchedulePickerOption is one academic-year (price_schedule) entry.
// Active/DateTimeStart are carried through so list/page.go's loadSchedules
// can replicate its existing "most recent active schedule" default rule.
type SchedulePickerOption struct {
	ID            string
	Name          string
	Active        bool
	DateTimeStart int64
}

// FormPageData carries every picker option the AY selector and Relink
// drawer need, sourced from a page-data use case authorized under
// rating_description_set_product_plan:read instead of separate
// product_plan:list / price_schedule:list / rating_description_set:list
// grants (codex-review-impl2.out.md finding #6 — live-confirmed:
// "loadSchedules swallows the permission error and presents no academic
// year"). Declared in this subpackage for the same import-cycle-avoidance
// reason as the sibling packages' form.FormPageData.
type FormPageData struct {
	ProductPlans   []PickerOption
	PriceSchedules []SchedulePickerOption
	PublishedSets  []PickerOption
}

// BuildProductPlanOptionsFromPageData converts FormPageData.ProductPlans
// into select options — the page-data-sourced counterpart to
// BuildProductPlanOptions below.
func BuildProductPlanOptionsFromPageData(data *FormPageData, selected string) []pyeza.SelectOption {
	if data == nil {
		return nil
	}
	opts := make([]pyeza.SelectOption, 0, len(data.ProductPlans))
	for _, pp := range data.ProductPlans {
		opts = append(opts, pyeza.SelectOption{Value: pp.ID, Label: pp.Name, Selected: pp.ID == selected})
	}
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// BuildPublishedSetOptionsFromPageData converts FormPageData.PublishedSets
// into select options — the page-data-sourced counterpart to
// BuildPublishedSetOptions below. The response already contains PUBLISHED
// sets only (server-side filtered).
func BuildPublishedSetOptionsFromPageData(data *FormPageData, selected string) []pyeza.SelectOption {
	if data == nil {
		return nil
	}
	opts := make([]pyeza.SelectOption, 0, len(data.PublishedSets))
	for _, s := range data.PublishedSets {
		opts = append(opts, pyeza.SelectOption{Value: s.ID, Label: s.Name, Selected: s.ID == selected})
	}
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// Data is the template-facing data shape for the Relink drawer. Used for
// both the first link (ExpectedCurrentLinkID empty) and a replacement
// (ExpectedCurrentLinkID = the current active link id).
type Data struct {
	FormAction             string
	WorkspaceID            string // injected by ViewAdapter for action_workspace_guard
	ProductPlanID          string
	ProductPlanLocked      bool // true when opened from an existing row (offering fixed)
	PriceScheduleID        string
	RatingDescriptionSetID string
	ExpectedCurrentLinkID  string
	Reason                 string
	ProductPlanOptions     []pyeza.SelectOption
	SetOptions             []pyeza.SelectOption
	Labels                 any
	CommonLabels           any
}

// BuildProductPlanOptions lists all product plans (offerings). A nil closure
// or a failed call yields an empty slice — the template falls back to the
// raw-id text input.
func BuildProductPlanOptions(ctx context.Context, listFn func(context.Context, *productplanpb.ListProductPlansRequest) (*productplanpb.ListProductPlansResponse, error), selected string) []pyeza.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &productplanpb.ListProductPlansRequest{})
	if err != nil || resp == nil {
		return nil
	}
	opts := make([]pyeza.SelectOption, 0, len(resp.GetData()))
	for _, pp := range resp.GetData() {
		if pp == nil || !pp.GetActive() {
			continue
		}
		label := pp.GetName()
		if label == "" {
			label = pp.GetId()
		}
		opts = append(opts, pyeza.SelectOption{Value: pp.GetId(), Label: label, Selected: pp.GetId() == selected})
	}
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

// BuildPublishedSetOptions lists only PUBLISHED rating description sets —
// the only status a Relink target may have (espyna's Relink use case
// enforces this server-side too; the picker just avoids offering an
// ineligible choice).
func BuildPublishedSetOptions(ctx context.Context, listFn func(context.Context, *setpb.ListRatingDescriptionSetsRequest) (*setpb.ListRatingDescriptionSetsResponse, error), selected string) []pyeza.SelectOption {
	if listFn == nil {
		return nil
	}
	resp, err := listFn(ctx, &setpb.ListRatingDescriptionSetsRequest{})
	if err != nil || resp == nil {
		return nil
	}
	opts := make([]pyeza.SelectOption, 0, len(resp.GetData()))
	for _, s := range resp.GetData() {
		if s == nil || s.GetVersionStatus() != enums.VersionStatus_VERSION_STATUS_PUBLISHED {
			continue
		}
		label := s.GetName()
		opts = append(opts, pyeza.SelectOption{Value: s.GetId(), Label: label, Selected: s.GetId() == selected})
	}
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}
