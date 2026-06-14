package form

import (
	"context"
	"log"
	"sort"

	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
)

// form.go — template-facing data for the polymorphic score-submission drawer
// (DF-1) + its dimension-slot fragment.
//
// Labels is typed `any` to avoid an import cycle between this sub-package and
// the parent evaluation package (templates read .Labels.* via reflection).
// Deps holds ONLY the narrow closures the form needs (the template-item loader)
// plus the slot route — it does NOT pull in the full ModuleDeps. The IDOR /
// servicing gates live in the parent action package + the espyna closures; the
// form is a pure read/render contract.

// Routes is the narrow route contract the slot fragment needs.
type Routes struct {
	DimensionSlotURL string
}

// Deps holds the narrow closures + config the form builder needs.
type Deps struct {
	ListEvaluationTemplateItems func(ctx context.Context, req *itempb.ListEvaluationTemplateItemsRequest) (*itempb.ListEvaluationTemplateItemsResponse, error)
	Routes                      Routes
	Labels                      any
}

// GetFormDataInput carries the GET-render parameters for the drawer.
type GetFormDataInput struct {
	FormAction  string
	IsEdit      bool
	ID          string
	TemplateID  string
	SeatID      string
	PeriodStart string
	PeriodEnd   string
	Narrative   string
}

// Data is the outer drawer-form data shape (scaffold + slot).
type Data struct {
	FormAction       string
	WorkspaceID      string // injected by the ViewAdapter (action_workspace_guard)
	IsEdit           bool
	ID               string
	TemplateID       string
	SeatID           string
	PeriodStart      string
	PeriodEnd        string
	Narrative        string
	DimensionSlotURL string
	Slot             SlotData
	Labels           any
	CommonLabels     any
}

// SlotData is the polymorphic dimension-slot fragment data (§A.2).
type SlotData struct {
	TemplateID string
	Dimensions []Dimension
	Labels     any
	CommonLabels any
}

// Dimension is one rubric row. Kind drives the template branch across all 5
// criteria_type values; the answer maps to a different evaluation_response
// column per branch (numeric/pass_fail/categorical/text).
type Dimension struct {
	CriteriaID    string  // OutcomeCriteria id — input name = "dim_" + CriteriaID
	Label         string  // question_label ?? criteria label
	Prompt        string  // optional question_prompt helper
	Kind          string  // NUMERIC_SCORE | NUMERIC_RANGE | PASS_FAIL | CATEGORICAL | TEXT | MULTI_CHECK
	Weight        float64 // weight_override ?? criteria weight (display-only chip)
	Required      bool
	MinScore      int32 // for NUMERIC_RANGE / NUMERIC_SCORE bars
	MaxScore      int32
	Options       []DimensionOption // CATEGORICAL select options
	Scored        bool              // numeric kinds are scored; others show "(not scored)"
	InputTestID   string            // dim-input-{criteria_id}
	BarTestID     string            // dim-bar-{criteria_id} (numeric kinds)
}

// DimensionOption is a CATEGORICAL select option.
type DimensionOption struct {
	Value string
	Label string
}

// GetFormData builds the outer drawer Data (loads the active template's items
// when a template is pre-selected). CommonLabels is injected by the ViewAdapter.
func GetFormData(ctx context.Context, deps *Deps, in GetFormDataInput) *Data {
	d := &Data{
		FormAction:       in.FormAction,
		IsEdit:           in.IsEdit,
		ID:               in.ID,
		TemplateID:       in.TemplateID,
		SeatID:           in.SeatID,
		PeriodStart:      in.PeriodStart,
		PeriodEnd:        in.PeriodEnd,
		Narrative:        in.Narrative,
		DimensionSlotURL: deps.Routes.DimensionSlotURL,
		Labels:           deps.Labels,
	}
	d.Slot = *GetDimensionSlot(ctx, deps, in.TemplateID)
	return d
}

// GetDimensionSlot builds the polymorphic slot fragment for a template. Items
// are loaded via the injected closure and rendered in sequence_order. When no
// template is selected the slot is empty (the picker drives the HTMX swap).
func GetDimensionSlot(ctx context.Context, deps *Deps, templateID string) *SlotData {
	slot := &SlotData{TemplateID: templateID, Labels: deps.Labels}
	if templateID == "" || deps.ListEvaluationTemplateItems == nil {
		return slot
	}

	resp, err := deps.ListEvaluationTemplateItems(ctx, &itempb.ListEvaluationTemplateItemsRequest{})
	if err != nil {
		log.Printf("Failed to list template items for %s: %v", templateID, err)
		return slot
	}

	items := []*itempb.EvaluationTemplateItem{}
	for _, it := range resp.GetData() {
		if it.GetEvaluationTemplateId() == templateID && it.GetActive() {
			items = append(items, it)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].GetSequenceOrder() < items[j].GetSequenceOrder()
	})

	for _, it := range items {
		cid := it.GetOutcomeCriteriaId()
		// criteria_type is surfaced from the linked OutcomeCriteria; the item
		// loader resolves it server-side. Until the join is wired we default
		// to a numeric scored kind (the dominant rubric branch); the Integrator
		// overlays the real criteria_type when ListEvaluationTemplateItems is
		// extended to carry it. This keeps the slot render polymorphic-ready.
		dim := Dimension{
			CriteriaID:  cid,
			Label:       it.GetQuestionLabel(),
			Prompt:      it.GetQuestionPrompt(),
			Kind:        "NUMERIC_SCORE",
			Weight:      it.GetWeightOverride(),
			Required:    it.GetRequiredOverride(),
			MinScore:    1,
			MaxScore:    5,
			Scored:      true,
			InputTestID: "dim-input-" + cid,
			BarTestID:   "dim-bar-" + cid,
		}
		slot.Dimensions = append(slot.Dimensions, dim)
	}
	return slot
}
