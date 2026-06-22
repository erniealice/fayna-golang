package reporting_checkpoint

import (
	"context"
	"log"
	"net/http"
	"strconv"

	checkpointform "github.com/erniealice/fayna-golang/domain/operation/reporting_checkpoint/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	checkpointpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/reporting_checkpoint"
)

// NewAddAction creates the reporting checkpoint add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("reporting_checkpoint", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("reporting-checkpoint-drawer-form", &checkpointform.Data{
				FormAction:           deps.Routes.AddURL,
				VersionStatusOptions: checkpointform.DefaultVersionStatusOptions(),
				Labels:               deps.Labels,
				CommonLabels:         nil, // injected by ViewAdapter
			})
		}

		// POST — create reporting checkpoint
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		seqOrder, _ := strconv.ParseInt(r.FormValue("sequence_order"), 10, 32)
		vsVal := enums.VersionStatus_value[r.FormValue("version_status")]

		req := &checkpointpb.CreateReportingCheckpointRequest{
			Data: &checkpointpb.ReportingCheckpoint{
				CheckpointGroupId: r.FormValue("checkpoint_group_id"),
				Label:             r.FormValue("label"),
				RoleCode:          r.FormValue("role_code"),
				SequenceOrder:     int32(seqOrder),
				IsTerminal:        r.FormValue("is_terminal") == "true" || r.FormValue("is_terminal") == "on",
				Active:            true,
				VersionStatus:     enums.VersionStatus(vsVal),
				WorkspaceId:       strPtrIfNotEmpty(r.FormValue("workspace_id")),
				PeriodId:          strPtrIfNotEmpty(r.FormValue("period_id")),
			},
		}

		_, err := deps.CreateReportingCheckpoint(ctx, req)
		if err != nil {
			log.Printf("Failed to create reporting checkpoint: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("reporting-checkpoint-table")
	})
}

// NewEditAction creates the reporting checkpoint edit action (GET = pre-filled form, POST = update).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("reporting_checkpoint", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}

		if viewCtx.Request.Method == http.MethodGet {
			if id == "" {
				return view.HTMXError(deps.Labels.Errors.IDRequired)
			}

			readResp, err := deps.ReadReportingCheckpoint(ctx, &checkpointpb.ReadReportingCheckpointRequest{
				Data: &checkpointpb.ReportingCheckpoint{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read reporting checkpoint %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			record := readData[0]

			return view.OK("reporting-checkpoint-drawer-form", &checkpointform.Data{
				FormAction:           route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:               true,
				ID:                   id,
				Label:                record.GetLabel(),
				CheckpointGroupID:    record.GetCheckpointGroupId(),
				RoleCode:             record.GetRoleCode(),
				SequenceOrder:        record.GetSequenceOrder(),
				WorkspaceID:          record.GetWorkspaceId(),
				PeriodID:             record.GetPeriodId(),
				IsTerminal:           record.GetIsTerminal(),
				VersionStatus:        record.GetVersionStatus().String(),
				VersionStatusOptions: checkpointform.DefaultVersionStatusOptions(),
				Labels:               deps.Labels,
				CommonLabels:         nil, // injected by ViewAdapter
			})
		}

		// POST — update reporting checkpoint
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request
		if id == "" {
			id = r.FormValue("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		seqOrder, _ := strconv.ParseInt(r.FormValue("sequence_order"), 10, 32)
		vsVal := enums.VersionStatus_value[r.FormValue("version_status")]

		_, err := deps.UpdateReportingCheckpoint(ctx, &checkpointpb.UpdateReportingCheckpointRequest{
			Data: &checkpointpb.ReportingCheckpoint{
				Id:            id,
				Label:         r.FormValue("label"),
				RoleCode:      r.FormValue("role_code"),
				SequenceOrder: int32(seqOrder),
				IsTerminal:    r.FormValue("is_terminal") == "true" || r.FormValue("is_terminal") == "on",
				VersionStatus: enums.VersionStatus(vsVal),
				WorkspaceId:   strPtrIfNotEmpty(r.FormValue("workspace_id")),
				PeriodId:      strPtrIfNotEmpty(r.FormValue("period_id")),
			},
		})
		if err != nil {
			log.Printf("Failed to update reporting checkpoint %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Trigger":  `{"formSuccess":true}`,
				"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", id),
			},
		}
	})
}

// NewDeleteAction creates the reporting checkpoint delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("reporting_checkpoint", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.DeleteReportingCheckpoint(ctx, &checkpointpb.DeleteReportingCheckpointRequest{
			Data: &checkpointpb.ReportingCheckpoint{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete reporting checkpoint %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("reporting-checkpoint-table")
	})
}

// NewBulkDeleteAction creates the reporting checkpoint bulk delete action (POST only).
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("reporting_checkpoint", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteReportingCheckpoint(ctx, &checkpointpb.DeleteReportingCheckpointRequest{
				Data: &checkpointpb.ReportingCheckpoint{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete reporting checkpoint %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("reporting-checkpoint-table")
	})
}

// strPtrIfNotEmpty returns a pointer to s if non-empty, otherwise nil.
func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
