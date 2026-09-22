package outcome_summary

import (
	"context"
	"errors"
	"testing"

	"github.com/erniealice/espyna-golang/consumer"

	workspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/workspace_user"
	sgwupb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_workspace_user"
)

func wuList(rows ...*workspaceuserpb.WorkspaceUser) ListWorkspaceUsersFunc {
	return func(context.Context, *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error) {
		return &workspaceuserpb.ListWorkspaceUsersResponse{Data: rows}, nil
	}
}

func grantList(rows ...*sgwupb.SubscriptionGroupWorkspaceUser) ListSubscriptionGroupWorkspaceUsersFunc {
	return func(context.Context, *sgwupb.ListSubscriptionGroupWorkspaceUsersRequest) (*sgwupb.ListSubscriptionGroupWorkspaceUsersResponse, error) {
		return &sgwupb.ListSubscriptionGroupWorkspaceUsersResponse{Data: rows}, nil
	}
}

func wu(id, userID string, active bool) *workspaceuserpb.WorkspaceUser {
	return &workspaceuserpb.WorkspaceUser{Id: id, UserId: userID, Active: active}
}

func grant(wuID, subscriptionGroupID string, active bool) *sgwupb.SubscriptionGroupWorkspaceUser {
	return &sgwupb.SubscriptionGroupWorkspaceUser{WorkspaceUserId: wuID, SubscriptionGroupId: subscriptionGroupID, Active: active}
}

func TestGrantedSubscriptionGroupIDs_AssignedOnly(t *testing.T) {
	ctx := consumer.WithUserID(context.Background(), "user-1")
	got, err := GrantedSubscriptionGroupIDs(
		ctx,
		wuList(
			wu("wu-1", "user-1", true),
			wu("wu-2", "user-2", true),
			wu("wu-inactive", "user-1", false),
		),
		grantList(
			grant("wu-1", "sg-own", true),
			grant("wu-1", "sg-inactive", false),
			grant("wu-2", "sg-other", true),
			grant("wu-inactive", "sg-inactive-wu", true),
		),
	)
	if err != nil {
		t.Fatalf("GrantedSubscriptionGroupIDs returned error: %v", err)
	}
	if len(got) != 1 || !got["sg-own"] {
		t.Fatalf("expected only the acting user's active grant, got %#v", got)
	}
}

func TestGrantedSubscriptionGroupIDs_NoIdentityIsEmpty(t *testing.T) {
	got, err := GrantedSubscriptionGroupIDs(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("GrantedSubscriptionGroupIDs returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no grants without an identity, got %#v", got)
	}
}

func TestGrantedSubscriptionGroupIDs_ReadErrorsPropagate(t *testing.T) {
	ctx := consumer.WithUserID(context.Background(), "user-1")
	wantWorkspaceUserErr := errors.New("workspace user read failed")
	_, err := GrantedSubscriptionGroupIDs(ctx, func(context.Context, *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error) {
		return nil, wantWorkspaceUserErr
	}, grantList())
	if !errors.Is(err, wantWorkspaceUserErr) {
		t.Fatalf("workspace_user error = %v, want %v", err, wantWorkspaceUserErr)
	}

	wantGrantErr := errors.New("subscription_group grant read failed")
	_, err = GrantedSubscriptionGroupIDs(ctx, wuList(wu("wu-1", "user-1", true)), func(context.Context, *sgwupb.ListSubscriptionGroupWorkspaceUsersRequest) (*sgwupb.ListSubscriptionGroupWorkspaceUsersResponse, error) {
		return nil, wantGrantErr
	})
	if !errors.Is(err, wantGrantErr) {
		t.Fatalf("subscription_group grant error = %v, want %v", err, wantGrantErr)
	}
}
