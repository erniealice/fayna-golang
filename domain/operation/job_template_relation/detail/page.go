package detail

import (
	"context"
	"fmt"
	"log"

	jtr "github.com/erniealice/fayna-golang/domain/operation/job_template_relation"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	relationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"
)

// PageData holds the data for the job template relation detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Record          map[string]any
	Labels          jtr.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem

	// Audit history tab
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// recordToMap converts a JobTemplateRelation protobuf to a map[string]any for template use.
func recordToMap(rel *relationpb.JobTemplateRelation) map[string]any {
	activeStatus := "inactive"
	if rel.GetActive() {
		activeStatus = "active"
	}
	statusVariant := "warning"
	if rel.GetActive() {
		statusVariant = "success"
	}
	return map[string]any{
		"id":                  rel.GetId(),
		"parent_template_id":  rel.GetParentTemplateId(),
		"child_template_id":   rel.GetChildTemplateId(),
		"relation_type":       rel.GetRelationType().String(),
		"sequence_order":      rel.GetSequenceOrder(),
		"active":              rel.GetActive(),
		"status":              activeStatus,
		"status_variant":      statusVariant,
		"date_created_string": rel.GetDateCreatedString(),
	}
}

// loadTabData populates tab-specific fields on pageData based on the active tab.
func loadTabData(ctx context.Context, deps *DetailViewDeps, pd *PageData, id string, viewCtx *view.ViewContext) {
	switch pd.ActiveTab {
	case "audit-history":
		if deps.ListAuditHistory != nil {
			cursor := viewCtx.QueryParams["cursor"]
			auditResp, err := deps.ListAuditHistory(ctx, &auditlog.ListAuditRequest{
				EntityType:  "job_template_relation",
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

// NewView creates the job template relation detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_relation", "read") {
			return view.Forbidden("job_template_relation:read")
		}
		_ = perms

		id := viewCtx.Request.PathValue("id")

		if deps.ReadJobTemplateRelation == nil {
			return view.Error(fmt.Errorf("job template relation read not available"))
		}
		resp, err := deps.ReadJobTemplateRelation(ctx, &relationpb.ReadJobTemplateRelationRequest{
			Data: &relationpb.JobTemplateRelation{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job template relation %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load job template relation: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Job template relation %s not found", id)
			return view.Error(fmt.Errorf("job template relation not found"))
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
				HeaderIcon:     "icon-git-branch",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-template-relation-detail-content",
			Record:          record,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		return view.OK("job-template-relation-detail", pageData)
	})
}

func buildTabItems(l jtr.Labels, id string, routes jtr.Routes) []pyeza.TabItem {
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
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_relation", "read") {
			return view.Forbidden("job_template_relation:read")
		}

		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		if deps.ReadJobTemplateRelation == nil {
			return view.Error(fmt.Errorf("job template relation read not available"))
		}
		resp, err := deps.ReadJobTemplateRelation(ctx, &relationpb.ReadJobTemplateRelationRequest{
			Data: &relationpb.JobTemplateRelation{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job template relation %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load job template relation: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Job template relation %s not found", id)
			return view.Error(fmt.Errorf("job template relation not found"))
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

		templateName := "job-template-relation-tab-" + tab
		if tab == "audit-history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}
