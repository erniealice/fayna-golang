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
	// SectionExport holds the consolidated subscription-group drawer/export
	// vocabulary. Canonical tags stay generic; vertical wording is a Lyngua value.
	SectionExport SectionExportLabels `json:"section_export"`
	// TemplateSettings holds the TB3 report-card template management surface
	// strings. Same snake_case-json-tag rule as LandingLabels — a missing tag
	// silently falls back to the compiled default.
	TemplateSettings TemplateSettingsLabels `json:"template_settings"`
	// SectionTemplateSettings is the separate subscription-group document family.
	SectionTemplateSettings SectionTemplateSettingsLabels `json:"section_template_settings"`
}

// SectionExportLabels holds the category × period × format drawer and its
// fail-loud export messages.
type SectionExportLabels struct {
	DrawerTitle               string `json:"drawer_title"`
	CategoryLabel             string `json:"category_label"`
	CategoryPlaceholder       string `json:"category_placeholder"`
	PeriodLabel               string `json:"period_label"`
	PeriodFinal               string `json:"period_final"`
	FormatLabel               string `json:"format_label"`
	FormatCSV                 string `json:"format_csv"`
	FormatPDF                 string `json:"format_pdf"`
	DownloadAction            string `json:"download_action"`
	NoTemplateError           string `json:"no_template_error"`
	NoProfileError            string `json:"no_profile_error"`
	IncompatibleTemplateError string `json:"incompatible_template_error"`
	AmbiguousPhaseError       string `json:"ambiguous_phase_error"`
	NotComputedError          string `json:"not_computed_error"`
	GroupingConfigError       string `json:"grouping_config_error"`
	DataUnavailableError      string `json:"data_unavailable_error"`
}

// SectionTemplateSettingsLabels holds the distinct subscription-group template
// management surface. Profile identifiers are canonical generic concepts; only
// their translated values may contain education vocabulary.
type SectionTemplateSettingsLabels struct {
	Title                                                 string `json:"title"`
	Subtitle                                              string `json:"subtitle"`
	NameColumn                                            string `json:"name_column"`
	ScheduleColumn                                        string `json:"schedule_column"`
	PlanColumn                                            string `json:"plan_column"`
	CategoryColumn                                        string `json:"category_column"`
	ProfileColumn                                         string `json:"profile_column"`
	ProfileSubscriptionGroupOutcomeMatrixSinglePeriod11V1 string `json:"profile_subscription_group_outcome_matrix_single_period_11_v1"`
	VersionColumn                                         string `json:"version_column"`
	StatusColumn                                          string `json:"status_column"`
	ValidityColumn                                        string `json:"validity_column"`
	UploadAction                                          string `json:"upload_action"`
	PublishAction                                         string `json:"publish_action"`
	DeleteAction                                          string `json:"delete_action"`
	EmptyTitle                                            string `json:"empty_title"`
	EmptyMessage                                          string `json:"empty_message"`
	UploadTitle                                           string `json:"upload_title"`
	NameLabel                                             string `json:"name_label"`
	ScheduleLabel                                         string `json:"schedule_label"`
	ScheduleFallback                                      string `json:"schedule_fallback"`
	PlanLabel                                             string `json:"plan_label"`
	PlanFallback                                          string `json:"plan_fallback"`
	CategoryLabel                                         string `json:"category_label"`
	CategoryFallback                                      string `json:"category_fallback"`
	CategoryRequiredForProfile                            string `json:"category_required_for_profile"`
	ValidityStartLabel                                    string `json:"validity_start_label"`
	ValidityEndLabel                                      string `json:"validity_end_label"`
	FileLabel                                             string `json:"file_label"`
	StatusDraft                                           string `json:"status_draft"`
	StatusPublished                                       string `json:"status_published"`
	StatusDeprecated                                      string `json:"status_deprecated"`
	PublishConfirm                                        string `json:"publish_confirm"`
	BroadScopeConfirm                                     string `json:"broad_scope_confirm"`
	DeleteConfirm                                         string `json:"delete_confirm"`
	NotConfigured                                         string `json:"not_configured"`
	InvalidFile                                           string `json:"invalid_file"`
	InvalidManifest                                       string `json:"invalid_manifest"`
	UploadFailed                                          string `json:"upload_failed"`
	CleanupFailed                                         string `json:"cleanup_failed"`
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
	// ApprovalStatus supplies the Phase-B status-distribution chip labels
	// (R9 W-B2; Q-R9-1/Q-R9-4). It REUSES R7's approval-ladder vocabulary
	// (outcome_matrix approval.status.* + approval.mixed) VERBATIM — the same
	// wording lives here too because the outcome_summary landing loads its OWN
	// lyngua namespace and cannot read the outcome_matrix keys directly. Do not
	// diverge the wording from R7 (lyngua.md); the badge VARIANT is a Go switch,
	// never lyngua (chips.go approvalChipVariant).
	ApprovalStatus ApprovalStatusChipLabels `json:"approval_status"`
}

// ApprovalStatusChipLabels holds the four approval-ladder status texts plus the
// mixed/attention overlay marker used by the Phase-B landing-cell chips. It
// MIRRORS outcome_matrix.ApprovalStatusLabels (+ its Mixed marker) — identical
// vocabulary, separate namespace (see LandingLabels.ApprovalStatus). Every
// field carries a snake_case json tag matching R7's keys.
type ApprovalStatusChipLabels struct {
	InProgress string `json:"in_progress"`
	ForReview  string `json:"for_review"`
	Verified   string `json:"verified"`
	Published  string `json:"published"`
	Mixed      string `json:"mixed"` // attention/mixed overlay (R7 approval.mixed)
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
			// Phase-B chip vocabulary — reuses R7's approval ladder wording
			// verbatim (outcome_matrix approval.status.* / approval.mixed).
			ApprovalStatus: ApprovalStatusChipLabels{
				InProgress: "In Progress",
				ForReview:  "For Review",
				Verified:   "Verified",
				Published:  "Published",
				Mixed:      "Attention — mixed",
			},
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
		SectionExport: SectionExportLabels{
			DrawerTitle:               "Download Group Outcomes",
			CategoryLabel:             "Category",
			CategoryPlaceholder:       "Select a category",
			PeriodLabel:               "Period",
			PeriodFinal:               "Final",
			FormatLabel:               "Format",
			FormatCSV:                 "CSV",
			FormatPDF:                 "PDF",
			DownloadAction:            "Download",
			NoTemplateError:           "No group template is configured for this selection.",
			NoProfileError:            "PDF is not configured for this category. CSV is still available.",
			IncompatibleTemplateError: "The group template does not match this outcome layout.",
			AmbiguousPhaseError:       "This period is configured inconsistently across items.",
			NotComputedError:          "No outcomes are available for this selection.",
			GroupingConfigError:       "Outcome export grouping is not configured correctly.",
			DataUnavailableError:      "The outcome export data is temporarily unavailable.",
		},
		Student: PeriodLabels{
			Title:             "Client outcomes",
			Subtitle:          "Outcomes by grading period",
			SubjectColumn:     "Item",
			Period1:           "Period 1",
			Period2:           "Period 2",
			YearColumn:        "Overall",
			ProgressColumn:    "Progress",
			FinalColumn:       "Final",
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
		SectionTemplateSettings: SectionTemplateSettingsLabels{
			Title:          "Group Templates",
			Subtitle:       "Manage templates used to print consolidated group outcomes",
			NameColumn:     "Template",
			ScheduleColumn: "Schedule",
			PlanColumn:     "Plan",
			CategoryColumn: "Category",
			ProfileColumn:  "Layout Profile",
			ProfileSubscriptionGroupOutcomeMatrixSinglePeriod11V1: "One period, 11 columns",
			VersionColumn:              "Version",
			StatusColumn:               "Status",
			ValidityColumn:             "Validity",
			UploadAction:               "Upload Template",
			PublishAction:              "Publish",
			DeleteAction:               "Delete",
			EmptyTitle:                 "No group templates",
			EmptyMessage:               "Upload a .docx template to enable consolidated PDF downloads.",
			UploadTitle:                "Upload Group Template",
			NameLabel:                  "Template Name",
			ScheduleLabel:              "Schedule",
			ScheduleFallback:           "All schedules",
			PlanLabel:                  "Plan",
			PlanFallback:               "All plans",
			CategoryLabel:              "Category",
			CategoryFallback:           "All categories",
			CategoryRequiredForProfile: "This layout profile requires a specific category.",
			ValidityStartLabel:         "Valid From",
			ValidityEndLabel:           "Valid Until",
			FileLabel:                  "Template File (.docx)",
			StatusDraft:                "Draft",
			StatusPublished:            "Published",
			StatusDeprecated:           "Deprecated",
			PublishConfirm:             "Publish this group template? The current template for the same scope will be superseded.",
			BroadScopeConfirm:          "Publish this broadly scoped template? It may be used when no more specific group template applies.",
			DeleteConfirm:              "Delete this group template? This cannot be undone.",
			NotConfigured:              "Group template management is not configured.",
			InvalidFile:                "Only .docx files are accepted.",
			InvalidManifest:            "This file does not match the required group template layout.",
			UploadFailed:               "Failed to upload group template.",
			CleanupFailed:              "The template operation failed and storage cleanup requires operator attention.",
		},
	}
}
