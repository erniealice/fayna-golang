package form

// form.go — template-facing data shape for the evaluation cycle create drawer.
//
// The cycle create drawer collects: name, subscription_id (engagement scope),
// period_start/end, sign_off_due_date, close_date (pages.md §F.1 primary action).
// Labels is typed any to avoid an import cycle with the parent package; templates
// read .Labels.* via Go template reflection.
type Data struct {
	FormAction  string
	WorkspaceID string // injected by ViewAdapter for action_workspace_guard
	IsEdit      bool
	ID          string

	Name           string
	SubscriptionID string
	PeriodStart    string
	PeriodEnd      string
	SignOffDueDate string
	CloseDate      string

	Labels       any
	CommonLabels any
}
