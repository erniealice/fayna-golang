package job_phase

// job_phase_labels.go — JobPhase label structs + DefaultJobPhaseLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// JobPhaseLabels holds all translatable strings for the job_phase module.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Buttons ButtonLabels  `json:"buttons"`
	Columns ColumnLabels  `json:"columns"`
	Empty   EmptyLabels   `json:"empty"`
	Form    FormLabels    `json:"form"`
	Actions ActionLabels  `json:"actions"`
	Detail  DetailLabels  `json:"detail"`
	Tabs    TabLabels     `json:"tabs"`
	Confirm ConfirmLabels `json:"confirm"`
	Errors  ErrorLabels   `json:"errors"`
}

type PageLabels struct {
	Heading          string `json:"heading"`
	Caption          string `json:"caption"`
	HeadingPending   string `json:"heading_pending"`
	HeadingActive    string `json:"heading_active"`
	HeadingCompleted string `json:"heading_completed"`
}

type ButtonLabels struct {
	AddPhase string `json:"add_phase"`
}

type ColumnLabels struct {
	Name         string `json:"name"`
	Job          string `json:"job"`
	PhaseOrder   string `json:"phase_order"`
	Status       string `json:"status"`
	PlannedStart string `json:"planned_start"`
	PlannedEnd   string `json:"planned_end"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type FormLabels struct {
	SectionPhase             string `json:"section_phase"`
	SectionResourceTiming    string `json:"section_resource_timing"`
	Name                     string `json:"name"`
	NamePlaceholder          string `json:"name_placeholder"`
	PhaseOrder               string `json:"phase_order"`
	Status                   string `json:"status"`
	PlannedStart             string `json:"planned_start"`
	PlannedEnd               string `json:"planned_end"`
	ActualStart              string `json:"actual_start"`
	ActualEnd                string `json:"actual_end"`
	TemplatePhasePlaceholder string `json:"template_phase_placeholder"`
	Resource                 string `json:"resource"`
	ResourcePlaceholder      string `json:"resource_placeholder"`
	SetupMinutes             string `json:"setup_minutes"`
	RunMinutesPerUnit        string `json:"run_minutes_per_unit"`
	PredecessorPhase         string `json:"predecessor_phase"`
	PredecessorPlaceholder   string `json:"predecessor_placeholder"`
}

type ActionLabels struct {
	View         string `json:"view"`
	Edit         string `json:"edit"`
	Delete       string `json:"delete"`
	MarkComplete string `json:"mark_complete"`
}

type DetailLabels struct {
	PageTitle         string `json:"page_title"`
	Name              string `json:"name"`
	Job               string `json:"job"`
	Status            string `json:"status"`
	PhaseOrder        string `json:"phase_order"`
	PlannedStart      string `json:"planned_start"`
	PlannedEnd        string `json:"planned_end"`
	ActualStart       string `json:"actual_start"`
	ActualEnd         string `json:"actual_end"`
	Resource          string `json:"resource"`
	SetupMinutes      string `json:"setup_minutes"`
	RunMinutesPerUnit string `json:"run_minutes_per_unit"`
}

type TabLabels struct {
	Info        string `json:"info"`
	Tasks       string `json:"tasks"`
	Activities  string `json:"activities"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
}

type ConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"delete_message"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permission_denied"`
	NotFound         string `json:"not_found"`
	IDRequired       string `json:"id_required"`
}

// DefaultJobPhaseLabels returns JobPhaseLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading:          "Job Phases",
			Caption:          "Manage execution phases across jobs",
			HeadingPending:   "Pending Phases",
			HeadingActive:    "Active Phases",
			HeadingCompleted: "Completed Phases",
		},
		Buttons: ButtonLabels{
			AddPhase: "Add Phase",
		},
		Columns: ColumnLabels{
			Name:         "Name",
			Job:          "Job",
			PhaseOrder:   "#",
			Status:       "Status",
			PlannedStart: "Planned Start",
			PlannedEnd:   "Planned End",
		},
		Empty: EmptyLabels{
			Title:   "No phases found",
			Message: "No job phases to display.",
		},
		Form: FormLabels{
			SectionPhase:             "Phase",
			SectionResourceTiming:    "Resource & Timing",
			Name:                     "Phase Name",
			NamePlaceholder:          "Enter phase name",
			PhaseOrder:               "Order",
			Status:                   "Status",
			PlannedStart:             "Planned Start",
			PlannedEnd:               "Planned End",
			ActualStart:              "Actual Start",
			ActualEnd:                "Actual End",
			TemplatePhasePlaceholder: "Search template phase...",
			Resource:                 "Resource",
			ResourcePlaceholder:      "Search resource...",
			SetupMinutes:             "Setup (min)",
			RunMinutesPerUnit:        "Run (min/unit)",
			PredecessorPhase:         "Predecessor Phase",
			PredecessorPlaceholder:   "Select predecessor...",
		},
		Actions: ActionLabels{
			View:         "View Phase",
			Edit:         "Edit Phase",
			Delete:       "Delete Phase",
			MarkComplete: "Mark Complete",
		},
		Detail: DetailLabels{
			PageTitle:         "Phase Details",
			Name:              "Name",
			Job:               "Job",
			Status:            "Status",
			PhaseOrder:        "Order",
			PlannedStart:      "Planned Start",
			PlannedEnd:        "Planned End",
			ActualStart:       "Actual Start",
			ActualEnd:         "Actual End",
			Resource:          "Resource",
			SetupMinutes:      "Setup (min)",
			RunMinutesPerUnit: "Run (min/unit)",
		},
		Tabs: TabLabels{
			Info:        "Information",
			Tasks:       "Tasks",
			Activities:  "Activities",
			Attachments: "Attachments",
			History:     "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Phase",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Phase not found",
			IDRequired:       "Phase ID is required",
		},
	}
}
