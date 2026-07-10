package outcome_summary

// outcome_summary_labels.go — OutcomeSummary label structs + DefaultOutcomeSummaryLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// OutcomeSummaryLabels holds all translatable strings for the outcome summary module.
type Labels struct {
	Page    PageLabels   `json:"page"`
	Buttons ButtonLabels `json:"buttons"`
	Columns ColumnLabels `json:"columns"`
	Empty   EmptyLabels  `json:"empty"`
	Detail  DetailLabels `json:"detail"`
	Errors  ErrorLabels  `json:"errors"`
}

type ColumnLabels struct {
	Job           string `json:"job"`
	Determination string `json:"determination"`
	Score         string `json:"score"`
	ScoringMethod string `json:"scoring_method"`
	Total         string `json:"total"`
	Pass          string `json:"pass"`
	Fail          string `json:"fail"`
	IssuedBy      string `json:"issued_by"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type PageLabels struct {
	JobHeading   string `json:"job_heading"`
	JobCaption   string `json:"job_caption"`
	PhaseHeading string `json:"phase_heading"`
	PhaseCaption string `json:"phase_caption"`
}

type ButtonLabels struct {
	GenerateSummary string `json:"generate_summary"`
}

type DetailLabels struct {
	OverallDetermination string `json:"overall_determination"`
	PhaseDetermination   string `json:"phase_determination"`
	Score                string `json:"score"`
	ScoringMethod        string `json:"scoring_method"`
	TotalCriteria        string `json:"total_criteria"`
	PassCount            string `json:"pass_count"`
	FailCount            string `json:"fail_count"`
	ConditionalCount     string `json:"conditional_count"`
	DeferredCount        string `json:"deferred_count"`
	NaCount              string `json:"na_count"`
	Narrative            string `json:"narrative"`
	IssuedBy             string `json:"issued_by"`
	IssuedDate           string `json:"issued_date"`
	ValidUntilDate       string `json:"valid_until_date"`
}

type ErrorLabels struct {
	NotFound         string `json:"not_found"`
	PermissionDenied string `json:"permission_denied"`
}

// DefaultOutcomeSummaryLabels returns OutcomeSummaryLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			JobHeading:   "Outcome Summary",
			JobCaption:   "Job-level outcome report card",
			PhaseHeading: "Phase Outcome Summary",
			PhaseCaption: "Phase-level outcome report card",
		},
		Buttons: ButtonLabels{
			GenerateSummary: "Generate Summary",
		},
		Columns: ColumnLabels{
			Job:           "Job",
			Determination: "Determination",
			Score:         "Score",
			ScoringMethod: "Scoring Method",
			Total:         "Total",
			Pass:          "Pass",
			Fail:          "Fail",
			IssuedBy:      "Issued By",
		},
		Empty: EmptyLabels{
			Title:   "No summaries",
			Message: "No outcome summaries have been generated yet.",
		},
		Detail: DetailLabels{
			OverallDetermination: "Overall Determination",
			PhaseDetermination:   "Phase Determination",
			Score:                "Score",
			ScoringMethod:        "Scoring Method",
			TotalCriteria:        "Total Criteria",
			PassCount:            "Pass",
			FailCount:            "Fail",
			ConditionalCount:     "Conditional",
			DeferredCount:        "Deferred",
			NaCount:              "N/A",
			Narrative:            "Narrative",
			IssuedBy:             "Issued By",
			IssuedDate:           "Issued Date",
			ValidUntilDate:       "Valid Until",
		},
		Errors: ErrorLabels{
			NotFound:         "Outcome summary not found",
			PermissionDenied: "You do not have permission to perform this action",
		},
	}
}
