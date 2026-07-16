package document

import (
	"context"
	"testing"

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
		row      subjectRow
		hasMarks bool
		want     bool // true = placeholder (suppressed from the DOCX)
	}{
		{
			name: "non-enrolled untaken elective all-zero scaffold suppressed",
			row: subjectRow{
				Name: "Korean", CritA: "0", CritB: "0", CritC: "0", CritD: "0",
				Total: "0", YearFinal: "1", // transmute-of-zero floor, not evidence
			},
			hasMarks: true,
			want:     true,
		},
		{
			name: "enrolled subject real zero protected by a positive criterion mark",
			row: subjectRow{
				Name: "Mathematics", CritA: "0", CritB: "0", CritC: "0", CritD: "5",
				Total: "5", YearFinal: "0",
			},
			hasMarks: true,
			want:     false,
		},
		{
			name: "enrolled subject all-zero criteria kept by a real non-floor semester band",
			row: subjectRow{
				Name: "Partial", CritA: "0", CritB: "0", CritC: "0", CritD: "0",
				Total: "0", Sem1Band: "6",
			},
			hasMarks: true,
			want:     false,
		},
		{
			name: "normal graded row rendered",
			row: subjectRow{
				Name: "Science", CritA: "6", CritB: "7", CritC: "5", CritD: "6",
				Total: "24", Sem1Band: "6", Sem2Band: "7", YearFinal: "7",
			},
			hasMarks: true,
			want:     false,
		},
		{
			name: "historical import no task_outcome but a real stored year-final kept",
			row: subjectRow{
				Name: "History", YearFinal: "6",
			},
			hasMarks: false,
			want:     false,
		},
		{
			name:     "fully blank row with no summary at all suppressed",
			row:      subjectRow{Name: "Blank"},
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
