package work_request

// routes.go -- WorkRequest route constants and Routes config struct.
//
// Route keys use dot.snake_case: work_request.list, work_request.detail, etc.
// Action prefix: /action/work_request/...
// Page prefix: /app/requests/... (workspace-keyed at route_config layer).

// Default route constants for work_request views.
const (
	ListURL      = "/app/requests/list/{status}"
	DetailURL    = "/app/requests/detail/{id}"
	TableURL     = "/action/work_request/table/{status}"
	TabActionURL = "/action/work_request/tab/{id}/{tab}"
	AddURL       = "/action/work_request/add"
	EditURL      = "/action/work_request/edit/{id}"
	SetStatusURL = "/action/work_request/set-status/{id}"
	AssignURL    = "/action/work_request/assign/{id}"
	BulkAssignURL = "/action/work_request/bulk-assign"
	ResolveURL   = "/action/work_request/resolve/{id}"
)

// Routes holds all route paths for work_request views and actions.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL       string `json:"list_url"`
	DetailURL     string `json:"detail_url"`
	TableURL      string `json:"table_url"`
	TabActionURL  string `json:"tab_action_url"`
	AddURL        string `json:"add_url"`
	EditURL       string `json:"edit_url"`
	SetStatusURL  string `json:"set_status_url"`
	AssignURL     string `json:"assign_url"`
	BulkAssignURL string `json:"bulk_assign_url"`
	ResolveURL    string `json:"resolve_url"`
}

// DefaultRoutes returns a Routes populated from the package-level
// route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "requests",
		ActiveSubNav: "requests",

		ListURL:       ListURL,
		DetailURL:     DetailURL,
		TableURL:      TableURL,
		TabActionURL:  TabActionURL,
		AddURL:        AddURL,
		EditURL:       EditURL,
		SetStatusURL:  SetStatusURL,
		AssignURL:     AssignURL,
		BulkAssignURL: BulkAssignURL,
		ResolveURL:    ResolveURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// work_request routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"work_request.list":        r.ListURL,
		"work_request.detail":      r.DetailURL,
		"work_request.table":       r.TableURL,
		"work_request.tab":         r.TabActionURL,
		"work_request.add":         r.AddURL,
		"work_request.edit":        r.EditURL,
		"work_request.set_status":  r.SetStatusURL,
		"work_request.assign":      r.AssignURL,
		"work_request.bulk_assign": r.BulkAssignURL,
		"work_request.resolve":     r.ResolveURL,
	}
}
