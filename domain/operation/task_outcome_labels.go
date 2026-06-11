package operation

// task_outcome_labels.go — TaskOutcome label structs + DefaultTaskOutcomeLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// TaskOutcomeLabels holds all translatable strings for the task outcome module.
type TaskOutcomeLabels struct {
	Page    TaskOutcomePageLabels    `json:"page"`
	Buttons TaskOutcomeButtonLabels  `json:"buttons"`
	Columns TaskOutcomeColumnLabels  `json:"columns"`
	Empty   TaskOutcomeEmptyLabels   `json:"empty"`
	Form    TaskOutcomeFormLabels    `json:"form"`
	Actions TaskOutcomeActionLabels  `json:"actions"`
	Detail  TaskOutcomeDetailLabels  `json:"detail"`
	Tabs    TaskOutcomeTabLabels     `json:"tabs"`
	Confirm TaskOutcomeConfirmLabels `json:"confirm"`
	Errors  TaskOutcomeErrorLabels   `json:"errors"`
}

// TaskOutcomeTabLabels holds tab labels for the task outcome detail page.
type TaskOutcomeTabLabels struct {
	Info        string `json:"info"`
	Attachments string `json:"attachments"`
}

type TaskOutcomePageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type TaskOutcomeButtonLabels struct {
	RecordOutcome string `json:"recordOutcome"`
}

type TaskOutcomeColumnLabels struct {
	Task          string `json:"task"`
	Criteria      string `json:"criteria"`
	Value         string `json:"value"`
	Determination string `json:"determination"`
	RecordedBy    string `json:"recordedBy"`
	Date          string `json:"date"`
}

type TaskOutcomeEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type TaskOutcomeFormLabels struct {
	Task             string `json:"task"`
	Criteria         string `json:"criteria"`
	Value            string `json:"value"`
	Notes            string `json:"notes"`
	NotesPlaceholder string `json:"notesPlaceholder"`
	Determination    string `json:"determination"`
}

type TaskOutcomeActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type TaskOutcomeDetailLabels struct {
	PageTitle           string `json:"pageTitle"`
	Task                string `json:"task"`
	Criteria            string `json:"criteria"`
	CriteriaType        string `json:"criteriaType"`
	Value               string `json:"value"`
	Determination       string `json:"determination"`
	DeterminationSource string `json:"determinationSource"`
	DeterminationNote   string `json:"determinationNote"`
	RecordedBy          string `json:"recordedBy"`
	RecordedDate        string `json:"recordedDate"`
	RevisionNumber      string `json:"revisionNumber"`
	CreatedDate         string `json:"createdDate"`
}

type TaskOutcomeConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type TaskOutcomeErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultTaskOutcomeLabels returns TaskOutcomeLabels with sensible English defaults.
func DefaultTaskOutcomeLabels() TaskOutcomeLabels {
	return TaskOutcomeLabels{
		Page: TaskOutcomePageLabels{
			Heading: "Outcome Recording",
			Caption: "Record and review task outcome evaluations",
		},
		Buttons: TaskOutcomeButtonLabels{
			RecordOutcome: "Record Outcome",
		},
		Columns: TaskOutcomeColumnLabels{
			Task:          "Task",
			Criteria:      "Criteria",
			Value:         "Value",
			Determination: "Determination",
			RecordedBy:    "Recorded By",
			Date:          "Date",
		},
		Empty: TaskOutcomeEmptyLabels{
			Title:   "No outcomes found",
			Message: "No task outcome records to display.",
		},
		Form: TaskOutcomeFormLabels{
			Task:             "Task",
			Criteria:         "Criteria",
			Value:            "Value",
			Notes:            "Notes",
			NotesPlaceholder: "Enter outcome notes...",
			Determination:    "Determination",
		},
		Actions: TaskOutcomeActionLabels{
			View:   "View Outcome",
			Edit:   "Edit Outcome",
			Delete: "Delete Outcome",
		},
		Detail: TaskOutcomeDetailLabels{
			PageTitle:           "Outcome Details",
			Task:                "Task",
			Criteria:            "Criteria",
			CriteriaType:        "Criteria Type",
			Value:               "Value",
			Determination:       "Determination",
			DeterminationSource: "Determination Source",
			DeterminationNote:   "Note",
			RecordedBy:          "Recorded By",
			RecordedDate:        "Recorded Date",
			RevisionNumber:      "Revision",
			CreatedDate:         "Created",
		},
		Tabs: TaskOutcomeTabLabels{
			Info:        "Information",
			Attachments: "Attachments",
		},
		Confirm: TaskOutcomeConfirmLabels{
			Delete:        "Delete Outcome",
			DeleteMessage: "Are you sure you want to delete this outcome record? This action cannot be undone.",
		},
		Errors: TaskOutcomeErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Outcome record not found",
			IDRequired:       "Outcome ID is required",
		},
	}
}
