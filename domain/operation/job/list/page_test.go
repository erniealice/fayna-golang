package list

// Phase 5 (UI permission reflection) — page-controller permission-gate tests.
//
// Job is the only fayna entity with full widget gating per the audit.
// Verifies buildTableRows applies the correct Disabled flag on every
// row action across the {viewer, editor, admin} matrix.

import (
	"testing"

	job "github.com/erniealice/fayna-golang/domain/operation/job"
	"github.com/erniealice/pyeza-golang/types"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

func jobTestLabels() job.Labels {
	l := job.DefaultLabels()
	if l.Errors.PermissionDenied == "" {
		l.Errors.PermissionDenied = "Missing permission"
	}
	if l.Errors.InUse == "" {
		l.Errors.InUse = "Cannot delete: in use"
	}
	return l
}

func jobTestRoutes() job.Routes {
	return job.DefaultRoutes()
}

func findJobAction(actions []types.TableAction, typ string) *types.TableAction {
	for i := range actions {
		if actions[i].Type == typ {
			return &actions[i]
		}
	}
	return nil
}

// TestBuildTableRows_JobPermissionMatrix exercises the
// {viewer, editor, admin} matrix against the job row actions.
func TestBuildTableRows_JobPermissionMatrix(t *testing.T) {
	t.Parallel()

	jobs := []*jobpb.Job{
		{Id: "job-1", Name: "Acme Service", Status: enums.JobStatus_JOB_STATUS_ACTIVE},
	}

	cases := []struct {
		name             string
		perms            []string
		wantEditDisabled bool
		wantDelDisabled  bool
	}{
		{
			name:             "viewer — edit and delete disabled",
			perms:            []string{"job:list", "job:read"},
			wantEditDisabled: true,
			wantDelDisabled:  true,
		},
		{
			name:             "editor (no delete)",
			perms:            []string{"job:list", "job:read", "job:create", "job:update"},
			wantEditDisabled: false,
			wantDelDisabled:  true,
		},
		{
			name:             "admin",
			perms:            []string{"job:list", "job:read", "job:create", "job:update", "job:delete"},
			wantEditDisabled: false,
			wantDelDisabled:  false,
		},
	}

	l := jobTestLabels()
	routes := jobTestRoutes()

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			perms := types.NewUserPermissions(tc.perms)
			rows := buildTableRows(jobs, l, routes, map[string]bool{}, perms)
			if len(rows) != 1 {
				t.Fatalf("rows = %d, want 1", len(rows))
			}
			actions := rows[0].Actions

			if edit := findJobAction(actions, "edit"); edit == nil {
				t.Fatalf("edit action not found")
			} else if edit.Disabled != tc.wantEditDisabled {
				t.Errorf("edit.Disabled = %v, want %v", edit.Disabled, tc.wantEditDisabled)
			}
			if del := findJobAction(actions, "delete"); del == nil {
				t.Fatalf("delete action not found")
			} else if del.Disabled != tc.wantDelDisabled {
				t.Errorf("delete.Disabled = %v, want %v", del.Disabled, tc.wantDelDisabled)
			}
		})
	}
}

// TestBuildTableRows_Job_InUseUsesInUseTooltip verifies the in-use tooltip
// takes priority over the permission tooltip on the delete action.
func TestBuildTableRows_Job_InUseUsesInUseTooltip(t *testing.T) {
	t.Parallel()

	jobs := []*jobpb.Job{
		{Id: "job-2", Name: "Linked Job", Status: enums.JobStatus_JOB_STATUS_ACTIVE},
	}
	l := jobTestLabels()
	routes := jobTestRoutes()

	// Admin perms — but in-use blocks delete.
	perms := types.NewUserPermissions([]string{"job:list", "job:read", "job:update", "job:delete"})
	rows := buildTableRows(jobs, l, routes, map[string]bool{"job-2": true}, perms)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	del := findJobAction(rows[0].Actions, "delete")
	if del == nil {
		t.Fatalf("delete action not found")
	}
	if !del.Disabled {
		t.Error("delete should be disabled when job is in use")
	}
	if del.DisabledTooltip != l.Errors.InUse {
		t.Errorf("delete.DisabledTooltip = %q, want %q (InUse)", del.DisabledTooltip, l.Errors.InUse)
	}
}

// TestBuildTableRows_Job_RendersEveryGivenJob verifies buildTableRows no
// longer re-filters by status client-side (2026-07-11, 20260710
// staff-class-list plan build spec §3 — the truncation-bug fix moved status
// filtering server-side into fetchScopedJobs' TypedFilter; the caller is now
// trusted to pass an already-scoped slice, and every row it's given renders).
func TestBuildTableRows_Job_RendersEveryGivenJob(t *testing.T) {
	t.Parallel()

	jobs := []*jobpb.Job{
		{Id: "job-draft", Name: "Draft Job", Status: enums.JobStatus_JOB_STATUS_DRAFT},
		{Id: "job-active", Name: "Active Job", Status: enums.JobStatus_JOB_STATUS_ACTIVE},
	}
	l := jobTestLabels()
	routes := jobTestRoutes()
	perms := types.NewUserPermissions([]string{"job:list", "job:read"})

	rows := buildTableRows(jobs, l, routes, nil, perms)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2 (buildTableRows renders every job it's given)", len(rows))
	}
}

// TestJobStatusFilterValue verifies the URL status segment -> JobStatus enum
// name mapping fetchScopedJobs uses for the server-side TypedFilter.
func TestJobStatusFilterValue(t *testing.T) {
	t.Parallel()

	cases := []struct{ in, want string }{
		{"active", "JOB_STATUS_ACTIVE"},
		{"draft", "JOB_STATUS_DRAFT"},
		{"paused", "JOB_STATUS_PAUSED"},
		{"completed", "JOB_STATUS_COMPLETED"},
		{"closed", "JOB_STATUS_CLOSED"},
		{"planned", "JOB_STATUS_PLANNED"},
		{"released", "JOB_STATUS_RELEASED"},
		{"bogus-status", "JOB_STATUS_ACTIVE"},
		{"", "JOB_STATUS_ACTIVE"},
	}
	for _, tc := range cases {
		if got := jobStatusFilterValue(tc.in); got != tc.want {
			t.Errorf("jobStatusFilterValue(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
