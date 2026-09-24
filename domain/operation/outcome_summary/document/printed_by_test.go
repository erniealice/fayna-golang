package document

import (
	"context"
	"errors"
	"testing"

	userpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/user"
	workspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/workspace_user"
)

func TestPrintedByNamePrefersOwnDisplayNameLastFirst(t *testing.T) {
	d := &Deps{
		ReadSelfDisplayName: func(context.Context) (string, string, error) { return "Junrey", "Tejas", nil },
		ListWorkspaceUsers: func(context.Context, *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error) {
			t.Fatal("workspace user list must not be needed when the self read succeeds")
			return nil, nil
		},
	}
	if got := printedByName(context.Background(), d, "u1"); got != "Tejas, Junrey" {
		t.Fatalf("printedByName = %q, want %q", got, "Tejas, Junrey")
	}
}

func TestPrintedByNameFallsBackToWorkspaceUserLastFirst(t *testing.T) {
	d := &Deps{
		ReadSelfDisplayName: func(context.Context) (string, string, error) { return "", "", errors.New("denied") },
		ListWorkspaceUsers: func(context.Context, *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error) {
			return &workspaceuserpb.ListWorkspaceUsersResponse{Data: []*workspaceuserpb.WorkspaceUser{
				{UserId: "u2", Active: true, User: &userpb.User{FirstName: "Other", LastName: "Person"}},
				{UserId: "u1", Active: true, User: &userpb.User{FirstName: "Nina", LastName: "Pareja"}},
			}}, nil
		},
	}
	if got := printedByName(context.Background(), d, "u1"); got != "Pareja, Nina" {
		t.Fatalf("printedByName = %q, want %q", got, "Pareja, Nina")
	}
}

func TestLastFirst(t *testing.T) {
	for _, tc := range []struct{ first, last, want string }{
		{"Junrey", "Tejas", "Tejas, Junrey"},
		{" Junrey ", "", "Junrey"},
		{"", "Tejas", "Tejas"},
		{"", "", ""},
	} {
		if got := lastFirst(tc.first, tc.last); got != tc.want {
			t.Errorf("lastFirst(%q, %q) = %q, want %q", tc.first, tc.last, got, tc.want)
		}
	}
}
