package fayna

// Default route constants for fayna views.
// Consumer apps can use these or define their own.
const (
	// Job (operational activity) routes
	JobListURL             = "/app/jobs/list/{status}"
	JobDetailURL           = "/app/jobs/detail/{id}"
	JobAddURL              = "/action/jobs/add"
	JobEditURL             = "/action/jobs/edit/{id}"
	JobDeleteURL           = "/action/jobs/delete"
	JobBulkDeleteURL       = "/action/jobs/bulk-delete"
	JobSetStatusURL        = "/action/jobs/set-status"
	JobBulkSetStatusURL    = "/action/jobs/bulk-set-status"
	JobTabActionURL        = "/action/jobs/detail/{id}/tab/{tab}"
	JobAttachmentUploadURL = "/action/jobs/detail/{id}/attachments/upload"
	JobAttachmentDeleteURL = "/action/jobs/detail/{id}/attachments/delete"

	// Job Template routes
	JobTemplateListURL             = "/app/job-templates/list/{status}"
	JobTemplateDetailURL           = "/app/job-templates/detail/{id}"
	JobTemplateAddURL              = "/action/job-templates/add"
	JobTemplateEditURL             = "/action/job-templates/edit/{id}"
	JobTemplateDeleteURL           = "/action/job-templates/delete"
	JobTemplateBulkDeleteURL       = "/action/job-templates/bulk-delete"
	JobTemplateSetStatusURL        = "/action/job-templates/set-status"
	JobTemplateBulkSetStatusURL    = "/action/job-templates/bulk-set-status"
	JobTemplateTabActionURL        = "/action/job-templates/detail/{id}/tab/{tab}"
	JobTemplateAttachmentUploadURL = "/action/job-templates/detail/{id}/attachments/upload"
	JobTemplateAttachmentDeleteURL = "/action/job-templates/detail/{id}/attachments/delete"

	// Job Activity (timesheet / cross-job activity log) routes
	JobActivityListURL    = "/app/activities"
	JobActivityDetailURL  = "/app/activities/detail/{id}"
	JobActivityCreateURL  = "/action/activities/create"
	JobActivityUpdateURL  = "/action/activities/update"
	JobActivityDeleteURL  = "/action/activities/delete"
	JobActivitySubmitURL  = "/action/activities/submit"
	JobActivityApproveURL = "/action/activities/approve"
	JobActivityRejectURL  = "/action/activities/reject"

	// Outcome Criteria routes (criteria library)
	OutcomeCriteriaListURL       = "/app/criteria/list/{status}"
	OutcomeCriteriaDetailURL     = "/app/criteria/detail/{id}"
	OutcomeCriteriaAddURL        = "/action/criteria/add"
	OutcomeCriteriaEditURL       = "/action/criteria/edit/{id}"
	OutcomeCriteriaDeleteURL     = "/action/criteria/delete"
	OutcomeCriteriaBulkDeleteURL = "/action/criteria/bulk-delete"
	OutcomeCriteriaTabActionURL  = "/action/criteria/detail/{id}/tab/{tab}"

	// Task Outcome routes (outcome recording on job tasks)
	TaskOutcomeListURL   = "/app/outcomes/list/{status}"
	TaskOutcomeDetailURL = "/app/outcomes/detail/{id}"
	TaskOutcomeAddURL    = "/action/outcomes/add"
	TaskOutcomeEditURL   = "/action/outcomes/edit/{id}"
	TaskOutcomeDeleteURL = "/action/outcomes/delete"

	// Outcome Summary routes (report cards)
	OutcomeSummaryListURL  = "/app/outcomes/summaries"
	OutcomeSummaryJobURL   = "/app/jobs/detail/{id}/summary"
	OutcomeSummaryPhaseURL = "/app/jobs/detail/{id}/phase/{phase_id}/summary"
)
