package outcome_summary

import (
	"context"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
)

func TestCanExplicitExportAllowsOwnerOperatorStaffAndStaff(t *testing.T) {
	perms := types.NewUserPermissions([]string{"subscription_group_outcome_export:read"})
	for _, tc := range []struct {
		name string
		kind int32
		want bool
	}{
		{name: "operator owner", kind: PrincipalKindOperatorOwner, want: true},
		{name: "operator staff", kind: PrincipalKindOperatorStaff, want: true},
		{name: "staff", kind: PrincipalKindStaff, want: true},
		{name: "unresolved", kind: 0},
		{name: "unknown", kind: 99},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanExplicitExport(perms, context.Background(), func(context.Context) int32 { return tc.kind }); got != tc.want {
				t.Fatalf("kind %d: CanExplicitExport = %v, want %v", tc.kind, got, tc.want)
			}
		})
	}
}

func TestCanExplicitExportFailsClosedWithoutCapabilityOrResolver(t *testing.T) {
	if CanExplicitExport(types.NewUserPermissions([]string{"job_outcome_summary:list"}), context.Background(), func(context.Context) int32 { return PrincipalKindOperatorOwner }) {
		t.Fatal("legacy list must not authorize explicit export")
	}
	if CanExplicitExport(types.NewUserPermissions([]string{"subscription_group_outcome_export:read"}), context.Background(), nil) {
		t.Fatal("missing principal resolver must fail closed")
	}
}

func TestCanLegacyDetailRequiresListAndRead(t *testing.T) {
	for _, tc := range []struct {
		name  string
		codes []string
		want  bool
	}{
		{name: "both", codes: []string{"job_outcome_summary:list", "job_outcome_summary:read"}, want: true},
		{name: "list only", codes: []string{"job_outcome_summary:list"}},
		{name: "read only", codes: []string{"job_outcome_summary:read"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanLegacyDetail(types.NewUserPermissions(tc.codes)); got != tc.want {
				t.Fatalf("CanLegacyDetail(%v) = %v, want %v", tc.codes, got, tc.want)
			}
		})
	}
}
