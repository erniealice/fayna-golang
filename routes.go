package fayna

// Default route constants for fayna views.
// Consumer apps can use these or define their own.
const (
	// Job dashboard route (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	JobDashboardURL = "/app/jobs/dashboard"

	// Job (operational activity) routes
	JobListURL             = "/app/jobs/list/{status}"
	JobDetailURL           = "/app/jobs/detail/{id}"
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

	// Job Template routes
	JobTemplateListURL             = "/app/job-templates/list/{status}"
	JobTemplateDetailURL           = "/app/job-templates/detail/{id}"
	JobTemplateAddURL              = "/action/job-template/add"
	JobTemplateEditURL             = "/action/job-template/edit/{id}"
	JobTemplateDeleteURL           = "/action/job-template/delete"
	JobTemplateBulkDeleteURL       = "/action/job-template/bulk-delete"
	JobTemplateSetStatusURL        = "/action/job-template/set-status"
	JobTemplateBulkSetStatusURL    = "/action/job-template/bulk-set-status"
	JobTemplateTabActionURL        = "/action/job-template/detail/{id}/tab/{tab}"
	JobTemplateAttachmentUploadURL = "/action/job-template/detail/{id}/attachments/upload"
	JobTemplateAttachmentDeleteURL = "/action/job-template/detail/{id}/attachments/delete"

	// Job Activity (timesheet / cross-job activity log) routes
	JobActivityListURL    = "/app/activities"
	JobActivityDetailURL  = "/app/activities/detail/{id}"
	JobActivityAddURL     = "/action/activity/add"
	JobActivityEditURL    = "/action/activity/edit/{id}"
	JobActivityDeleteURL              = "/action/activity/delete"
	JobActivitySubmitURL              = "/action/activity/submit"
	JobActivityApproveURL             = "/action/activity/approve"
	JobActivityRejectURL              = "/action/activity/reject"
	JobActivityPostURL                = "/action/activity/post"
	JobActivityReverseURL             = "/action/activity/reverse"
	JobActivityBulkGenerateInvoiceURL = "/action/activity/bulk-generate-invoice"

	// Outcome Criteria routes (criteria library)
	OutcomeCriteriaListURL       = "/app/criteria/list/{status}"
	OutcomeCriteriaDetailURL     = "/app/criteria/detail/{id}"
	OutcomeCriteriaAddURL        = "/action/criterion/add"
	OutcomeCriteriaEditURL       = "/action/criterion/edit/{id}"
	OutcomeCriteriaDeleteURL     = "/action/criterion/delete"
	OutcomeCriteriaBulkDeleteURL = "/action/criterion/bulk-delete"
	OutcomeCriteriaTabActionURL  = "/action/criterion/detail/{id}/tab/{tab}"

	// Task Outcome routes (outcome recording on job tasks)
	TaskOutcomeListURL   = "/app/outcomes/list/{status}"
	TaskOutcomeDetailURL = "/app/outcomes/detail/{id}"
	TaskOutcomeAddURL    = "/action/outcome/add"
	TaskOutcomeEditURL   = "/action/outcome/edit/{id}"
	TaskOutcomeDeleteURL = "/action/outcome/delete"

	// Outcome Summary routes (report cards)
	OutcomeSummaryListURL  = "/app/outcomes/summaries"
	OutcomeSummaryJobURL   = "/app/jobs/detail/{id}/summary"
	OutcomeSummaryPhaseURL = "/app/jobs/detail/{id}/phase/{phase_id}/summary"
)

// Fulfillment routes
const (
	// Fulfillment dashboard route (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	FulfillmentDashboardURL = "/app/fulfillment/dashboard"

	FulfillmentListURL       = "/app/fulfillment/list/{status}"
	FulfillmentDetailURL     = "/app/fulfillment/detail/{id}"
	FulfillmentAddURL        = "/action/fulfillment/add"
	FulfillmentEditURL       = "/action/fulfillment/edit/{id}"
	FulfillmentDeleteURL     = "/action/fulfillment/delete"
	FulfillmentTransitionURL = "/action/fulfillment/transition/{id}"
	FulfillmentReturnURL     = "/action/fulfillment/return/{id}"
)
