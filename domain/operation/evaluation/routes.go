package evaluation

// routes.go — Evaluation route constants and Routes config struct.
//
// Route keys use dot.snake_case on the proto domain name (evaluation.*) —
// NEVER a lyngua skin (route-map.md / pages.md §0.1). Page routes embed
// /app/... internally (rewritten to /w/{ws}/... by the workspace-path
// middleware). Action routes (/action/*) and portal routes (/portal/*) are
// not workspace-rewritten.
//
// Surfaces: §C list, §D detail, §A polymorphic score-submission drawer (DF-1),
// §H client-portal "Rate My Team".

const (
	// Page routes (workspace-keyed at the route_config layer)
	ListURL   = "/app/evaluations/list/{status}"
	DetailURL = "/app/evaluations/detail/{id}"

	// HTMX swaps
	TableURL     = "/action/evaluation/table/{status}"
	TabActionURL = "/action/evaluation/tab/{id}/{tab}"

	// Drawer-form (DF-1) + polymorphic dimension slot
	AddURL           = "/action/evaluation/add"
	EditURL          = "/action/evaluation/edit/{id}"
	DimensionSlotURL = "/action/evaluation/dimension-slot"

	// Named lifecycle actions
	SignOffURL = "/action/evaluation/{id}/sign-off"
	ArchiveURL = "/action/evaluation/{id}/archive"
	DeleteURL  = "/action/evaluation/{id}/delete"

	// Bulk
	BulkArchiveURL = "/action/evaluation/bulk/archive"

	// Client-portal "Rate My Team" (§H) — NOT workspace-rewritten
	PortalURL      = "/portal/performance"
	PortalTableURL = "/action/evaluation/portal/table"
)

// Routes holds all route paths for the evaluation views and actions.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`

	TableURL     string `json:"table_url"`
	TabActionURL string `json:"tab_action_url"`

	AddURL           string `json:"add_url"`
	EditURL          string `json:"edit_url"`
	DimensionSlotURL string `json:"dimension_slot_url"`

	SignOffURL string `json:"sign_off_url"`
	ArchiveURL string `json:"archive_url"`
	DeleteURL  string `json:"delete_url"`

	BulkArchiveURL string `json:"bulk_archive_url"`

	PortalURL      string `json:"portal_url"`
	PortalTableURL string `json:"portal_table_url"`
}

// DefaultRoutes returns a Routes populated from the package-level route
// constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "reviews",
		ActiveSubNav: "reviews",

		ListURL:   ListURL,
		DetailURL: DetailURL,

		TableURL:     TableURL,
		TabActionURL: TabActionURL,

		AddURL:           AddURL,
		EditURL:          EditURL,
		DimensionSlotURL: DimensionSlotURL,

		SignOffURL: SignOffURL,
		ArchiveURL: ArchiveURL,
		DeleteURL:  DeleteURL,

		BulkArchiveURL: BulkArchiveURL,

		PortalURL:      PortalURL,
		PortalTableURL: PortalTableURL,
	}
}

// RouteMap returns a map of dot-notation keys (evaluation.*) to route paths.
// Tab keys are nested (evaluation.tab.info etc.) and are resolved at the
// container/route_config layer from TabActionURL.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"evaluation.list":           r.ListURL,
		"evaluation.detail":         r.DetailURL,
		"evaluation.table":          r.TableURL,
		"evaluation.tab":            r.TabActionURL,
		"evaluation.add":            r.AddURL,
		"evaluation.edit":           r.EditURL,
		"evaluation.dimension_slot": r.DimensionSlotURL,
		"evaluation.sign_off":       r.SignOffURL,
		"evaluation.archive":        r.ArchiveURL,
		"evaluation.delete":         r.DeleteURL,
		"evaluation.bulk_archive":   r.BulkArchiveURL,
		"evaluation.portal":         r.PortalURL,
		"evaluation.portal_table":   r.PortalTableURL,
	}
}
