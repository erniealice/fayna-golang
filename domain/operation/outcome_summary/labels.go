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
	Student PeriodLabels  `json:"student"`
	// TemplateSettings holds the TB3 report-card template management surface
	// strings. Same snake_case-json-tag rule as LandingLabels — a missing tag
	// silently falls back to the compiled default.
	TemplateSettings TemplateSettingsLabels `json:"template_settings"`
}

// TemplateSettingsLabels holds the report-card template settings page strings
// (list + upload drawer + publish/delete). Generic identifiers; the vertical
// wording ("Report Card Template") lives only in lyngua values.
type TemplateSettingsLabels struct {
	Title          string `json:"title"`
	Subtitle       string `json:"subtitle"`
	NameColumn     string `json:"name_column"`
	ScheduleColumn string `json:"schedule_column"`
	VersionColumn  string `json:"version_column"`
	StatusColumn   string `json:"status_column"`
	ValidityColumn string `json:"validity_column"`
	UploadAction   string `json:"upload_action"`
	PublishAction  string `json:"publish_action"`
	DeleteAction   string `json:"delete_action"`
	EmptyTitle     string `json:"empty_title"`
	EmptyMessage   string `json:"empty_message"`
	// Upload drawer.
	UploadTitle        string `json:"upload_title"`
	NameLabel          string `json:"name_label"`
	ScheduleLabel      string `json:"schedule_label"`
	ScheduleHint       string `json:"schedule_hint"`
	ScheduleFallback   string `json:"schedule_fallback"`
	ValidityStartLabel string `json:"validity_start_label"`
	ValidityEndLabel   string `json:"validity_end_label"`
	FileLabel          string `json:"file_label"`
	// Status badges (VersionStatus enum).
	StatusDraft      string `json:"status_draft"`
	StatusPublished  string `json:"status_published"`
	StatusDeprecated string `json:"status_deprecated"`
	// Confirms + errors.
	PublishConfirm string `json:"publish_confirm"`
	DeleteConfirm  string `json:"delete_confirm"`
	NotConfigured  string `json:"not_configured"`
	InvalidFile    string `json:"invalid_file"`
	UploadFailed   string `json:"upload_failed"`
}

// PeriodLabels holds the view-3 (per-client report card) strings, grouped by
// grading period. Same snake_case-json-tag rule as LandingLabels/SectionLabels —
// a per-tier override silently falls back to the compiled default without the
// tag. (Renamed from StudentLabels{Semester1,Semester2}: generic identifiers,
// vertical wording — "Semester 1/2", "student" — lives in lyngua values only.)
type PeriodLabels struct {
	Title          string `json:"title"`
	Subtitle       string `json:"subtitle"`
	SubjectColumn  string `json:"subject_column"`
	Period1        string `json:"period_1"`
	Period2        string `json:"period_2"`
	YearColumn     string `json:"year_column"`
	ProgressColumn string `json:"progress_column"`
	FinalColumn    string `json:"final_column"`
	ViewAction     string `json:"view_action"`
	// DownloadAction labels the per-student report-card PDF download link (W5,
	// ?format=pdf). Siblings LandingLabels/SectionLabels already carry a
	// DownloadAction (their CSV export); PeriodLabels lacked one. Generic
	// identifier — the vertical wording lives in the lyngua value.
	DownloadAction string `json:"download_action"`
	// StaffLabel / StaffPluralLabel prefix the per-subject staff line on the
	// report-card document ("Teacher:" / "Teachers:" on the education tier).
	// Generic identifiers — the vertical wording lives in the label values.
	StaffLabel       string `json:"staff_label"`
	StaffPluralLabel string `json:"staff_plural_label"`
	// UncategorizedBand titles the client card's single NULL/foreign-category
	// band (R9 W-A6) when the card is banded by job_category — templates whose
	// effective category is NULL/stale/foreign count here, never dropped or
	// duplicated. Generic identifier; the wording lives in the lyngua value. On
	// education1 (zero NULL-category jobs) this band is a defensive path.
	UncategorizedBand string `json:"uncategorized_band"`
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
	// CellViewAction is the aria-label FRAME for each category cell's eye link
	// (R9 W-A2). It MUST name BOTH nouns — the {category} and {section}
	// placeholders are substituted from DATA at render (job_category.name /
	// subscription_group.name) so the link's accessible name carries both
	// dimensions (pyeza renders the first cell as a plain <td>, not a row
	// header, so the section is NOT automatically in the accessible name). A
	// frame missing either placeholder falls back to the typed cell's default
	// "{category} — {section}" composition (fail-safe a11y).
	CellViewAction string `json:"cell_view_action"`
	// UncategorizedColumn heads the single NULL-category bucket column (plan
	// §3.0): templates whose effective category is NULL are never dropped and
	// never duplicated across categories — they count under this one column.
	UncategorizedColumn string `json:"uncategorized_column"`
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
	// CategoryTabsAriaLabel is the aria-label on the section's ?jc= category
	// tabstrip <nav> (R9 W-A3). Without it the strip falls back to the pyeza
	// tabs component's generic "Tabs" default (harmless but generic). Generic
	// identifier — the vertical wording lives only in the lyngua value. The tab
	// LABELS themselves are job_category.name DATA, never lyngua keys.
	CategoryTabsAriaLabel string `json:"category_tabs_aria_label"`
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
	// RenderGate is the D5 409 message shown when a card's grades have entered
	// the approval workflow but are not yet fully published (plan §4.4).
	RenderGate string `json:"render_gate"`
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
			RenderGate:       "This report card cannot be generated yet — its grades are still being reviewed and have not been published. Please try again after the grading period is published.",
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
			// Both placeholders are DATA-substituted at render; the frame must
			// name category AND section (codex §4 pt 8).
			CellViewAction:      "View {category} for {section}",
			UncategorizedColumn: "Uncategorized",
		},
		Section: SectionLabels{
			Title:                 "Group outcomes",
			ClientColumn:          "Client",
			RatingEmpty:           "—",
			DownloadAction:        "Download outcomes (CSV)",
			DetailLink:            "View group",
			NotComputedBanner:     "Final outcomes have not been computed yet.",
			CategoryTabsAriaLabel: "Categories",
		},
		Student: PeriodLabels{
			Title:            "Client outcomes",
			Subtitle:         "Outcomes by grading period",
			SubjectColumn:    "Item",
			Period1:          "Period 1",
			Period2:          "Period 2",
			YearColumn:       "Overall",
			ProgressColumn:   "Progress",
			FinalColumn:      "Final",
			ViewAction:        "View outcomes",
			DownloadAction:    "Download PDF",
			StaffLabel:        "Staff:",
			StaffPluralLabel:  "Staff:",
			UncategorizedBand: "Uncategorized",
		},
		TemplateSettings: TemplateSettingsLabels{
			Title:              "Outcome Report Templates",
			Subtitle:           "Upload and publish the document template used to render outcome reports per schedule",
			NameColumn:         "Template",
			ScheduleColumn:     "Schedule",
			VersionColumn:      "Version",
			StatusColumn:       "Status",
			ValidityColumn:     "Valid",
			UploadAction:       "Upload Template",
			PublishAction:      "Publish",
			DeleteAction:       "Delete",
			EmptyTitle:         "No templates",
			EmptyMessage:       "Upload a document template to render outcome reports.",
			UploadTitle:        "Upload Template",
			NameLabel:          "Template Name",
			ScheduleLabel:      "Schedule",
			ScheduleHint:       "Leave blank to use this template for every schedule (workspace default).",
			ScheduleFallback:   "Workspace default",
			ValidityStartLabel: "Valid From",
			ValidityEndLabel:   "Valid Until",
			FileLabel:          "Template File (.docx)",
			StatusDraft:        "Draft",
			StatusPublished:    "Published",
			StatusDeprecated:   "Deprecated",
			PublishConfirm:     "Publish this template? The previously published template for this schedule will be superseded.",
			DeleteConfirm:      "Delete this template binding? This cannot be undone.",
			NotConfigured:      "Template management is not configured.",
			InvalidFile:        "Only .docx files are accepted.",
			UploadFailed:       "Failed to upload template.",
		},
	}
}
