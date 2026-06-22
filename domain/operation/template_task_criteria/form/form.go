package form

// Data is the template-facing data shape for the template task criteria drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent template_task_criteria package. Templates read .Labels.* via
// Go template reflection — no cast required.
type Data struct {
	FormAction  string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit      bool
	ID          string

	// FK fields
	JobTemplateTaskID string
	OutcomeCriteriaID string
	SequenceOrder     int32

	Labels       any
	CommonLabels any
}
