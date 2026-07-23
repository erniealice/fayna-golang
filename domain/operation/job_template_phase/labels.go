package job_template_phase

// job_template_phase_labels.go — JobTemplatePhase label structs + DefaultJobTemplatePhaseLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// JobTemplatePhaseLabels holds all translatable strings for the job_template_phase
// drawer-only view module. This module has no list page, no sidebar entry, and no
// standalone detail page — operators reach it only via the JobTemplate detail Phases tab.
type Labels struct {
	Columns ColumnLabels `json:"columns"`
	Form    FormLabels   `json:"form"`
	Actions ActionLabels `json:"actions"`
	Errors  ErrorLabels  `json:"errors"`
}

type ColumnLabels struct {
	Name        string `json:"name"`
	PhaseOrder  string `json:"phase_order"`
	EstDuration string `json:"est_duration"`
}

type FormLabels struct {
	SectionPhase           string `json:"section_phase"`
	SectionResource        string `json:"section_resource"`
	SectionDependencies    string `json:"section_dependencies"`
	Name                   string `json:"name"`
	NamePlaceholder        string `json:"name_placeholder"`
	Code                   string `json:"code"`
	CodeHint               string `json:"code_hint"`
	PhaseOrder             string `json:"phase_order"`
	EstDurationMinutes     string `json:"est_duration_minutes"`
	Resource               string `json:"resource"`
	ResourcePlaceholder    string `json:"resource_placeholder"`
	PredecessorPhase       string `json:"predecessor_phase"`
	PredecessorPlaceholder string `json:"predecessor_placeholder"`
	// SectionScoring / ScoringScheme — generic wording here; the education
	// tier overrides ScoringScheme to "Grading Scheme" via lyngua.
	SectionScoring           string `json:"section_scoring"`
	ScoringScheme            string `json:"scoring_scheme"`
	ScoringSchemeInfo        string `json:"scoring_scheme_info"`
	ScoringSchemePlaceholder string `json:"scoring_scheme_placeholder"`
}

type ActionLabels struct {
	Add    string `json:"add"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permission_denied"`
	NotFound         string `json:"not_found"`
	IDRequired       string `json:"id_required"`
}

// DefaultJobTemplatePhaseLabels returns JobTemplatePhaseLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Columns: ColumnLabels{
			Name:        "Name",
			PhaseOrder:  "#",
			EstDuration: "Est. Duration (min)",
		},
		Form: FormLabels{
			SectionPhase:             "Phase",
			SectionResource:          "Resource",
			SectionDependencies:      "Dependencies",
			Name:                     "Phase Name",
			NamePlaceholder:          "Enter phase name",
			Code:                     "Code",
			CodeHint:                 "Optional stable key: lowercase letters, digits, underscore.",
			PhaseOrder:               "Order",
			EstDurationMinutes:       "Estimated Duration (min)",
			Resource:                 "Resource",
			ResourcePlaceholder:      "Search resource...",
			PredecessorPhase:         "Predecessor Phase",
			PredecessorPlaceholder:   "Select predecessor...",
			SectionScoring:           "Scoring",
			ScoringScheme:            "Scoring Scheme",
			ScoringSchemeInfo:        "Optional. Scheme this phase's outcomes roll up under.",
			ScoringSchemePlaceholder: "Select scoring scheme...",
		},
		Actions: ActionLabels{
			Add:    "+ Add Phase",
			Edit:   "Edit Phase",
			Delete: "Delete Phase",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Template phase not found",
			IDRequired:       "Template phase ID is required",
		},
	}
}
