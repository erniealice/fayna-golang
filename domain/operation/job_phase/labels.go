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
	HeadingPending   string `json:"headingPending"`
	HeadingActive    string `json:"headingActive"`
	HeadingCompleted string `json:"headingCompleted"`
}

type ButtonLabels struct {
	AddPhase string `json:"addPhase"`
}

type ColumnLabels struct {
	Name         string `json:"name"`
	Job          string `json:"job"`
	PhaseOrder   string `json:"phaseOrder"`
	Status       string `json:"status"`
	PlannedStart string `json:"plannedStart"`
	PlannedEnd   string `json:"plannedEnd"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type FormLabels struct {
	SectionPhase             string `json:"sectionPhase"`
	SectionResourceTiming    string `json:"sectionResourceTiming"`
	Name                     string `json:"name"`
	NamePlaceholder          string `json:"namePlaceholder"`
	PhaseOrder               string `json:"phaseOrder"`
	Status                   string `json:"status"`
	PlannedStart             string `json:"plannedStart"`
	PlannedEnd               string `json:"plannedEnd"`
	ActualStart              string `json:"actualStart"`
	ActualEnd                string `json:"actualEnd"`
	TemplatePhasePlaceholder string `json:"templatePhasePlaceholder"`
	Resource                 string `json:"resource"`
	ResourcePlaceholder      string `json:"resourcePlaceholder"`
	SetupMinutes             string `json:"setupMinutes"`
	RunMinutesPerUnit        string `json:"runMinutesPerUnit"`
	PredecessorPhase         string `json:"predecessorPhase"`
	PredecessorPlaceholder   string `json:"predecessorPlaceholder"`
}

type ActionLabels struct {
	View         string `json:"view"`
	Edit         string `json:"edit"`
	Delete       string `json:"delete"`
	MarkComplete string `json:"markComplete"`
}

type DetailLabels struct {
	PageTitle         string `json:"pageTitle"`
	Name              string `json:"name"`
	Job               string `json:"job"`
	Status            string `json:"status"`
	PhaseOrder        string `json:"phaseOrder"`
	PlannedStart      string `json:"plannedStart"`
	PlannedEnd        string `json:"plannedEnd"`
	ActualStart       string `json:"actualStart"`
	ActualEnd         string `json:"actualEnd"`
	Resource          string `json:"resource"`
	SetupMinutes      string `json:"setupMinutes"`
	RunMinutesPerUnit string `json:"runMinutesPerUnit"`
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
	DeleteMessage string `json:"deleteMessage"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
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
