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

		"job.task.assign":       r.TaskAssignURL,
		"job.phase.set_status":  r.PhaseSetStatusURL,
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
	ListURL    string `json:"list_url"`
	DetailURL  string `json:"detail_url"`
	AddURL     string `json:"add_url"`
	EditURL    string `json:"edit_url"`
	DeleteURL  string `json:"delete_url"`
	SubmitURL  string `json:"submit_url"`
	ApproveURL string `json:"approve_url"`
	RejectURL  string `json:"reject_url"`
	PostURL    string `json:"post_url"`
	ReverseURL string `json:"reverse_url"`

	BulkGenerateInvoiceURL string `json:"bulk_generate_invoice_url"`
}

// DefaultJobActivityRoutes returns a JobActivityRoutes populated from the
// package-level route constants defined in routes.go.
func DefaultJobActivityRoutes() JobActivityRoutes {
	return JobActivityRoutes{
		ListURL:    JobActivityListURL,
		DetailURL:  JobActivityDetailURL,
		AddURL:     JobActivityAddURL,
		EditURL:    JobActivityEditURL,
		DeleteURL:  JobActivityDeleteURL,
		SubmitURL:  JobActivitySubmitURL,
		ApproveURL: JobActivityApproveURL,
		RejectURL:  JobActivityRejectURL,
		PostURL:    JobActivityPostURL,
		ReverseURL: JobActivityReverseURL,

		BulkGenerateInvoiceURL: JobActivityBulkGenerateInvoiceURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job activity routes.
func (r JobActivityRoutes) RouteMap() map[string]string {
	return map[string]string{
		"job_activity.list":                   r.ListURL,
		"job_activity.detail":                 r.DetailURL,
		"job_activity.add":                    r.AddURL,
		"job_activity.edit":                   r.EditURL,
		"job_activity.delete":                 r.DeleteURL,
		"job_activity.submit":                 r.SubmitURL,
		"job_activity.approve":                r.ApproveURL,
		"job_activity.reject":                 r.RejectURL,
		"job_activity.post":                   r.PostURL,
		"job_activity.reverse":                r.ReverseURL,
		"job_activity.bulk_generate_invoice":  r.BulkGenerateInvoiceURL,
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
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// outcome criteria routes.
func (r OutcomeCriteriaRoutes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_criteria.list":        r.ListURL,
		"outcome_criteria.detail":      r.DetailURL,
		"outcome_criteria.add":         r.AddURL,
		"outcome_criteria.edit":        r.EditURL,
		"outcome_criteria.delete":      r.DeleteURL,
		"outcome_criteria.bulk_delete": r.BulkDeleteURL,
		"outcome_criteria.tab_action":  r.TabActionURL,
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
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// task outcome routes.
func (r TaskOutcomeRoutes) RouteMap() map[string]string {
	return map[string]string{
		"task_outcome.list":   r.ListURL,
		"task_outcome.detail": r.DetailURL,
		"task_outcome.add":    r.AddURL,
		"task_outcome.edit":   r.EditURL,
		"task_outcome.delete": r.DeleteURL,
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
	}
}

// RouteMap returns all fulfillment routes as a map for template URL resolution.
func (r FulfillmentRoutes) RouteMap() map[string]string {
	return map[string]string{
		"fulfillment.dashboard":  r.DashboardURL,
		"fulfillment.list":       r.ListURL,
		"fulfillment.detail":     r.DetailURL,
		"fulfillment.add":        r.AddURL,
		"fulfillment.edit":       r.EditURL,
		"fulfillment.delete":     r.DeleteURL,
		"fulfillment.transition": r.TransitionURL,
		"fulfillment.return":     r.ReturnURL,
	}
}
