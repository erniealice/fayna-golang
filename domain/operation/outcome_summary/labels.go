package outcome_summary

// outcome_summary_labels.go — OutcomeSummary label structs + DefaultOutcomeSummaryLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// OutcomeSummaryLabels holds all translatable strings for the outcome summary module.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Buttons ButtonLabels  `json:"buttons"`
	Columns ColumnLabels  `json:"columns"`
	Empty   EmptyLabels   `json:"empty"`
	Detail  DetailLabels  `json:"detail"`
	Errors  ErrorLabels   `json:"errors"`
	Landing LandingLabels `json:"landing"`
	Section SectionLabels `json:"section"`
	Student StudentLabels `json:"student"`
}

// StudentLabels holds the view-3 (per-student report card) strings. Same
// snake_case-json-tag rule as LandingLabels/SectionLabels — a per-tier override
// silently falls back to the compiled default without the tag.
type StudentLabels struct {
	Title          string `json:"title"`
	Subtitle       string `json:"subtitle"`
	SubjectColumn  string `json:"subject_column"`
	Semester1      string `json:"semester_1"`
	Semester2      string `json:"semester_2"`
	YearColumn     string `json:"year_column"`
	ProgressColumn string `json:"progress_column"`
	FinalColumn    string `json:"final_column"`
	ViewAction     string `json:"view_action"`
}

// LandingLabels holds the view-1 (report-cards landing) strings. Every field
// carries a snake_case json tag matching the lyngua keys — without it a
// per-tier override silently falls back to the compiled default (the
// CellGridLabels lesson, 2026-07-12).
type LandingLabels struct {
	Title           string `json:"title"`
	Subtitle        string `json:"subtitle"`
	GroupColumn     string `json:"group_column"`
	MembersColumn   string `json:"members_column"`
	TemplatesColumn string `json:"templates_column"`
	ViewAction      string `json:"view_action"`
	DownloadAction  string `json:"download_action"`
	TabsAriaLabel   string `json:"tabs_aria_label"`
	InactiveSuffix  string `json:"inactive_suffix"`
}

// SectionLabels holds the view-2 (per-section report-card grid) strings. Same
// snake_case-json-tag rule as LandingLabels.
type SectionLabels struct {
	Title             string `json:"title"`
	ClientColumn      string `json:"client_column"`
	RatingEmpty       string `json:"rating_empty"`
	DownloadAction    string `json:"download_action"`
	DetailLink        string `json:"detail_link"`
	NotComputedBanner string `json:"not_computed_banner"`
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
		Landing: LandingLabels{
			Title:           "Outcome Reports",
			Subtitle:        "Groups by schedule",
			GroupColumn:     "Group",
			MembersColumn:   "Members",
			TemplatesColumn: "Items",
			ViewAction:      "View outcomes",
			DownloadAction:  "Download outcomes (CSV)",
			TabsAriaLabel:   "Schedules",
			InactiveSuffix:  "(inactive)",
		},
		Section: SectionLabels{
			Title:             "Group outcomes",
			ClientColumn:      "Client",
			RatingEmpty:       "—",
			DownloadAction:    "Download outcomes (CSV)",
			DetailLink:        "View group",
			NotComputedBanner: "Final outcomes have not been computed yet.",
		},
		Student: StudentLabels{
			Title:          "Client outcomes",
			Subtitle:       "Outcomes by grading period",
			SubjectColumn:  "Item",
			Semester1:      "Period 1",
			Semester2:      "Period 2",
			YearColumn:     "Overall",
			ProgressColumn: "Progress",
			FinalColumn:    "Final",
			ViewAction:     "View outcomes",
		},
	}
}
