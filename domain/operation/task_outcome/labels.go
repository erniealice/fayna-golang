package task_outcome

// task_outcome_labels.go — TaskOutcome label structs + DefaultTaskOutcomeLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// TaskOutcomeLabels holds all translatable strings for the task outcome module.
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

// TabLabels holds tab labels for the task outcome detail page.
type TabLabels struct {
	Info        string `json:"info"`
	Attachments string `json:"attachments"`
}

type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ButtonLabels struct {
	RecordOutcome string `json:"record_outcome"`
}

type ColumnLabels struct {
	Task          string `json:"task"`
	Criteria      string `json:"criteria"`
	Value         string `json:"value"`
	Determination string `json:"determination"`
	RecordedBy    string `json:"recorded_by"`
	Date          string `json:"date"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type FormLabels struct {
	Task             string `json:"task"`
	Criteria         string `json:"criteria"`
	Value            string `json:"value"`
	Notes            string `json:"notes"`
	NotesPlaceholder string `json:"notes_placeholder"`
	Determination    string `json:"determination"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle           string `json:"page_title"`
	Task                string `json:"task"`
	Criteria            string `json:"criteria"`
	CriteriaType        string `json:"criteria_type"`
	Value               string `json:"value"`
	Determination       string `json:"determination"`
	DeterminationSource string `json:"determination_source"`
	DeterminationNote   string `json:"determination_note"`
	RecordedBy          string `json:"recorded_by"`
	RecordedDate        string `json:"recorded_date"`
	RevisionNumber      string `json:"revision_number"`
	CreatedDate         string `json:"created_date"`
}

type ConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"delete_message"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permission_denied"`
	InvalidFormData  string `json:"invalid_form_data"`
	NotFound         string `json:"not_found"`
	IDRequired       string `json:"id_required"`
}

// DefaultTaskOutcomeLabels returns TaskOutcomeLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Outcome Recording",
			Caption: "Record and review task outcome evaluations",
		},
		Buttons: ButtonLabels{
			RecordOutcome: "Record Outcome",
		},
		Columns: ColumnLabels{
			Task:          "Task",
			Criteria:      "Criteria",
			Value:         "Value",
			Determination: "Determination",
			RecordedBy:    "Recorded By",
			Date:          "Date",
		},
		Empty: EmptyLabels{
			Title:   "No outcomes found",
			Message: "No task outcome records to display.",
		},
		Form: FormLabels{
			Task:             "Task",
			Criteria:         "Criteria",
			Value:            "Value",
			Notes:            "Notes",
			NotesPlaceholder: "Enter outcome notes...",
			Determination:    "Determination",
		},
		Actions: ActionLabels{
			View:   "View Outcome",
			Edit:   "Edit Outcome",
			Delete: "Delete Outcome",
		},
		Detail: DetailLabels{
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
		Tabs: TabLabels{
			Info:        "Information",
			Attachments: "Attachments",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Outcome",
			DeleteMessage: "Are you sure you want to delete this outcome record? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Outcome record not found",
			IDRequired:       "Outcome ID is required",
		},
	}
}
