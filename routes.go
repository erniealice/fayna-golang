package fayna

// Default route constants for fayna views.
// Consumer apps can use these or define their own.
const (
	// Job dashboard route (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	JobDashboardURL = "/jobs/dashboard"

	// Job (operational activity) routes
	JobListURL             = "/jobs/list/{status}"
	JobDetailURL           = "/jobs/detail/{id}"
	JobAddURL              = "/action/job/add"
	JobEditURL             = "/action/job/edit/{id}"
	JobDeleteURL           = "/action/job/delete"
	JobBulkDeleteURL       = "/action/job/bulk-delete"
	JobSetStatusURL        = "/action/job/set-status"
	JobBulkSetStatusURL    = "/action/job/bulk-set-status"
	JobTabActionURL        = "/action/job/detail/{id}/tab/{tab}"
	JobAttachmentUploadURL = "/action/job/detail/{id}/attachments/upload"
	JobAttachmentDeleteURL = "/action/job/detail/{id}/attachments/delete"
	JobTaskAssignURL       = "/action/job/{id}/task/{taskId}/assign"
	// JobPhaseSetStatusURL — operator-facing phase status flip (PENDING ↔ ACTIVE
	// ↔ COMPLETED). Reads `id` and `status` from query string. Drives the
	// milestone-billing flow: COMPLETED transitions fire the espyna
	// `OnJobPhaseCompleted` hook (BillingEvent → READY).
	// 2026-04-29 milestone-billing plan §4.
	JobPhaseSetStatusURL = "/action/job-phase/set-status"

	// JobPhase standalone module routes.
	// The list page is a power-user/debugging surface — no sidebar entry.
	// The detail page is the canonical single-phase view with tab strip.
	JobPhaseListURL             = "/job-phases/list/{status}"
	JobPhaseDetailURL           = "/job-phase/{id}"
	JobPhaseAddURL              = "/action/job-phase/add"
	JobPhaseEditURL             = "/action/job-phase/edit/{id}"
	JobPhaseDeleteURL           = "/action/job-phase/delete"
	JobPhaseBulkDeleteURL       = "/action/job-phase/bulk-delete"
	JobPhaseBulkSetStatusURL    = "/action/job-phase/bulk-set-status"
	JobPhaseTabActionURL        = "/action/job-phase/detail/{id}/tab/{tab}"
	JobPhaseResourceSearchURL   = "/action/job-phase/search/resources"
	JobPhaseAttachmentUploadURL = "/action/job-phase/detail/{id}/attachments/upload"
	JobPhaseAttachmentDeleteURL = "/action/job-phase/detail/{id}/attachments/delete"

	// JobTask standalone module routes.
	// The list page is a power-user/debugging surface — no sidebar entry.
	// The detail page is reached via JobPhase detail's Tasks tab deep links.
	JobTaskListURL               = "/job-tasks/list/{status}"
	JobTaskDetailURL             = "/job-task/{id}"
	JobTaskAddURL                = "/action/job-task/add"
	JobTaskEditURL               = "/action/job-task/edit/{id}"
	JobTaskDeleteURL             = "/action/job-task/delete"
	JobTaskBulkDeleteURL         = "/action/job-task/bulk-delete"
	JobTaskSetStatusURL          = "/action/job-task/set-status"
	JobTaskBulkSetStatusURL      = "/action/job-task/bulk-set-status"
	JobTaskTabActionURL          = "/action/job-task/detail/{id}/tab/{tab}"
	JobTaskStaffSearchURL        = "/action/job-task/search/staff"
	JobTaskResourceSearchURL     = "/action/job-task/search/resources"
	JobTaskTemplateTaskSearchURL = "/action/job-task/search/template-tasks"
	JobTaskAttachmentUploadURL   = "/action/job-task/detail/{id}/attachments/upload"
	JobTaskAttachmentDeleteURL   = "/action/job-task/detail/{id}/attachments/delete"

	// Job auto-complete search endpoints.
	// Accept ?q= and return [{"value":"id","label":"Name"},...] JSON.
	// Registered by the fayna block; consumed by the job drawer form.
	JobClientSearchURL   = "/action/job/search/clients"
	JobLocationSearchURL = "/action/job/search/locations"

	// Job Template routes
	JobTemplateListURL             = "/job-templates/list/{status}"
	JobTemplateDetailURL           = "/job-templates/detail/{id}"
	JobTemplateAddURL              = "/action/job-template/add"
	JobTemplateEditURL             = "/action/job-template/edit/{id}"
	JobTemplateDeleteURL           = "/action/job-template/delete"
	JobTemplateBulkDeleteURL       = "/action/job-template/bulk-delete"
	JobTemplateSetStatusURL        = "/action/job-template/set-status"
	JobTemplateBulkSetStatusURL    = "/action/job-template/bulk-set-status"
	JobTemplateTabActionURL        = "/action/job-template/detail/{id}/tab/{tab}"
	JobTemplateAttachmentUploadURL = "/action/job-template/detail/{id}/attachments/upload"
	JobTemplateAttachmentDeleteURL = "/action/job-template/detail/{id}/attachments/delete"

	// JobTemplatePhase drawer-only module routes.
	// No list page, no detail page, no sidebar entry.
	// Reached via JobTemplate detail Phases tab Add/Edit/Delete CTAs.
	JobTemplatePhaseAddURL        = "/action/job-template-phase/add"
	JobTemplatePhaseEditURL       = "/action/job-template-phase/edit/{id}"
	JobTemplatePhaseDeleteURL     = "/action/job-template-phase/delete"
	JobTemplatePhaseBulkDeleteURL = "/action/job-template-phase/bulk-delete"
	// JobTemplatePhaseResourceSearchURL reuses the job_phase resource search endpoint
	// (same underlying resource entity — no separate handler needed).
	JobTemplatePhaseResourceSearchURL = JobPhaseResourceSearchURL

	// JobTemplateTask drawer-only module routes.
	// No list page, no detail page, no sidebar entry.
	// Reached via JobTemplate detail Tasks tab Add/Edit/Delete CTAs.
	JobTemplateTaskAddURL        = "/action/job-template-task/add"
	JobTemplateTaskEditURL       = "/action/job-template-task/edit/{id}"
	JobTemplateTaskDeleteURL     = "/action/job-template-task/delete"
	JobTemplateTaskBulkDeleteURL = "/action/job-template-task/bulk-delete"
	// JobTemplateTaskResourceSearchURL reuses the job_phase resource search endpoint.
	JobTemplateTaskResourceSearchURL = JobPhaseResourceSearchURL

	// Job Activity (timesheet / cross-job activity log) routes
	JobActivityListURL                = "/activities"
	JobActivityDetailURL              = "/activities/detail/{id}"
	JobActivityAddURL                 = "/action/activity/add"
	JobActivityEditURL                = "/action/activity/edit/{id}"
	JobActivityDeleteURL              = "/action/activity/delete"
	JobActivitySubmitURL              = "/action/activity/submit"
	JobActivityApproveURL             = "/action/activity/approve"
	JobActivityRejectURL              = "/action/activity/reject"
	JobActivityPostURL                = "/action/activity/post"
	JobActivityReverseURL             = "/action/activity/reverse"
	JobActivityBulkDeleteURL          = "/action/activity/bulk-delete"
	JobActivityBulkGenerateInvoiceURL = "/action/activity/bulk-generate-invoice"
	JobActivityTabActionURL           = "/action/activity/detail/{id}/tab/{tab}"
	JobActivityAttachmentUploadURL    = "/action/activity/detail/{id}/attachments/upload"
	JobActivityAttachmentDeleteURL    = "/action/activity/detail/{id}/attachments/delete"

	// Outcome Criteria routes (criteria library)
	OutcomeCriteriaListURL             = "/criteria/list/{status}"
	OutcomeCriteriaDetailURL           = "/criteria/detail/{id}"
	OutcomeCriteriaAddURL              = "/action/criterion/add"
	OutcomeCriteriaEditURL             = "/action/criterion/edit/{id}"
	OutcomeCriteriaDeleteURL           = "/action/criterion/delete"
	OutcomeCriteriaBulkDeleteURL       = "/action/criterion/bulk-delete"
	OutcomeCriteriaTabActionURL        = "/action/criterion/detail/{id}/tab/{tab}"
	OutcomeCriteriaAttachmentUploadURL = "/action/criterion/detail/{id}/attachments/upload"
	OutcomeCriteriaAttachmentDeleteURL = "/action/criterion/detail/{id}/attachments/delete"

	// Task Outcome routes (outcome recording on job tasks)
	TaskOutcomeListURL             = "/outcomes/list/{status}"
	TaskOutcomeDetailURL           = "/outcomes/detail/{id}"
	TaskOutcomeAddURL              = "/action/outcome/add"
	TaskOutcomeEditURL             = "/action/outcome/edit/{id}"
	TaskOutcomeDeleteURL           = "/action/outcome/delete"
	TaskOutcomeTabActionURL        = "/action/outcome/detail/{id}/tab/{tab}"
	TaskOutcomeAttachmentUploadURL = "/action/outcome/detail/{id}/attachments/upload"
	TaskOutcomeAttachmentDeleteURL = "/action/outcome/detail/{id}/attachments/delete"

	// Outcome Summary routes (report cards)
	OutcomeSummaryListURL  = "/outcomes/summaries"
	OutcomeSummaryJobURL   = "/jobs/detail/{id}/summary"
	OutcomeSummaryPhaseURL = "/jobs/detail/{id}/phase/{phase_id}/summary"
)

// Activity Labor routes (charge detail for ENTRY_TYPE_LABOR job activities).
// ActivityLaborListURL is registered but NOT in the sidebar — power-user / debug only.
// The primary surface is the JobActivity detail page's charge tab.
const (
	ActivityLaborListURL        = "/activity-labor/list"
	ActivityLaborDetailURL      = "/activity-labor/{id}"
	ActivityLaborAddURL         = "/action/activity-labor/add"
	ActivityLaborEditURL        = "/action/activity-labor/edit/{id}"
	ActivityLaborDeleteURL      = "/action/activity-labor/delete"
	ActivityLaborStaffSearchURL = "/action/activity-labor/search/staff"
)

// Activity Material routes (charge detail for ENTRY_TYPE_MATERIAL job activities).
// ActivityMaterialListURL is registered but NOT in the sidebar — power-user / debug only.
// The primary surface is the JobActivity detail page's charge tab.
const (
	ActivityMaterialListURL           = "/activity-material/list"
	ActivityMaterialDetailURL         = "/activity-material/{id}"
	ActivityMaterialAddURL            = "/action/activity-material/add"
	ActivityMaterialEditURL           = "/action/activity-material/edit/{id}"
	ActivityMaterialDeleteURL         = "/action/activity-material/delete"
	ActivityMaterialProductSearchURL  = "/action/activity-material/search/products"
	ActivityMaterialLocationSearchURL = "/action/activity-material/search/locations"
)

// Activity Expense routes (charge detail for ENTRY_TYPE_EXPENSE job activities).
// ActivityExpenseListURL is registered but NOT in the sidebar — power-user / debug only.
// The primary surface is the JobActivity detail page's charge tab.
const (
	ActivityExpenseListURL                  = "/activity-expense/list"
	ActivityExpenseDetailURL                = "/activity-expense/{id}"
	ActivityExpenseAddURL                   = "/action/activity-expense/add"
	ActivityExpenseEditURL                  = "/action/activity-expense/edit/{id}"
	ActivityExpenseDeleteURL                = "/action/activity-expense/delete"
	ActivityExpenseExpenseCategorySearchURL = "/action/activity-expense/search/expense-categories"
)

// Fulfillment routes
const (
	// Fulfillment dashboard route (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	FulfillmentDashboardURL = "/fulfillment/dashboard"

	FulfillmentListURL             = "/fulfillment/list/{status}"
	FulfillmentDetailURL           = "/fulfillment/detail/{id}"
	FulfillmentAddURL              = "/action/fulfillment/add"
	FulfillmentEditURL             = "/action/fulfillment/edit/{id}"
	FulfillmentDeleteURL           = "/action/fulfillment/delete"
	FulfillmentTransitionURL       = "/action/fulfillment/transition/{id}"
	FulfillmentReturnURL           = "/action/fulfillment/return/{id}"
	FulfillmentTabActionURL        = "/action/fulfillment/detail/{id}/tab/{tab}"
	FulfillmentAttachmentUploadURL = "/action/fulfillment/detail/{id}/attachments/upload"
	FulfillmentAttachmentDeleteURL = "/action/fulfillment/detail/{id}/attachments/delete"
)
