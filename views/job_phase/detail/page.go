// Package detail provides the job_phase detail page with tab strip:
// info | tasks | activities | attachments | history.
package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
)

// PageData holds the data for the job_phase detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Phase           map[string]any
	Labels          fayna.JobPhaseLabels
	ActiveTab       string
	TabItems        []pyeza.TabItem
	TasksTable      *types.TableConfig
	ActivitiesTable *types.TableConfig
	AttachmentTable *types.TableConfig
	// Audit history
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// phaseToMap converts a JobPhase protobuf message to a template-friendly map.
func phaseToMap(p *jobphasepb.JobPhase) map[string]any {
	status := "pending"
	statusVariant := "warning"
	switch p.GetStatus() {
	case jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE:
		status = "active"
		statusVariant = "success"
	case jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED:
		status = "completed"
		statusVariant = "info"
	}

	jobName := ""
	if j := p.GetJob(); j != nil {
		jobName = j.GetName()
	}

	resourceID := ""
	if p.ResourceId != nil {
		resourceID = *p.ResourceId
	}
	setupMinutes := int32(0)
	if p.SetupMinutes != nil {
		setupMinutes = *p.SetupMinutes
	}
	runMinutesPerUnit := float64(0)
	if p.RunMinutesPerUnit != nil {
		runMinutesPerUnit = *p.RunMinutesPerUnit
	}

	return map[string]any{
		"id":                   p.GetId(),
		"name":                 p.GetName(),
		"job_id":               p.GetJobId(),
		"job_name":             jobName,
		"phase_order":          fmt.Sprintf("%d", p.GetPhaseOrder()),
		"status":               status,
		"status_variant":       statusVariant,
		"resource_id":          resourceID,
		"setup_minutes":        fmt.Sprintf("%d", setupMinutes),
		"run_minutes_per_unit": fmt.Sprintf("%.2f", runMinutesPerUnit),
		"planned_start_string": p.GetPlannedStartString(),
		"planned_end_string":   p.GetPlannedEndString(),
		"actual_start_string":  p.GetActualStartString(),
		"actual_end_string":    p.GetActualEndString(),
		"date_created_string":  p.GetDateCreatedString(),
		"date_modified_string": p.GetDateModifiedString(),
	}
}

func buildTabItems(l fayna.JobPhaseLabels, id string, routes fayna.JobPhaseRoutes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "tasks", Label: l.Tabs.Tasks, Href: base + "?tab=tasks", HxGet: action + "tasks", Icon: "icon-check-square"},
		{Key: "activities", Label: l.Tabs.Activities, Href: base + "?tab=activities", HxGet: action + "activities", Icon: "icon-activity"},
		{Key: "attachments", Label: l.Tabs.Attachments, Href: base + "?tab=attachments", HxGet: action + "attachments", Icon: "icon-paperclip"},
		{Key: "history", Label: l.Tabs.History, Href: base + "?tab=history", HxGet: action + "history", Icon: "icon-clock"},
	}
}

// NewView creates the job_phase detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}

		if deps.ReadJobPhase == nil {
			return view.Error(fmt.Errorf("read phase not available"))
		}
		resp, err := deps.ReadJobPhase(ctx, &jobphasepb.ReadJobPhaseRequest{
			Data: &jobphasepb.JobPhase{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job phase %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load phase: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("phase not found"))
		}
		p := data[0]
		phaseMap := phaseToMap(p)

		name, _ := phaseMap["name"].(string)
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
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    name,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-layers",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-phase-detail-content",
			Phase:           phaseMap,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		loadTabData(ctx, deps, pageData, p, activeTab, viewCtx)

		return view.OK("job-phase-detail", pageData)
	})
}

// NewTabAction creates the tab partial action for the job_phase detail page.
// Returns only the tab content partial for HTMX tab switching.
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		if deps.ReadJobPhase == nil {
			return view.Error(fmt.Errorf("read phase not available"))
		}
		resp, err := deps.ReadJobPhase(ctx, &jobphasepb.ReadJobPhaseRequest{
			Data: &jobphasepb.JobPhase{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job phase %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load phase: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("phase not found"))
		}
		p := data[0]
		phaseMap := phaseToMap(p)

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Phase:     phaseMap,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, p, tab, viewCtx)

		templateName := "jp-tab-" + tab
		if tab == "attachments" {
			templateName = "attachment-tab"
		}
		if tab == "history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}

func loadTabData(ctx context.Context, deps *DetailViewDeps, pageData *PageData, p *jobphasepb.JobPhase, activeTab string, viewCtx *view.ViewContext) {
	id := p.GetId()
	switch activeTab {
	case "info":
		// all info fields available from phaseToMap

	case "tasks":
		loadTasksTab(ctx, deps, pageData, id)

	case "activities":
		loadActivitiesTab(ctx, deps, pageData, p)

	case "attachments":
		if deps.ListAttachments != nil {
			cfg := attachmentConfig(deps)
			resp, err := deps.ListAttachments(ctx, cfg.EntityType, id)
			if err != nil {
				log.Printf("Failed to list attachments for phase %s: %v", id, err)
			}
			var items []*attachmentpb.Attachment
			if resp != nil {
				items = resp.GetData()
			}
			pageData.AttachmentTable = attachment.BuildTable(items, cfg, id)
		}

	case "history":
		loadAuditHistoryTab(ctx, deps, pageData, id, viewCtx.QueryParams["cursor"])
	}
}
