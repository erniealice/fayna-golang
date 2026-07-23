package form

import "github.com/erniealice/pyeza-golang/types"

// Context discriminates the drawer's entry point (see ui-drawer-form-
// template-anatomy § Context Discriminator Pattern).
type Context string

const (
	// ContextStandalone — opened from the module's own list page (or a bare
	// /action/template-task-criteria/add call with no template context).
	// JobTemplateTaskID falls back to a raw-id text input — there is no
	// template to scope a picker against.
	ContextStandalone Context = "Standalone"
	// ContextTemplate — opened from a job_template detail Standards tab
	// (?job_template_id=). JobTemplateTaskID becomes a picker scoped to that
	// template's tasks (BuildTemplateTaskOptions); TemplateID rides along as
	// a hidden input so the picker options and the post-success table
	// refresh both stay template-scoped.
	ContextTemplate Context = "Template"
)

// Data is the template-facing data shape for the template task criteria drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent template_task_criteria package. Templates read .Labels.* via
// Go template reflection — no cast required.
type Data struct {
	FormAction  string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit      bool
	ID          string

	// Context discriminates Standalone vs Template entry — see Context above.
	Context Context
	// TemplateID is the parent job_template id when Context == ContextTemplate.
	// Hidden input; used to scope the JobTemplateTaskID picker and to route
	// the post-success table refresh back to the Standards tab.
	TemplateID string

	// FK fields
	JobTemplateTaskID string
	// TaskOptions is non-nil only when Context == ContextTemplate — the
	// template renders a picker when populated, else a raw-id text input.
	TaskOptions       []types.SelectOption
	OutcomeCriteriaID string
	// CriteriaOptions is non-nil when ListOutcomeCriterias is wired — the
	// template renders a picker when populated, else a raw-id text input.
	CriteriaOptions []types.SelectOption
	SequenceOrder   int32
	Active          bool

	Labels       any
	CommonLabels any
}
