package client_card

import (
	"context"
	"testing"

	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

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
