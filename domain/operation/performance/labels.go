package performance

// labels.go — PerformancePanelLabels (panel headings + group headings + per-row
// + cycle-banner labels). View-local label struct (Option-B): the lyngua provider
// overlays translations/en/general/performance.json (camelCase root key
// "performance", LBL trap) onto these defaults at boot.

// Labels holds all translatable strings for the performance admin panel.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Groups  GroupLabels   `json:"groups"`
	Columns ColumnLabels  `json:"columns"`
	Actions ActionLabels  `json:"actions"`
	Rating  RatingLabels  `json:"rating"`
	Banner  BannerLabels  `json:"banner"`
	Empty   EmptyLabels   `json:"empty"`
	Errors  ErrorLabels   `json:"errors"`
}

type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

// GroupLabels are the three cycle-status matrix section headings (pages.md §G.1).
type GroupLabels struct {
	Overdue   string `json:"overdue"`
	Due       string `json:"due"`
	UpToDate  string `json:"upToDate"`
}

type ColumnLabels struct {
	Associate string `json:"associate"`
	Client    string `json:"client"`
	Rating    string `json:"rating"`
	Status    string `json:"status"`
}

type ActionLabels struct {
	StartReview    string `json:"startReview"`
	ViewLastReview string `json:"viewLastReview"`
}

type RatingLabels struct {
	None    string `json:"none"`    // shown when no latest evaluation exists yet
	Summary string `json:"summary"` // rating-summary region heading
}

// BannerLabels feed the "X of Y" cycle-progress banner rendered atop the panel
// (the shared read-projection, §F.2). v1, Phase E.
type BannerLabels struct {
	Progress    string `json:"progress"`    // "%d of %d complete"
	SignOffsDue string `json:"signOffsDue"` // "sign-offs due %s"
	Closes      string `json:"closes"`      // "closes %s"
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	LoadFailed       string `json:"loadFailed"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Performance",
			Caption: "Reviews due, ratings, and cycle progress across your engagements",
		},
		Groups: GroupLabels{
			Overdue:  "Overdue",
			Due:      "Due this period",
			UpToDate: "Up to date",
		},
		Columns: ColumnLabels{
			Associate: "Associate",
			Client:    "Client",
			Rating:    "Rating",
			Status:    "Status",
		},
		Actions: ActionLabels{
			StartReview:    "Start review",
			ViewLastReview: "View last review",
		},
		Rating: RatingLabels{
			None:    "Not yet rated",
			Summary: "Rating summary",
		},
		Banner: BannerLabels{
			Progress:    "%d of %d complete",
			SignOffsDue: "sign-offs due %s",
			Closes:      "closes %s",
		},
		Empty: EmptyLabels{
			Title:   "No reviews to show",
			Message: "Associates on engagements you service will appear here once seats are active.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to view the performance panel",
			LoadFailed:       "Failed to load the performance panel",
		},
	}
}
