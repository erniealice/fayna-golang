package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/fayna-golang/domain/operation/score_scale"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	scalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
)

// PageData holds the data for the score scale detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Scale           map[string]any
	Labels          score_scale.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem
}

// scaleToMap converts a ScoreScale protobuf to a map[string]any for template use.
func scaleToMap(s *scalepb.ScoreScale) map[string]any {
	var inputMin, inputMax interface{}
	if s.InputMin != nil {
		inputMin = *s.InputMin
	}
	if s.InputMax != nil {
		inputMax = *s.InputMax
	}
	return map[string]any{
		"id":                     s.GetId(),
		"name":                   s.GetName(),
		"scale_group_id":         s.GetScaleGroupId(),
		"version":                s.GetVersion(),
		"version_status":         versionStatusString(s.GetVersionStatus()),
		"version_status_variant": versionStatusVariant(s.GetVersionStatus()),
		"scale_kind":             scaleKindString(s.GetScaleKind()),
		"scale_kind_variant":     scaleKindVariant(s.GetScaleKind()),
		"input_unit":             s.GetInputUnit(),
		"input_min":              inputMin,
		"input_max":              inputMax,
		"output_unit":            s.GetOutputUnit(),
		"active":                 s.GetActive(),
		"created_by":             s.GetCreatedBy(),
		"date_created_string":    s.GetDateCreatedString(),
		"date_modified_string":   s.GetDateModifiedString(),
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

func scaleKindString(k enums.ScaleKind) string {
	switch k {
	case enums.ScaleKind_SCALE_KIND_RANGE_MAP:
		return "Range Map"
	case enums.ScaleKind_SCALE_KIND_EXACT_MAP:
		return "Exact Map"
	default:
		return "Unspecified"
	}
}

func scaleKindVariant(k enums.ScaleKind) string {
	switch k {
	case enums.ScaleKind_SCALE_KIND_RANGE_MAP:
		return "info"
	case enums.ScaleKind_SCALE_KIND_EXACT_MAP:
		return "default"
	default:
		return "default"
	}
}

// NewView creates the score scale detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale", "read") {
			return view.Forbidden("score_scale:read")
		}
		_ = perms

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadScoreScale(ctx, &scalepb.ReadScoreScaleRequest{
			Data: &scalepb.ScoreScale{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read score scale %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load score scale: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Score scale %s not found", id)
			return view.Error(fmt.Errorf("score scale not found"))
		}
		scale := scaleToMap(data[0])

		name, _ := scale["name"].(string)
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
				HeaderIcon:     "icon-bar-chart-2",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "score-scale-detail-content",
			Scale:           scale,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		return view.OK("score-scale-detail", pageData)
	})
}

func buildTabItems(l score_scale.Labels, id string, routes score_scale.Routes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "history", Label: l.Tabs.History, Href: base + "?tab=history", HxGet: action + "history", Icon: "icon-clock"},
	}
}

// NewTabAction creates the tab action view (partial — returns only the tab content).
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		resp, err := deps.ReadScoreScale(ctx, &scalepb.ReadScoreScaleRequest{
			Data: &scalepb.ScoreScale{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read score scale %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load score scale: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Score scale %s not found", id)
			return view.Error(fmt.Errorf("score scale not found"))
		}
		scale := scaleToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Scale:     scale,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		templateName := "score-scale-tab-" + tab
		return view.OK(templateName, pageData)
	})
}
