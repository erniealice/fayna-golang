// Package action contains HTTP/HTMX handlers for the job_task view module.
// Dependency-bearing helpers that need full Deps live here; pure-function
// builders live in the sibling form/ package.
package action

import (
	"context"

	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	fayna "github.com/erniealice/fayna-golang"
)

// Deps holds dependencies shared across all job_task action handlers.
type Deps struct {
	Routes fayna.JobTaskRoutes
	Labels fayna.JobTaskLabels

	// Job task CRUD
	CreateJobTask func(ctx context.Context, req *jobtaskpb.CreateJobTaskRequest) (*jobtaskpb.CreateJobTaskResponse, error)
	ReadJobTask   func(ctx context.Context, req *jobtaskpb.ReadJobTaskRequest) (*jobtaskpb.ReadJobTaskResponse, error)
	UpdateJobTask func(ctx context.Context, req *jobtaskpb.UpdateJobTaskRequest) (*jobtaskpb.UpdateJobTaskResponse, error)
	DeleteJobTask func(ctx context.Context, req *jobtaskpb.DeleteJobTaskRequest) (*jobtaskpb.DeleteJobTaskResponse, error)
	ListJobTasks  func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)

	// Search URLs for the auto-complete pickers in the Add/Edit drawer.
	StaffSearchURL        string
	ResourceSearchURL     string
	TemplateTaskSearchURL string
}

// taskStatusToEnum converts a status string (proto enum name or shorthand)
// to the protobuf TaskStatus enum.
// Canonical conversion — also used by set_status.go and bulk_set_status.
func taskStatusToEnum(status string) jobtaskpb.TaskStatus {
	switch status {
	case "TASK_STATUS_PENDING", "pending":
		return jobtaskpb.TaskStatus_TASK_STATUS_PENDING
	case "TASK_STATUS_IN_PROGRESS", "in_progress":
		return jobtaskpb.TaskStatus_TASK_STATUS_IN_PROGRESS
	case "TASK_STATUS_COMPLETED", "completed":
		return jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED
	case "TASK_STATUS_SKIPPED", "skipped":
		return jobtaskpb.TaskStatus_TASK_STATUS_SKIPPED
	case "TASK_STATUS_HOLD", "hold":
		return jobtaskpb.TaskStatus_TASK_STATUS_HOLD
	case "TASK_STATUS_REWORK", "rework":
		return jobtaskpb.TaskStatus_TASK_STATUS_REWORK
	default:
		return jobtaskpb.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

// taskStatusString converts a TaskStatus enum to a lowercase display string.
func taskStatusString(s jobtaskpb.TaskStatus) string {
	switch s {
	case jobtaskpb.TaskStatus_TASK_STATUS_PENDING:
		return "pending"
	case jobtaskpb.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return "in_progress"
	case jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED:
		return "completed"
	case jobtaskpb.TaskStatus_TASK_STATUS_SKIPPED:
		return "skipped"
	case jobtaskpb.TaskStatus_TASK_STATUS_HOLD:
		return "hold"
	case jobtaskpb.TaskStatus_TASK_STATUS_REWORK:
		return "rework"
	default:
		return "pending"
	}
}
