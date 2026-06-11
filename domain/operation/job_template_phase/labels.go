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
	PhaseOrder  string `json:"phaseOrder"`
	EstDuration string `json:"estDuration"`
}

type FormLabels struct {
	SectionPhase           string `json:"sectionPhase"`
	SectionResource        string `json:"sectionResource"`
	SectionDependencies    string `json:"sectionDependencies"`
	Name                   string `json:"name"`
	NamePlaceholder        string `json:"namePlaceholder"`
	PhaseOrder             string `json:"phaseOrder"`
	EstDurationMinutes     string `json:"estDurationMinutes"`
	Resource               string `json:"resource"`
	ResourcePlaceholder    string `json:"resourcePlaceholder"`
	PredecessorPhase       string `json:"predecessorPhase"`
	PredecessorPlaceholder string `json:"predecessorPlaceholder"`
}

type ActionLabels struct {
	Add    string `json:"add"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
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
			SectionPhase:           "Phase",
			SectionResource:        "Resource",
			SectionDependencies:    "Dependencies",
			Name:                   "Phase Name",
			NamePlaceholder:        "Enter phase name",
			PhaseOrder:             "Order",
			EstDurationMinutes:     "Estimated Duration (min)",
			Resource:               "Resource",
			ResourcePlaceholder:    "Search resource...",
			PredecessorPhase:       "Predecessor Phase",
			PredecessorPlaceholder: "Select predecessor...",
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
