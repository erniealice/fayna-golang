// Package detail provides the job_task detail page with tab strip:
// info | activities | attachments | history.
package detail

import (
	"context"
	"fmt"
	"log"

	operation "github.com/erniealice/fayna-golang/domain/operation"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// PageData holds the data for the job_task detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Task            map[string]any
	Labels          operation.JobTaskLabels
	ActiveTab       string
	TabItems        []pyeza.TabItem
	ActivitiesTable *types.TableConfig
	AttachmentTable *types.TableConfig
	// Audit history
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// taskToMap converts a JobTask protobuf message to a template-friendly map.
func taskToMap(t *jobtaskpb.JobTask) map[string]any {
	status := "pending"
	statusVariant := "warning"
	switch t.GetStatus() {
	case jobtaskpb.TaskStatus_TASK_STATUS_IN_PROGRESS:
		status = "in_progress"
		statusVariant = "info"
	case jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED:
		status = "completed"
		statusVariant = "success"
	case jobtaskpb.TaskStatus_TASK_STATUS_SKIPPED:
		status = "skipped"
		statusVariant = "default"
	case jobtaskpb.TaskStatus_TASK_STATUS_HOLD:
		status = "hold"
		statusVariant = "warning"
	case jobtaskpb.TaskStatus_TASK_STATUS_REWORK:
		status = "rework"
		statusVariant = "danger"
	}

	phaseName := ""
	if p := t.GetJobPhase(); p != nil {
		phaseName = p.GetName()
	}
	assignedTo := ""
	if t.AssignedTo != nil {
		assignedTo = *t.AssignedTo
	}
	resourceID := ""
	if t.ResourceId != nil {
		resourceID = *t.ResourceId
	}
	templateTaskID := ""
	if t.TemplateTaskId != nil {
		templateTaskID = *t.TemplateTaskId
	}
	plannedQty := float64(0)
	if t.PlannedQuantity != nil {
		plannedQty = *t.PlannedQuantity
	}
	completedQty := float64(0)
	if t.CompletedQuantity != nil {
		completedQty = *t.CompletedQuantity
	}
	percentComplete := float64(0)
	if t.PercentComplete != nil {
		percentComplete = *t.PercentComplete
	}
	allowParallel := false
	if t.AllowParallel != nil {
		allowParallel = *t.AllowParallel
	}

	return map[string]any{
		"id":                   t.GetId(),
		"name":                 t.GetName(),
		"job_phase_id":         t.GetJobPhaseId(),
		"phase_name":           phaseName,
		"step_order":           fmt.Sprintf("%d", t.GetStepOrder()),
		"status":               status,
		"status_variant":       statusVariant,
		"is_ad_hoc":            t.GetIsAdHoc(),
		"assigned_to":          assignedTo,
		"resource_id":          resourceID,
		"template_task_id":     templateTaskID,
		"planned_quantity":     fmt.Sprintf("%.2f", plannedQty),
		"completed_quantity":   fmt.Sprintf("%.2f", completedQty),
		"percent_complete":     fmt.Sprintf("%.0f", percentComplete),
		"allow_parallel":       allowParallel,
		"actual_start_string":  t.GetActualStartString(),
		"actual_end_string":    t.GetActualEndString(),
		"date_created_string":  t.GetDateCreatedString(),
		"date_modified_string": t.GetDateModifiedString(),
	}
}

func buildTabItems(l operation.JobTaskLabels, id string, routes operation.JobTaskRoutes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "activities", Label: l.Tabs.Activities, Href: base + "?tab=activities", HxGet: action + "activities", Icon: "icon-activity"},
		{Key: "attachments", Label: l.Tabs.Attachments, Href: base + "?tab=attachments", HxGet: action + "attachments", Icon: "icon-paperclip"},
		{Key: "history", Label: l.Tabs.History, Href: base + "?tab=history", HxGet: action + "history", Icon: "icon-clock"},
	}
}

// NewView creates the job_task detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// 2026-05-14 permission-gates P2a.
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_task", "read") {
			return view.Forbidden("job_task:read")
		}
		_ = perms

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}

		if deps.ReadJobTask == nil {
			return view.Error(fmt.Errorf("read task not available"))
		}
		resp, err := deps.ReadJobTask(ctx, &jobtaskpb.ReadJobTaskRequest{
			Data: &jobtaskpb.JobTask{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job task %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load task: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("task not found"))
		}
		t := data[0]
		taskMap := taskToMap(t)

		name, _ := taskMap["name"].(string)
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
				HeaderIcon:     "icon-check-square",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-task-detail-content",
			Task:            taskMap,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		loadTabData(ctx, deps, pageData, t, activeTab, viewCtx)

		return view.OK("job-task-detail", pageData)
	})
}

// NewTabAction creates the tab partial action for the job_task detail page.
// Returns only the tab content partial for HTMX tab switching.
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		if deps.ReadJobTask == nil {
			return view.Error(fmt.Errorf("read task not available"))
		}
		resp, err := deps.ReadJobTask(ctx, &jobtaskpb.ReadJobTaskRequest{
			Data: &jobtaskpb.JobTask{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job task %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load task: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			return view.Error(fmt.Errorf("task not found"))
		}
		t := data[0]
		taskMap := taskToMap(t)

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Task:      taskMap,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		loadTabData(ctx, deps, pageData, t, tab, viewCtx)

		templateName := "jt-tab-" + tab
		if tab == "attachments" {
			templateName = "attachment-tab"
		}
		if tab == "history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}

func loadTabData(ctx context.Context, deps *DetailViewDeps, pageData *PageData, t *jobtaskpb.JobTask, activeTab string, viewCtx *view.ViewContext) {
	id := t.GetId()
	switch activeTab {
	case "info":
		// all info fields available from taskToMap

	case "activities":
		loadActivitiesTab(ctx, deps, pageData, t)

	case "attachments":
		if deps.ListAttachments != nil {
			cfg := attachmentConfig(deps)
			resp, err := deps.ListAttachments(ctx, cfg.EntityType, id)
			if err != nil {
				log.Printf("Failed to list attachments for task %s: %v", id, err)
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
