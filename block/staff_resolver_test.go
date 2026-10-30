package block

import (
	"context"
	"errors"
	"testing"

	"github.com/erniealice/espyna-golang/shared/identity"
	principaltypepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/principal_type"
	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
)

// errStaffListDenied stands in for the AUTHZ_DENIED a teacher role without
// staff:list gets from ListStaffs (the 20261030 grade-sheet 422 root cause).
var errStaffListDenied = errors.New("permission denied: staff:list")

type staffListSpy struct {
	calls int
	resp  *staffpb.ListStaffsResponse
	err   error
}

func (s *staffListSpy) list(context.Context, *staffpb.ListStaffsRequest) (*staffpb.ListStaffsResponse, error) {
	s.calls++
	return s.resp, s.err
}

func identityCtx(userID string, principalType principaltypepb.PrincipalType, principalID string) context.Context {
	return identity.WithRequestIdentity(context.Background(), &identity.RequestIdentity{
		UserID:        userID,
		WorkspaceID:   "ws-1",
		PrincipalType: int32(principalType),
		PrincipalID:   principalID,
	})
}

// TestStaffResolver_StaffPersonaWithoutStaffList pins the fix: a STAFF-persona
// teacher whose role lacks staff:list still resolves to the session-bound
// staff.id, so grade-sheet cell writes are not denied — and staff:list is never
// consulted.
func TestStaffResolver_StaffPersonaWithoutStaffList(t *testing.T) {
	t.Parallel()
	spy := &staffListSpy{err: errStaffListDenied}
	resolve := newStaffResolver(spy.list)

	got, err := resolve(identityCtx("user-1", principaltypepb.PrincipalType_PRINCIPAL_TYPE_STAFF, "staff-1"))
	if err != nil {
		t.Fatalf("resolve: unexpected error %v", err)
	}
	if got != "staff-1" {
		t.Fatalf("resolve = %q, want the session-bound staff id %q", got, "staff-1")
	}
	if spy.calls != 0 {
		t.Fatalf("ListStaffs called %d times for a staff persona, want 0 (no staff:list dependency)", spy.calls)
	}
}

// TestStaffResolver_FailsClosed covers every unresolvable identity: each must
// return "" so the record action denies the write.
func TestStaffResolver_FailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		ctx  context.Context
		spy  *staffListSpy
	}{
		{"no identity", context.Background(), &staffListSpy{}},
		{"staff persona with empty principal", identityCtx("user-1", principaltypepb.PrincipalType_PRINCIPAL_TYPE_STAFF, ""), &staffListSpy{}},
		{"operator persona denied staff:list", identityCtx("user-1", principaltypepb.PrincipalType_PRINCIPAL_TYPE_OPERATOR_STAFF, "op-1"), &staffListSpy{err: errStaffListDenied}},
		{"operator persona with no staff row", identityCtx("user-1", principaltypepb.PrincipalType_PRINCIPAL_TYPE_OPERATOR_STAFF, "op-1"), &staffListSpy{resp: &staffpb.ListStaffsResponse{}}},
		{"operator persona with another user's row only", identityCtx("user-1", principaltypepb.PrincipalType_PRINCIPAL_TYPE_OPERATOR_STAFF, "op-1"), &staffListSpy{resp: &staffpb.ListStaffsResponse{
			Data: []*staffpb.Staff{{Id: "staff-x", UserId: "user-2", Active: true}},
		}}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, _ := newStaffResolver(c.spy.list)(c.ctx)
			if got != "" {
				t.Fatalf("resolve = %q, want \"\" (fail closed)", got)
			}
		})
	}
}

// TestStaffResolver_OperatorPersonaUsesStaffList keeps the operator path: the
// user → staff mapping still goes through the typed list use case, re-verified
// to this user's active rows and picked deterministically.
func TestStaffResolver_OperatorPersonaUsesStaffList(t *testing.T) {
	t.Parallel()
	spy := &staffListSpy{resp: &staffpb.ListStaffsResponse{Data: []*staffpb.Staff{
		{Id: "staff-b", UserId: "user-1", Active: true},
		{Id: "staff-a", UserId: "user-1", Active: true},
		{Id: "staff-0", UserId: "user-1", Active: false},
	}}}
	got, err := newStaffResolver(spy.list)(identityCtx("user-1", principaltypepb.PrincipalType_PRINCIPAL_TYPE_OPERATOR_STAFF, "op-1"))
	if err != nil {
		t.Fatalf("resolve: unexpected error %v", err)
	}
	if got != "staff-a" || spy.calls != 1 {
		t.Fatalf("resolve = %q after %d ListStaffs calls, want %q after 1", got, spy.calls, "staff-a")
	}
}
