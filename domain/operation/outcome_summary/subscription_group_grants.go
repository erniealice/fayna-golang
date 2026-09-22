package outcome_summary

import (
	"context"
	"strings"

	"github.com/erniealice/espyna-golang/consumer"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	workspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/workspace_user"
	subscriptiongroupworkspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_workspace_user"
)

// ListWorkspaceUsersFunc and ListSubscriptionGroupWorkspaceUsersFunc are the two
// reads that resolve the acting principal's subscription_group assignments (access
// axis: workspace_user -> subscription_group_workspace_user).
type (
	ListWorkspaceUsersFunc func(ctx context.Context, req *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error)

	ListSubscriptionGroupWorkspaceUsersFunc func(ctx context.Context, req *subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersRequest) (*subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersResponse, error)
)

// GrantedSubscriptionGroupIDs returns the subscription_group ids on which the
// acting session user holds an ACTIVE servicing grant. It is the shared resolver
// behind the landing's subscription_group visibility.
//
// The workspace_user list adapter ignores request filters (known limitation), so
// active workspace_users are matched on user_id in code; grants are re-checked
// against the acting workspace_user ids in case the subscription_group grant
// adapter ignores its filter. Any failure returns an error — callers must treat
// that as "no access".
func GrantedSubscriptionGroupIDs(ctx context.Context, listWorkspaceUsers ListWorkspaceUsersFunc, listGrants ListSubscriptionGroupWorkspaceUsersFunc) (map[string]bool, error) {
	granted := map[string]bool{}
	userID := strings.TrimSpace(consumer.GetUserIDFromContext(ctx))
	if userID == "" {
		return granted, nil // no identity → no grants (fail-closed)
	}
	wUResp, err := listWorkspaceUsers(ctx, &workspaceuserpb.ListWorkspaceUsersRequest{})
	if err != nil {
		return nil, err
	}
	myWU := map[string]bool{}
	for _, wu := range wUResp.GetData() {
		if wu.GetActive() && wu.GetUserId() == userID && wu.GetId() != "" {
			myWU[wu.GetId()] = true
		}
	}
	if len(myWU) == 0 {
		return granted, nil // no workspace_user for this principal
	}
	wuIDs := make([]string, 0, len(myWU))
	for id := range myWU {
		wuIDs = append(wuIDs, id)
	}
	sgResp, err := listGrants(ctx, &subscriptiongroupworkspaceuserpb.ListSubscriptionGroupWorkspaceUsersRequest{
		Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
			{
				Field: "workspace_user_id",
				FilterType: &commonpb.TypedFilter_ListFilter{
					ListFilter: &commonpb.ListFilter{Values: wuIDs, Operator: commonpb.ListOperator_LIST_IN},
				},
			},
			{
				Field:      "active",
				FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: true}},
			},
		}},
	})
	if err != nil {
		return nil, err
	}
	for _, g := range sgResp.GetData() {
		// Defense-in-depth against a filter-ignoring adapter: only honor grants
		// whose workspace_user_id is actually the acting principal's.
		if g.GetActive() && myWU[g.GetWorkspaceUserId()] {
			if subscriptionGroupID := g.GetSubscriptionGroupId(); subscriptionGroupID != "" {
				granted[subscriptionGroupID] = true
			}
		}
	}
	return granted, nil
}
