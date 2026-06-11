package activity_labor

// activity_labor_labels.go — ActivityLabor label structs + DefaultActivityLaborLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// ActivityLaborLabels holds all translatable strings for the activity labor module.
// ActivityLabor is the charge detail for ENTRY_TYPE_LABOR job activities.
// TODO(P7 lyngua sweep): add lyngua JSON files for these labels.
type Labels struct {
	Page    PageLabels   `json:"page"`
	Buttons ButtonLabels `json:"buttons"`
	Columns ColumnLabels `json:"columns"`
	Empty   EmptyLabels  `json:"empty"`
	Form    FormLabels   `json:"form"`
	Detail  DetailLabels `json:"detail"`
	Errors  ErrorLabels  `json:"errors"`
}

type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ButtonLabels struct {
	Add  string `json:"add"`
	Edit string `json:"edit"`
}

type ColumnLabels struct {
	Staff     string `json:"staff"`
	Hours     string `json:"hours"`
	RateType  string `json:"rateType"`
	TimeStart string `json:"timeStart"`
	TimeEnd   string `json:"timeEnd"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type FormLabels struct {
	SectionCharge    string `json:"sectionCharge"`
	Staff            string `json:"staff"`
	Hours            string `json:"hours"`
	RateType         string `json:"rateType"`
	TimeStart        string `json:"timeStart"`
	TimeEnd          string `json:"timeEnd"`
	RateTypeStandard string `json:"rateTypeStandard"`
	RateTypeOvertime string `json:"rateTypeOvertime"`
	RateTypePremium  string `json:"rateTypePremium"`
	SubmitCreate     string `json:"submitCreate"`
	SubmitUpdate     string `json:"submitUpdate"`
}

type DetailLabels struct {
	PageTitle   string `json:"pageTitle"`
	TitlePrefix string `json:"titlePrefix"`
	Staff       string `json:"staff"`
	Hours       string `json:"hours"`
	RateType    string `json:"rateType"`
	TimeStart   string `json:"timeStart"`
	TimeEnd     string `json:"timeEnd"`
}

type ErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultActivityLaborLabels returns ActivityLaborLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading: "Labor Charges",
			Caption: "Labor time entries per activity",
		},
		Buttons: ButtonLabels{
			Add:  "Add Labor",
			Edit: "Edit Labor",
		},
		Columns: ColumnLabels{
			Staff:     "Staff",
			Hours:     "Hours",
			RateType:  "Rate Type",
			TimeStart: "Start",
			TimeEnd:   "End",
		},
		Empty: EmptyLabels{
			Title:   "No labor entries",
			Message: "No labor charge recorded for this activity.",
		},
		Form: FormLabels{
			SectionCharge:    "Charge",
			Staff:            "Staff",
			Hours:            "Hours",
			RateType:         "Rate Type",
			TimeStart:        "Start Time",
			TimeEnd:          "End Time",
			RateTypeStandard: "Standard",
			RateTypeOvertime: "Overtime",
			RateTypePremium:  "Premium",
			SubmitCreate:     "Save",
			SubmitUpdate:     "Update",
		},
		Detail: DetailLabels{
			PageTitle:   "Labor Charge",
			TitlePrefix: "Labor: ",
			Staff:       "Staff",
			Hours:       "Hours",
			RateType:    "Rate Type",
			TimeStart:   "Start Time",
			TimeEnd:     "End Time",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Labor charge record not found",
			IDRequired:       "Activity ID is required",
		},
	}
}
