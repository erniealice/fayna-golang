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
