package outcome_matrix

// labels.go — OutcomeMatrix label structs + DefaultLabels constructor.
//
// The label tree mirrors the lyngua files at
// translations/en/{general,education}/outcome_matrix.json (root key
// "outcomeMatrix"). The "grid" section embeds pyeza's CellGridLabels verbatim
// so the shared strings (saveButton/savedBanner/clientColumn/...) round-trip
// straight onto the component config; the three grid keys that CellGridLabels
// does NOT model (criterionColumn / scoreColumn / readOnlyTooltip) are added
// alongside it via struct embedding, so they unmarshal from the SAME "grid"
// JSON object level.

import (
	pyezatypes "github.com/erniealice/pyeza-golang/types"
)

// Labels holds all translatable strings for the outcome matrix module.
type Labels struct {
	Page             PageLabels             `json:"page"`
	Picker           PickerLabels           `json:"picker"`
	Scope            ScopeLabels            `json:"scope"`
	Grid             GridLabels             `json:"grid"`
	Errors           ErrorLabels            `json:"errors"`
	Approval         ApprovalLabels         `json:"approval"`
	Columns          ColumnsLabels          `json:"columns"`
	Export           ExportLabels           `json:"export"`
	Narrative        NarrativeLabels        `json:"narrative"`
	TemplateSettings TemplateSettingsLabels `json:"template_settings"`
}

// NarrativeLabels — the per-cell narrative drawer strings (N-1 LOCKED). The
// drawer view (add/edit) and its read-only variant (other-grader / frozen) draw
// from here; the icon affordance + its state-bearing aria are composed in the
// view from the pyeza CellGridLabels, not here. All wording lives in lyngua
// (snake_case keys); education overrides only wording. FIELD-LEVEL ONLY — no
// vertical noun ("grade"/"student") belongs in a generic label default.
type NarrativeLabels struct {
	Title        string `json:"title"`          // drawer heading / fallback sheet title
	FieldLabel   string `json:"field_label"`    // textarea label
	Placeholder  string `json:"placeholder"`    // textarea placeholder (empty note)
	SaveButton   string `json:"save_button"`    // primary submit
	ClearButton  string `json:"clear_button"`   // secondary — clears the note (empty save)
	EmptyText    string `json:"empty_text"`     // read-only variant, no note recorded
	ReadOnlyHint string `json:"read_only_hint"` // why the drawer is view-only

	// Icon + drawer-title composition templates. The view fills {name} (the
	// resolved entity/row name) and {column} (the resolved leaf-column label) —
	// GENERIC placeholder tokens, never vertical nouns; the pyeza component
	// receives only the composed strings. The verb varies with editability +
	// whether a narrative exists, so assistive tech hears the state (the glyph
	// fill is invisible to a screen reader).
	AriaAdd       string `json:"aria_add"`       // editable, no note:  "Add narrative for {name}, {column}"
	AriaEdit      string `json:"aria_edit"`      // editable, has note: "Edit narrative for {name}, {column}"
	AriaView      string `json:"aria_view"`      // read-only:          "View narrative for {name}, {column} (read only)"
	TitleTemplate string `json:"title_template"` // dialog title:       "{name} — {column}"
}

// TemplateSettingsLabels — the grade-sheet template-management settings surface
// (Wave C / P4): list table (Name/Category/Schedule/Version/Status/Validity —
// Category is the NEW sheet-shape axis), upload drawer (category + schedule
// selects), status badges, publish/delete confirms, and fail-closed error
// strings. Mirrors the JOSDT outcome_summary.TemplateSettingsLabels 28-field
// shape + the category axis. All wording lives in lyngua (snake_case keys);
// education overrides only wording ("Grade Sheet …", "Academic Year").
type TemplateSettingsLabels struct {
	// Page heading.
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	// List columns. CategoryColumn is the NEW sheet-shape axis (job_category);
	// ScheduleColumn is relabeled "Academic Year" on the education tier.
	NameColumn     string `json:"column_name"`
	CategoryColumn string `json:"column_category"`
	ScheduleColumn string `json:"column_schedule"`
	VersionColumn  string `json:"column_version"`
	StatusColumn   string `json:"column_status"`
	ValidityColumn string `json:"column_validity"`
	// Row / primary actions.
	UploadAction  string `json:"upload_action"`
	PublishAction string `json:"publish_action"`
	DeleteAction  string `json:"delete_action"`
	// Empty state.
	EmptyTitle   string `json:"empty_title"`
	EmptyMessage string `json:"empty_message"`
	// Upload drawer. CategoryPlaceholder / SchedulePlaceholder are the blank
	// "Any …" <option> labels (also the list-cell fallback for an unscoped
	// binding). NotesLabel maps to the document_template description.
	UploadTitle         string `json:"upload_title"`
	NameLabel           string `json:"name_label"`
	CategoryLabel       string `json:"category_label"`
	CategoryPlaceholder string `json:"category_placeholder"`
	ScheduleLabel       string `json:"schedule_label"`
	SchedulePlaceholder string `json:"schedule_placeholder"`
	FileLabel           string `json:"file_label"`
	FileHint            string `json:"file_hint"`
	NotesLabel          string `json:"notes_label"`
	// Status badges (VersionStatus enum).
	StatusDraft      string `json:"status_draft"`
	StatusPublished  string `json:"status_published"`
	StatusDeprecated string `json:"status_deprecated"`
	// Confirms + fail-closed errors.
	PublishConfirm string `json:"publish_confirm"`
	DeleteConfirm  string `json:"delete_confirm"`
	NotConfigured  string `json:"not_configured"`
	InvalidFile    string `json:"invalid_file"`
	UploadFailed   string `json:"upload_failed"`
}

// ColumnsLabels — the toolbar columns-selector dropdown strings. StateShown/
// StateHidden are visually-hidden state suffixes on each toggle link (the mark
// glyph is aria-hidden — AT needs the state as text); HiddenSuffix follows the
// trigger's hidden-count pip for the same reason.
type ColumnsLabels struct {
	Button       string `json:"button"`        // dropdown trigger text
	Title        string `json:"title"`         // menu heading
	ShowAll      string `json:"show_all"`      // clear-all-hiding link
	StateShown   string `json:"state_shown"`   // sr-only toggle state
	StateHidden  string `json:"state_hidden"`  // sr-only toggle state
	HiddenSuffix string `json:"hidden_suffix"` // sr-only after the count pip
}

// ExportLabels — sheet-level download actions (toolbar trigger + drawer form).
// The drawer replaces the bare CSV anchor with a Period × Format download form
// (20260720 export drawer). Semester option LABELS come from phase rows
// (PhaseColumn.Label, DB data) — lyngua mints only the reserved All/Final
// options + chrome. CSVButton is retained for any plain-anchor fallback.
type ExportLabels struct {
	CSVButton string `json:"csv_button"`

	DrawerButton    string `json:"drawer_button"`     // toolbar trigger text
	DrawerTitle     string `json:"drawer_title"`      // sheet header
	PeriodLabel     string `json:"period_label"`      // period select label
	PeriodAll       string `json:"period_all"`        // "all periods" option
	PeriodFinal     string `json:"period_final"`      // reserved "Final" option + composite column
	FormatLabel     string `json:"format_label"`      // format select label
	FormatCSV       string `json:"format_csv"`        // csv option
	FormatPDF       string `json:"format_pdf"`        // pdf option
	PDFPeriodHint   string `json:"pdf_period_hint"`   // shown when format=pdf (period locked)
	DownloadButton  string `json:"download_button"`   // submit button text
	NoTemplateError string `json:"no_template_error"` // 503 body when no PDF template configured
}

// ApprovalLabels holds the per-phase approval-bar strings (plan §4.5 / lyngua.md).
// Badge VARIANT (default/warning/info/success) is a Go switch, NOT lyngua — only
// the status TEXT lives here. Vocabulary is relabeled by the education overlay.
type ApprovalLabels struct {
	Bar     ApprovalBarLabels     `json:"bar"`
	Status  ApprovalStatusLabels  `json:"status"`
	Actions ApprovalActionLabels  `json:"actions"`
	Chip    ApprovalChipLabels    `json:"chip"`
	Confirm ApprovalConfirmLabels `json:"confirm"`
	Errors  ApprovalErrorLabels   `json:"errors"`

	// Derived-overlay + hint strings (codex label additions).
	Mixed          string `json:"mixed"`            // mixed/Attention marker
	NotStarted     string `json:"not_started"`      // IN_PROGRESS && no data
	LockedHint     string `json:"locked_hint"`      // workflow-locked (advanced, not frozen)
	HardFrozenHint string `json:"hard_frozen_hint"` // finalized / closed schedule
}

// ApprovalBarLabels — the bar heading.
type ApprovalBarLabels struct {
	Title string `json:"title"`
}

// ApprovalStatusLabels — the four ladder status texts (chip/badge label).
type ApprovalStatusLabels struct {
	InProgress string `json:"in_progress"`
	ForReview  string `json:"for_review"`
	Verified   string `json:"verified"`
	Published  string `json:"published"`
}

// ApprovalActionLabels — the transition button texts + the return reason field.
type ApprovalActionLabels struct {
	Submit               string `json:"submit"`
	Verify               string `json:"verify"`
	Publish              string `json:"publish"`
	Return               string `json:"return"`
	ReturnReason         string `json:"return_reason"`          // input label / placeholder (optional case)
	ReturnReasonRequired string `json:"return_reason_required"` // placeholder when a published row forces it
}

// ApprovalChipLabels — chip aria + count framing.
type ApprovalChipLabels struct {
	Aria           string `json:"aria"`            // aria-label for the chip
	PublishedCount string `json:"published_count"` // "{n} students" target-count caption
}

// ApprovalConfirmLabels — hx-confirm dialog messages. Submit carries a "{count}"
// placeholder the view substitutes with the blank required-cell count (D6).
type ApprovalConfirmLabels struct {
	Submit  string `json:"submit"` // "{count}" substituted with the blank required-cell count
	Verify  string `json:"verify"`
	Publish string `json:"publish"`
	Return  string `json:"return"`
}

// ApprovalErrorLabels — the single generic transition-failure message (raw
// server errors are never echoed to the client — fail-closed messaging).
type ApprovalErrorLabels struct {
	ActionFailed string `json:"action_failed"`
}

// PageLabels — page heading.
type PageLabels struct {
	Title string `json:"title"`
}

// PickerLabels — the template/subject picker widget label.
type PickerLabels struct {
	Template string `json:"template"`
}

// ScopeLabels — the scope-toggle options ("mine" vs "all").
type ScopeLabels struct {
	Mine string `json:"mine"`
	All  string `json:"all"`
}

// GridLabels embeds pyeza's CellGridLabels (the fields that map 1:1 onto the
// cell-grid component config) and adds the three grid-object keys that
// CellGridLabels does not model. Because CellGridLabels is embedded anonymously,
// its json-tagged fields promote to the same object level, so a single "grid"
// JSON object populates both.
type GridLabels struct {
	pyezatypes.CellGridLabels
	CriterionColumn string `json:"criterion_column"`
	ScoreColumn     string `json:"score_column"`
	ReadOnlyTooltip string `json:"read_only_tooltip"`
}

// ErrorLabels — generic permission-denied string.
type ErrorLabels struct {
	PermissionDenied string `json:"permission_denied"`
}

// DefaultLabels returns Labels with English (general tier) defaults, byte-for-byte
// matching translations/en/general/outcome_matrix.json.
func DefaultLabels() Labels {
	return Labels{
		Page:   PageLabels{Title: "Outcome Matrix"},
		Picker: PickerLabels{Template: "Template"},
		Scope: ScopeLabels{
			Mine: "My clients",
			All:  "All clients",
		},
		Grid: GridLabels{
			CellGridLabels: pyezatypes.CellGridLabels{
				ClientColumn:   "Client",
				SaveButton:     "Save scores",
				SavingButton:   "Saving…",
				SavedBanner:    "Scores saved.",
				ErrorBanner:    "Save failed — please try again.",
				ReadOnlyMarker: "(read only)",
				EmptyGrid:      "No rows to display.",
				// W2 grade-sheet edit-mode (AutoSave) per-cell + notice strings.
				CellSaving:  "Saving…",
				CellSaved:   "Saved",
				CellError:   "Save failed",
				RatingStale: "Rating not yet recomputed",
				UnsavedWarn: "You have unsaved changes",
				RetryButton: "Retry unsaved",
			},
			CriterionColumn: "Criterion",
			ScoreColumn:     "Score",
			ReadOnlyTooltip: "Recorded by another staff member — read only",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
		},
		Approval: ApprovalLabels{
			Bar: ApprovalBarLabels{Title: "Phase Approvals"},
			Status: ApprovalStatusLabels{
				InProgress: "In Progress",
				ForReview:  "For Review",
				Verified:   "Verified",
				Published:  "Published",
			},
			Actions: ApprovalActionLabels{
				Submit:               "Submit for Review",
				Verify:               "Verify",
				Publish:              "Publish",
				Return:               "Return",
				ReturnReason:         "Return reason (optional)",
				ReturnReasonRequired: "Return reason (required)",
			},
			Chip: ApprovalChipLabels{
				Aria:           "Phase approval status",
				PublishedCount: "{count} in this phase",
			},
			Confirm: ApprovalConfirmLabels{
				Submit:  "Submit this phase for review? {count} required cells are still blank. Editing locks until it is returned.",
				Verify:  "Verify this phase?",
				Publish: "Publish this phase?",
				Return:  "Return this phase to In Progress? Editing reopens.",
			},
			Errors: ApprovalErrorLabels{
				ActionFailed: "This phase could not be updated — it may be locked, finalized, or you may lack permission.",
			},
			Mixed:          "Attention — mixed",
			NotStarted:     "Not started",
			LockedHint:     "This phase is locked — return it to edit",
			HardFrozenHint: "This phase is finalized and can no longer be edited",
		},
		Columns: ColumnsLabels{
			Button:       "Columns",
			Title:        "Show columns",
			ShowAll:      "Show all",
			StateShown:   "shown",
			StateHidden:  "hidden",
			HiddenSuffix: "hidden",
		},
		Export: ExportLabels{
			CSVButton:       "Export CSV",
			DrawerButton:    "Download",
			DrawerTitle:     "Download sheet",
			PeriodLabel:     "Period",
			PeriodAll:       "All periods",
			PeriodFinal:     "Final",
			FormatLabel:     "Format",
			FormatCSV:       "CSV",
			FormatPDF:       "PDF",
			PDFPeriodHint:   "PDF prints the full sheet",
			DownloadButton:  "Download",
			NoTemplateError: "No sheet template is configured for this document",
		},
		Narrative: NarrativeLabels{
			Title:         "Narrative",
			FieldLabel:    "Narrative",
			Placeholder:   "Add a narrative for this cell…",
			SaveButton:    "Save narrative",
			ClearButton:   "Clear",
			EmptyText:     "No narrative recorded.",
			ReadOnlyHint:  "This cell is read only — you can view but not edit its narrative.",
			AriaAdd:       "Add narrative for {name}, {column}",
			AriaEdit:      "Edit narrative for {name}, {column}",
			AriaView:      "View narrative for {name}, {column} (read only)",
			TitleTemplate: "{name} — {column}",
		},
		TemplateSettings: TemplateSettingsLabels{
			Title:               "Sheet Templates",
			Subtitle:            "Manage the document templates used to print sheets",
			NameColumn:          "Template",
			CategoryColumn:      "Category",
			ScheduleColumn:      "Schedule",
			VersionColumn:       "Version",
			StatusColumn:        "Status",
			ValidityColumn:      "Validity",
			UploadAction:        "Upload template",
			PublishAction:       "Publish",
			DeleteAction:        "Delete",
			EmptyTitle:          "No templates yet",
			EmptyMessage:        "Upload a .docx template to enable PDF downloads",
			UploadTitle:         "Upload sheet template",
			NameLabel:           "Template name",
			CategoryLabel:       "Category",
			CategoryPlaceholder: "Any category",
			ScheduleLabel:       "Schedule",
			SchedulePlaceholder: "Any schedule",
			FileLabel:           "Template file (.docx)",
			FileHint:            "Upload a .docx template (max 10 MB).",
			NotesLabel:          "Notes",
			StatusDraft:         "Draft",
			StatusPublished:     "Published",
			StatusDeprecated:    "Deprecated",
			PublishConfirm:      "Publish this template version? The current published version's validity will close.",
			DeleteConfirm:       "Delete this draft?",
			NotConfigured:       "Sheet template management is not configured.",
			InvalidFile:         "Only .docx files are accepted.",
			UploadFailed:        "Failed to upload template.",
		},
	}
}
