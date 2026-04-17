package detail

import (
	"context"
	"fmt"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"
	lynguaV1 "github.com/erniealice/lyngua/golang/v1"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// PageData holds the data for the job detail page.
type PageData struct {
	types.PageData
	ContentTemplate     string
	Job                 map[string]any
	Labels              fayna.JobLabels
	ActiveTab           string
	TabItems            []pyeza.TabItem
	PhasesTable         *types.TableConfig
	ActivitiesTable     *types.TableConfig
	SettlementTable     *types.TableConfig
	OutcomesTable       *types.TableConfig
	AttachmentTable     *types.TableConfig
	AttachmentUploadURL string
	// Audit history tab
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string
}

// jobToMap converts a Job protobuf to a map[string]any for template use.
func jobToMap(j *jobpb.Job) map[string]any {
	// Build client display name
	clientName := ""
	if c := j.GetClient(); c != nil {
		if u := c.GetUser(); u != nil {
			first := u.GetFirstName()
			last := u.GetLastName()
			if first != "" || last != "" {
				clientName = first + " " + last
			}
		}
		if clientName == "" {
			clientName = c.GetName()
		}
	}

	// Build location name
	locationName := ""
	if loc := j.GetLocation(); loc != nil {
		locationName = loc.GetName()
	}

	return map[string]any{
		"id":                   j.GetId(),
		"name":                 j.GetName(),
		"client_id":            j.GetClientId(),
		"client":               clientName,
		"location_id":          j.GetLocationId(),
		"location":             locationName,
		"job_template_id":      j.GetJobTemplateId(),
		"origin_type":          j.GetOriginType().String(),
		"origin_id":            j.GetOriginId(),
		"demand_type":          j.GetDemandType().String(),
		"fulfillment_type":     j.GetFulfillmentType().String(),
		"cost_flow_type":       j.GetCostFlowType().String(),
		"billing_rule_type":    j.GetBillingRuleType().String(),
		"status":               jobStatusString(j.GetStatus()),
		"status_variant":       jobStatusVariant(j.GetStatus()),
		"approval_status":      approvalStatusString(j.GetApprovalStatus()),
		"approval_variant":     approvalStatusVariant(j.GetApprovalStatus()),
		"posting_status":       j.GetPostingStatus().String(),
		"billing_status":       j.GetBillingStatus().String(),
		"active":               j.GetActive(),
		"created_by":           j.GetCreatedBy(),
		"date_created_string":  j.GetDateCreatedString(),
		"date_modified_string": j.GetDateModifiedString(),
	}
}

func jobStatusString(s enums.JobStatus) string {
	switch s {
	case enums.JobStatus_JOB_STATUS_DRAFT:
		return "draft"
	case enums.JobStatus_JOB_STATUS_PENDING:
		return "pending"
	case enums.JobStatus_JOB_STATUS_ACTIVE:
		return "active"
	case enums.JobStatus_JOB_STATUS_PAUSED:
		return "paused"
	case enums.JobStatus_JOB_STATUS_COMPLETED:
		return "completed"
	case enums.JobStatus_JOB_STATUS_CLOSED:
		return "closed"
	default:
		return "draft"
	}
}

func jobStatusVariant(s enums.JobStatus) string {
	switch s {
	case enums.JobStatus_JOB_STATUS_DRAFT:
		return "default"
	case enums.JobStatus_JOB_STATUS_PENDING:
		return "warning"
	case enums.JobStatus_JOB_STATUS_ACTIVE:
		return "success"
	case enums.JobStatus_JOB_STATUS_PAUSED:
		return "warning"
	case enums.JobStatus_JOB_STATUS_COMPLETED:
		return "info"
	case enums.JobStatus_JOB_STATUS_CLOSED:
		return "default"
	default:
		return "default"
	}
}

func approvalStatusString(s enums.ApprovalStatus) string {
	switch s {
	case enums.ApprovalStatus_APPROVAL_STATUS_NOT_REQUIRED:
		return "not_required"
	case enums.ApprovalStatus_APPROVAL_STATUS_PENDING_APPROVAL:
		return "pending_approval"
	case enums.ApprovalStatus_APPROVAL_STATUS_APPROVED:
		return "approved"
	case enums.ApprovalStatus_APPROVAL_STATUS_REJECTED:
		return "rejected"
	default:
		return "not_required"
	}
}

func approvalStatusVariant(s enums.ApprovalStatus) string {
	switch s {
	case enums.ApprovalStatus_APPROVAL_STATUS_NOT_REQUIRED:
		return "default"
	case enums.ApprovalStatus_APPROVAL_STATUS_PENDING_APPROVAL:
		return "warning"
	case enums.ApprovalStatus_APPROVAL_STATUS_APPROVED:
		return "success"
	case enums.ApprovalStatus_APPROVAL_STATUS_REJECTED:
		return "danger"
	default:
		return "default"
	}
}

// NewView creates the job detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadJob(ctx, &jobpb.ReadJobRequest{
			Data: &jobpb.Job{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load job: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Job %s not found", id)
			return view.Error(fmt.Errorf("job not found"))
		}
		job := jobToMap(data[0])

		jobName, _ := job["name"].(string)
		headerTitle := jobName

		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "info"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          headerTitle,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    headerTitle,
				HeaderSubtitle: l.Detail.PageTitle,
				HeaderIcon:     "icon-briefcase",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-detail-content",
			Job:             job,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		// KB help content
		if viewCtx.Translations != nil {
			if provider, ok := viewCtx.Translations.(*lynguaV1.TranslationProvider); ok {
				if kb, _ := provider.LoadKBIfExists(viewCtx.Lang, viewCtx.BusinessType, "job-detail"); kb != nil {
					pageData.HasHelp = true
					pageData.HelpContent = kb.Body
				}
			}
		}

		// Load tab-specific data
		loadTabData(ctx, deps, pageData, id, activeTab)
		if activeTab == "audit-history" {
			loadAuditHistoryTab(ctx, deps, pageData, id, viewCtx.QueryParams["cursor"])
		}

		return view.OK("job-detail", pageData)
	})
}

func buildTabItems(l fayna.JobLabels, id string, routes fayna.JobRoutes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "phases", Label: l.Tabs.Phases, Href: base + "?tab=phases", HxGet: action + "phases", Icon: "icon-list"},
		{Key: "activities", Label: l.Tabs.Activities, Href: base + "?tab=activities", HxGet: action + "activities", Icon: "icon-clock"},
		{Key: "settlement", Label: l.Tabs.Settlement, Href: base + "?tab=settlement", HxGet: action + "settlement", Icon: "icon-wallet"},
		{Key: "outcomes", Label: l.Tabs.Outcomes, Href: base + "?tab=outcomes", HxGet: action + "outcomes", Icon: "icon-check-circle"},
		{Key: "attachments", Label: l.Tabs.Attachments, Href: base + "?tab=attachments", HxGet: action + "attachments", Icon: "icon-paperclip"},
		{Key: "audit-history", Label: "History", Href: base + "?tab=audit-history", HxGet: action + "audit-history", Icon: "icon-clock"},
	}
}

// loadTabData populates the PageData with tab-specific data.
func loadTabData(ctx context.Context, deps *DetailViewDeps, pageData *PageData, id string, activeTab string) {
	switch activeTab {
	case "info":
		// Job map has all the info fields
	case "phases":
		loadPhasesTab(ctx, deps, pageData, id)
	case "activities":
		loadActivitiesTab(ctx, deps, pageData, id)
	case "settlement":
		loadSettlementTab(ctx, deps, pageData, id)
	case "outcomes":
		// TODO: wire ListTaskOutcomesByJob when available
		pageData.OutcomesTable = nil
	case "attachments":
		if deps.ListAttachments != nil {
			cfg := attachmentConfig(deps)
			resp, err := deps.ListAttachments(ctx, cfg.EntityType, id)
			if err != nil {
				log.Printf("Failed to list attachments for job %s: %v", id, err)
			}
			var items []*attachmentpb.Attachment
			if resp != nil {
				items = resp.GetData()
			}
			pageData.AttachmentTable = attachment.BuildTable(items, cfg, id)
		}
		pageData.AttachmentUploadURL = route.ResolveURL(deps.Routes.AttachmentUploadURL, "id", id)
	}
}

// loadAuditHistoryTab populates the audit history fields on PageData for the job entity.
func loadAuditHistoryTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, id string, cursor string) {
	if deps.ListAuditHistory == nil {
		return
	}
	auditResp, err := deps.ListAuditHistory(ctx, &auditlog.ListAuditRequest{
		EntityType:  "job",
		EntityID:    id,
		Limit:       20,
		CursorToken: cursor,
	})
	if err != nil {
		log.Printf("Failed to load audit history: %v", err)
	}
	if auditResp != nil {
		pageData.AuditEntries = auditResp.Entries
		pageData.AuditHasNext = auditResp.HasNext
		pageData.AuditNextCursor = auditResp.NextCursor
	}
	pageData.AuditHistoryURL = route.ResolveURL(deps.Routes.TabActionURL, "id", id, "tab", "") + "audit-history"
}

// NewTabAction creates the tab action view (partial -- returns only the tab content).
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "info"
		}

		resp, err := deps.ReadJob(ctx, &jobpb.ReadJobRequest{
			Data: &jobpb.Job{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read job %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load job: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Job %s not found", id)
			return view.Error(fmt.Errorf("job not found"))
		}
		job := jobToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Job:       job,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		// Load tab-specific data
		loadTabData(ctx, deps, pageData, id, tab)
		if tab == "audit-history" {
			loadAuditHistoryTab(ctx, deps, pageData, id, viewCtx.QueryParams["cursor"])
		}

		templateName := "job-tab-" + tab
		if tab == "attachments" {
			templateName = "attachment-tab"
		}
		if tab == "audit-history" {
			templateName = "audit-history-tab"
		}
		return view.OK(templateName, pageData)
	})
}

// NewAssignTaskAction handles POST /action/job/{id}/task/{taskId}/assign.
// It reads staff_id from the form, updates JobTask.assigned_to,
// and returns an HTMX response that refreshes the phases tab.
func NewAssignTaskAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		taskID := viewCtx.Request.PathValue("taskId")

		if err := viewCtx.Request.ParseForm(); err != nil {
			log.Printf("Failed to parse form for assign task: %v", err)
			return view.Error(fmt.Errorf("invalid form data"))
		}
		staffID := viewCtx.Request.FormValue("staff_id")
		if staffID == "" {
			log.Printf("Assign task: staff_id is required for task %s on job %s", taskID, id)
			return view.Error(fmt.Errorf("staff_id is required"))
		}

		if deps.UpdateJobTask == nil {
			log.Printf("Assign task: UpdateJobTask dependency is not wired")
			return view.Error(fmt.Errorf("task assignment not available"))
		}

		_, err := deps.UpdateJobTask(ctx, &jobtaskpb.UpdateJobTaskRequest{
			Data: &jobtaskpb.JobTask{
				Id:         taskID,
				AssignedTo: &staffID,
			},
		})
		if err != nil {
			log.Printf("Failed to assign task %s to staff %s: %v", taskID, staffID, err)
			return view.Error(fmt.Errorf("failed to assign task: %w", err))
		}

		// Return HTMX redirect to re-render the phases tab.
		tabActionURL := route.ResolveURL(deps.Routes.TabActionURL, "id", id, "tab", "phases")
		return view.ViewResult{
			StatusCode: http.StatusNoContent,
			Headers: map[string]string{
				"HX-Redirect": tabActionURL,
			},
		}
	})
}
