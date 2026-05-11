package fayna

// Three-level routing system for fayna views:
//
// Level 1: Generic defaults from Go consts (this file).
//   DefaultXxxRoutes() constructors return structs populated from the route
//   constants defined in routes.go. These are sensible defaults that work
//   out of the box for any app.
//
// Level 2: Industry-specific overrides via JSON (loaded by consumer apps).
//   Consumer apps can load a JSON config that partially overrides the
//   default routes. Struct fields carry json tags for unmarshalling.
//
// Level 3: App-specific overrides via Go field assignment (optional).
//   After loading defaults and/or JSON, consumer apps can programmatically
//   set individual fields to further customize routing.
//
// Each route struct also exposes a RouteMap() method that returns a
// map[string]string keyed by dot-notation identifiers (e.g. "job.list"),
// useful for template rendering, URL resolution, and debugging.

// JobRoutes holds all route paths for job (operational activity) views and actions.
type JobRoutes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	// DashboardURL — read-only Job dashboard (Phase 3 — Pyeza dashboard block plan).
	DashboardURL string `json:"dashboard_url"`

	ListURL          string `json:"list_url"`
	DetailURL        string `json:"detail_url"`
	AddURL           string `json:"add_url"`
	EditURL          string `json:"edit_url"`
	DeleteURL        string `json:"delete_url"`
	BulkDeleteURL    string `json:"bulk_delete_url"`
	SetStatusURL     string `json:"set_status_url"`
	BulkSetStatusURL string `json:"bulk_set_status_url"`

	TabActionURL string `json:"tab_action_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`

	// Task action routes
	TaskAssignURL string `json:"task_assign_url"`

	// Phase action routes (2026-04-29 milestone-billing plan §4)
	PhaseSetStatusURL string `json:"phase_set_status_url"`

	// Auto-complete search endpoints for the job drawer form (client + location pickers).
	// Accept ?q= and return [{"value":"id","label":"Name"},...] JSON.
	// Registered and served by the fayna block.
	ClientSearchURL   string `json:"client_search_url"`
	LocationSearchURL string `json:"location_search_url"`
}

// DefaultJobRoutes returns a JobRoutes populated from the package-level
// route constants defined in routes.go.
func DefaultJobRoutes() JobRoutes {
	return JobRoutes{
		ActiveNav:    "job",
		ActiveSubNav: "jobs",

		DashboardURL: JobDashboardURL,

		ListURL:          JobListURL,
		DetailURL:        JobDetailURL,
		AddURL:           JobAddURL,
		EditURL:          JobEditURL,
		DeleteURL:        JobDeleteURL,
		BulkDeleteURL:    JobBulkDeleteURL,
		SetStatusURL:     JobSetStatusURL,
		BulkSetStatusURL: JobBulkSetStatusURL,

		TabActionURL: JobTabActionURL,

		AttachmentUploadURL: JobAttachmentUploadURL,
		AttachmentDeleteURL: JobAttachmentDeleteURL,

		TaskAssignURL: JobTaskAssignURL,

		PhaseSetStatusURL: JobPhaseSetStatusURL,

		ClientSearchURL:   JobClientSearchURL,
		LocationSearchURL: JobLocationSearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job routes.
func (r JobRoutes) RouteMap() map[string]string {
	return map[string]string{
		"job.dashboard":       r.DashboardURL,
		"job.list":            r.ListURL,
		"job.detail":          r.DetailURL,
		"job.add":             r.AddURL,
		"job.edit":            r.EditURL,
		"job.delete":          r.DeleteURL,
		"job.bulk_delete":     r.BulkDeleteURL,
		"job.set_status":      r.SetStatusURL,
		"job.bulk_set_status": r.BulkSetStatusURL,

		"job.tab_action": r.TabActionURL,

		"job.attachment.upload": r.AttachmentUploadURL,
		"job.attachment.delete": r.AttachmentDeleteURL,

		"job.task.assign":      r.TaskAssignURL,
		"job.phase.set_status": r.PhaseSetStatusURL,

		"job.search.client":   r.ClientSearchURL,
		"job.search.location": r.LocationSearchURL,
	}
}

// JobPhaseRoutes holds all route paths for the JobPhase standalone module.
// The list page is a power-user/debugging surface with no sidebar entry.
// The detail page is reached via Job detail's Phases tab deep links.
type JobPhaseRoutes struct {
	// Sidebar navigation context — mirrors the Job module (phases anchor to the job nav).
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL          string `json:"list_url"`
	DetailURL        string `json:"detail_url"`
	AddURL           string `json:"add_url"`
	EditURL          string `json:"edit_url"`
	DeleteURL        string `json:"delete_url"`
	BulkDeleteURL    string `json:"bulk_delete_url"`
	SetStatusURL     string `json:"set_status_url"`
	BulkSetStatusURL string `json:"bulk_set_status_url"`

	TabActionURL string `json:"tab_action_url"`

	// ResourceSearchURL — action-mode auto-complete for the resource FK picker.
	ResourceSearchURL string `json:"resource_search_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`

	// JobDetailURL — used for breadcrumb back to the parent job.
	JobDetailURL string `json:"job_detail_url"`
}

// DefaultJobPhaseRoutes returns a JobPhaseRoutes populated from the
// package-level route constants defined in routes.go.
func DefaultJobPhaseRoutes() JobPhaseRoutes {
	return JobPhaseRoutes{
		ActiveNav:    "job",
		ActiveSubNav: "jobs",

		ListURL:          JobPhaseListURL,
		DetailURL:        JobPhaseDetailURL,
		AddURL:           JobPhaseAddURL,
		EditURL:          JobPhaseEditURL,
		DeleteURL:        JobPhaseDeleteURL,
		BulkDeleteURL:    JobPhaseBulkDeleteURL,
		SetStatusURL:     JobPhaseSetStatusURL,
		BulkSetStatusURL: JobPhaseBulkSetStatusURL,

		TabActionURL: JobPhaseTabActionURL,

		ResourceSearchURL: JobPhaseResourceSearchURL,

		AttachmentUploadURL: JobPhaseAttachmentUploadURL,
		AttachmentDeleteURL: JobPhaseAttachmentDeleteURL,

		JobDetailURL: JobDetailURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job phase routes.
func (r JobPhaseRoutes) RouteMap() map[string]string {
	return map[string]string{
		"job_phase.list":              r.ListURL,
		"job_phase.detail":            r.DetailURL,
		"job_phase.add":               r.AddURL,
		"job_phase.edit":              r.EditURL,
		"job_phase.delete":            r.DeleteURL,
		"job_phase.bulk_delete":       r.BulkDeleteURL,
		"job_phase.set_status":        r.SetStatusURL,
		"job_phase.bulk_set_status":   r.BulkSetStatusURL,
		"job_phase.tab_action":        r.TabActionURL,
		"job_phase.search.resource":   r.ResourceSearchURL,
		"job_phase.attachment.upload": r.AttachmentUploadURL,
		"job_phase.attachment.delete": r.AttachmentDeleteURL,
	}
}

// JobTemplateRoutes holds all route paths for job template views and actions.
type JobTemplateRoutes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL          string `json:"list_url"`
	DetailURL        string `json:"detail_url"`
	AddURL           string `json:"add_url"`
	EditURL          string `json:"edit_url"`
	DeleteURL        string `json:"delete_url"`
	BulkDeleteURL    string `json:"bulk_delete_url"`
	SetStatusURL     string `json:"set_status_url"`
	BulkSetStatusURL string `json:"bulk_set_status_url"`

	TabActionURL string `json:"tab_action_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultJobTemplateRoutes returns a JobTemplateRoutes populated from the
// package-level route constants defined in routes.go.
func DefaultJobTemplateRoutes() JobTemplateRoutes {
	return JobTemplateRoutes{
		ActiveNav:    "job",
		ActiveSubNav: "job-templates",

		ListURL:          JobTemplateListURL,
		DetailURL:        JobTemplateDetailURL,
		AddURL:           JobTemplateAddURL,
		EditURL:          JobTemplateEditURL,
		DeleteURL:        JobTemplateDeleteURL,
		BulkDeleteURL:    JobTemplateBulkDeleteURL,
		SetStatusURL:     JobTemplateSetStatusURL,
		BulkSetStatusURL: JobTemplateBulkSetStatusURL,

		TabActionURL: JobTemplateTabActionURL,

		AttachmentUploadURL: JobTemplateAttachmentUploadURL,
		AttachmentDeleteURL: JobTemplateAttachmentDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job template routes.
func (r JobTemplateRoutes) RouteMap() map[string]string {
	return map[string]string{
		"job_template.list":            r.ListURL,
		"job_template.detail":          r.DetailURL,
		"job_template.add":             r.AddURL,
		"job_template.edit":            r.EditURL,
		"job_template.delete":          r.DeleteURL,
		"job_template.bulk_delete":     r.BulkDeleteURL,
		"job_template.set_status":      r.SetStatusURL,
		"job_template.bulk_set_status": r.BulkSetStatusURL,

		"job_template.tab_action": r.TabActionURL,

		"job_template.attachment.upload": r.AttachmentUploadURL,
		"job_template.attachment.delete": r.AttachmentDeleteURL,
	}
}

// JobActivityRoutes holds all route paths for the job activity (timesheet)
// views and actions.
type JobActivityRoutes struct {
	ListURL       string `json:"list_url"`
	DetailURL     string `json:"detail_url"`
	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	BulkDeleteURL string `json:"bulk_delete_url"`
	SubmitURL     string `json:"submit_url"`
	ApproveURL    string `json:"approve_url"`
	RejectURL     string `json:"reject_url"`
	PostURL       string `json:"post_url"`
	ReverseURL    string `json:"reverse_url"`

	BulkGenerateInvoiceURL string `json:"bulk_generate_invoice_url"`

	TabActionURL string `json:"tab_action_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultJobActivityRoutes returns a JobActivityRoutes populated from the
// package-level route constants defined in routes.go.
func DefaultJobActivityRoutes() JobActivityRoutes {
	return JobActivityRoutes{
		ListURL:       JobActivityListURL,
		DetailURL:     JobActivityDetailURL,
		AddURL:        JobActivityAddURL,
		EditURL:       JobActivityEditURL,
		DeleteURL:     JobActivityDeleteURL,
		BulkDeleteURL: JobActivityBulkDeleteURL,
		SubmitURL:     JobActivitySubmitURL,
		ApproveURL:    JobActivityApproveURL,
		RejectURL:     JobActivityRejectURL,
		PostURL:       JobActivityPostURL,
		ReverseURL:    JobActivityReverseURL,

		BulkGenerateInvoiceURL: JobActivityBulkGenerateInvoiceURL,

		TabActionURL: JobActivityTabActionURL,

		AttachmentUploadURL: JobActivityAttachmentUploadURL,
		AttachmentDeleteURL: JobActivityAttachmentDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job activity routes.
func (r JobActivityRoutes) RouteMap() map[string]string {
	return map[string]string{
		"job_activity.list":                  r.ListURL,
		"job_activity.detail":                r.DetailURL,
		"job_activity.add":                   r.AddURL,
		"job_activity.edit":                  r.EditURL,
		"job_activity.delete":                r.DeleteURL,
		"job_activity.bulk_delete":           r.BulkDeleteURL,
		"job_activity.submit":                r.SubmitURL,
		"job_activity.approve":               r.ApproveURL,
		"job_activity.reject":                r.RejectURL,
		"job_activity.post":                  r.PostURL,
		"job_activity.reverse":               r.ReverseURL,
		"job_activity.bulk_generate_invoice": r.BulkGenerateInvoiceURL,
		"job_activity.tab_action":            r.TabActionURL,
		"job_activity.attachment.upload":     r.AttachmentUploadURL,
		"job_activity.attachment.delete":     r.AttachmentDeleteURL,
	}
}

// OutcomeCriteriaRoutes holds all route paths for the outcome criteria (criteria library) views.
type OutcomeCriteriaRoutes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL       string `json:"list_url"`
	DetailURL     string `json:"detail_url"`
	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	BulkDeleteURL string `json:"bulk_delete_url"`

	TabActionURL string `json:"tab_action_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultOutcomeCriteriaRoutes returns an OutcomeCriteriaRoutes populated from
// the package-level route constants defined in routes.go.
func DefaultOutcomeCriteriaRoutes() OutcomeCriteriaRoutes {
	return OutcomeCriteriaRoutes{
		ActiveNav:    "job",
		ActiveSubNav: "criteria",

		ListURL:       OutcomeCriteriaListURL,
		DetailURL:     OutcomeCriteriaDetailURL,
		AddURL:        OutcomeCriteriaAddURL,
		EditURL:       OutcomeCriteriaEditURL,
		DeleteURL:     OutcomeCriteriaDeleteURL,
		BulkDeleteURL: OutcomeCriteriaBulkDeleteURL,

		TabActionURL: OutcomeCriteriaTabActionURL,

		AttachmentUploadURL: OutcomeCriteriaAttachmentUploadURL,
		AttachmentDeleteURL: OutcomeCriteriaAttachmentDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// outcome criteria routes.
func (r OutcomeCriteriaRoutes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_criteria.list":              r.ListURL,
		"outcome_criteria.detail":            r.DetailURL,
		"outcome_criteria.add":               r.AddURL,
		"outcome_criteria.edit":              r.EditURL,
		"outcome_criteria.delete":            r.DeleteURL,
		"outcome_criteria.bulk_delete":       r.BulkDeleteURL,
		"outcome_criteria.tab_action":        r.TabActionURL,
		"outcome_criteria.attachment.upload": r.AttachmentUploadURL,
		"outcome_criteria.attachment.delete": r.AttachmentDeleteURL,
	}
}

// TaskOutcomeRoutes holds all route paths for task outcome (outcome recording) views.
type TaskOutcomeRoutes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`

	TabActionURL string `json:"tab_action_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultTaskOutcomeRoutes returns a TaskOutcomeRoutes populated from the
// package-level route constants defined in routes.go.
func DefaultTaskOutcomeRoutes() TaskOutcomeRoutes {
	return TaskOutcomeRoutes{
		ActiveNav:    "job",
		ActiveSubNav: "outcomes",

		ListURL:   TaskOutcomeListURL,
		DetailURL: TaskOutcomeDetailURL,
		AddURL:    TaskOutcomeAddURL,
		EditURL:   TaskOutcomeEditURL,
		DeleteURL: TaskOutcomeDeleteURL,

		TabActionURL: TaskOutcomeTabActionURL,

		AttachmentUploadURL: TaskOutcomeAttachmentUploadURL,
		AttachmentDeleteURL: TaskOutcomeAttachmentDeleteURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// task outcome routes.
func (r TaskOutcomeRoutes) RouteMap() map[string]string {
	return map[string]string{
		"task_outcome.list":              r.ListURL,
		"task_outcome.detail":            r.DetailURL,
		"task_outcome.add":               r.AddURL,
		"task_outcome.edit":              r.EditURL,
		"task_outcome.delete":            r.DeleteURL,
		"task_outcome.tab_action":        r.TabActionURL,
		"task_outcome.attachment.upload": r.AttachmentUploadURL,
		"task_outcome.attachment.delete": r.AttachmentDeleteURL,
	}
}

// OutcomeSummaryRoutes holds all route paths for outcome summary (report card) views.
type OutcomeSummaryRoutes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	// ListActiveSubNav overrides ActiveSubNav for the standalone list page.
	// Job/phase summary pages highlight "jobs" while the list page highlights "report-cards".
	ListActiveSubNav string `json:"list_active_sub_nav"`

	ListURL         string `json:"list_url"`
	JobSummaryURL   string `json:"job_summary_url"`
	PhaseSummaryURL string `json:"phase_summary_url"`
}

// DefaultOutcomeSummaryRoutes returns an OutcomeSummaryRoutes populated from
// the package-level route constants defined in routes.go.
func DefaultOutcomeSummaryRoutes() OutcomeSummaryRoutes {
	return OutcomeSummaryRoutes{
		ActiveNav:        "job",
		ActiveSubNav:     "jobs",
		ListActiveSubNav: "report-cards",

		ListURL:         OutcomeSummaryListURL,
		JobSummaryURL:   OutcomeSummaryJobURL,
		PhaseSummaryURL: OutcomeSummaryPhaseURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// outcome summary routes.
func (r OutcomeSummaryRoutes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_summary.list":  r.ListURL,
		"outcome_summary.job":   r.JobSummaryURL,
		"outcome_summary.phase": r.PhaseSummaryURL,
	}
}

// JobTemplatePhaseRoutes holds URL patterns for the job_template_phase drawer-only module.
// No list page, no detail page, no sidebar entry.
// Operators reach this only via the JobTemplate detail Phases tab Add/Edit/Delete CTAs.
type JobTemplatePhaseRoutes struct {
	// No ActiveNav/ActiveSubNav — not in sidebar.

	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	BulkDeleteURL string `json:"bulk_delete_url"`

	// ResourceSearchURL — action-mode auto-complete for the resource FK picker.
	// Reuses the job_phase resource search endpoint (same resource entity).
	ResourceSearchURL string `json:"resource_search_url"`
}

// DefaultJobTemplatePhaseRoutes returns a JobTemplatePhaseRoutes populated from
// the package-level route constants defined in routes.go.
func DefaultJobTemplatePhaseRoutes() JobTemplatePhaseRoutes {
	return JobTemplatePhaseRoutes{
		AddURL:        JobTemplatePhaseAddURL,
		EditURL:       JobTemplatePhaseEditURL,
		DeleteURL:     JobTemplatePhaseDeleteURL,
		BulkDeleteURL: JobTemplatePhaseBulkDeleteURL,

		ResourceSearchURL: JobTemplatePhaseResourceSearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job template phase routes.
func (r JobTemplatePhaseRoutes) RouteMap() map[string]string {
	return map[string]string{
		"job_template_phase.add":             r.AddURL,
		"job_template_phase.edit":            r.EditURL,
		"job_template_phase.delete":          r.DeleteURL,
		"job_template_phase.bulk_delete":     r.BulkDeleteURL,
		"job_template_phase.search.resource": r.ResourceSearchURL,
	}
}

// JobTemplateTaskRoutes holds URL patterns for the job_template_task drawer-only module.
// No list page, no detail page, no sidebar entry.
// Operators reach this only via the JobTemplate detail Tasks tab Add/Edit/Delete CTAs.
type JobTemplateTaskRoutes struct {
	// No ActiveNav/ActiveSubNav — not in sidebar.

	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	BulkDeleteURL string `json:"bulk_delete_url"`

	// ResourceSearchURL — action-mode auto-complete for the resource FK picker.
	// Reuses the job_phase resource search endpoint (same resource entity).
	ResourceSearchURL string `json:"resource_search_url"`
}

// DefaultJobTemplateTaskRoutes returns a JobTemplateTaskRoutes populated from
// the package-level route constants defined in routes.go.
func DefaultJobTemplateTaskRoutes() JobTemplateTaskRoutes {
	return JobTemplateTaskRoutes{
		AddURL:        JobTemplateTaskAddURL,
		EditURL:       JobTemplateTaskEditURL,
		DeleteURL:     JobTemplateTaskDeleteURL,
		BulkDeleteURL: JobTemplateTaskBulkDeleteURL,

		ResourceSearchURL: JobTemplateTaskResourceSearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job template task routes.
func (r JobTemplateTaskRoutes) RouteMap() map[string]string {
	return map[string]string{
		"job_template_task.add":             r.AddURL,
		"job_template_task.edit":            r.EditURL,
		"job_template_task.delete":          r.DeleteURL,
		"job_template_task.bulk_delete":     r.BulkDeleteURL,
		"job_template_task.search.resource": r.ResourceSearchURL,
	}
}

// ActivityLaborRoutes holds URL patterns for the activity labor sibling module.
// ActiveNav is intentionally empty — this module is NOT in the sidebar.
// Entry point is the JobActivity detail page's charge tab (entry_type=LABOR).
type ActivityLaborRoutes struct {
	// No ActiveNav/ActiveSubNav — not in sidebar.
	ActiveNav string `json:"active_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`

	// StaffSearchURL — JSON endpoint for the staff auto-complete picker.
	// Returns [{value, label}]. Empty when staff use case is unavailable.
	StaffSearchURL string `json:"staff_search_url"`
}

// DefaultActivityLaborRoutes returns an ActivityLaborRoutes populated from
// the package-level route constants defined in routes.go.
func DefaultActivityLaborRoutes() ActivityLaborRoutes {
	return ActivityLaborRoutes{
		ActiveNav: "", // not in sidebar

		ListURL:   ActivityLaborListURL,
		DetailURL: ActivityLaborDetailURL,
		AddURL:    ActivityLaborAddURL,
		EditURL:   ActivityLaborEditURL,
		DeleteURL: ActivityLaborDeleteURL,

		StaffSearchURL: ActivityLaborStaffSearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// activity labor routes.
func (r ActivityLaborRoutes) RouteMap() map[string]string {
	return map[string]string{
		"activity_labor.list":         r.ListURL,
		"activity_labor.detail":       r.DetailURL,
		"activity_labor.add":          r.AddURL,
		"activity_labor.edit":         r.EditURL,
		"activity_labor.delete":       r.DeleteURL,
		"activity_labor.search.staff": r.StaffSearchURL,
	}
}

// ActivityMaterialRoutes holds URL patterns for the activity material sibling module.
// ActiveNav is intentionally empty — this module is NOT in the sidebar.
// Entry point is the JobActivity detail page's charge tab (entry_type=MATERIAL).
type ActivityMaterialRoutes struct {
	// No ActiveNav/ActiveSubNav — not in sidebar.
	ActiveNav string `json:"active_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`

	// ProductSearchURL — JSON endpoint for the product auto-complete picker.
	// Returns [{value, label}]. Empty when product use case is unavailable.
	ProductSearchURL string `json:"product_search_url"`

	// LocationSearchURL — JSON endpoint for the location auto-complete picker.
	// Returns [{value, label}]. Empty when location use case is unavailable.
	LocationSearchURL string `json:"location_search_url"`
}

// DefaultActivityMaterialRoutes returns an ActivityMaterialRoutes populated from
// the package-level route constants defined in routes.go.
func DefaultActivityMaterialRoutes() ActivityMaterialRoutes {
	return ActivityMaterialRoutes{
		ActiveNav: "", // not in sidebar

		ListURL:   ActivityMaterialListURL,
		DetailURL: ActivityMaterialDetailURL,
		AddURL:    ActivityMaterialAddURL,
		EditURL:   ActivityMaterialEditURL,
		DeleteURL: ActivityMaterialDeleteURL,

		ProductSearchURL:  ActivityMaterialProductSearchURL,
		LocationSearchURL: ActivityMaterialLocationSearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// activity material routes.
func (r ActivityMaterialRoutes) RouteMap() map[string]string {
	return map[string]string{
		"activity_material.list":            r.ListURL,
		"activity_material.detail":          r.DetailURL,
		"activity_material.add":             r.AddURL,
		"activity_material.edit":            r.EditURL,
		"activity_material.delete":          r.DeleteURL,
		"activity_material.search.product":  r.ProductSearchURL,
		"activity_material.search.location": r.LocationSearchURL,
	}
}

// ActivityExpenseRoutes holds URL patterns for the activity expense sibling module.
// ActiveNav is intentionally empty — this module is NOT in the sidebar.
// Entry point is the JobActivity detail page's charge tab (entry_type=EXPENSE).
type ActivityExpenseRoutes struct {
	// No ActiveNav/ActiveSubNav — not in sidebar.
	ActiveNav string `json:"active_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`

	// ExpenseCategorySearchURL — JSON endpoint for the expense category auto-complete picker.
	// Returns [{value, label}]. Empty when expenditure use case is unavailable.
	ExpenseCategorySearchURL string `json:"expense_category_search_url"`
}

// DefaultActivityExpenseRoutes returns an ActivityExpenseRoutes populated from
// the package-level route constants defined in routes.go.
func DefaultActivityExpenseRoutes() ActivityExpenseRoutes {
	return ActivityExpenseRoutes{
		ActiveNav: "", // not in sidebar

		ListURL:   ActivityExpenseListURL,
		DetailURL: ActivityExpenseDetailURL,
		AddURL:    ActivityExpenseAddURL,
		EditURL:   ActivityExpenseEditURL,
		DeleteURL: ActivityExpenseDeleteURL,

		ExpenseCategorySearchURL: ActivityExpenseExpenseCategorySearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// activity expense routes.
func (r ActivityExpenseRoutes) RouteMap() map[string]string {
	return map[string]string{
		"activity_expense.list":                    r.ListURL,
		"activity_expense.detail":                  r.DetailURL,
		"activity_expense.add":                     r.AddURL,
		"activity_expense.edit":                    r.EditURL,
		"activity_expense.delete":                  r.DeleteURL,
		"activity_expense.search.expense_category": r.ExpenseCategorySearchURL,
	}
}

// JobTaskRoutes holds all route paths for the JobTask standalone module.
// The list page is a power-user/debugging surface with no sidebar entry.
// The detail page is reached via JobPhase detail's Tasks tab deep links.
type JobTaskRoutes struct {
	// Sidebar navigation context — mirrors the Job module (tasks anchor to the job nav).
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL          string `json:"list_url"`
	DetailURL        string `json:"detail_url"`
	AddURL           string `json:"add_url"`
	EditURL          string `json:"edit_url"`
	DeleteURL        string `json:"delete_url"`
	BulkDeleteURL    string `json:"bulk_delete_url"`
	SetStatusURL     string `json:"set_status_url"`
	BulkSetStatusURL string `json:"bulk_set_status_url"`

	TabActionURL string `json:"tab_action_url"`

	// Search endpoints for the task drawer form pickers.
	StaffSearchURL        string `json:"staff_search_url"`
	ResourceSearchURL     string `json:"resource_search_url"`
	TemplateTaskSearchURL string `json:"template_task_search_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`

	// JobPhaseDetailURL — used for breadcrumb back to the parent phase.
	JobPhaseDetailURL string `json:"job_phase_detail_url"`
}

// DefaultJobTaskRoutes returns a JobTaskRoutes populated from the
// package-level route constants defined in routes.go.
func DefaultJobTaskRoutes() JobTaskRoutes {
	return JobTaskRoutes{
		ActiveNav:    "job",
		ActiveSubNav: "jobs",

		ListURL:          JobTaskListURL,
		DetailURL:        JobTaskDetailURL,
		AddURL:           JobTaskAddURL,
		EditURL:          JobTaskEditURL,
		DeleteURL:        JobTaskDeleteURL,
		BulkDeleteURL:    JobTaskBulkDeleteURL,
		SetStatusURL:     JobTaskSetStatusURL,
		BulkSetStatusURL: JobTaskBulkSetStatusURL,

		TabActionURL: JobTaskTabActionURL,

		StaffSearchURL:        JobTaskStaffSearchURL,
		ResourceSearchURL:     JobTaskResourceSearchURL,
		TemplateTaskSearchURL: JobTaskTemplateTaskSearchURL,

		AttachmentUploadURL: JobTaskAttachmentUploadURL,
		AttachmentDeleteURL: JobTaskAttachmentDeleteURL,

		JobPhaseDetailURL: JobPhaseDetailURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job task routes.
func (r JobTaskRoutes) RouteMap() map[string]string {
	return map[string]string{
		"job_task.list":              r.ListURL,
		"job_task.detail":            r.DetailURL,
		"job_task.add":               r.AddURL,
		"job_task.edit":              r.EditURL,
		"job_task.delete":            r.DeleteURL,
		"job_task.bulk_delete":       r.BulkDeleteURL,
		"job_task.set_status":        r.SetStatusURL,
		"job_task.bulk_set_status":   r.BulkSetStatusURL,
		"job_task.tab_action":        r.TabActionURL,
		"job_task.search.staff":      r.StaffSearchURL,
		"job_task.search.resource":   r.ResourceSearchURL,
		"job_task.search.template":   r.TemplateTaskSearchURL,
		"job_task.attachment.upload": r.AttachmentUploadURL,
		"job_task.attachment.delete": r.AttachmentDeleteURL,
	}
}

// FulfillmentRoutes holds URL patterns for fulfillment views.
type FulfillmentRoutes struct {
	// DashboardURL — read-only Fulfillment dashboard (Phase 3 — Pyeza dashboard block plan).
	DashboardURL  string `json:"dashboard_url"`
	ListURL       string `json:"list_url"`
	DetailURL     string `json:"detail_url"`
	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	DeleteURL     string `json:"delete_url"`
	TransitionURL string `json:"transition_url"`
	ReturnURL     string `json:"return_url"`

	TabActionURL string `json:"tab_action_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
}

// DefaultFulfillmentRoutes returns the standard fulfillment route configuration.
func DefaultFulfillmentRoutes() FulfillmentRoutes {
	return FulfillmentRoutes{
		DashboardURL:  FulfillmentDashboardURL,
		ListURL:       FulfillmentListURL,
		DetailURL:     FulfillmentDetailURL,
		AddURL:        FulfillmentAddURL,
		EditURL:       FulfillmentEditURL,
		DeleteURL:     FulfillmentDeleteURL,
		TransitionURL: FulfillmentTransitionURL,
		ReturnURL:     FulfillmentReturnURL,

		TabActionURL: FulfillmentTabActionURL,

		AttachmentUploadURL: FulfillmentAttachmentUploadURL,
		AttachmentDeleteURL: FulfillmentAttachmentDeleteURL,
	}
}

// RouteMap returns all fulfillment routes as a map for template URL resolution.
func (r FulfillmentRoutes) RouteMap() map[string]string {
	return map[string]string{
		"fulfillment.dashboard":         r.DashboardURL,
		"fulfillment.list":              r.ListURL,
		"fulfillment.detail":            r.DetailURL,
		"fulfillment.add":               r.AddURL,
		"fulfillment.edit":              r.EditURL,
		"fulfillment.delete":            r.DeleteURL,
		"fulfillment.transition":        r.TransitionURL,
		"fulfillment.return":            r.ReturnURL,
		"fulfillment.tab_action":        r.TabActionURL,
		"fulfillment.attachment.upload": r.AttachmentUploadURL,
		"fulfillment.attachment.delete": r.AttachmentDeleteURL,
	}
}
