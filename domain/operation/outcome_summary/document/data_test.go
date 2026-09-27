package document

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
	userpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/user"
	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	productplanpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product_plan"
	productplanstaffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product_plan_staff"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
	sgppspb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_product_plan_staff"
)

func TestPlanLevelAndGroupLabel_MatchesLegacyGradeParse(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ group, plan, level, label string }{
		{"Grade 9 Gold", "Grade 9", "Grade 9", "Gold"},
		{"Grade 10 Palladium", "Grade 10", "Grade 10", "Palladium"},
		{"Grade 9 Gold", "", "", "Grade 9 Gold"},
		{"Palladium", "", "", "Palladium"},
	} {
		level, label := planLevelAndGroupLabel(tc.group, tc.plan)
		if level != tc.level || label != tc.label {
			t.Errorf("planLevelAndGroupLabel(%q, %q) = (%q, %q), want (%q, %q)", tc.group, tc.plan, level, label, tc.level, tc.label)
		}
	}
}

// Generic plan names have no equivalent in the old Grade-only parser.
func TestPlanLevelAndGroupLabel_GenericPlanName(t *testing.T) {
	level, label := planLevelAndGroupLabel("Level 2 Blue", "Level 2")
	if level != "Level 2" || label != "Blue" {
		t.Fatalf("got (%q, %q)", level, label)
	}
}

func TestStripScheduleSuffix_OnlyKnownSchedules(t *testing.T) {
	t.Parallel()
	known := []string{"AY 2025-2026", "AY 2026-2027"}
	for _, tc := range []struct{ input, want string }{
		{"Mathematics — AY 2025-2026", "Mathematics"},
		{"Science – AY 2026-2027", "Science"},
		{"Arts - AY 2025-2026", "Arts"},
		{"Mathematics — AY 2024-2025", "Mathematics — AY 2024-2025"},
		{"Mathematics — Unknown", "Mathematics — Unknown"},
		{"Mathematics", "Mathematics"},
	} {
		if got := stripScheduleSuffix(tc.input, known); got != tc.want {
			t.Errorf("stripScheduleSuffix(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSplitGroupQualifier_PrefixFromOptions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, prefix, wantName, wantPeriod string }{
		{"Grade 9 Gold (AY 2025-26)", "AY", "Grade 9 Gold", "2025-26"},
		{"Grade 9 Gold (AY 2025-26)", "", "Grade 9 Gold (AY 2025-26)", ""},
		{"Grade 9 Gold (Q1)", "AY", "Grade 9 Gold (Q1)", ""},
		{"Palladium", "AY", "Palladium", ""},
	} {
		options := outcome_summary.Options{GroupPeriodQualifierPrefix: tc.prefix}
		name, period := splitGroupQualifier(tc.name, options.GroupPeriodQualifierPrefix)
		if name != tc.wantName || period != tc.wantPeriod {
			t.Errorf("splitGroupQualifier(%q, %q) = (%q, %q), want (%q, %q)", tc.name, tc.prefix, name, period, tc.wantName, tc.wantPeriod)
		}
	}
}

// A non-education qualifier is generic-only; the legacy AY parser kept it.
func TestSplitGroupQualifier_GenericConfiguredPrefix(t *testing.T) {
	name, period := splitGroupQualifier("Group One (Q1)", "Q")
	if name != "Group One" || period != "1" {
		t.Fatalf("got (%q, %q)", name, period)
	}
}

func TestFetchScheduleNames_PagesActiveAndInactive(t *testing.T) {
	pages := map[bool][]int32{true: {}, false: {}}
	d := &Deps{ListPriceSchedules: func(_ context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error) {
		if req.GetSort() == nil {
			t.Fatal("schedule pagination requires a stable sort")
		}
		active := true
		for _, filter := range req.GetFilters().GetFilters() {
			if filter.GetField() == "active" {
				active = filter.GetBooleanFilter().GetValue()
			}
		}
		page := req.GetPagination().GetOffset().GetPage()
		pages[active] = append(pages[active], page)
		if page == 1 {
			data := make([]*priceschedulepb.PriceSchedule, pageLimit)
			for i := range data {
				data[i] = &priceschedulepb.PriceSchedule{Name: fmt.Sprintf("schedule-%t-%d", active, i)}
			}
			return &priceschedulepb.ListPriceSchedulesResponse{Data: data}, nil
		}
		return &priceschedulepb.ListPriceSchedulesResponse{Data: []*priceschedulepb.PriceSchedule{{Name: fmt.Sprintf("last-%t", active)}}}, nil
	}}
	names := fetchScheduleNames(context.Background(), d)
	if len(names) != 2*(pageLimit+1) || !reflect.DeepEqual(pages[true], []int32{1, 2}) || !reflect.DeepEqual(pages[false], []int32{1, 2}) {
		t.Fatalf("names=%d pages=%v", len(names), pages)
	}
}

// M4 (audit T5): the report-card .docx builder duplicates the client_card IDOR
// gates verbatim (fetchGroup + memberSubscription, "mirror student_card").
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

func TestFetchGroup_ForeignGroup_Nil(t *testing.T) {
	d := &Deps{ListSubscriptionGroups: groupsFn()}
	if g := fetchGroup(context.Background(), d, "sec-1"); g != nil {
		t.Fatalf("foreign group must resolve nil, got %v", g)
	}
}

func TestFetchGroup_Present(t *testing.T) {
	d := &Deps{ListSubscriptionGroups: groupsFn(&subscriptiongrouppb.SubscriptionGroup{Id: "sec-1", Active: true})}
	g := fetchGroup(context.Background(), d, "sec-1")
	if g == nil || g.GetId() != "sec-1" {
		t.Fatalf("present group must resolve, got %v", g)
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
		t.Fatalf("inactive member in a live group must resolve empty, got %q", sub)
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
func depsForGroupJobs(n int) (d *Deps, group, client string) {
	const sub, catAcad, catHome, staffID = "sub-1", "cat-acad", "cat-home", "staff-1"
	group, client = "sec-1", "cli-1"

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
			Id: group, Active: true, Name: "Grade 7 Nickel (AY 2025-2026)",
		}),
		ListSubscriptionGroupMembers: membersFn(&subscriptiongroupmemberpb.SubscriptionGroupMember{
			SubscriptionGroupId: group, ClientId: client, SubscriptionId: sub, Active: true,
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
	return d, group, client
}

// The bulk closures must feed every consumer of the old per-job reads from one
// card-local result, including the repeated strict/semester/item-rating walks.
func TestCollectCard_BulkPhaseParityAndCount(t *testing.T) {
	for _, n := range []int{1, 12, 24} {
		t.Run(fmt.Sprintf("%d-group-jobs", n), func(t *testing.T) {
			legacy, group, client := depsForGroupJobs(n)
			listPhases := legacy.ListJobPhases
			legacy.ListJobPhases = func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error) {
				resp, err := listPhases(ctx, req)
				for _, phase := range resp.GetData() {
					phase.TemplatePhaseId = strp("tp-1")
				}
				return resp, err
			}
			var rows []*phasesumpb.PhaseOutcomeSummary
			for i := 1; i <= n; i++ {
				label := fmt.Sprintf("label-%d", i)
				rows = append(rows, &phasesumpb.PhaseOutcomeSummary{
					JobId: fmt.Sprintf("jh-%d", i), JobPhaseId: fmt.Sprintf("ph-%d", i),
					Active: true, ScaledLabel: &label,
				})
			}
			legacyPhaseCalls, legacyTemplateCalls := 0, 0
			legacy.ListPhaseOutcomeSummarysByJob = func(_ context.Context, req *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error) {
				legacyPhaseCalls++
				var out []*phasesumpb.PhaseOutcomeSummary
				for _, row := range rows {
					if row.GetJobId() == req.GetJobId() {
						out = append(out, row)
					}
				}
				return &phasesumpb.ListPhaseOutcomeSummarysByJobResponse{PhaseOutcomeSummarys: out}, nil
			}
			legacy.ListJobTemplatePhasesByTemplate = func(context.Context, *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error) {
				legacyTemplateCalls++
				code := "phase-one"
				return &jobtemplatephasepb.ListByJobTemplateResponse{JobTemplatePhases: []*jobtemplatephasepb.JobTemplatePhase{{Id: "tp-1", Code: &code}}}, nil
			}
			before, ok := collectCard(context.Background(), legacy, group, client)
			if !ok {
				t.Fatal("legacy card gate failed")
			}
			beforePhaseCalls, beforeTemplateCalls := legacyPhaseCalls, legacyTemplateCalls
			bulk := *legacy
			phaseCalls, templateCalls := 0, 0
			bulk.ListPhaseOutcomeSummariesByJobs = func(_ context.Context, ids []string) ([]*phasesumpb.PhaseOutcomeSummary, error) {
				phaseCalls++
				if len(ids) != n {
					t.Fatalf("bulk phase ids = %d, want %d", len(ids), n)
				}
				return rows, nil
			}
			bulk.ListJobTemplatePhasesByTemplates = func(_ context.Context, ids []string) ([]*jobtemplatephasepb.JobTemplatePhase, error) {
				templateCalls++
				if len(ids) != 1 || ids[0] != "tmpl-h" {
					t.Fatalf("bulk template ids = %v, want [tmpl-h]", ids)
				}
				code := "phase-one"
				return []*jobtemplatephasepb.JobTemplatePhase{{Id: "tp-1", Code: &code}}, nil
			}
			after, ok := collectCard(context.Background(), &bulk, group, client)
			if !ok {
				t.Fatal("bulk card gate failed")
			}
			if phaseCalls != 1 || templateCalls != 1 {
				t.Fatalf("bulk calls phase=%d template=%d, want 1 each", phaseCalls, templateCalls)
			}
			if legacyPhaseCalls != beforePhaseCalls || legacyTemplateCalls != beforeTemplateCalls {
				t.Fatalf("bulk card fell back to per-item reads: phase=%d template=%d", legacyPhaseCalls-beforePhaseCalls, legacyTemplateCalls-beforeTemplateCalls)
			}
			t.Logf("dependency calls for %d group jobs: phase %d→1, template %d→1", n, beforePhaseCalls, beforeTemplateCalls)
			if a, b := buildClientOutcomeSummaryData(*before), buildClientOutcomeSummaryData(*after); !reflect.DeepEqual(a, b) {
				t.Fatalf("card data changed for %d group jobs", n)
			}
		})
	}
}

// The singleton-cardinality gate for the block-layout root alias
// (lead_staff_name_display): it is populated ONLY when the group category has
// EXACTLY one job. With 2+ homeroom jobs the first-picked adviser is arbitrary and
// must NOT leak onto the cover/headers (the nested singleton already blanks), yet
// the FROZEN v1/v2 "adviser" key keeps its first-job behavior. 0/1/2-job cases
// through collectCard + buildClientOutcomeSummaryData, asserting root alias + nested
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
			d, group, client := depsForGroupJobs(tc.n)
			rc, ok := collectCard(context.Background(), d, group, client)
			if !ok {
				t.Fatalf("collectCard returned !ok")
			}
			data := buildClientOutcomeSummaryData(*rc)

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
// wraps the shared outcome_summary.IsPlaceholderOutcomeCell predicate. It pins the
// row-level contract collectCard relies on at data.go:199 — a subject the
// client never took (an all-zero active scaffold, e.g. the untaken half of
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

// DP-10/11: every active primary teacher appears in the fallback; secondary,
// inactive and foreign-group edges cannot leak into the document.
func TestFetchProductPlanEdgeStaff_AllPrimariesOnly(t *testing.T) {
	edge := func(id, staffID, role, group string, active bool) *sgppspb.SubscriptionGroupProductPlanStaff {
		return &sgppspb.SubscriptionGroupProductPlanStaff{
			Id: id, StaffId: staffID, Role: role, ProductPlanId: "pp-1", SubscriptionGroupId: group, Active: active,
		}
	}
	depsFor := func(edges []*sgppspb.SubscriptionGroupProductPlanStaff) *Deps {
		return &Deps{
			ListSubscriptionGroupProductPlanStaffs: func(context.Context, *sgppspb.ListSubscriptionGroupProductPlanStaffsRequest) (*sgppspb.ListSubscriptionGroupProductPlanStaffsResponse, error) {
				return &sgppspb.ListSubscriptionGroupProductPlanStaffsResponse{Data: edges}, nil
			},
			ListProductPlans: func(context.Context, *productplanpb.ListProductPlansRequest) (*productplanpb.ListProductPlansResponse, error) {
				return &productplanpb.ListProductPlansResponse{Data: []*productplanpb.ProductPlan{{Id: "pp-1", ProductId: "prod-1"}}}, nil
			},
		}
	}
	prod := "prod-1"
	job := &jobpb.Job{Id: "job-1", OutputProductId: &prod}
	edges := []*sgppspb.SubscriptionGroupProductPlanStaff{
		edge("e-b", "staff-B", "primary", "sec-1", true),
		edge("e-a", "staff-A", "primary", "sec-1", true),
		edge("e-a-duplicate", "staff-A", "primary", "sec-1", true),
		edge("e-secondary", "staff-Tejas", "secondary", "sec-1", true),
		edge("e-inactive", "staff-X", "primary", "sec-1", false),
		edge("e-foreign", "staff-Y", "primary", "sec-2", true),
	}
	for i, ordered := range [][]*sgppspb.SubscriptionGroupProductPlanStaff{edges, {edges[5], edges[4], edges[3], edges[2], edges[1], edges[0]}} {
		got := fetchProductPlanEdgeStaff(context.Background(), depsFor(ordered), "sec-1", []*jobpb.Job{job})
		if !reflect.DeepEqual([]string{"staff-A", "staff-B"}, got["job-1"]) {
			t.Fatalf("order %d: primary teachers = %v", i, got["job-1"])
		}
	}
}

// T-A6: the class-edge eligibility gate mirrors espyna's
// classEdgeEligibilityLivePredicate (job_template_summary_query.go:
// "AND (e.product_plan_staff_id IS NULL OR pps.active)") — a linked edge
// (product_plan_staff_id set) is honored only while that row is active; an
// unlinked edge is unaffected; a missing row or a list error drops linked
// edges (fail closed); a nil ListProductPlanStaffs dep skips the gate
// entirely (today's un-gated behavior, preserved for callers that haven't
// wired the dep yet).
func TestFetchProductPlanEdgeStaff_EligibilityGate(t *testing.T) {
	prod := "prod-1"
	job := &jobpb.Job{Id: "job-1", OutputProductId: &prod}

	plans := func(context.Context, *productplanpb.ListProductPlansRequest) (*productplanpb.ListProductPlansResponse, error) {
		return &productplanpb.ListProductPlansResponse{Data: []*productplanpb.ProductPlan{{Id: "pp-1", ProductId: "prod-1"}}}, nil
	}
	// staff-A's edge carries no product_plan_staff_id (unlinked — never gated).
	// staff-B's edge links to "pps-B" (gated on that row's active flag).
	edgesFn := func(context.Context, *sgppspb.ListSubscriptionGroupProductPlanStaffsRequest) (*sgppspb.ListSubscriptionGroupProductPlanStaffsResponse, error) {
		return &sgppspb.ListSubscriptionGroupProductPlanStaffsResponse{Data: []*sgppspb.SubscriptionGroupProductPlanStaff{
			{Id: "e-unlinked", StaffId: "staff-A", Role: "primary", ProductPlanId: "pp-1", SubscriptionGroupId: "sec-1", Active: true},
			{Id: "e-linked", StaffId: "staff-B", Role: "primary", ProductPlanId: "pp-1", SubscriptionGroupId: "sec-1", Active: true, ProductPlanStaffId: strp("pps-B")},
		}}, nil
	}

	t.Run("revoked eligibility drops only the linked teacher; the unlinked primary stays", func(t *testing.T) {
		d := &Deps{
			ListSubscriptionGroupProductPlanStaffs: edgesFn,
			ListProductPlans:                       plans,
			ListProductPlanStaffs: func(context.Context, *productplanstaffpb.ListProductPlanStaffsRequest) (*productplanstaffpb.ListProductPlanStaffsResponse, error) {
				return &productplanstaffpb.ListProductPlanStaffsResponse{Data: []*productplanstaffpb.ProductPlanStaff{
					{Id: "pps-B", Active: false},
				}}, nil
			},
		}
		got := fetchProductPlanEdgeStaff(context.Background(), d, "sec-1", []*jobpb.Job{job})
		if !reflect.DeepEqual([]string{"staff-A"}, got["job-1"]) {
			t.Fatalf("teachers = %v, want [staff-A] (staff-B's revoked eligibility must drop only staff-B)", got["job-1"])
		}
	})

	t.Run("active eligibility keeps the linked teacher alongside the unlinked one", func(t *testing.T) {
		d := &Deps{
			ListSubscriptionGroupProductPlanStaffs: edgesFn,
			ListProductPlans:                       plans,
			ListProductPlanStaffs: func(context.Context, *productplanstaffpb.ListProductPlanStaffsRequest) (*productplanstaffpb.ListProductPlanStaffsResponse, error) {
				return &productplanstaffpb.ListProductPlanStaffsResponse{Data: []*productplanstaffpb.ProductPlanStaff{
					{Id: "pps-B", Active: true},
				}}, nil
			},
		}
		got := fetchProductPlanEdgeStaff(context.Background(), d, "sec-1", []*jobpb.Job{job})
		if !reflect.DeepEqual([]string{"staff-A", "staff-B"}, got["job-1"]) {
			t.Fatalf("teachers = %v, want [staff-A staff-B]", got["job-1"])
		}
	})

	t.Run("missing eligibility row drops the linked teacher", func(t *testing.T) {
		d := &Deps{
			ListSubscriptionGroupProductPlanStaffs: edgesFn,
			ListProductPlans:                       plans,
			ListProductPlanStaffs: func(context.Context, *productplanstaffpb.ListProductPlanStaffsRequest) (*productplanstaffpb.ListProductPlanStaffsResponse, error) {
				return &productplanstaffpb.ListProductPlanStaffsResponse{}, nil // "pps-B" never comes back
			},
		}
		got := fetchProductPlanEdgeStaff(context.Background(), d, "sec-1", []*jobpb.Job{job})
		if !reflect.DeepEqual([]string{"staff-A"}, got["job-1"]) {
			t.Fatalf("teachers = %v, want [staff-A] (a missing eligibility row must drop the linked teacher)", got["job-1"])
		}
	})

	t.Run("list error fails closed and drops every linked teacher", func(t *testing.T) {
		d := &Deps{
			ListSubscriptionGroupProductPlanStaffs: edgesFn,
			ListProductPlans:                       plans,
			ListProductPlanStaffs: func(context.Context, *productplanstaffpb.ListProductPlanStaffsRequest) (*productplanstaffpb.ListProductPlanStaffsResponse, error) {
				return nil, errors.New("boom")
			},
		}
		got := fetchProductPlanEdgeStaff(context.Background(), d, "sec-1", []*jobpb.Job{job})
		if !reflect.DeepEqual([]string{"staff-A"}, got["job-1"]) {
			t.Fatalf("teachers = %v, want [staff-A] (a list error must fail closed and drop staff-B)", got["job-1"])
		}
	})

	t.Run("nil ListProductPlanStaffs dep skips the gate (today's un-gated behavior)", func(t *testing.T) {
		d := &Deps{
			ListSubscriptionGroupProductPlanStaffs: edgesFn,
			ListProductPlans:                       plans,
			// ListProductPlanStaffs intentionally left nil.
		}
		got := fetchProductPlanEdgeStaff(context.Background(), d, "sec-1", []*jobpb.Job{job})
		if !reflect.DeepEqual([]string{"staff-A", "staff-B"}, got["job-1"]) {
			t.Fatalf("teachers = %v, want [staff-A staff-B] (nil dep must not gate)", got["job-1"])
		}
	})
}
