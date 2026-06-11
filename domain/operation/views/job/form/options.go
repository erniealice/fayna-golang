package form

import (
	"github.com/erniealice/pyeza-golang/types"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
)

// BuildStatusOptions returns select options for the job status field.
// Values match the proto enum string names (JOB_STATUS_*); UNSPECIFIED is skipped.
// The option whose value equals current is marked Selected.
func BuildStatusOptions(current string) []types.SelectOption {
	rows := []struct {
		value string
		enum  enums.JobStatus
		label string
	}{
		{"JOB_STATUS_DRAFT", enums.JobStatus_JOB_STATUS_DRAFT, "Draft"},
		{"JOB_STATUS_PENDING", enums.JobStatus_JOB_STATUS_PENDING, "Pending"},
		{"JOB_STATUS_PLANNED", enums.JobStatus_JOB_STATUS_PLANNED, "Planned"},
		{"JOB_STATUS_RELEASED", enums.JobStatus_JOB_STATUS_RELEASED, "Released"},
		{"JOB_STATUS_ACTIVE", enums.JobStatus_JOB_STATUS_ACTIVE, "Active"},
		{"JOB_STATUS_PAUSED", enums.JobStatus_JOB_STATUS_PAUSED, "Paused"},
		{"JOB_STATUS_COMPLETED", enums.JobStatus_JOB_STATUS_COMPLETED, "Completed"},
		{"JOB_STATUS_CLOSED", enums.JobStatus_JOB_STATUS_CLOSED, "Closed"},
	}
	opts := make([]types.SelectOption, 0, len(rows))
	for _, r := range rows {
		_ = r.enum // import-guard: ensures enum const is referenced
		opts = append(opts, types.SelectOption{
			Value:    r.value,
			Label:    r.label,
			Selected: r.value == current,
		})
	}
	return opts
}

// BuildBillingRuleOptions returns select options for the billing_rule_type field.
// Values match the proto enum string names (BILLING_RULE_TYPE_*); UNSPECIFIED is skipped.
func BuildBillingRuleOptions(current string) []types.SelectOption {
	rows := []struct {
		value string
		enum  enums.BillingRuleType
		label string
	}{
		{"BILLING_RULE_TYPE_T_AND_M", enums.BillingRuleType_BILLING_RULE_TYPE_T_AND_M, "Time & Materials"},
		{"BILLING_RULE_TYPE_FIXED_FEE", enums.BillingRuleType_BILLING_RULE_TYPE_FIXED_FEE, "Fixed Fee"},
		{"BILLING_RULE_TYPE_MILESTONE", enums.BillingRuleType_BILLING_RULE_TYPE_MILESTONE, "Milestone"},
		{"BILLING_RULE_TYPE_INCLUDED", enums.BillingRuleType_BILLING_RULE_TYPE_INCLUDED, "Included"},
		{"BILLING_RULE_TYPE_NON_BILLABLE", enums.BillingRuleType_BILLING_RULE_TYPE_NON_BILLABLE, "Non-Billable"},
	}
	opts := make([]types.SelectOption, 0, len(rows))
	for _, r := range rows {
		_ = r.enum // import-guard: ensures enum const is referenced
		opts = append(opts, types.SelectOption{
			Value:    r.value,
			Label:    r.label,
			Selected: r.value == current,
		})
	}
	return opts
}
