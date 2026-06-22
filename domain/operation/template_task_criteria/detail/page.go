package detail

import (
	"context"
	"fmt"
	"log"

	ttc "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
)

// PageData holds the data for the template task criteria detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Record          map[string]any
	Labels          ttc.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem

	// Audit history tab
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// recordToMap converts a TemplateTaskCriteria protobuf to a map[string]any for template use.
func recordToMap(c *ttcpb.TemplateTaskCriteria) map[string]any {
	activeStatus := "inactive"
	if c.GetActive() {
		activeStatus = "active"
	}
	statusVariant := "warning"
	if c.GetActive() {
		statusVariant = "success"
	}
	requiredOverride := ""
	if c.RequiredOverride != nil {
		if c.GetRequiredOverride() {
			requiredOverride = "true"
		} else {
			requiredOverride = "false"
		}
	}
	return map[string]any{
		"id":                    c.GetId(),
		"job_template_task_id":  c.GetJobTemplateTaskId(),
		"outcome_criteria_id":   c.GetOutcomeCriteriaId(),
		"sequence_order":        c.GetSequenceOrder(),
		"required_override":     requiredOverride,
		"active":                c.GetActive(),
		"status":                activeStatus,
		"status_variant":        statusVariant,
		"date_created_string":   c.GetDateCreatedString(),
	}
}

// loadTabData populates tab-specific fields on pageData based on the active tab.
func loadTabData(ctx context.Context, deps *DetailViewDeps, pd *PageData, id string, viewCtx *view.ViewContext) {
	switch pd.ActiveTab {
	case "audit-history":
		if deps.ListAuditHistory != nil {
			cursor := viewCtx.QueryParams["cursor"]
			auditResp, err := deps.ListAuditHistory(ctx, &auditlog.ListAuditRequest{
				EntityType:  "template_task_criteria",
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

// NewView creates the template task criteria detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("template_task_criteria", "read") {
			return view.Forbidden("template_task_criteria:read")
		}
		_ = perms

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadTemplateTaskCriteria(ctx, &ttcpb.ReadTemplateTaskCriteriaRequest{
			Data: &ttcpb.TemplateTaskCriteria{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read template task criteria %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load template task criteria: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Template task criteria %s not found", id)
			return view.Error(fmt.Errorf("template task criteria not found"))
		}
		record := recordToMap(data[0])

		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          id,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    id,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-link",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "template-task-criteria-detail-content",
			Record:          record,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		return view.OK("template-task-criteria-detail", pageData)
	})
}

func buildTabItems(l ttc.Labels, id string, routes ttc.Routes) []pyeza.TabItem {
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

		resp, err := deps.ReadTemplateTaskCriteria(ctx, &ttcpb.ReadTemplateTaskCriteriaRequest{
			Data: &ttcpb.TemplateTaskCriteria{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read template task criteria %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load template task criteria: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Template task criteria %s not found", id)
			return view.Error(fmt.Errorf("template task criteria not found"))
		}
		record := recordToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Record:    record,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		templateName := "template-task-criteria-tab-" + tab
		if tab == "audit-history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}
