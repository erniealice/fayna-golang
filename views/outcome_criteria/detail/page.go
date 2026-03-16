package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
)

// PageData holds the data for the outcome criteria detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Criteria        map[string]any
	Labels          fayna.OutcomeCriteriaLabels
	ActiveTab       string
	TabItems        []pyeza.TabItem
}

// criteriaToMap converts an OutcomeCriteria protobuf to a map[string]any for template use.
func criteriaToMap(c *criteriapb.OutcomeCriteria) map[string]any {
	return map[string]any{
		"id":                    c.GetId(),
		"name":                  c.GetName(),
		"description":           c.GetDescription(),
		"criteria_type":         criteriaTypeString(c.GetCriteriaType()),
		"scope":                 criteriaScopeString(c.GetScope()),
		"version":               c.GetVersion(),
		"version_status":        versionStatusString(c.GetVersionStatus()),
		"version_status_variant": versionStatusVariant(c.GetVersionStatus()),
		"required":              c.GetRequired(),
		"weight":                c.GetWeight(),
		"active":                c.GetActive(),
		"created_by":            c.GetCreatedBy(),
		"date_created_string":   c.GetDateCreatedString(),
		"date_modified_string":  c.GetDateModifiedString(),
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

func criteriaScopeString(s enums.CriteriaScope) string {
	switch s {
	case enums.CriteriaScope_CRITERIA_SCOPE_SYSTEM:
		return "System"
	case enums.CriteriaScope_CRITERIA_SCOPE_INDUSTRY:
		return "Industry"
	case enums.CriteriaScope_CRITERIA_SCOPE_WORKSPACE:
		return "Workspace"
	default:
		return "Unspecified"
	}
}

func versionStatusString(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "draft"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "published"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "deprecated"
	default:
		return "draft"
	}
}

func versionStatusVariant(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "default"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "success"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "warning"
	default:
		return "default"
	}
}

// NewView creates the outcome criteria detail view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadOutcomeCriteria(ctx, &criteriapb.ReadOutcomeCriteriaRequest{
			Data: &criteriapb.OutcomeCriteria{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read outcome criteria %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load criterion: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Outcome criteria %s not found", id)
			return view.Error(fmt.Errorf("criterion not found"))
		}
		criteria := criteriaToMap(data[0])

		name, _ := criteria["name"].(string)
		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          name,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    name,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-check-circle",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "outcome-criteria-detail-content",
			Criteria:        criteria,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		return view.OK("outcome-criteria-detail", pageData)
	})
}

func buildTabItems(l fayna.OutcomeCriteriaLabels, id string, routes fayna.OutcomeCriteriaRoutes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "thresholds", Label: l.Tabs.Thresholds, Href: base + "?tab=thresholds", HxGet: action + "thresholds", Icon: "icon-sliders"},
		{Key: "options", Label: l.Tabs.Options, Href: base + "?tab=options", HxGet: action + "options", Icon: "icon-list"},
		{Key: "templates", Label: l.Tabs.Templates, Href: base + "?tab=templates", HxGet: action + "templates", Icon: "icon-file"},
		{Key: "versions", Label: l.Tabs.Versions, Href: base + "?tab=versions", HxGet: action + "versions", Icon: "icon-clock"},
	}
}

// NewTabAction creates the tab action view (partial — returns only the tab content).
func NewTabAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		resp, err := deps.ReadOutcomeCriteria(ctx, &criteriapb.ReadOutcomeCriteriaRequest{
			Data: &criteriapb.OutcomeCriteria{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read outcome criteria %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load criterion: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Outcome criteria %s not found", id)
			return view.Error(fmt.Errorf("criterion not found"))
		}
		criteria := criteriaToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Criteria:  criteria,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		templateName := "outcome-criteria-tab-" + tab
		return view.OK(templateName, pageData)
	})
}
