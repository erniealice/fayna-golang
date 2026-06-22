package form

import (
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the score scale drawer form.
// Used by both Add (IsEdit = false) and Edit (IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent score_scale package.
type Data struct {
	FormAction           string
	WorkspaceID          string // injected by ViewAdapter for action_workspace_guard
	IsEdit               bool
	ID                   string
	Name                 string
	ScaleKind            string // proto enum string name e.g. "SCALE_KIND_RANGE_MAP"
	VersionStatus        string // proto enum string name e.g. "VERSION_STATUS_DRAFT"
	InputUnit            string
	InputMin             string // string so optional: empty = unset
	InputMax             string // string so optional: empty = unset
	OutputUnit           string
	ScaleGroupId         string
	ScaleKindOptions     []pyeza.SelectOption
	VersionStatusOptions []pyeza.SelectOption
	Labels               any
	CommonLabels         any
}

// DefaultScaleKindOptions returns selectable scale kind options.
// Values are proto enum string names for round-trip fidelity.
func DefaultScaleKindOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "SCALE_KIND_UNSPECIFIED", Label: "Unspecified"},
		{Value: "SCALE_KIND_RANGE_MAP", Label: "Range Map"},
		{Value: "SCALE_KIND_EXACT_MAP", Label: "Exact Map"},
	}
}

// DefaultVersionStatusOptions returns selectable version status options.
func DefaultVersionStatusOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "VERSION_STATUS_DRAFT", Label: "Draft"},
		{Value: "VERSION_STATUS_PUBLISHED", Label: "Published"},
		{Value: "VERSION_STATUS_DEPRECATED", Label: "Deprecated"},
	}
}
