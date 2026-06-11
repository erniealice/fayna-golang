package job_template_task

// job_template_task_labels.go — JobTemplateTask label structs + DefaultJobTemplateTaskLabels constructor.
//
// Extracted verbatim from packages/fayna-golang/labels.go (operation domain, W1).
// Pure structural move — no behaviour change; strings are byte-identical.

// JobTemplateTaskLabels holds all translatable strings for the job_template_task
// drawer-only view module. This module has no list page, no sidebar entry, and no
// standalone detail page — operators reach it only via the JobTemplate detail Tasks tab.
type Labels struct {
	Columns ColumnLabels `json:"columns"`
	Form    FormLabels   `json:"form"`
	Actions ActionLabels `json:"actions"`
	Errors  ErrorLabels  `json:"errors"`
}

type ColumnLabels struct {
	Name        string `json:"name"`
	StepOrder   string `json:"stepOrder"`
	EstDuration string `json:"estDuration"`
	Phase       string `json:"phase"`
}

type FormLabels struct {
	SectionTask         string `json:"sectionTask"`
	SectionResource     string `json:"sectionResource"`
	Name                string `json:"name"`
	NamePlaceholder     string `json:"namePlaceholder"`
	StepOrder           string `json:"stepOrder"`
	EstDurationMinutes  string `json:"estDurationMinutes"`
	Resource            string `json:"resource"`
	ResourcePlaceholder string `json:"resourcePlaceholder"`
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

// DefaultJobTemplateTaskLabels returns JobTemplateTaskLabels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Columns: ColumnLabels{
			Name:        "Name",
			StepOrder:   "#",
			EstDuration: "Est. Duration (min)",
			Phase:       "Phase",
		},
		Form: FormLabels{
			SectionTask:         "Task",
			SectionResource:     "Resource",
			Name:                "Task Name",
			NamePlaceholder:     "Enter task name",
			StepOrder:           "Order",
			EstDurationMinutes:  "Estimated Duration (min)",
			Resource:            "Resource",
			ResourcePlaceholder: "Search resource...",
		},
		Actions: ActionLabels{
			Add:    "+ Add Task",
			Edit:   "Edit Task",
			Delete: "Delete Task",
		},
		Errors: ErrorLabels{
			PermissionDenied: "You do not have permission to perform this action",
			NotFound:         "Template task not found",
			IDRequired:       "Template task ID is required",
		},
	}
}
