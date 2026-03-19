// Package fayna provides translatable label structs for the operations/work domain.
//
// This file contains all label types for job templates, job execution, and job activities.
// Labels are loaded from lyngua translation files and injected into views at startup.
package fayna

// ---------------------------------------------------------------------------
// Job Template labels
// ---------------------------------------------------------------------------

// JobTemplateLabels holds all translatable strings for the job template module.
type JobTemplateLabels struct {
	Page    JobTemplatePageLabels    `json:"page"`
	Buttons JobTemplateButtonLabels  `json:"buttons"`
	Columns JobTemplateColumnLabels  `json:"columns"`
	Empty   JobTemplateEmptyLabels   `json:"empty"`
	Form    JobTemplateFormLabels    `json:"form"`
	Actions JobTemplateActionLabels  `json:"actions"`
	Detail  JobTemplateDetailLabels  `json:"detail"`
	Tabs    JobTemplateTabLabels     `json:"tabs"`
	Confirm JobTemplateConfirmLabels `json:"confirm"`
	Errors  JobTemplateErrorLabels   `json:"errors"`
}

type JobTemplatePageLabels struct {
	Heading         string `json:"heading"`
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type JobTemplateButtonLabels struct {
	AddJobTemplate string `json:"addJobTemplate"`
}

type JobTemplateColumnLabels struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type JobTemplateEmptyLabels struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type JobTemplateActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type JobTemplateErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
	NoPermission     string `json:"noPermission"`
}

type JobTemplateFormLabels struct {
	Name            string `json:"name"`
	NamePlaceholder string `json:"namePlaceholder"`
	Description     string `json:"description"`
	DescPlaceholder string `json:"descriptionPlaceholder"`
	Active          string `json:"active"`
}

type JobTemplateDetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
}

type JobTemplateTabLabels struct {
	Info        string `json:"info"`
	Phases      string `json:"phases"`
	Attachments string `json:"attachments"`
	AuditTrail  string `json:"auditTrail"`
}

type JobTemplateConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

// DefaultJobTemplateLabels returns JobTemplateLabels with sensible English defaults.
func DefaultJobTemplateLabels() JobTemplateLabels {
	return JobTemplateLabels{
		Page: JobTemplatePageLabels{
			Heading:         "Job Templates",
			HeadingActive:   "Active Job Templates",
			HeadingInactive: "Inactive Job Templates",
			Caption:         "Manage reusable execution plans",
			CaptionActive:   "Manage your active job templates",
			CaptionInactive: "View inactive or archived job templates",
		},
		Buttons: JobTemplateButtonLabels{
			AddJobTemplate: "Add Template",
		},
		Columns: JobTemplateColumnLabels{
			Name:        "Name",
			Description: "Description",
			Status:      "Status",
		},
		Empty: JobTemplateEmptyLabels{
			Title:           "No job templates found",
			Message:         "No job templates to display.",
			ActiveTitle:     "No active job templates",
			ActiveMessage:   "Create your first job template to get started.",
			InactiveTitle:   "No inactive job templates",
			InactiveMessage: "Deactivated templates will appear here.",
		},
		Form: JobTemplateFormLabels{
			Name:            "Template Name",
			NamePlaceholder: "Enter template name",
			Description:     "Description",
			DescPlaceholder: "Enter template description...",
			Active:          "Active",
		},
		Actions: JobTemplateActionLabels{
			View:   "View Template",
			Edit:   "Edit Template",
			Delete: "Delete Template",
		},
		Detail: JobTemplateDetailLabels{
			PageTitle:    "Job Template Details",
			Name:         "Name",
			Description:  "Description",
			Status:       "Status",
			CreatedDate:  "Created",
			ModifiedDate: "Last Modified",
		},
		Tabs: JobTemplateTabLabels{
			Info:        "Information",
			Phases:      "Phases",
			Attachments: "Attachments",
			AuditTrail:  "Audit Trail",
		},
		Confirm: JobTemplateConfirmLabels{
			Delete:        "Delete Template",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: JobTemplateErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Job template not found",
			IDRequired:       "Job template ID is required",
			NoPermission:     "No permission",
		},
	}
}

// ---------------------------------------------------------------------------
// Job Activity labels (cross-job timesheet / activity log)
// ---------------------------------------------------------------------------

// JobActivityLabels holds all translatable strings for the job activity module.
type JobActivityLabels struct {
	Page    JobActivityPageLabels    `json:"page"`
	Buttons JobActivityButtonLabels  `json:"buttons"`
	Columns JobActivityColumnLabels  `json:"columns"`
	Empty   JobActivityEmptyLabels   `json:"empty"`
	Form    JobActivityFormLabels    `json:"form"`
	Actions JobActivityActionLabels  `json:"actions"`
	Detail  JobActivityDetailLabels  `json:"detail"`
	Status  JobActivityStatusLabels  `json:"status"`
	Errors  JobActivityErrorLabels   `json:"errors"`
}

type JobActivityPageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type JobActivityButtonLabels struct {
	AddActivity string `json:"addActivity"`
}

type JobActivityColumnLabels struct {
	Date        string `json:"date"`
	Job         string `json:"job"`
	EntryType   string `json:"entryType"`
	Description string `json:"description"`
	Quantity    string `json:"quantity"`
	Amount      string `json:"amount"`
	Status      string `json:"status"`
}

type JobActivityEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type JobActivityFormLabels struct {
	Job            string `json:"job"`
	Task           string `json:"task"`
	EntryType      string `json:"entryType"`
	Description    string `json:"description"`
	BillableStatus string `json:"billableStatus"`
	Hours          string `json:"hours"`
	HourlyRate     string `json:"hourlyRate"`
	Product        string `json:"product"`
	Quantity       string `json:"quantity"`
	UnitCost       string `json:"unitCost"`
	Amount         string `json:"amount"`
	Category       string `json:"category"`
}

type JobActivityActionLabels struct {
	View    string `json:"view"`
	Edit    string `json:"edit"`
	Delete  string `json:"delete"`
	Submit  string `json:"submit"`
	Approve string `json:"approve"`
	Reject  string `json:"reject"`
}

type JobActivityDetailLabels struct {
	PageTitle      string `json:"pageTitle"`
	TitlePrefix    string `json:"titlePrefix"`
	Job            string `json:"job"`
	EntryType      string `json:"entryType"`
	EntryDate      string `json:"entryDate"`
	Description    string `json:"description"`
	Quantity       string `json:"quantity"`
	UnitCost       string `json:"unitCost"`
	TotalCost      string `json:"totalCost"`
	Currency       string `json:"currency"`
	BillableStatus string `json:"billableStatus"`
	ApprovalStatus string `json:"approvalStatus"`
	PostingStatus  string `json:"postingStatus"`
	CreatedDate    string `json:"createdDate"`
	// Labor subtype
	Staff     string `json:"staff"`
	Hours     string `json:"hours"`
	RateType  string `json:"rateType"`
	TimeStart string `json:"timeStart"`
	TimeEnd   string `json:"timeEnd"`
	// Material subtype
	Product       string `json:"product"`
	UnitOfMeasure string `json:"unitOfMeasure"`
	LotNumber     string `json:"lotNumber"`
	Location      string `json:"location"`
	// Expense subtype
	ExpenseCategory string `json:"expenseCategory"`
	Vendor          string `json:"vendor"`
	Receipt         string `json:"receipt"`
	Reimbursable    string `json:"reimbursable"`
}

type JobActivityStatusLabels struct {
	Draft     string `json:"draft"`
	Submitted string `json:"submitted"`
	Approved  string `json:"approved"`
	Rejected  string `json:"rejected"`
}

type JobActivityErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
}

// DefaultJobActivityLabels returns JobActivityLabels with sensible English defaults.
func DefaultJobActivityLabels() JobActivityLabels {
	return JobActivityLabels{
		Page: JobActivityPageLabels{
			Heading: "Activities",
			Caption: "Cross-job timesheet and activity log",
		},
		Buttons: JobActivityButtonLabels{
			AddActivity: "Add Activity",
		},
		Columns: JobActivityColumnLabels{
			Date:        "Date",
			Job:         "Job",
			EntryType:   "Type",
			Description: "Description",
			Quantity:    "Hrs/Qty",
			Amount:      "Amount",
			Status:      "Status",
		},
		Empty: JobActivityEmptyLabels{
			Title:   "No activities found",
			Message: "No activity entries to display.",
		},
		Form: JobActivityFormLabels{
			Job:            "Job",
			Task:           "Task",
			EntryType:      "Entry Type",
			Description:    "Description",
			BillableStatus: "Billable Status",
			Hours:          "Hours",
			HourlyRate:     "Hourly Rate",
			Product:        "Product",
			Quantity:       "Quantity",
			UnitCost:       "Unit Cost",
			Amount:         "Amount",
			Category:       "Category",
		},
		Actions: JobActivityActionLabels{
			View:    "View Activity",
			Edit:    "Edit Activity",
			Delete:  "Delete Activity",
			Submit:  "Submit for Approval",
			Approve: "Approve",
			Reject:  "Reject",
		},
		Detail: JobActivityDetailLabels{
			PageTitle:       "Activity Details",
			TitlePrefix:     "Activity ",
			Job:             "Job",
			EntryType:       "Entry Type",
			EntryDate:       "Date",
			Description:     "Description",
			Quantity:        "Quantity",
			UnitCost:        "Unit Cost",
			TotalCost:       "Total Cost",
			Currency:        "Currency",
			BillableStatus:  "Billable Status",
			ApprovalStatus:  "Approval Status",
			PostingStatus:   "Posting Status",
			CreatedDate:     "Created",
			Staff:           "Staff",
			Hours:           "Hours",
			RateType:        "Rate Type",
			TimeStart:       "Start Time",
			TimeEnd:         "End Time",
			Product:         "Product",
			UnitOfMeasure:   "Unit of Measure",
			LotNumber:       "Lot Number",
			Location:        "Location",
			ExpenseCategory: "Expense Category",
			Vendor:          "Vendor",
			Receipt:         "Receipt",
			Reimbursable:    "Reimbursable",
		},
		Status: JobActivityStatusLabels{
			Draft:     "Draft",
			Submitted: "Submitted",
			Approved:  "Approved",
			Rejected:  "Rejected",
		},
		Errors: JobActivityErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Activity not found",
			IDRequired:       "Activity ID is required",
		},
	}
}

// ---------------------------------------------------------------------------
// Job (operational activity) labels
// ---------------------------------------------------------------------------

// JobLabels holds all translatable strings for the job module.
type JobLabels struct {
	Page    JobPageLabels    `json:"page"`
	Buttons JobButtonLabels  `json:"buttons"`
	Columns JobColumnLabels  `json:"columns"`
	Empty   JobEmptyLabels   `json:"empty"`
	Form    JobFormLabels    `json:"form"`
	Actions JobActionLabels  `json:"actions"`
	Detail  JobDetailLabels  `json:"detail"`
	Tabs    JobTabLabels     `json:"tabs"`
	Confirm JobConfirmLabels `json:"confirm"`
	Errors  JobErrorLabels   `json:"errors"`
}

type JobPageLabels struct {
	Heading          string `json:"heading"`
	Caption          string `json:"caption"`
	HeadingDraft     string `json:"headingDraft"`
	CaptionDraft     string `json:"captionDraft"`
	HeadingActive    string `json:"headingActive"`
	CaptionActive    string `json:"captionActive"`
	HeadingCompleted string `json:"headingCompleted"`
	CaptionCompleted string `json:"captionCompleted"`
	HeadingClosed    string `json:"headingClosed"`
	CaptionClosed    string `json:"captionClosed"`
}

type JobButtonLabels struct {
	AddJob string `json:"addJob"`
}

type JobColumnLabels struct {
	Name     string `json:"name"`
	Client   string `json:"client"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Location string `json:"location"`
}

type JobEmptyLabels struct {
	Title              string `json:"title"`
	Message            string `json:"message"`
	PhasesTitle        string `json:"phasesTitle"`
	PhasesMessage      string `json:"phasesMessage"`
	ActivitiesTitle    string `json:"activitiesTitle"`
	ActivitiesMessage  string `json:"activitiesMessage"`
	SettlementTitle    string `json:"settlementTitle"`
	SettlementMessage  string `json:"settlementMessage"`
	OutcomesTitle      string `json:"outcomesTitle"`
	OutcomesMessage    string `json:"outcomesMessage"`
}

type JobFormLabels struct {
	NamePlaceholder     string `json:"namePlaceholder"`
	ClientPlaceholder   string `json:"clientPlaceholder"`
	LocationPlaceholder string `json:"locationPlaceholder"`
}

type JobActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type JobDetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	TitlePrefix  string `json:"titlePrefix"`
	SectionInfo  string `json:"sectionInfo"`
	Approval     string `json:"approval"`
	Description  string `json:"description"`
	Quantity     string `json:"quantity"`
	UnitCost     string `json:"unitCost"`
	TotalCost    string `json:"totalCost"`
	EntryDate    string `json:"entryDate"`
	EntryType    string `json:"entryType"`
	PhaseName    string `json:"phaseName"`
	PhaseOrder   string `json:"phaseOrder"`
	PhaseStatus  string `json:"phaseStatus"`
	TaskName     string `json:"taskName"`
	TaskOrder    string `json:"taskOrder"`
	TaskStatus   string `json:"taskStatus"`
	AssignedTo   string `json:"assignedTo"`
	TargetType   string `json:"targetType"`
	TargetID     string `json:"targetId"`
	AllocatedAmt string `json:"allocatedAmount"`
	SettleDate   string `json:"settlementDate"`
	SettleStatus string `json:"settlementStatus"`
}

type JobTabLabels struct {
	Info        string `json:"info"`
	Phases      string `json:"phases"`
	Activities  string `json:"activities"`
	Settlement  string `json:"settlement"`
	Outcomes    string `json:"outcomes"`
	Attachments string `json:"attachments"`
}

type JobConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type JobErrorLabels struct {
	NotFound         string `json:"notFound"`
	PermissionDenied string `json:"permissionDenied"`
}

// ---------------------------------------------------------------------------
// Outcome Criteria labels (criteria library)
// ---------------------------------------------------------------------------

// OutcomeCriteriaLabels holds all translatable strings for the outcome criteria module.
type OutcomeCriteriaLabels struct {
	Page    OutcomeCriteriaPageLabels    `json:"page"`
	Buttons OutcomeCriteriaButtonLabels  `json:"buttons"`
	Columns OutcomeCriteriaColumnLabels  `json:"columns"`
	Empty   OutcomeCriteriaEmptyLabels   `json:"empty"`
	Form    OutcomeCriteriaFormLabels    `json:"form"`
	Actions OutcomeCriteriaActionLabels  `json:"actions"`
	Detail  OutcomeCriteriaDetailLabels  `json:"detail"`
	Tabs    OutcomeCriteriaTabLabels     `json:"tabs"`
	Confirm OutcomeCriteriaConfirmLabels `json:"confirm"`
	Errors  OutcomeCriteriaErrorLabels   `json:"errors"`
}

type OutcomeCriteriaPageLabels struct {
	Heading         string `json:"heading"`
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type OutcomeCriteriaButtonLabels struct {
	AddCriterion string `json:"addCriterion"`
}

type OutcomeCriteriaColumnLabels struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Scope   string `json:"scope"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

type OutcomeCriteriaEmptyLabels struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type OutcomeCriteriaFormLabels struct {
	Name            string `json:"name"`
	NamePlaceholder string `json:"namePlaceholder"`
	Type            string `json:"type"`
	Scope           string `json:"scope"`
	Description     string `json:"description"`
	DescPlaceholder string `json:"descriptionPlaceholder"`
	Required        string `json:"required"`
	Weight          string `json:"weight"`
}

type OutcomeCriteriaActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type OutcomeCriteriaDetailLabels struct {
	PageTitle    string `json:"pageTitle"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	Scope        string `json:"scope"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	Required     string `json:"required"`
	Weight       string `json:"weight"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
}

type OutcomeCriteriaTabLabels struct {
	Info       string `json:"info"`
	Thresholds string `json:"thresholds"`
	Options    string `json:"options"`
	Templates  string `json:"templates"`
	Versions   string `json:"versions"`
}

type OutcomeCriteriaConfirmLabels struct {
	Delete        string `json:"delete"`
	DeleteMessage string `json:"deleteMessage"`
}

type OutcomeCriteriaErrorLabels struct {
	PermissionDenied string `json:"permissionDenied"`
	InvalidFormData  string `json:"invalidFormData"`
	NotFound         string `json:"notFound"`
	IDRequired       string `json:"idRequired"`
	NoPermission     string `json:"noPermission"`
}

// DefaultOutcomeCriteriaLabels returns OutcomeCriteriaLabels with sensible English defaults.
func DefaultOutcomeCriteriaLabels() OutcomeCriteriaLabels {
	return OutcomeCriteriaLabels{
		Page: OutcomeCriteriaPageLabels{
			Heading:         "Criteria Library",
			HeadingActive:   "Active Criteria",
			HeadingInactive: "Inactive Criteria",
			Caption:         "Manage reusable outcome evaluation criteria",
			CaptionActive:   "Manage your active outcome criteria",
			CaptionInactive: "View inactive or deprecated criteria",
		},
		Buttons: OutcomeCriteriaButtonLabels{
			AddCriterion: "Add Criterion",
		},
		Columns: OutcomeCriteriaColumnLabels{
			Name:    "Name",
			Type:    "Type",
			Scope:   "Scope",
			Version: "Version",
			Status:  "Status",
		},
		Empty: OutcomeCriteriaEmptyLabels{
			Title:           "No criteria found",
			Message:         "No outcome criteria to display.",
			ActiveTitle:     "No active criteria",
			ActiveMessage:   "Create your first outcome criterion to get started.",
			InactiveTitle:   "No inactive criteria",
			InactiveMessage: "Deactivated criteria will appear here.",
		},
		Form: OutcomeCriteriaFormLabels{
			Name:            "Name",
			NamePlaceholder: "Enter criterion name",
			Type:            "Criteria Type",
			Scope:           "Scope",
			Description:     "Description",
			DescPlaceholder: "Enter criterion description...",
			Required:        "Required",
			Weight:          "Weight",
		},
		Actions: OutcomeCriteriaActionLabels{
			View:   "View Criterion",
			Edit:   "Edit Criterion",
			Delete: "Delete Criterion",
		},
		Detail: OutcomeCriteriaDetailLabels{
			PageTitle:    "Criterion Details",
			Name:         "Name",
			Description:  "Description",
			Type:         "Criteria Type",
			Scope:        "Scope",
			Version:      "Version",
			Status:       "Status",
			Required:     "Required",
			Weight:       "Weight",
			CreatedDate:  "Created",
			ModifiedDate: "Last Modified",
		},
		Tabs: OutcomeCriteriaTabLabels{
			Info:       "Information",
			Thresholds: "Thresholds",
			Options:    "Options",
			Templates:  "Templates",
			Versions:   "Versions",
		},
		Confirm: OutcomeCriteriaConfirmLabels{
			Delete:        "Delete Criterion",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: OutcomeCriteriaErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			InvalidFormData:  "Invalid form data. Please check your inputs and try again.",
			NotFound:         "Outcome criterion not found",
			IDRequired:       "Criterion ID is required",
			NoPermission:     "No permission",
		},
	}
}

// ---------------------------------------------------------------------------
// Task Outcome labels (outcome recording on job tasks)
// ---------------------------------------------------------------------------

// TaskOutcomeLabels holds all translatable strings for the task outcome module.
type TaskOutcomeLabels struct {
	Page    TaskOutcomePageLabels    `json:"page"`
	Buttons TaskOutcomeButtonLabels  `json:"buttons"`
	Columns TaskOutcomeColumnLabels  `json:"columns"`
	Empty   TaskOutcomeEmptyLabels   `json:"empty"`
	Form    TaskOutcomeFormLabels    `json:"form"`
	Actions TaskOutcomeActionLabels  `json:"actions"`
	Detail  TaskOutcomeDetailLabels  `json:"detail"`
	Confirm TaskOutcomeConfirmLabels `json:"confirm"`
	Errors  TaskOutcomeErrorLabels   `json:"errors"`
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
	PageTitle             string `json:"pageTitle"`
	Task                  string `json:"task"`
	Criteria              string `json:"criteria"`
	CriteriaType          string `json:"criteriaType"`
	Value                 string `json:"value"`
	Determination         string `json:"determination"`
	DeterminationSource   string `json:"determinationSource"`
	DeterminationNote     string `json:"determinationNote"`
	RecordedBy            string `json:"recordedBy"`
	RecordedDate          string `json:"recordedDate"`
	RevisionNumber        string `json:"revisionNumber"`
	CreatedDate           string `json:"createdDate"`
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

// ---------------------------------------------------------------------------
// Outcome Summary labels (job/phase report cards)
// ---------------------------------------------------------------------------

// OutcomeSummaryLabels holds all translatable strings for the outcome summary module.
type OutcomeSummaryLabels struct {
	Page    OutcomeSummaryPageLabels    `json:"page"`
	Buttons OutcomeSummaryButtonLabels  `json:"buttons"`
	Columns OutcomeSummaryColumnLabels  `json:"columns"`
	Empty   OutcomeSummaryEmptyLabels   `json:"empty"`
	Detail  OutcomeSummaryDetailLabels  `json:"detail"`
	Errors  OutcomeSummaryErrorLabels   `json:"errors"`
}

type OutcomeSummaryColumnLabels struct {
	Job           string `json:"job"`
	Determination string `json:"determination"`
	Score         string `json:"score"`
	ScoringMethod string `json:"scoringMethod"`
	Total         string `json:"total"`
	Pass          string `json:"pass"`
	Fail          string `json:"fail"`
	IssuedBy      string `json:"issuedBy"`
}

type OutcomeSummaryEmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type OutcomeSummaryPageLabels struct {
	JobHeading   string `json:"jobHeading"`
	JobCaption   string `json:"jobCaption"`
	PhaseHeading string `json:"phaseHeading"`
	PhaseCaption string `json:"phaseCaption"`
}

type OutcomeSummaryButtonLabels struct {
	GenerateSummary string `json:"generateSummary"`
}

type OutcomeSummaryDetailLabels struct {
	OverallDetermination string `json:"overallDetermination"`
	PhaseDetermination   string `json:"phaseDetermination"`
	Score                string `json:"score"`
	ScoringMethod        string `json:"scoringMethod"`
	TotalCriteria        string `json:"totalCriteria"`
	PassCount            string `json:"passCount"`
	FailCount            string `json:"failCount"`
	ConditionalCount     string `json:"conditionalCount"`
	DeferredCount        string `json:"deferredCount"`
	NaCount              string `json:"naCount"`
	Narrative            string `json:"narrative"`
	IssuedBy             string `json:"issuedBy"`
	IssuedDate           string `json:"issuedDate"`
	ValidUntilDate       string `json:"validUntilDate"`
}

type OutcomeSummaryErrorLabels struct {
	NotFound         string `json:"notFound"`
	PermissionDenied string `json:"permissionDenied"`
}

// DefaultOutcomeSummaryLabels returns OutcomeSummaryLabels with sensible English defaults.
func DefaultOutcomeSummaryLabels() OutcomeSummaryLabels {
	return OutcomeSummaryLabels{
		Page: OutcomeSummaryPageLabels{
			JobHeading:   "Outcome Summary",
			JobCaption:   "Job-level outcome report card",
			PhaseHeading: "Phase Outcome Summary",
			PhaseCaption: "Phase-level outcome report card",
		},
		Buttons: OutcomeSummaryButtonLabels{
			GenerateSummary: "Generate Summary",
		},
		Columns: OutcomeSummaryColumnLabels{
			Job:           "Job",
			Determination: "Determination",
			Score:         "Score",
			ScoringMethod: "Scoring Method",
			Total:         "Total",
			Pass:          "Pass",
			Fail:          "Fail",
			IssuedBy:      "Issued By",
		},
		Empty: OutcomeSummaryEmptyLabels{
			Title:   "No summaries",
			Message: "No outcome summaries have been generated yet.",
		},
		Detail: OutcomeSummaryDetailLabels{
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
		Errors: OutcomeSummaryErrorLabels{
			NotFound:         "Outcome summary not found",
			PermissionDenied: "You do not have permission to perform this action",
		},
	}
}

// DefaultJobLabels returns JobLabels with sensible English defaults.
func DefaultJobLabels() JobLabels {
	return JobLabels{
		Page: JobPageLabels{
			Heading:          "Jobs",
			Caption:          "Manage operational jobs",
			HeadingDraft:     "Draft Jobs",
			CaptionDraft:     "Review jobs that are still being prepared",
			HeadingActive:    "Active Jobs",
			CaptionActive:    "Track work currently in progress",
			HeadingCompleted: "Completed Jobs",
			CaptionCompleted: "Review jobs that have been completed",
			HeadingClosed:    "Closed Jobs",
			CaptionClosed:    "View archived or closed jobs",
		},
		Buttons: JobButtonLabels{
			AddJob: "Add Job",
		},
		Columns: JobColumnLabels{
			Name:     "Name",
			Client:   "Client",
			Status:   "Status",
			Created:  "Created",
			Location: "Location",
		},
		Empty: JobEmptyLabels{
			Title:             "No jobs found",
			Message:           "No jobs to display.",
			PhasesTitle:       "No phases",
			PhasesMessage:     "This job has no phases defined yet.",
			ActivitiesTitle:   "No activities",
			ActivitiesMessage: "No activity entries have been recorded for this job yet.",
			SettlementTitle:   "No settlements",
			SettlementMessage: "No cost allocations have been settled for this job yet.",
			OutcomesTitle:     "No outcomes",
			OutcomesMessage:   "No outcome evaluations have been recorded for this job yet.",
		},
		Form: JobFormLabels{
			NamePlaceholder:     "Enter job name",
			ClientPlaceholder:   "Select client",
			LocationPlaceholder: "Select location",
		},
		Actions: JobActionLabels{
			View:   "View Job",
			Edit:   "Edit Job",
			Delete: "Delete Job",
		},
		Detail: JobDetailLabels{
			PageTitle:    "Job Details",
			TitlePrefix:  "Job ",
			SectionInfo:  "Job Information",
			Approval:     "Approval",
			Description:  "Description",
			Quantity:     "Quantity",
			UnitCost:     "Unit Cost",
			TotalCost:    "Total Cost",
			EntryDate:    "Date",
			EntryType:    "Entry Type",
			PhaseName:    "Phase",
			PhaseOrder:   "Order",
			PhaseStatus:  "Status",
			TaskName:     "Task",
			TaskOrder:    "Task Order",
			TaskStatus:   "Task Status",
			AssignedTo:   "Assigned To",
			TargetType:   "Target Type",
			TargetID:     "Target ID",
			AllocatedAmt: "Allocated Amount",
			SettleDate:   "Settlement Date",
			SettleStatus: "Settlement Status",
		},
		Tabs: JobTabLabels{
			Info:        "Information",
			Phases:      "Phases",
			Activities:  "Activities",
			Settlement:  "Settlement",
			Outcomes:    "Outcomes",
			Attachments: "Attachments",
		},
		Confirm: JobConfirmLabels{
			Delete:        "Delete Job",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: JobErrorLabels{
			NotFound:         "Job not found",
			PermissionDenied: "You do not have permission to perform this action",
		},
	}
}
