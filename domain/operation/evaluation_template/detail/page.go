package detail

import (
	"context"
	"fmt"
	"log"
	"sort"

	evaluation_template "github.com/erniealice/fayna-golang/domain/operation/evaluation_template"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	templatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template"
)

// DetailViewDeps holds view dependencies for the evaluation template detail.
type DetailViewDeps struct {
	Routes                      evaluation_template.Routes
	Labels                      evaluation_template.Labels
	CommonLabels                pyeza.CommonLabels
	TableLabels                 types.TableLabels
	ReadEvaluationTemplate      func(ctx context.Context, req *templatepb.ReadEvaluationTemplateRequest) (*templatepb.ReadEvaluationTemplateResponse, error)
	ListEvaluationTemplateItems func(ctx context.Context, req *itempb.ListEvaluationTemplateItemsRequest) (*itempb.ListEvaluationTemplateItemsResponse, error)
	ListOutcomeCriterias        func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)
	// Item-drawer routes used by the rubric builder Items tab (Add Question /
	// edit / remove). Wired by the integrator from the evaluation_template_item
	// module's routes so the builder mounts its drawer endpoints.
	ItemAddURL    string
	ItemEditURL   string
	ItemRemoveURL string
}

// PageData holds the data for the evaluation template detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Template        map[string]any
	Labels          evaluation_template.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem

	// Items tab — the rubric builder ordered list.
	Items         []map[string]any
	ItemAddURL    string
	ItemRemoveURL string
}

func templateToMap(t *templatepb.EvaluationTemplate) map[string]any {
	return map[string]any{
		"id":                   t.GetId(),
		"name":                 t.GetName(),
		"description":          t.GetDescription(),
		"evaluation_type":      evaluationTypeString(t.GetEvaluationType()),
		"relationship_type":    relationshipTypeString(t.GetRelationshipType()),
		"visibility":           visibilityString(t.GetVisibilityType()),
		"version":              t.GetVersion(),
		"status":               statusString(t.GetStatus()),
		"status_variant":       statusVariant(t.GetStatus()),
		"copied_from_id":       t.GetCopiedFromId(),
		"date_created_string":  t.GetDateCreatedString(),
		"date_modified_string": t.GetDateModifiedString(),
	}
}

// NewView creates the evaluation template detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "read") {
			return view.Forbidden("evaluation_template:read")
		}

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadEvaluationTemplate(ctx, &templatepb.ReadEvaluationTemplateRequest{
			Data: &templatepb.EvaluationTemplate{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read evaluation template %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load template: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("template not found"))
		}
		tmpl := templateToMap(data[0])
		name, _ := tmpl["name"].(string)
		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          name,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    name,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-clipboard",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "evaluation_template-detail-content",
			Template:        tmpl,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, id)

		return view.OK("evaluation_template-detail", pageData)
	})
}

// NewTabAction returns the tab-content partial (HTMX tab-swap).
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("evaluation_template", "read") {
			return view.Forbidden("evaluation_template:read")
		}

		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		resp, err := deps.ReadEvaluationTemplate(ctx, &templatepb.ReadEvaluationTemplateRequest{
			Data: &templatepb.EvaluationTemplate{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read evaluation template %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load template: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("template not found"))
		}

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Template:  templateToMap(data[0]),
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, id)

		return view.OK("evaluation_template-tab-"+tab, pageData)
	})
}

// loadTabData populates tab-specific fields. Only the Items tab does extra work.
func loadTabData(ctx context.Context, deps *DetailViewDeps, pd *PageData, id string) {
	if deps.ItemAddURL != "" {
		pd.ItemAddURL = route.ResolveURL(deps.ItemAddURL, "template_id", id)
	}
	pd.ItemRemoveURL = deps.ItemRemoveURL

	if pd.ActiveTab != "items" {
		return
	}
	if deps.ListEvaluationTemplateItems == nil {
		return
	}
	resp, err := deps.ListEvaluationTemplateItems(ctx, &itempb.ListEvaluationTemplateItemsRequest{})
	if err != nil {
		log.Printf("Failed to list items for template %s: %v", id, err)
		return
	}

	// Criterion type lookup so the builder can surface the response type and
	// the "(not scored)" hint on non-numeric criteria.
	critType := loadCriteriaTypes(ctx, deps)

	var items []*itempb.EvaluationTemplateItem
	for _, it := range resp.GetData() {
		if it.GetEvaluationTemplateId() == id && it.GetActive() {
			items = append(items, it)
		}
	}
	sort.SliceStable(items, func(a, b int) bool {
		return items[a].GetSequenceOrder() < items[b].GetSequenceOrder()
	})

	l := deps.Labels
	for _, it := range items {
		ct := critType[it.GetOutcomeCriteriaId()]
		label := it.GetQuestionLabel()
		weight := ""
		if it.WeightOverride != nil {
			weight = fmt.Sprintf("%g", it.GetWeightOverride())
		}
		pd.Items = append(pd.Items, map[string]any{
			"id":                  it.GetId(),
			"sequence_order":      it.GetSequenceOrder(),
			"outcome_criteria_id": it.GetOutcomeCriteriaId(),
			"question_label":      label,
			"question_prompt":     it.GetQuestionPrompt(),
			"weight":              weight,
			"required":            it.GetRequiredOverride(),
			"criteria_type":       criteriaTypeString(ct),
			"not_scored":          !isNumeric(ct),
			"edit_url":            route.ResolveURL(deps.ItemEditURL, "item_id", it.GetId()),
			"remove_url":          route.ResolveURL(deps.ItemRemoveURL, "item_id", it.GetId()),
			"not_scored_label":    l.Items.NotScored,
		})
	}
}

// loadCriteriaTypes returns criteria_type keyed by criterion id (single query).
func loadCriteriaTypes(ctx context.Context, deps *DetailViewDeps) map[string]enums.CriteriaType {
	out := map[string]enums.CriteriaType{}
	if deps.ListOutcomeCriterias == nil {
		return out
	}
	resp, err := deps.ListOutcomeCriterias(ctx, &criteriapb.ListOutcomeCriteriasRequest{})
	if err != nil {
		log.Printf("Failed to load criteria for rubric builder: %v", err)
		return out
	}
	for _, c := range resp.GetData() {
		out[c.GetId()] = c.GetCriteriaType()
	}
	return out
}

func isNumeric(t enums.CriteriaType) bool {
	return t == enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE ||
		t == enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE
}

func buildTabItems(l evaluation_template.Labels, id string, routes evaluation_template.Routes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "items", Label: l.Tabs.Items, Href: base + "?tab=items", HxGet: action + "items", Icon: "icon-list"},
	}
}

// --- enum string helpers (mirror list/page.go; detail is a separate package) ---

func statusString(s templatepb.EvaluationTemplateStatus) string {
	switch s {
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DRAFT:
		return "draft"
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_ACTIVE:
		return "active"
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DEPRECATED:
		return "deprecated"
	default:
		return "draft"
	}
}

func statusVariant(s templatepb.EvaluationTemplateStatus) string {
	switch s {
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_ACTIVE:
		return "success"
	case templatepb.EvaluationTemplateStatus_EVALUATION_TEMPLATE_STATUS_DEPRECATED:
		return "warning"
	default:
		return "default"
	}
}

func evaluationTypeString(t evalpb.EvaluationType) string {
	switch t {
	case evalpb.EvaluationType_EVALUATION_TYPE_PERFORMANCE_REVIEW:
		return "Performance Review"
	case evalpb.EvaluationType_EVALUATION_TYPE_CSAT:
		return "CSAT"
	case evalpb.EvaluationType_EVALUATION_TYPE_COURSE_EVAL:
		return "Course Eval"
	case evalpb.EvaluationType_EVALUATION_TYPE_VENDOR_SCORECARD:
		return "Vendor Scorecard"
	default:
		return "Unspecified"
	}
}

func relationshipTypeString(t evalpb.RelationshipType) string {
	switch t {
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_CLIENT_TO_ASSOCIATE:
		return "Client -> Associate"
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_STAFF_TO_CLIENT:
		return "Staff -> Client"
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_SELF:
		return "Self"
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_PEER:
		return "Peer"
	case evalpb.RelationshipType_RELATIONSHIP_TYPE_MANAGER:
		return "Manager"
	default:
		return "Unspecified"
	}
}

func visibilityString(v evalpb.VisibilityType) string {
	switch v {
	case evalpb.VisibilityType_VISIBILITY_TYPE_INTERNAL_ONLY:
		return "Internal Only"
	case evalpb.VisibilityType_VISIBILITY_TYPE_VISIBLE_TO_SUBJECT:
		return "Visible to Subject"
	case evalpb.VisibilityType_VISIBILITY_TYPE_VISIBLE_TO_SUBJECT_ANON:
		return "Visible (Anon)"
	default:
		return "Unspecified"
	}
}

func criteriaTypeString(t enums.CriteriaType) string {
	switch t {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE:
		return "Numeric Range"
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return "Numeric Score"
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		return "Pass/Fail"
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return "Categorical"
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		return "Text"
	case enums.CriteriaType_CRITERIA_TYPE_MULTI_CHECK:
		return "Multi-Check"
	default:
		return "Unspecified"
	}
}
