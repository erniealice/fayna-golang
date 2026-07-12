package outcome_summary

// routes.go — OutcomeSummary route constants and Routes config struct.
//
// Extracted from packages/fayna-golang/domain/operation/routes.go.
// Pure structural move — route string values are byte-identical.

// Outcome Summary routes (report cards)
const (
	ListURL = "/outcomes/summaries"
	JobURL  = "/jobs/detail/{id}/summary"
	// SectionURL is the per-section report-card grid (view-2): {id} is a
	// subscription_group id. Generic default here; the education tier overrides
	// it to /report-cards/section/{id} via education/route.json.
	SectionURL = "/outcomes/summaries/section/{id}"
	PhaseURL   = "/jobs/detail/{id}/phase/{phase_id}/summary"
)

// Routes holds all route paths for outcome summary (report card) views.
type Routes struct {
	// Sidebar navigation context
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	// ListActiveSubNav overrides ActiveSubNav for the standalone list page.
	// Job/phase summary pages highlight "jobs" while the list page highlights "report-cards".
	ListActiveSubNav string `json:"list_active_sub_nav"`

	ListURL         string `json:"list_url"`
	JobSummaryURL   string `json:"job_summary_url"`
	SectionURL      string `json:"section_url"`
	PhaseSummaryURL string `json:"phase_summary_url"`
}

// DefaultRoutes returns a Routes populated from
// the package-level route constants defined in this file.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:        "job",
		ActiveSubNav:     "jobs",
		ListActiveSubNav: "report-cards",

		ListURL:         ListURL,
		JobSummaryURL:   JobURL,
		SectionURL:      SectionURL,
		PhaseSummaryURL: PhaseURL,
	}
}

// RouteMap returns a map of dot-notation keys to route paths for all
// outcome summary routes.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"outcome_summary.list":    r.ListURL,
		"outcome_summary.job":     r.JobSummaryURL,
		"outcome_summary.section": r.SectionURL,
		"outcome_summary.phase":   r.PhaseSummaryURL,
	}
}
