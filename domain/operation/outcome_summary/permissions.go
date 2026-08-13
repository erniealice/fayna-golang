package outcome_summary

import (
	"context"

	"github.com/erniealice/pyeza-golang/types"
)

// Permission families for the two deliberately separate report surfaces.
// The legacy family protects the landing/detail graph; the export family is
// the narrow, period-scoped composite download capability.
const (
	LegacyPermissionEntity = "job_outcome_summary"
	LegacyListAction       = "list"
	LegacyReadAction       = "read"
	ExportPermissionEntity = "subscription_group_outcome_export"
	ExportReadAction       = "read"
)

// Export principal kinds are the only operator/persona bindings allowed to
// receive the consolidated section export. Keep these numeric values aligned
// with the canonical PrincipalType enum without importing the HTTP/session
// layer into Fayna.
const (
	PrincipalKindOperatorOwner int32 = 1
	PrincipalKindOperatorStaff int32 = 2
	PrincipalKindStaff         int32 = 7
)

func CanLegacyLanding(perms *types.UserPermissions) bool {
	return perms != nil && perms.Can(LegacyPermissionEntity, LegacyListAction)
}

func CanLegacyDetail(perms *types.UserPermissions) bool {
	return CanLegacyLanding(perms) && perms.Can(LegacyPermissionEntity, LegacyReadAction)
}

func CanExplicitExport(perms *types.UserPermissions, ctx context.Context, resolve func(context.Context) int32) bool {
	if perms == nil || !perms.Can(ExportPermissionEntity, ExportReadAction) || resolve == nil {
		return false
	}
	switch resolve(ctx) {
	case PrincipalKindOperatorOwner, PrincipalKindOperatorStaff, PrincipalKindStaff:
		return true
	default:
		return false
	}
}
