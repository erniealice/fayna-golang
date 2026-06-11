package job_task

// job_task_labels.go — JobTask label structs + DefaultJobTaskLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// JobTaskLabels holds all translatable strings for the job_task module.
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
	Heading           string `json:"heading"`
	Caption           string `json:"caption"`
	HeadingPending    string `json:"headingPending"`
	HeadingInProgress string `json:"headingInProgress"`
	HeadingCompleted  string `json:"headingCompleted"`
}

type ButtonLabels struct {
	AddTask string `json:"addTask"`
}

type ColumnLabels struct {
	Name              string `json:"name"`
	Phase             string `json:"phase"`
	StepOrder         string `json:"stepOrder"`
	Status            string `json:"status"`
	AssignedTo        string `json:"assignedTo"`
	PercentComplete   string `json:"percentComplete"`
	PlannedQuantity   string `json:"plannedQuantity"`
	CompletedQuantity string `json:"completedQuantity"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type FormLabels struct {
	SectionTask               string `json:"sectionTask"`
	SectionAssignmentResource string `json:"sectionAssignmentResource"`
	SectionSchedule           string `json:"sectionSchedule"`
	SectionActuals            string `json:"sectionActuals"`
	Name                      string `json:"name"`
	NamePlaceholder           string `json:"namePlaceholder"`
	StepOrder                 string `json:"stepOrder"`
	Status                    string `json:"status"`
	IsAdHoc                   string `json:"isAdHoc"`
	AssignedTo                string `json:"assignedTo"`
	AssignedToPlaceholder     string `json:"assignedToPlaceholder"`
	ResourceID                string `json:"resourceId"`
	ResourcePlaceholder       string `json:"resourcePlaceholder"`
	TemplateTaskID            string `json:"templateTaskId"`
	TemplateTaskPlaceholder   string `json:"templateTaskPlaceholder"`
	PlannedQuantity           string `json:"plannedQuantity"`
	CompletedQuantity         string `json:"completedQuantity"`
	PercentComplete           string `json:"percentComplete"`
	AllowParallel             string `json:"allowParallel"`
	ActualStart               string `json:"actualStart"`
	ActualEnd                 string `json:"actualEnd"`
}

type ActionLabels struct {
	View   string `json:"view"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
}

type DetailLabels struct {
	PageTitle         string `json:"pageTitle"`
	Name              string `json:"name"`
	Phase             string `json:"phase"`
	StepOrder         string `json:"stepOrder"`
	Status            string `json:"status"`
	IsAdHoc           string `json:"isAdHoc"`
	AssignedTo        string `json:"assignedTo"`
	Resource          string `json:"resource"`
	TemplateTask      string `json:"templateTask"`
	PlannedQuantity   string `json:"plannedQuantity"`
	CompletedQuantity string `json:"completedQuantity"`
	PercentComplete   string `json:"percentComplete"`
	AllowParallel     string `json:"allowParallel"`
	ActualStart       string `json:"actualStart"`
	ActualEnd         string `json:"actualEnd"`
}

type TabLabels struct {
	Info        string `json:"info"`
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

// DefaultJobTaskLabels returns JobTaskLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Page: PageLabels{
			Heading:           "Job Tasks",
			Caption:           "Manage execution tasks across job phases",
			HeadingPending:    "Pending Tasks",
			HeadingInProgress: "In-Progress Tasks",
			HeadingCompleted:  "Completed Tasks",
		},
		Buttons: ButtonLabels{
			AddTask: "Add Task",
		},
		Columns: ColumnLabels{
			Name:              "Name",
			Phase:             "Phase",
			StepOrder:         "#",
			Status:            "Status",
			AssignedTo:        "Assigned To",
			PercentComplete:   "% Done",
			PlannedQuantity:   "Planned Qty",
			CompletedQuantity: "Completed Qty",
		},
		Empty: EmptyLabels{
			Title:   "No tasks found",
			Message: "No job tasks to display.",
		},
		Form: FormLabels{
			SectionTask:               "Task",
			SectionAssignmentResource: "Assignment & Resource",
			SectionSchedule:           "Schedule",
			SectionActuals:            "Actuals",
			Name:                      "Task Name",
			NamePlaceholder:           "Enter task name",
			StepOrder:                 "Order",
			Status:                    "Status",
			IsAdHoc:                   "Ad Hoc",
			AssignedTo:                "Assigned To",
			AssignedToPlaceholder:     "Search staff...",
			ResourceID:                "Resource",
			ResourcePlaceholder:       "Search resource...",
			TemplateTaskID:            "Template Task",
			TemplateTaskPlaceholder:   "Search template task...",
			PlannedQuantity:           "Planned Qty",
			CompletedQuantity:         "Completed Qty",
			PercentComplete:           "% Complete",
			AllowParallel:             "Allow Parallel",
			ActualStart:               "Actual Start",
			ActualEnd:                 "Actual End",
		},
		Actions: ActionLabels{
			View:   "View Task",
			Edit:   "Edit Task",
			Delete: "Delete Task",
		},
		Detail: DetailLabels{
			PageTitle:         "Task Details",
			Name:              "Name",
			Phase:             "Phase",
			StepOrder:         "Order",
			Status:            "Status",
			IsAdHoc:           "Ad Hoc",
			AssignedTo:        "Assigned To",
			Resource:          "Resource",
			TemplateTask:      "Template Task",
			PlannedQuantity:   "Planned Qty",
			CompletedQuantity: "Completed Qty",
			PercentComplete:   "% Complete",
			AllowParallel:     "Allow Parallel",
			ActualStart:       "Actual Start",
			ActualEnd:         "Actual End",
		},
		Tabs: TabLabels{
			Info:        "Information",
			Activities:  "Activities",
			Attachments: "Attachments",
			History:     "History",
		},
		Confirm: ConfirmLabels{
			Delete:        "Delete Task",
			DeleteMessage: "Are you sure you want to delete \"%s\"? This action cannot be undone.",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Task not found",
			IDRequired:       "Task ID is required",
		},
	}
}
