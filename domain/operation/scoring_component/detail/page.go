package detail

import (
	"context"
	"fmt"
	"log"

	scoring_component "github.com/erniealice/fayna-golang/domain/operation/scoring_component"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	scoringpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component"
)

// PageData holds the data for the scoring component detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Component       map[string]any
	Labels          scoring_component.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem

	// Audit history tab
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// componentToMap converts a ScoringComponent protobuf to a map[string]any for template use.
func componentToMap(c *scoringpb.ScoringComponent) map[string]any {
	return map[string]any{
		"id":                   c.GetId(),
		"scoring_scheme_id":    c.GetScoringSchemeId(),
		"code":                 c.GetCode(),
		"label":                c.GetLabel(),
		"weight":               c.GetWeight(),
		"sequence_order":       c.GetSequenceOrder(),
		"parent_component_id":  c.GetParentComponentId(),
		"active":               c.GetActive(),
		"date_created_string":  c.GetDateCreatedString(),
		"date_modified_string": c.GetDateModifiedString(),
	}
}

// loadTabData populates tab-specific data on pageData based on the active tab.
func loadTabData(ctx context.Context, deps *DetailViewDeps, pd *PageData, id string, viewCtx *view.ViewContext) {
	switch pd.ActiveTab {
	case "audit-history":
		if deps.ListAuditHistory != nil {
			cursor := viewCtx.QueryParams["cursor"]
			auditResp, err := deps.ListAuditHistory(ctx, &auditlog.ListAuditRequest{
				EntityType:  "scoring_component",
				EntityID:    id,
				Limit:       20,
				CursorToken: cursor,
			})
			if err != nil {
				log.Printf("Failed to load audit history: %v", err)
			}
			if auditResp != nil {
				pd.AuditEntries = auditResp.Entries
				pd.AuditHasNext = auditResp.HasNext
				pd.AuditNextCursor = auditResp.NextCursor
			}
		}
		pd.AuditHistoryURL = route.ResolveURL(deps.Routes.TabActionURL, "id", id, "tab", "") + "audit-history"
	}
}

// NewView creates the scoring component detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component", "read") {
			return view.Forbidden("scoring_component:read")
		}
		_ = perms

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadScoringComponent(ctx, &scoringpb.ReadScoringComponentRequest{
			Data: &scoringpb.ScoringComponent{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read scoring component %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load scoring component: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Scoring component %s not found", id)
			return view.Error(fmt.Errorf("scoring component not found"))
		}
		component := componentToMap(data[0])

		label, _ := component["label"].(string)
		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          label,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    label,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-bar-chart-2",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "scoring-component-detail-content",
			Component:       component,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		return view.OK("scoring-component-detail", pageData)
	})
}

func buildTabItems(l scoring_component.Labels, id string, routes scoring_component.Routes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "audit-history", Label: l.Tabs.History, Href: base + "?tab=audit-history", HxGet: action + "audit-history", Icon: "icon-clock"},
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

		resp, err := deps.ReadScoringComponent(ctx, &scoringpb.ReadScoringComponentRequest{
			Data: &scoringpb.ScoringComponent{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read scoring component %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load scoring component: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Scoring component %s not found", id)
			return view.Error(fmt.Errorf("scoring component not found"))
		}
		component := componentToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Component: component,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		templateName := "scoring-component-tab-" + tab
		if tab == "audit-history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}
