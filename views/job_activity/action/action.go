package action

import (
	"context"
	"strconv"

	fayna "github.com/erniealice/fayna-golang"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// Deps holds the dependencies needed by the job activity action handlers.
type Deps struct {
	Routes fayna.JobActivityRoutes
	Labels fayna.JobActivityLabels
	NewID  func() string

	// Use case functions
	CreateJobActivity func(ctx context.Context, req *jobactivitypb.CreateJobActivityRequest) (*jobactivitypb.CreateJobActivityResponse, error)
	ReadJobActivity   func(ctx context.Context, req *jobactivitypb.ReadJobActivityRequest) (*jobactivitypb.ReadJobActivityResponse, error)
	UpdateJobActivity func(ctx context.Context, req *jobactivitypb.UpdateJobActivityRequest) (*jobactivitypb.UpdateJobActivityResponse, error)
	DeleteJobActivity func(ctx context.Context, req *jobactivitypb.DeleteJobActivityRequest) (*jobactivitypb.DeleteJobActivityResponse, error)

	SubmitForApproval func(ctx context.Context, req *jobactivitypb.SubmitForApprovalRequest) (*jobactivitypb.SubmitForApprovalResponse, error)
	ApproveActivity   func(ctx context.Context, req *jobactivitypb.ApproveJobActivityRequest) (*jobactivitypb.ApproveJobActivityResponse, error)
	RejectActivity    func(ctx context.Context, req *jobactivitypb.RejectJobActivityRequest) (*jobactivitypb.RejectJobActivityResponse, error)
	PostActivity      func(ctx context.Context, req *jobactivitypb.PostJobActivityRequest) (*jobactivitypb.PostJobActivityResponse, error)
	ReverseActivity   func(ctx context.Context, req *jobactivitypb.ReverseJobActivityRequest) (*jobactivitypb.ReverseJobActivityResponse, error)

	// GenerateInvoiceFromActivities creates a revenue record from a set of
	// activity IDs. Returns the new revenue ID on success.
	GenerateInvoiceFromActivities func(ctx context.Context, activityIDs []string, clientID, locationID, currency, name string) (string, error)
}

// parseFormFloat parses a float64 from a form value, returning 0 on error.
func parseFormFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseEntryType converts a form string to the EntryType enum. The drawer
// form posts canonical proto enum names (ENTRY_TYPE_*); legacy lowercase
// shorthand ("labor", "material", ...) is still accepted so older tests and
// any non-form caller don't break.
func parseEntryType(s string) jobactivitypb.EntryType {
	switch s {
	case "ENTRY_TYPE_LABOR", "LABOR", "labor":
		return jobactivitypb.EntryType_ENTRY_TYPE_LABOR
	case "ENTRY_TYPE_MATERIAL", "MATERIAL", "material":
		return jobactivitypb.EntryType_ENTRY_TYPE_MATERIAL
	case "ENTRY_TYPE_EXPENSE", "EXPENSE", "expense":
		return jobactivitypb.EntryType_ENTRY_TYPE_EXPENSE
	case "ENTRY_TYPE_EQUIPMENT", "EQUIPMENT", "equipment":
		return jobactivitypb.EntryType_ENTRY_TYPE_EQUIPMENT
	case "ENTRY_TYPE_SUBCONTRACT", "SUBCONTRACT", "subcontract":
		return jobactivitypb.EntryType_ENTRY_TYPE_SUBCONTRACT
	default:
		return jobactivitypb.EntryType_ENTRY_TYPE_UNSPECIFIED
	}
}

// parseBillableStatus converts a form string to the BillableStatus enum.
// Accepts both shorthand ("billable") and proto enum form
// ("BILLABLE_STATUS_BILLABLE") so e2e selectors and form submits stay
// flexible.
func parseBillableStatus(s string) jobactivitypb.BillableStatus {
	switch s {
	case "billable", "BILLABLE_STATUS_BILLABLE":
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_BILLABLE
	case "non_billable", "BILLABLE_STATUS_NON_BILLABLE":
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_NON_BILLABLE
	case "included", "BILLABLE_STATUS_INCLUDED":
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_INCLUDED
	case "write_off", "BILLABLE_STATUS_WRITE_OFF":
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_WRITE_OFF
	default:
		return jobactivitypb.BillableStatus_BILLABLE_STATUS_UNSPECIFIED
	}
}
