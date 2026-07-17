package document

import (
	"context"
	"fmt"
	"testing"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	userpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/user"
	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

// M4 (audit T5): the report-card .docx builder duplicates the client_card IDOR
// gates verbatim (fetchSection + memberSubscription, "mirror student_card").
// These pin the SAME fail-closed contract on this second copy so a drift in
// either package is caught.

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

func TestFetchSection_ForeignSection_Nil(t *testing.T) {
	d := &Deps{ListSubscriptionGroups: groupsFn()}
	if g := fetchSection(context.Background(), d, "sec-1"); g != nil {
		t.Fatalf("foreign section must resolve nil, got %v", g)
	}
}

func TestFetchSection_Present(t *testing.T) {
	d := &Deps{ListSubscriptionGroups: groupsFn(&subscriptiongrouppb.SubscriptionGroup{Id: "sec-1", Active: true})}
	g := fetchSection(context.Background(), d, "sec-1")
	if g == nil || g.GetId() != "sec-1" {
		t.Fatalf("present section must resolve, got %v", g)
	}
}

func TestMemberSubscription_NonMember_Empty(t *testing.T) {
	d := &Deps{ListSubscriptionGroupMembers: membersFn(
		&subscriptiongroupmemberpb.SubscriptionGroupMember{ClientId: "other", SubscriptionId: "sub-x", Active: true},
	)}
	if sub := memberSubscription(context.Background(), d, "sec-1", "target", false); sub != "" {
		t.Fatalf("non-member must resolve empty, got %q", sub)
	}
}

func TestMemberSubscription_Member_Resolves(t *testing.T) {
	d := &Deps{ListSubscriptionGroupMembers: membersFn(
		&subscriptiongroupmemberpb.SubscriptionGroupMember{ClientId: "target", SubscriptionId: "sub-1", Active: true},
	)}
	if sub := memberSubscription(context.Background(), d, "sec-1", "target", false); sub != "sub-1" {
		t.Fatalf("active member must resolve sub-1, got %q", sub)
	}
}

func TestMemberSubscription_InactiveInActiveGroup_Empty(t *testing.T) {
	d := &Deps{ListSubscriptionGroupMembers: membersFn(
		&subscriptiongroupmemberpb.SubscriptionGroupMember{ClientId: "target", SubscriptionId: "sub-1", Active: false},
	)}
	if sub := memberSubscription(context.Background(), d, "sec-1", "target", false); sub != "" {
		t.Fatalf("inactive member in a live section must resolve empty, got %q", sub)
	}
}

func TestMemberSubscription_HistoricalAccepted(t *testing.T) {
	d := &Deps{ListSubscriptionGroupMembers: membersFn(
		&subscriptiongroupmemberpb.SubscriptionGroupMember{ClientId: "target", SubscriptionId: "sub-1", Active: false},
	)}
	if sub := memberSubscription(context.Background(), d, "sec-1", "target", true); sub != "sub-1" {
		t.Fatalf("historical mode must accept the frozen inactive member, got %q", sub)
	}
}

// depsForGroupJobs builds a collectCard Deps whose enrollment carries exactly n
// jobs in the configured GROUP (homeroom) category, each advised by the SAME
// staff. Used to exercise the singleton-cardinality gate for the block-layout
// root alias. No academic jobs (the transcript is irrelevant to the lead gate).
func depsForGroupJobs(n int) (d *Deps, section, client string) {
	const sub, catAcad, catHome, staffID = "sub-1", "cat-acad", "cat-home", "staff-1"
	section, client = "sec-1", "cli-1"

	var jobs []*jobpb.Job
	var phases []*jobphasepb.JobPhase
	var tasks []*jobtaskpb.JobTask
	for i := 1; i <= n; i++ {
		jid, pid, tid := fmt.Sprintf("jh-%d", i), fmt.Sprintf("ph-%d", i), fmt.Sprintf("jt-%d", i)
		jobs = append(jobs, &jobpb.Job{
			Id: jid, Active: true,
			OriginType:    enums.OriginType_ORIGIN_TYPE_SUBSCRIPTION,
			OriginId:      strp(sub),
			JobCategoryId: strp(catHome),
			JobTemplateId: strp("tmpl-h"),
		})
		phases = append(phases, &jobphasepb.JobPhase{Id: pid, JobId: jid, PhaseOrder: 1})
		tasks = append(tasks, &jobtaskpb.JobTask{Id: tid, JobPhaseId: pid, Active: true, AssignedTo: strp(staffID)})
	}

	d = &Deps{
		CategoryFilter: "academic",
		DocOptions:     outcome_summary.DocumentOptions{GroupCategoryFilter: "homeroom_deportment"},
		ListSubscriptionGroups: groupsFn(&subscriptiongrouppb.SubscriptionGroup{
			Id: section, Active: true, Name: "Grade 7 Nickel (AY 2025-2026)",
		}),
		ListSubscriptionGroupMembers: membersFn(&subscriptiongroupmemberpb.SubscriptionGroupMember{
			SubscriptionGroupId: section, ClientId: client, SubscriptionId: sub, Active: true,
		}),
		ListJobs: func(_ context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
			// The inactive-subject probe filters active=false — return nothing so
			// those names never pollute the rotation merge.
			for _, f := range req.GetFilters().GetFilters() {
				if f.GetBooleanFilter() != nil {
					return &jobpb.ListJobsResponse{}, nil
				}
			}
			return &jobpb.ListJobsResponse{Data: jobs}, nil
		},
		ListJobCategories: func(context.Context, *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error) {
			return &jobcategorypb.ListJobCategoriesResponse{Data: []*jobcategorypb.JobCategory{
				{Id: catAcad, Name: "Academic", Code: strp("academic")},
				{Id: catHome, Name: "Homeroom Deportment", Code: strp("homeroom_deportment")},
			}}, nil
		},
		ListJobPhases: func(context.Context, *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
			return &jobphasepb.ListJobPhasesResponse{Data: phases}, nil
		},
		ListJobTasks: func(context.Context, *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error) {
			return &jobtaskpb.ListJobTasksResponse{Data: tasks}, nil
		},
		ListTaskOutcomes: func(context.Context, *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error) {
			return &taskoutcomepb.ListTaskOutcomesResponse{}, nil
		},
		GetStaffListPageData: func(context.Context, *staffpb.GetStaffListPageDataRequest) (*staffpb.GetStaffListPageDataResponse, error) {
			return &staffpb.GetStaffListPageDataResponse{StaffList: []*staffpb.Staff{
				{Id: staffID, User: &userpb.User{FirstName: "Adviser", LastName: "One"}},
			}}, nil
		},
	}
	return d, section, client
}

// The singleton-cardinality gate for the block-layout root alias
// (lead_staff_name_display): it is populated ONLY when the group category has
// EXACTLY one job. With 2+ homeroom jobs the first-picked adviser is arbitrary and
// must NOT leak onto the cover/headers (the nested singleton already blanks), yet
// the FROZEN v1/v2 "adviser" key keeps its first-job behavior. 0/1/2-job cases
// through collectCard + buildReportCardData, asserting root alias + nested
// projection consistency.
func TestCollectCard_GroupLeadSingletonGate(t *testing.T) {
	cases := []struct {
		n           int
		wantLead    string // root block alias + nested singleton projection
		wantAdviser string // FROZEN v1/v2 key (first-job behavior)
	}{
		{n: 0, wantLead: "", wantAdviser: ""},
		{n: 1, wantLead: "Adviser One", wantAdviser: "Adviser One"},
		{n: 2, wantLead: "", wantAdviser: "Adviser One"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%d-jobs", tc.n), func(t *testing.T) {
			d, section, client := depsForGroupJobs(tc.n)
			rc, ok := collectCard(context.Background(), d, section, client)
			if !ok {
				t.Fatalf("collectCard returned !ok")
			}
			data := buildReportCardData(*rc)

			// Root block alias — gated on exactly-one group job.
			assertLeaf(t, data, "lead_staff_name_display", tc.wantLead)
			// Frozen v1/v2 adviser — first-job behavior, untouched by the gate.
			assertLeaf(t, data, "adviser", tc.wantAdviser)

			// Nested singleton projection: emitted only for exactly one job, and
			// then equal to the root alias (consistency); blank/absent otherwise.
			nested, nestedOK := resolvePath(data, "job_categories.homeroom_deportment.lead_staff_name_display")
			if tc.n == 1 {
				if !nestedOK || nested != tc.wantLead {
					t.Fatalf("nested singleton lead = %#v ok=%v, want %q", nested, nestedOK, tc.wantLead)
				}
			} else if nestedOK && nested != "" {
				t.Fatalf("nested singleton must be blank/absent for %d jobs, got %#v", tc.n, nested)
			}
		})
	}
}

// TestIsNonEnrolledPlaceholder is the backlogged B1 unit test (GOAL.md B1 row /
// progress.md "B1 unit test"): the DOCX-layer row→evidence adaptation that
// wraps the shared outcome_summary.IsNonEnrolledCell predicate. It pins the
// row-level contract collectCard relies on at data.go:199 — a subject the
// student never took (an all-zero active scaffold, e.g. the untaken half of
// an English/Filipino-style language pair) is suppressed, while a REAL zero
// for an enrolled subject (a positive per-criterion mark somewhere, or a real
// >1 stored band) is protected and still renders. NEVER blank a real grade.
func TestIsNonEnrolledPlaceholder(t *testing.T) {
	cases := []struct {
		name     string
		row      itemRow
		hasMarks bool
		want     bool // true = placeholder (suppressed from the DOCX)
	}{
		{
			name: "non-enrolled untaken elective all-zero scaffold suppressed",
			row: itemRow{
				Name: "Korean", CritA: "0", CritB: "0", CritC: "0", CritD: "0",
				Total: "0", YearFinal: "1", // transmute-of-zero floor, not evidence
			},
			hasMarks: true,
			want:     true,
		},
		{
			name: "enrolled subject real zero protected by a positive criterion mark",
			row: itemRow{
				Name: "Mathematics", CritA: "0", CritB: "0", CritC: "0", CritD: "5",
				Total: "5", YearFinal: "0",
			},
			hasMarks: true,
			want:     false,
		},
		{
			name: "enrolled subject all-zero criteria kept by a real non-floor semester band",
			row: itemRow{
				Name: "Partial", CritA: "0", CritB: "0", CritC: "0", CritD: "0",
				Total: "0", Sem1Band: "6",
			},
			hasMarks: true,
			want:     false,
		},
		{
			name: "normal graded row rendered",
			row: itemRow{
				Name: "Science", CritA: "6", CritB: "7", CritC: "5", CritD: "6",
				Total: "24", Sem1Band: "6", Sem2Band: "7", YearFinal: "7",
			},
			hasMarks: true,
			want:     false,
		},
		{
			name: "historical import no task_outcome but a real stored year-final kept",
			row: itemRow{
				Name: "History", YearFinal: "6",
			},
			hasMarks: false,
			want:     false,
		},
		{
			name:     "fully blank row with no summary at all suppressed",
			row:      itemRow{Name: "Blank"},
			hasMarks: false,
			want:     true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isNonEnrolledPlaceholder(c.row, c.hasMarks); got != c.want {
				t.Fatalf("isNonEnrolledPlaceholder(%+v, hasMarks=%v) = %v, want %v", c.row, c.hasMarks, got, c.want)
			}
		})
	}
}
