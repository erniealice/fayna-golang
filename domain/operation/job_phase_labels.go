package operation

// job_phase_labels.go — JobPhase label structs + DefaultJobPhaseLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// JobPhaseLabels holds all translatable strings for the job_phase module.
type JobPhaseLabels struct {
	Page    JobPhasePageLabels    `json:"page"`
	Buttons JobPhaseButtonLabels  `json:"buttons"`
	Columns JobPhaseColumnLabels  `json:"columns"`
	Empty   JobPhaseEmptyLabels   `json:"empty"`
	Form    JobPhaseFormLabels    `json:"form"`
	Actions JobPhaseActionLabels  `json:"actions"`
	Detail  JobPhaseDetailLabels  `json:"detail"`
	Tabs    JobPhaseTabLabels     `json:"tabs"`
	Confirm JobPhaseConfirmLabels `json:"confirm"`
	Errors  JobPhaseErrorLabels   `json:"errors"`
}

type JobPhasePageLabels struct {
	Heading          string `json:"heading"`
	Caption          string `json:"caption"`
	HeadingPending   string `json:"headingPending"`
	HeadingActive    string `json:"headingActive"`
	HeadingCompleted string `json:"headingCompleted"`
}

type JobPhaseButtonLabels struct {
	AddPhase string `json:"addPhase"`
}

type JobPhaseColumnLabels struct {
	Name         string `json:"name"`
	Job          string `json:"job"`
	PhaseOrder   string `json:"phaseOrder"`
	Status       string `json:"status"`
	PlannedStart string `json:"plannedStart"`
	PlannedEnd   string `json:"plannedEnd"`
}

type JobPhaseEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type JobPhaseFormLabels struct {
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

type JobPhaseActionLabels struct {
	View         string `json:"view"`
	Edit         string `json:"edit"`
	Delete       string `json:"delete"`
	MarkComplete string `json:"markComplete"`
}

type JobPhaseDetailLabels struct {
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

type JobPhaseTabLabels struct {
	Info        string `json:"info"`
	Tasks       string `json:"tasks"`
	Activities  string `json:"activities"`
	Attachments string `json:"attachments"`
	History     string `json:"history"`
}

type JobPhaseConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type JobPhaseErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultJobPhaseLabels returns JobPhaseLabels with sensible English defaults.
func DefaultJobPhaseLabels() JobPhaseLabels {
	return JobPhaseLabels{
		Page: JobPhasePageLabels{
			Heading:          "Job Phases",
			Caption:          "Manage execution phases across jobs",
			HeadingPending:   "Pending Phases",
			HeadingActive:    "Active Phases",
			HeadingCompleted: "Completed Phases",
		},
		Buttons: JobPhaseButtonLabels{
			AddPhase: "Add Phase",
		},
		Columns: JobPhaseColumnLabels{
			Name:         "Name",
			Job:          "Job",
			PhaseOrder:   "#",
			Status:       "Status",
			PlannedStart: "Planned Start",
			PlannedEnd:   "Planned End",
		},
		Empty: JobPhaseEmptyLabels{
			Title:   "No phases found",
			Message: "No job phases to display.",
		},
		Form: JobPhaseFormLabels{
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
		Actions: JobPhaseActionLabels{
			View:         "View Phase",
			Edit:         "Edit Phase",
			Delete:       "Delete Phase",
			MarkComplete: "Mark Complete",
		},
		Detail: JobPhaseDetailLabels{
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
		Tabs: JobPhaseTabLabels{
			Info:        "Information",
			Tasks:       "Tasks",
			Activities:  "Activities",
			Attachments: "Attachments",
			History:     "History",
		},
		Confirm: JobPhaseConfirmLabels{
			Delete:        "Delete Phase",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: JobPhaseErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Phase not found",
			IDRequired:       "Phase ID is required",
		},
	}
}
