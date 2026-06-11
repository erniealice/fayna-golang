package operation

// outcome_summary_labels.go — OutcomeSummary label structs + DefaultOutcomeSummaryLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// OutcomeSummaryLabels holds all translatable strings for the outcome summary module.
type OutcomeSummaryLabels struct {
	Page    OutcomeSummaryPageLabels   `json:"page"`
	Buttons OutcomeSummaryButtonLabels `json:"buttons"`
	Columns OutcomeSummaryColumnLabels `json:"columns"`
	Empty   OutcomeSummaryEmptyLabels  `json:"empty"`
	Detail  OutcomeSummaryDetailLabels `json:"detail"`
	Errors  OutcomeSummaryErrorLabels  `json:"errors"`
}

type OutcomeSummaryColumnLabels struct {
	Job           string `json:"job"`
	Determination string `json:"determination"`
	Score         string `json:"score"`
	ScoringMethod string `json:"scoringMethod"`
	Total         string `json:"total"`
	Pass          string `json:"pass"`
	Fail          string `json:"fail"`
	IssuedBy      string `json:"issuedBy"`
}

type OutcomeSummaryEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type OutcomeSummaryPageLabels struct {
	JobHeading   string `json:"jobHeading"`
	JobCaption   string `json:"jobCaption"`
	PhaseHeading string `json:"phaseHeading"`
	PhaseCaption string `json:"phaseCaption"`
}

type OutcomeSummaryButtonLabels struct {
	GenerateSummary string `json:"generateSummary"`
}

type OutcomeSummaryDetailLabels struct {
	OverallDetermination string `json:"overallDetermination"`
	PhaseDetermination   string `json:"phaseDetermination"`
	Score                string `json:"score"`
	ScoringMethod        string `json:"scoringMethod"`
	TotalCriteria        string `json:"totalCriteria"`
	PassCount            string `json:"passCount"`
	FailCount            string `json:"failCount"`
	ConditionalCount     string `json:"conditionalCount"`
	DeferredCount        string `json:"deferredCount"`
	NaCount              string `json:"naCount"`
	Narrative            string `json:"narrative"`
	IssuedBy             string `json:"issuedBy"`
	IssuedDate           string `json:"issuedDate"`
	ValidUntilDate       string `json:"validUntilDate"`
}

type OutcomeSummaryErrorLabels struct {
	NotFound         string `json:"notFound"`
	PermissionDenied string `json:"permissionDenied"`
}

// DefaultOutcomeSummaryLabels returns OutcomeSummaryLabels with sensible English defaults.
func DefaultOutcomeSummaryLabels() OutcomeSummaryLabels {
	return OutcomeSummaryLabels{
		Page: OutcomeSummaryPageLabels{
			JobHeading:   "Outcome Summary",
			JobCaption:   "Job-level outcome report card",
			PhaseHeading: "Phase Outcome Summary",
			PhaseCaption: "Phase-level outcome report card",
		},
		Buttons: OutcomeSummaryButtonLabels{
			GenerateSummary: "Generate Summary",
		},
		Columns: OutcomeSummaryColumnLabels{
			Job:           "Job",
			Determination: "Determination",
			Score:         "Score",
			ScoringMethod: "Scoring Method",
			Total:         "Total",
			Pass:          "Pass",
			Fail:          "Fail",
			IssuedBy:      "Issued By",
		},
		Empty: OutcomeSummaryEmptyLabels{
			Title:   "No summaries",
			Message: "No outcome summaries have been generated yet.",
		},
		Detail: OutcomeSummaryDetailLabels{
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
		Errors: OutcomeSummaryErrorLabels{
			NotFound:         "Outcome summary not found",
			PermissionDenied: "You do not have permission to perform this action",
		},
	}
}
