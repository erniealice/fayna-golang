package form

import (
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the evaluation template header
// drawer form (DF-2 template header). Used by both Add (FormAction = AddURL,
// IsEdit = false) and Edit (FormAction = EditURL, IsEdit = true).
//
// Labels is typed any to avoid an import cycle between the form sub-package and
// the parent evaluation_template package. Templates read .Labels.* via Go
// template reflection — no cast required.
type Data struct {
	FormAction       string
	WorkspaceID      string // injected by ViewAdapter for the action_workspace_guard
	IsEdit           bool
	ID               string
	Name             string
	Description      string
	EvaluationType   string
	RelationshipType string
	VisibilityType   string

	EvaluationTypeOptions []pyeza.SelectOption
	RelationshipOptions   []pyeza.SelectOption
	VisibilityOptions     []pyeza.SelectOption

	Labels       any
	CommonLabels any
}

// DefaultEvaluationTypeOptions returns the selectable evaluation_type options.
// Values are the proto enum short names (parsed back via the action's
// parseEvaluationType which prefixes EVALUATION_TYPE_).
func DefaultEvaluationTypeOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "PERFORMANCE_REVIEW", Label: "Performance Review"},
		{Value: "CSAT", Label: "CSAT"},
		{Value: "COURSE_EVAL", Label: "Course Evaluation"},
		{Value: "VENDOR_SCORECARD", Label: "Vendor Scorecard"},
	}
}

// DefaultRelationshipOptions returns the selectable relationship_type options.
func DefaultRelationshipOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "CLIENT_TO_ASSOCIATE", Label: "Client rates Associate"},
		{Value: "STAFF_TO_CLIENT", Label: "Staff rates Client"},
		{Value: "SELF", Label: "Self"},
		{Value: "PEER", Label: "Peer"},
		{Value: "MANAGER", Label: "Manager"},
	}
}

// DefaultVisibilityOptions returns the selectable visibility_type options.
func DefaultVisibilityOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "INTERNAL_ONLY", Label: "Internal Only"},
		{Value: "VISIBLE_TO_SUBJECT", Label: "Visible to Subject"},
		{Value: "VISIBLE_TO_SUBJECT_ANON", Label: "Visible to Subject (Anonymous)"},
	}
}
