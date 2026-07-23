package client_card

import (
	"context"
	"strings"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

func sptr(s string) *string   { return &s }
func fptr(v float64) *float64 { return &v }

// M4 (audit T5): the two IDOR gates on the per-student report card —
// fetchSection (workspace EXISTS gate) + memberSubscription (section-membership
// gate) — are security-critical and were untested. These pin the fail-closed
// behavior with fake, injectable closures.

func groupsFn(groups ...*subscriptiongrouppb.SubscriptionGroup) func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
	return func(context.Context, *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error) {
		return &subscriptiongrouppb.ListSubscriptionGroupsResponse{Data: groups}, nil
	}
}

func membersFn(members ...*subscriptiongroupmemberpb.SubscriptionGroupMember) func(context.Context, *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
	return func(context.Context, *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error) {
		return &subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse{Data: members}, nil
	}
}

// TestFetchSection_ForeignSection_Nil: the workspace-scoped adapter returns no
// rows for a foreign/missing id → fail-closed nil.
func TestFetchSection_ForeignSection_Nil(t *testing.T) {
	deps := &Deps{ListSubscriptionGroups: groupsFn()}
	if g := fetchSection(context.Background(), deps, "sec-1"); g != nil {
		t.Fatalf("foreign section must resolve nil, got %v", g)
	}
}

// TestFetchSection_Present: an in-workspace section resolves.
func TestFetchSection_Present(t *testing.T) {
	deps := &Deps{ListSubscriptionGroups: groupsFn(&subscriptiongrouppb.SubscriptionGroup{Id: "sec-1", Active: true})}
	g := fetchSection(context.Background(), deps, "sec-1")
	if g == nil || g.GetId() != "sec-1" {
		t.Fatalf("present section must resolve, got %v", g)
	}
}

// TestMemberSubscription_NonMember_Empty: a client that is not a member of the
// section resolves to "" (the IDOR fail-closed signal).
func TestMemberSubscription_NonMember_Empty(t *testing.T) {
	deps := &Deps{ListSubscriptionGroupMembers: membersFn(
		&subscriptiongroupmemberpb.SubscriptionGroupMember{ClientId: "other", SubscriptionId: "sub-x", Active: true},
	)}
	if sub := memberSubscription(context.Background(), deps, "sec-1", "target", false); sub != "" {
		t.Fatalf("non-member must resolve empty, got %q", sub)
	}
}

// TestMemberSubscription_Member_Resolves: an active member resolves its sub id.
func TestMemberSubscription_Member_Resolves(t *testing.T) {
	deps := &Deps{ListSubscriptionGroupMembers: membersFn(
		&subscriptiongroupmemberpb.SubscriptionGroupMember{ClientId: "target", SubscriptionId: "sub-1", Active: true},
	)}
	if sub := memberSubscription(context.Background(), deps, "sec-1", "target", false); sub != "sub-1" {
		t.Fatalf("active member must resolve sub-1, got %q", sub)
	}
}

// TestMemberSubscription_InactiveInActiveGroup_Empty: an inactive membership row
// in a live (non-historical) section is skipped → "".
func TestMemberSubscription_InactiveInActiveGroup_Empty(t *testing.T) {
	deps := &Deps{ListSubscriptionGroupMembers: membersFn(
		&subscriptiongroupmemberpb.SubscriptionGroupMember{ClientId: "target", SubscriptionId: "sub-1", Active: false},
	)}
	if sub := memberSubscription(context.Background(), deps, "sec-1", "target", false); sub != "" {
		t.Fatalf("inactive member in a live section must resolve empty, got %q", sub)
	}
}

// TestMemberSubscription_HistoricalAccepted: in historical (frozen) mode an
// inactive membership row IS accepted (the frozen roster).
func TestMemberSubscription_HistoricalAccepted(t *testing.T) {
	deps := &Deps{ListSubscriptionGroupMembers: membersFn(
		&subscriptiongroupmemberpb.SubscriptionGroupMember{ClientId: "target", SubscriptionId: "sub-1", Active: false},
	)}
	if sub := memberSubscription(context.Background(), deps, "sec-1", "target", true); sub != "sub-1" {
		t.Fatalf("historical mode must accept the frozen inactive member, got %q", sub)
	}
}

// --- buildTable phantom-blank invariant --------------------------------------

func jobsFn(jobs ...*jobpb.Job) func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
	return func(context.Context, *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
		return &jobpb.ListJobsResponse{Data: jobs}, nil
	}
}

func templatesFn(tmpls ...*jobtemplatepb.JobTemplate) func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error) {
	return func(context.Context, *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error) {
		return &jobtemplatepb.ListJobTemplatesResponse{Data: tmpls}, nil
	}
}

func jobPhasesFn(phases ...*jobphasepb.JobPhase) func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
	return func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
		return &jobphasepb.ListJobPhasesResponse{Data: phases}, nil
	}
}

func phaseSummariesByJobFn(byJob map[string][]*phasesumpb.PhaseOutcomeSummary) func(context.Context, *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error) {
	return func(_ context.Context, req *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error) {
		return &phasesumpb.ListPhaseOutcomeSummarysByJobResponse{PhaseOutcomeSummarys: byJob[req.GetJobId()]}, nil
	}
}

func yearSummariesFn(sums ...*jobsumpb.JobOutcomeSummary) func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
	return func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
		return &jobsumpb.ListJobOutcomeSummarysResponse{Data: sums}, nil
	}
}

func jobTasksFn(tasks ...*jobtaskpb.JobTask) func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
	return func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
		return &jobtaskpb.ListJobTasksResponse{Data: tasks}, nil
	}
}

func taskOutcomesFn(outcomes ...*taskoutcomepb.TaskOutcome) func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
	return func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
		return &taskoutcomepb.ListTaskOutcomesResponse{Data: outcomes}, nil
	}
}

// --- W-A7: Download-PDF wired into the table primary action ------------------

// TestOkPage_DownloadWiredIntoPrimaryAction pins the moved Download-PDF: when a
// document URL exists it lands in TableConfig.PrimaryAction with Download=true,
// the preserved rc-download-pdf test id, and the student.download_action label.
func TestOkPage_DownloadWiredIntoPrimaryAction(t *testing.T) {
	deps := &Deps{
		Routes: outcome_summary.Routes{
			ClientDocumentURL: "/rc/section/{id}/client/{client_id}/doc",
			SectionURL:        "/rc/section/{id}",
			ActiveNav:         "reports",
		},
		Labels: outcome_summary.Labels{
			Student: outcome_summary.PeriodLabels{DownloadAction: "Download report card (PDF)"},
		},
	}
	group := &subscriptiongrouppb.SubscriptionGroup{Id: "sec-1", Name: "Grade 5 Diamond"}
	viewCtx := &view.ViewContext{CacheVersion: "v1", CurrentPath: "/rc/section/sec-1/client/c-1"}
	table := &types.TableConfig{ID: "report-cards-student"}

	res := okPage(viewCtx, deps, group, "c-1", "Ada Lovelace", table)
	pd, ok := res.Data.(*PageData)
	if !ok {
		t.Fatalf("Data is not *PageData: %T", res.Data)
	}
	pa := pd.Table.PrimaryAction
	if pa == nil {
		t.Fatal("PrimaryAction must be wired when a document URL exists")
	}
	if pa.Label != "Download report card (PDF)" {
		t.Errorf("Label = %q, want the student.download_action label", pa.Label)
	}
	if !pa.Download {
		t.Error("Download must be true so the body-boosted app does not intercept the download")
	}
	if pa.TestID != "rc-download-pdf" {
		t.Errorf("TestID = %q, want rc-download-pdf (preserved)", pa.TestID)
	}
	if !strings.HasSuffix(pa.Href, "?format=pdf") {
		t.Errorf("Href = %q, want the resolved document URL + ?format=pdf", pa.Href)
	}
}

// TestOkPage_NoDocumentURL_NoPrimaryAction preserves the former
// {{if .DocumentDownloadURL}} gate: no document URL ⇒ no primary action.
func TestOkPage_NoDocumentURL_NoPrimaryAction(t *testing.T) {
	deps := &Deps{
		Routes: outcome_summary.Routes{SectionURL: "/rc/section/{id}"}, // ClientDocumentURL empty
		Labels: outcome_summary.Labels{Student: outcome_summary.PeriodLabels{DownloadAction: "x"}},
	}
	group := &subscriptiongrouppb.SubscriptionGroup{Id: "sec-1", Name: "S"}
	viewCtx := &view.ViewContext{CacheVersion: "v1"}
	table := &types.TableConfig{ID: "report-cards-student"}

	res := okPage(viewCtx, deps, group, "c-1", "N", table)
	pd := res.Data.(*PageData)
	if pd.Table.PrimaryAction != nil {
		t.Errorf("no document URL → no primary action; got %+v", pd.Table.PrimaryAction)
	}
}

// TestOkPage_NilTable_NoPanic confirms a blank card (table == nil ⇒ NotComputed)
// takes the no-URL path without a nil-pointer deref on table.PrimaryAction.
func TestOkPage_NilTable_NoPanic(t *testing.T) {
	deps := &Deps{
		Routes: outcome_summary.Routes{ClientDocumentURL: "/rc/{id}/{client_id}", SectionURL: "/rc/{id}"},
		Labels: outcome_summary.Labels{Section: outcome_summary.SectionLabels{NotComputedBanner: "not yet"}},
	}
	group := &subscriptiongrouppb.SubscriptionGroup{Id: "sec-1", Name: "S"}
	viewCtx := &view.ViewContext{CacheVersion: "v1"}

	res := okPage(viewCtx, deps, group, "c-1", "N", nil)
	pd := res.Data.(*PageData)
	if pd.Table != nil {
		t.Errorf("nil table must stay nil; got %+v", pd.Table)
	}
	if !pd.NotComputed {
		t.Error("nil table → NotComputed true")
	}
}

// findRow returns the row whose subject-name cell equals name.
func findRow(t *testing.T, table *types.TableConfig, name string) types.TableRow {
	t.Helper()
	for _, r := range table.Rows {
		if len(r.Cells) > 0 && r.Cells[0].Value == name {
			return r
		}
	}
	t.Fatalf("row %q not found", name)
	return types.TableRow{}
}

// TestBuildTable_PhantomBlank_RealShown pins the phantom-blank invariant on the
// per-student client card. Cells per row: [0]=subject, [1]=S1 progress, [2]=S1
// final, [3]=S2 progress, [4]=S2 final, [5]=year final.
//
//   - Korean is an untaken-elective all-zero scaffold (all task marks 0, bands
//     floored to "1") → its S1/S2/Year grade cells render BLANK "".
//   - English is genuinely enrolled with a real year-final of "1" (a positive
//     task mark) → its rating "1" is KEPT (a real grade must never blank).
//   - Math is genuinely enrolled with a real "0" (positive task mark) → KEPT.
func TestBuildTable_PhantomBlank_RealShown(t *testing.T) {
	jobR := &jobpb.Job{Id: "jobR", JobTemplateId: sptr("tEng"), OriginId: sptr("sub1"), OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION, Active: true}
	jobP := &jobpb.Job{Id: "jobP", JobTemplateId: sptr("tKor"), OriginId: sptr("sub1"), OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION, Active: true}
	jobM := &jobpb.Job{Id: "jobM", JobTemplateId: sptr("tMath"), OriginId: sptr("sub1"), OriginType: enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION, Active: true}

	deps := &Deps{
		ListJobs: jobsFn(jobR, jobP, jobM),
		ListJobTemplates: templatesFn(
			&jobtemplatepb.JobTemplate{Id: "tEng", Name: "English"},
			&jobtemplatepb.JobTemplate{Id: "tKor", Name: "Korean"},
			&jobtemplatepb.JobTemplate{Id: "tMath", Name: "Math"},
		),
		ListJobPhases: jobPhasesFn(
			&jobphasepb.JobPhase{Id: "phR1", JobId: "jobR", PhaseOrder: 1, Active: true},
			&jobphasepb.JobPhase{Id: "phP1", JobId: "jobP", PhaseOrder: 1, Active: true},
			&jobphasepb.JobPhase{Id: "phP2", JobId: "jobP", PhaseOrder: 2, Active: true},
			&jobphasepb.JobPhase{Id: "phM1", JobId: "jobM", PhaseOrder: 1, Active: true},
		),
		ListPhaseOutcomeSummarysByJob: phaseSummariesByJobFn(map[string][]*phasesumpb.PhaseOutcomeSummary{
			"jobR": {{JobPhaseId: "phR1", ScaledLabel: sptr("1"), Active: true}},
			"jobP": {{JobPhaseId: "phP1", ScaledLabel: sptr("1"), Active: true}, {JobPhaseId: "phP2", ScaledLabel: sptr("1"), Active: true}},
			"jobM": {{JobPhaseId: "phM1", ScaledLabel: sptr("0"), Active: true}},
		}),
		ListJobOutcomeSummarys: yearSummariesFn(
			&jobsumpb.JobOutcomeSummary{JobId: "jobR", ScaledLabel: sptr("1"), Active: true},
			&jobsumpb.JobOutcomeSummary{JobId: "jobP", ScaledLabel: sptr("1"), Active: true},
			&jobsumpb.JobOutcomeSummary{JobId: "jobM", ScaledLabel: sptr("0"), Active: true},
		),
		ListJobTasks: jobTasksFn(
			&jobtaskpb.JobTask{Id: "tkR", JobPhaseId: "phR1", Active: true},
			&jobtaskpb.JobTask{Id: "tkP1", JobPhaseId: "phP1", Active: true},
			&jobtaskpb.JobTask{Id: "tkP2", JobPhaseId: "phP2", Active: true},
			&jobtaskpb.JobTask{Id: "tkM", JobPhaseId: "phM1", Active: true},
		),
		ListTaskOutcomes: taskOutcomesFn(
			&taskoutcomepb.TaskOutcome{JobTaskId: "tkR", NumericValue: fptr(6), Active: true},  // real
			&taskoutcomepb.TaskOutcome{JobTaskId: "tkP1", NumericValue: fptr(0), Active: true}, // scaffold
			&taskoutcomepb.TaskOutcome{JobTaskId: "tkP2", NumericValue: fptr(0), Active: true}, // scaffold
			&taskoutcomepb.TaskOutcome{JobTaskId: "tkM", NumericValue: fptr(3), Active: true},  // real
		),
	}

	table := buildTable(context.Background(), deps, "sub1", false)
	if table == nil {
		t.Fatal("buildTable returned nil")
	}
	if len(table.Rows) != 3 {
		t.Fatalf("want 3 subject rows, got %d", len(table.Rows))
	}

	// English (real 1): S1 final + Year final KEEP "1".
	eng := findRow(t, table, "English")
	if eng.Cells[2].Value != "1" || eng.Cells[5].Value != "1" {
		t.Errorf("English (real 1) must show its rating: S1final=%q yearfinal=%q, want \"1\"/\"1\"", eng.Cells[2].Value, eng.Cells[5].Value)
	}

	// Math (real 0): S1 final + Year final KEEP "0".
	math := findRow(t, table, "Math")
	if math.Cells[2].Value != "0" || math.Cells[5].Value != "0" {
		t.Errorf("Math (real 0) must show its rating: S1final=%q yearfinal=%q, want \"0\"/\"0\"", math.Cells[2].Value, math.Cells[5].Value)
	}

	// Korean (phantom): all three grade cells BLANK "" (not "1", not "—").
	kor := findRow(t, table, "Korean")
	for _, i := range []int{2, 4, 5} {
		if kor.Cells[i].Value != "" {
			t.Errorf("Korean (phantom) cell[%d] = %q, want \"\" (blank)", i, kor.Cells[i].Value)
		}
	}
}
