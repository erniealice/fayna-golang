package detail

import (
	"context"
	"fmt"
	"log"

	reporting_checkpoint "github.com/erniealice/fayna-golang/domain/operation/reporting_checkpoint"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	checkpointpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/reporting_checkpoint"
)

// PageData holds the data for the reporting checkpoint detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Checkpoint      map[string]any
	Labels          reporting_checkpoint.Labels
	ActiveTab       string
	TabItems        []pyeza.TabItem

	// Audit history tab
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// checkpointToMap converts a ReportingCheckpoint protobuf to a map[string]any for template use.
func checkpointToMap(c *checkpointpb.ReportingCheckpoint) map[string]any {
	return map[string]any{
		"id":                     c.GetId(),
		"label":                  c.GetLabel(),
		"checkpoint_group_id":    c.GetCheckpointGroupId(),
		"role_code":              c.GetRoleCode(),
		"sequence_order":         c.GetSequenceOrder(),
		"version":                c.GetVersion(),
		"version_status":         versionStatusString(c.GetVersionStatus()),
		"version_status_variant": versionStatusVariant(c.GetVersionStatus()),
		"workspace_id":           c.GetWorkspaceId(),
		"period_id":              c.GetPeriodId(),
		"is_terminal":            c.GetIsTerminal(),
		"active":                 c.GetActive(),
		"date_created_string":    c.GetDateCreatedString(),
		"date_modified_string":   c.GetDateModifiedString(),
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

// loadTabData populates tab-specific fields on pageData based on the active tab.
func loadTabData(ctx context.Context, deps *DetailViewDeps, pd *PageData, id string, viewCtx *view.ViewContext) {
	switch pd.ActiveTab {
	case "audit-history":
		if deps.ListAuditHistory != nil {
			cursor := viewCtx.QueryParams["cursor"]
			auditResp, err := deps.ListAuditHistory(ctx, &auditlog.ListAuditRequest{
				EntityType:  "reporting_checkpoint",
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

// NewView creates the reporting checkpoint detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("reporting_checkpoint", "read") {
			return view.Forbidden("reporting_checkpoint:read")
		}
		_ = perms

		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadReportingCheckpoint(ctx, &checkpointpb.ReadReportingCheckpointRequest{
			Data: &checkpointpb.ReportingCheckpoint{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read reporting checkpoint %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load checkpoint: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Reporting checkpoint %s not found", id)
			return view.Error(fmt.Errorf("checkpoint not found"))
		}
		checkpoint := checkpointToMap(data[0])

		label, _ := checkpoint["label"].(string)
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
				HeaderIcon:     "icon-flag",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "reporting-checkpoint-detail-content",
			Checkpoint:      checkpoint,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		return view.OK("reporting-checkpoint-detail", pageData)
	})
}

func buildTabItems(l reporting_checkpoint.Labels, id string, routes reporting_checkpoint.Routes) []pyeza.TabItem {
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

		resp, err := deps.ReadReportingCheckpoint(ctx, &checkpointpb.ReadReportingCheckpointRequest{
			Data: &checkpointpb.ReportingCheckpoint{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read reporting checkpoint %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load checkpoint: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Reporting checkpoint %s not found", id)
			return view.Error(fmt.Errorf("checkpoint not found"))
		}
		checkpoint := checkpointToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Checkpoint: checkpoint,
			Labels:     l,
			ActiveTab:  tab,
			TabItems:   buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, id, viewCtx)

		templateName := "reporting-checkpoint-tab-" + tab
		if tab == "audit-history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}
