// Package deliverygroup resolves the generic "delivery group" (a section /
// cohort — subscription_group) name and its schedule (academic-year /
// period — the group's nested price_schedule) name from a subscription id.
//
// Extracted 2026-07-11 (20260710-staff-class-list plan, round 4 item 2) out
// of job/list/template_summary.go — both that file's education-tier
// template-grain delivery summary AND outcome_matrix/list/page.go's page
// header need the SAME subscription_group_member -> subscription_group
// chain ("resolve Group via the job's OWN membership row — never via
// subscription_group_product_plan_staff(plan, staff), which over-counts to
// cohort grain"). This package is that one shared resolution.
//
// Deliberately a thin LEAF package: it takes closures (never touches
// espyna/consumer types directly) and does nothing beyond the
// subscription_id -> group -> schedule lookup — no job/status filtering, no
// aggregation, no pagination-loop orchestration for its CALLERS' data (each
// caller still owns gathering its own subscription id(s)). S6 (a
// server-side aggregate use case) is expected to replace
// template_summary.go's data assembly; this package should need no changes
// when that lands, because it was never tangled into that assembly.
//
// Import-shape note: this is its own package, not domain/operation itself,
// because domain/operation/{job_module,outcome_matrix_module}.go (package
// operation) already import job/list and outcome_matrix/list — putting this
// helper in package operation would create an import cycle (operation ->
// job/list -> operation via this helper). Both list packages import this
// leaf package instead; it imports neither of them.
package deliverygroup

import (
	"context"
	"log"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

// PageLimit bounds every ID-batch fetch this package makes. Several bare-
// List RPCs in this codebase silently ignore PaginationRequest (verified
// against the postgres adapters — see
// docs/plan/20260710-staff-class-list/s3b-view.md), so PageLimit chunks the
// INPUT id set instead of relying on server-side pagination — the only
// reliable way to keep every call's OUTPUT under the adapter's 100-row
// default regardless of which RPC ends up wired.
const PageLimit = 100

// ListSubscriptionGroupMembersFunc is the espyna closure shape both callers
// already have wired (block/wiring.go u.Subscription.SubscriptionGroupMember.
// ListSubscriptionGroupMembers).
type ListSubscriptionGroupMembersFunc func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)

// ListSubscriptionGroupsFunc is the espyna closure shape both callers
// already have wired (block/wiring.go u.Subscription.SubscriptionGroup.
// ListSubscriptionGroups).
type ListSubscriptionGroupsFunc func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)

// ResolveGroupIDs resolves subscription_id -> subscription_group_id for the
// given (deduplicated) subscription ids, chunked into ListFilter(IN)
// batches of PageLimit ids per call so each call's result set stays bounded
// regardless of whether listMembers honors PaginationRequest.
func ResolveGroupIDs(ctx context.Context, listMembers ListSubscriptionGroupMembersFunc, subscriptionIDs []string) map[string]string {
	out := map[string]string{}
	if listMembers == nil || len(subscriptionIDs) == 0 {
		return out
	}
	for start := 0; start < len(subscriptionIDs); start += PageLimit {
		end := start + PageLimit
		if end > len(subscriptionIDs) {
			end = len(subscriptionIDs)
		}
		chunk := subscriptionIDs[start:end]
		resp, err := listMembers(ctx, &subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest{
			Filters: &commonpb.FilterRequest{
				Filters: []*commonpb.TypedFilter{{
					Field: "subscription_id",
					FilterType: &commonpb.TypedFilter_ListFilter{
						ListFilter: &commonpb.ListFilter{Values: chunk, Operator: commonpb.ListOperator_LIST_IN},
					},
				}},
			},
		})
		if err != nil {
			log.Printf("deliverygroup: list subscription group members: %v", err)
			continue
		}
		for _, m := range resp.GetData() {
			sid := m.GetSubscriptionId()
			if sid == "" {
				continue
			}
			if _, ok := out[sid]; !ok {
				out[sid] = m.GetSubscriptionGroupId()
			}
		}
	}
	return out
}

// ListAll returns every subscription_group in one call, keyed by id. The
// espyna adapter hydrates each group's nested PriceSchedule (STATUS-AGNOSTIC
// join — "a section may reference an archived AY"), so callers get the
// schedule (academic-year) name for free via
// SubscriptionGroup.GetPriceSchedule().GetName() — no separate fetch needed.
// One unfiltered call — safe while the workspace's subscription_group count
// stays under PageLimit (ground truth: 16 rows, education1).
func ListAll(ctx context.Context, listGroups ListSubscriptionGroupsFunc) map[string]*subscriptiongrouppb.SubscriptionGroup {
	out := map[string]*subscriptiongrouppb.SubscriptionGroup{}
	if listGroups == nil {
		return out
	}
	resp, err := listGroups(ctx, &subscriptiongrouppb.ListSubscriptionGroupsRequest{})
	if err != nil {
		log.Printf("deliverygroup: list subscription groups: %v", err)
		return out
	}
	for _, g := range resp.GetData() {
		if id := g.GetId(); id != "" {
			out[id] = g
		}
	}
	return out
}

// ResolveOne is the single-subscription convenience: resolve one origin
// subscription id straight to its delivery-group name + schedule
// (academic-year) name. Empty strings mean "unresolved" (nil-safe deps, an
// unmapped subscription, or no matching group) — callers render blank.
func ResolveOne(ctx context.Context, listMembers ListSubscriptionGroupMembersFunc, listGroups ListSubscriptionGroupsFunc, subscriptionID string) (groupName, scheduleName string) {
	if subscriptionID == "" {
		return "", ""
	}
	ids := ResolveGroupIDs(ctx, listMembers, []string{subscriptionID})
	groupID := ids[subscriptionID]
	if groupID == "" {
		return "", ""
	}
	groups := ListAll(ctx, listGroups)
	g := groups[groupID]
	if g == nil {
		return "", ""
	}
	return g.GetName(), g.GetPriceSchedule().GetName()
}
