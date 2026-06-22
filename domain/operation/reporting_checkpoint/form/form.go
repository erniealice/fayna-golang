package form

import (
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the reporting checkpoint drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent reporting_checkpoint package. Templates read .Labels.* via
// Go template reflection — no cast required.
type Data struct {
	FormAction           string
	WorkspaceID          string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit               bool
	ID                   string
	Label                string
	CheckpointGroupID    string
	RoleCode             string
	SequenceOrder        int32
	WorkspaceIDField     string // optional FK — empty = unset (NULL)
	PeriodID             string // optional FK — empty = unset (NULL)
	IsTerminal           bool
	VersionStatus        string
	VersionStatusOptions []pyeza.SelectOption
	Labels               any
	CommonLabels         any
}

// DefaultVersionStatusOptions returns the selectable version status options.
func DefaultVersionStatusOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "VERSION_STATUS_DRAFT", Label: "Draft"},
		{Value: "VERSION_STATUS_PUBLISHED", Label: "Published"},
		{Value: "VERSION_STATUS_DEPRECATED", Label: "Deprecated"},
	}
}
