package form

// Data is the template-facing data shape for the job category drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and Edit (FormAction =
// EditURL, IsEdit = true). Labels is typed any to avoid an import cycle with the
// parent job_category package; templates read .Labels.* via reflection.
type Data struct {
	FormAction   string
	WorkspaceID  string // injected by ViewAdapter for the action_workspace_guard
	IsEdit       bool
	ID           string
	Name         string
	Code         string
	SortOrder    int32
	Active       bool
	Labels       any
	CommonLabels any
}
