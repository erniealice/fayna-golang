package action

import (
	"context"

	job "github.com/erniealice/fayna-golang/domain/operation/job"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// Deps holds dependencies for job action handlers.
type Deps struct {
	Routes    job.Routes
	Labels    job.Labels
	CreateJob func(ctx context.Context, req *jobpb.CreateJobRequest) (*jobpb.CreateJobResponse, error)
	ReadJob   func(ctx context.Context, req *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error)
	UpdateJob func(ctx context.Context, req *jobpb.UpdateJobRequest) (*jobpb.UpdateJobResponse, error)
	DeleteJob func(ctx context.Context, req *jobpb.DeleteJobRequest) (*jobpb.DeleteJobResponse, error)
	ListJobs  func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)

	// Search endpoints for the job drawer client + location pickers.
	// Set from job.Routes.ClientSearchURL / LocationSearchURL by the fayna block.
	ClientSearchURL   string
	LocationSearchURL string
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}

// jobStatusToEnum converts a status string to the protobuf JobStatus enum.
// Accepts proto enum name (JOB_STATUS_*) and legacy lowercase shorthands.
func jobStatusToEnum(status string) enums.JobStatus {
	switch status {
	case "JOB_STATUS_DRAFT", "draft":
		return enums.JobStatus_JOB_STATUS_DRAFT
	case "JOB_STATUS_PENDING", "pending":
		return enums.JobStatus_JOB_STATUS_PENDING
	case "JOB_STATUS_PLANNED", "planned":
		return enums.JobStatus_JOB_STATUS_PLANNED
	case "JOB_STATUS_RELEASED", "released":
		return enums.JobStatus_JOB_STATUS_RELEASED
	case "JOB_STATUS_ACTIVE", "active":
		return enums.JobStatus_JOB_STATUS_ACTIVE
	case "JOB_STATUS_PAUSED", "paused":
		return enums.JobStatus_JOB_STATUS_PAUSED
	case "JOB_STATUS_COMPLETED", "completed":
		return enums.JobStatus_JOB_STATUS_COMPLETED
	case "JOB_STATUS_CLOSED", "closed":
		return enums.JobStatus_JOB_STATUS_CLOSED
	default:
		return enums.JobStatus_JOB_STATUS_UNSPECIFIED
	}
}

// billingRuleTypeToEnum converts a billing_rule_type form string to the
// protobuf BillingRuleType enum.
func billingRuleTypeToEnum(s string) enums.BillingRuleType {
	switch s {
	case "BILLING_RULE_TYPE_T_AND_M":
		return enums.BillingRuleType_BILLING_RULE_TYPE_T_AND_M
	case "BILLING_RULE_TYPE_FIXED_FEE":
		return enums.BillingRuleType_BILLING_RULE_TYPE_FIXED_FEE
	case "BILLING_RULE_TYPE_MILESTONE":
		return enums.BillingRuleType_BILLING_RULE_TYPE_MILESTONE
	case "BILLING_RULE_TYPE_INCLUDED":
		return enums.BillingRuleType_BILLING_RULE_TYPE_INCLUDED
	case "BILLING_RULE_TYPE_NON_BILLABLE":
		return enums.BillingRuleType_BILLING_RULE_TYPE_NON_BILLABLE
	default:
		return enums.BillingRuleType_BILLING_RULE_TYPE_UNSPECIFIED
	}
}
