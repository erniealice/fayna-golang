package form

import "github.com/erniealice/pyeza-golang/types"

// Context discriminates the drawer's entry point (see ui-drawer-form-
// template-anatomy § Context Discriminator Pattern).
type Context string

const (
	// ContextStandalone — opened from the module's own list page. Both the
	// Parent and Child template pickers are shown.
	ContextStandalone Context = "Standalone"
	// ContextTemplate — opened from a job_template detail Spawn Graph tab
	// (?parent_template_id=). ParentTemplateID is pre-locked (hidden input
	// above sheet-body, display-only row inside); only the Child Template
	// picker is editable.
	ContextTemplate Context = "Template"
)

// Data is the template-facing data shape for the job template relation
// drawer form. Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent job_template_relation package. Templates read .Labels.* via
// Go template reflection — no cast required.
type Data struct {
	FormAction  string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit      bool
	ID          string

	// Context discriminates Standalone vs Template entry — see Context above.
	Context Context

	// ParentTemplateID / ChildTemplateID — both FK pickers, populated via
	// ListJobTemplates (BuildTemplateOptions in options.go). When Context ==
	// ContextTemplate, ParentTemplateID is pre-locked (hidden input) and
	// ParentTemplateName is shown as read-only display text.
	ParentTemplateID      string
	ParentTemplateName    string
	ParentTemplateOptions []types.SelectOption
	ChildTemplateID       string
	ChildTemplateOptions  []types.SelectOption

	// RelationType — JobTemplateRelationType enum name (SUB_TEMPLATE /
	// ONCE_AT_ENGAGEMENT_START). See options.go BuildRelationTypeOptions.
	RelationType        string
	RelationTypeOptions []types.SelectOption

	SequenceOrder int32
	Active        bool

	Labels       any
	CommonLabels any
}
