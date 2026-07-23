package form

import (
	"github.com/erniealice/fayna-golang/domain/operation/job_template"
	"github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the job template drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// No mapper: the Labels field is job_template.Labels verbatim — templates
// read .Labels.Columns.Name, .Labels.Form.NamePlaceholder, etc. directly.
type Data struct {
	FormAction  string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit      bool
	ID          string
	Name        string
	Description string
	Active      bool

	// CategoryID / OutputProductID — both optional FKs. Templates for
	// non-education tiers may leave either unset. CategoryOptions /
	// OutputProductOptions are pre-built select options (BuildCategoryOptions /
	// BuildOutputProductOptions in options.go); an empty slice renders an
	// otherwise-empty picker rather than failing the drawer.
	CategoryID           string
	CategoryOptions      []types.SelectOption
	OutputProductID      string
	OutputProductOptions []types.SelectOption

	// InitialStatus — the JobStatus enum name a spawned Job takes when this
	// template materializes (job_template.proto field 50). Optional; NULL
	// falls back to JOB_STATUS_PLANNED at spawn time. See form/options.go
	// BuildInitialStatusOptions.
	InitialStatus        string
	InitialStatusOptions []types.SelectOption

	// VersionStatus — the template's publication state (VersionStatus enum name,
	// e.g. "VERSION_STATUS_PUBLISHED"). Defaults to Draft on create. See
	// form/options.go BuildVersionStatusOptions. No approval-ladder mechanics.
	VersionStatus        string
	VersionStatusOptions []types.SelectOption

	Labels       job_template.Labels
	CommonLabels any
}
