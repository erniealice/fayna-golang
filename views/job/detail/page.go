package detail

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

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
	subscriptionpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription"
)

// BudgetTask is a per-task row in the budget snapshot (hours-only for v1;
// money is deferred until resource→PriceProduct→bill_rate chain is wired).
//
// TODO(Wave3): replace Hours with Rate+Subtotal once the resource→PriceProduct
// bill_rate lookup is available via deps (referenced in job_template_task proto
// but not yet exposed in fayna detail deps).
type BudgetTask struct {
	Name  string
	Hours float64 // estimated_duration_minutes / 60
}

// BudgetPhase is a per-phase section in the budget snapshot.
type BudgetPhase struct {
	Name       string
	Tasks      []BudgetTask
	PhaseHours float64 // sum of task hours
}

// BudgetSnapshot is the v1 template-derived budget view-model.
// HasBudget=false when job_template_id is empty/nil.
type BudgetSnapshot struct {
	Phases     []BudgetPhase
	TotalHours float64
	HasBudget  bool
}

// ActualsRow is one row in the actuals rollup table.
type ActualsRow struct {
	EntryType  string
	Count      int32
	TotalCost  string // formatted display value (centavos ÷ 100)
	Currency   string
}

// PageData holds the data for the job detail page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Job             map[string]any
	Labels          fayna.JobLabels
	ActiveTab       string
	TabItems        []pyeza.TabItem
	PhasesTable     *types.TableConfig
	ActivitiesTable *types.TableConfig
	// 2026-04-29 milestone-billing plan §5/§6 — Activities tab additions:
	// per-row mini list (with billable-status badge + edit CTA) and the add
	// CTA URL. Both empty when the JobActivityRoutes deps are not wired.
	ActivitiesList    []ActivityRow
	JobActivityLabels fayna.JobActivityLabels
	AddActivityURL    string
	EditActivityURL   string
	SettlementTable *types.TableConfig
	OutcomesTable   *types.TableConfig
	AttachmentTable *types.TableConfig
	// PhasesList is the per-phase mini-list rendered above the denormalized
	// task table on the Phases tab. Drives the "Mark Complete" CTA that
	// flips JobPhase.status to COMPLETED. 2026-04-29 milestone-billing §4.
	PhasesList []PhaseRow
	// Audit history tab
	AuditEntries    []auditlog.AuditEntryView
	AuditHasNext    bool
	AuditNextCursor string
	AuditHistoryURL string

	// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 — origin
	// breadcrumb. Populated only when Job.origin_type = SUBSCRIPTION and
	// the consuming app supplied SubscriptionDetailURL via deps.
	OriginSubscriptionShown bool
	OriginSubscriptionURL   string
	OriginSubscriptionCode  string

	// Budget tab — v1 hours-per-phase rollup from JobTemplate.
	Budget BudgetSnapshot

	// Actuals tab — cost rollup from GetJobActivityRollup.
	ActualsRows       []ActualsRow
	ActualsGrandTotal string // formatted display value (centavos ÷ 100)
	ActualsCurrency   string
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

		// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 — populate
		// the subscription-origin breadcrumb when applicable. Best-effort:
		// nil deps or read failures keep the breadcrumb hidden.
		applyOriginSubscriptionData(ctx, deps, pageData, data[0])

		// Load tab-specific data
		loadTabData(ctx, deps, pageData, id, activeTab)
		if activeTab == "audit-history" {
			loadAuditHistoryTab(ctx, deps, pageData, id, viewCtx.QueryParams["cursor"])
		}

		return view.OK("job-detail", pageData)
	})
}

// cmpLabelOrDefault returns s if non-empty, otherwise returns fallback.
// Used for tab labels that don't yet have a lyngua key
// (Budget, Actuals, History — lyngua sweep is P7).
func cmpLabelOrDefault(s, fallback string) string {
	if s != "" {
		return s
	}
	return fallback
}

// buildTabItems returns the visible tab strip for the job detail page.
// Tab order: info | phases | activities | budget | actuals | attachments | audit-history.
// NOTE: settlement and outcomes are intentionally excluded from the visible
// tab strip (user request). Their loader files (settlement.go) remain
// compiled and callable from loadTabData. They will be folded into actuals
// in a later phase.
func buildTabItems(l fayna.JobLabels, id string, routes fayna.JobRoutes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "info", Label: l.Tabs.Info, Href: base + "?tab=info", HxGet: action + "info", Icon: "icon-info"},
		{Key: "phases", Label: l.Tabs.Phases, Href: base + "?tab=phases", HxGet: action + "phases", Icon: "icon-list"},
		{Key: "activities", Label: l.Tabs.Activities, Href: base + "?tab=activities", HxGet: action + "activities", Icon: "icon-clock"},
		{Key: "budget", Label: cmpLabelOrDefault("", "Budget"), Href: base + "?tab=budget", HxGet: action + "budget", Icon: "icon-target"},
		{Key: "actuals", Label: cmpLabelOrDefault("", "Actuals"), Href: base + "?tab=actuals", HxGet: action + "actuals", Icon: "icon-trending-up"},
		{Key: "attachments", Label: l.Tabs.Attachments, Href: base + "?tab=attachments", HxGet: action + "attachments", Icon: "icon-paperclip"},
		{Key: "audit-history", Label: cmpLabelOrDefault("", "History"), Href: base + "?tab=audit-history", HxGet: action + "audit-history", Icon: "icon-clock"},
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
	case "budget":
		templateID, _ := pageData.Job["job_template_id"].(string)
		loadBudgetTab(ctx, deps, pageData, id, templateID)
	case "actuals":
		loadActualsTab(ctx, deps, pageData, id)
	// settlement and outcomes are excluded from the visible tab strip but their
	// loaders remain here to keep the files compiling. They will be folded into
	// the actuals tab in a later phase.
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
		// budget and actuals follow the standard "job-tab-{tab}" naming
		// (job-tab-budget, job-tab-actuals) — no special case needed.
		return view.OK(templateName, pageData)
	})
}

// applyOriginSubscriptionData populates the subscription-origin breadcrumb
// fields on PageData when the Job was spawned from a Subscription. Best-
// effort: silent no-op when origin_type is not SUBSCRIPTION, when origin_id
// is empty, when ReadSubscription is nil, or when the subscription read
// fails.
//
// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4.
func applyOriginSubscriptionData(ctx context.Context, deps *DetailViewDeps, pageData *PageData, j *jobpb.Job) {
	if j == nil || j.GetOriginType() != enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION {
		return
	}
	subscriptionID := j.GetOriginId()
	if subscriptionID == "" || deps.SubscriptionDetailURL == "" {
		return
	}
	pageData.OriginSubscriptionShown = true
	pageData.OriginSubscriptionURL = strings.ReplaceAll(deps.SubscriptionDetailURL, "{id}", subscriptionID)
	if deps.ReadSubscription != nil {
		if resp, err := deps.ReadSubscription(ctx, &subscriptionpb.ReadSubscriptionRequest{
			Data: &subscriptionpb.Subscription{Id: subscriptionID},
		}); err == nil && resp != nil && len(resp.GetData()) > 0 {
			if code := resp.GetData()[0].GetCode(); code != "" {
				pageData.OriginSubscriptionCode = code
			}
		}
	}
	if pageData.OriginSubscriptionCode == "" {
		pageData.OriginSubscriptionCode = subscriptionID
	}
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
