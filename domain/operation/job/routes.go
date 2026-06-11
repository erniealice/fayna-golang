package job

// routes.go — Job route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Default route constants for job views.
const (
	// DashboardURL — Job dashboard route (Phase 3 — Pyeza dashboard block + per-app live dashboards plan).
	DashboardURL = "/jobs/dashboard"

	// Job (operational activity) routes
	ListURL             = "/jobs/list/{status}"
	DetailURL           = "/jobs/detail/{id}"
	AddURL              = "/action/job/add"
	EditURL             = "/action/job/edit/{id}"
	DeleteURL           = "/action/job/delete"
	BulkDeleteURL       = "/action/job/bulk-delete"
	SetStatusURL        = "/action/job/set-status"
	BulkSetStatusURL    = "/action/job/bulk-set-status"
	TabActionURL        = "/action/job/detail/{id}/tab/{tab}"
	AttachmentUploadURL = "/action/job/detail/{id}/attachments/upload"
	AttachmentDeleteURL = "/action/job/detail/{id}/attachments/delete"
	TaskAssignURL       = "/action/job/{id}/task/{taskId}/assign"
	// PhaseSetStatusURL — operator-facing phase status flip (PENDING ↔ ACTIVE
	// ↔ COMPLETED). Reads `id` and `status` from query string. Drives the
	// milestone-billing flow: COMPLETED transitions fire the espyna
	// `OnJobPhaseCompleted` hook (BillingEvent → READY).
	// 2026-04-29 milestone-billing plan §4.
	PhaseSetStatusURL = "/action/job-phase/set-status"

	// Auto-complete search endpoints for the job drawer form (client + location pickers).
	// Accept ?q= and return [{"value":"id","label":"Name"},...] JSON.
	// Registered by the fayna block; consumed by the job drawer form.
	ClientSearchURL   = "/action/job/search/clients"
	LocationSearchURL = "/action/job/search/locations"
)

// Routes holds all route paths for job (operational activity) views and actions.
type Routes struct {
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

// DefaultRoutes returns a Routes populated from the package-level
// route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "job",
		ActiveSubNav: "jobs",

		DashboardURL: DashboardURL,

		ListURL:          ListURL,
		DetailURL:        DetailURL,
		AddURL:           AddURL,
		EditURL:          EditURL,
		DeleteURL:        DeleteURL,
		BulkDeleteURL:    BulkDeleteURL,
		SetStatusURL:     SetStatusURL,
		BulkSetStatusURL: BulkSetStatusURL,

		TabActionURL: TabActionURL,

		AttachmentUploadURL: AttachmentUploadURL,
		AttachmentDeleteURL: AttachmentDeleteURL,

		TaskAssignURL: TaskAssignURL,

		PhaseSetStatusURL: PhaseSetStatusURL,

		ClientSearchURL:   ClientSearchURL,
		LocationSearchURL: LocationSearchURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// job routes.
func (r Routes) RouteMap() map[string]string {
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
