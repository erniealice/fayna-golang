package job_outcome_line

import (
	"context"
	"log"
	"net/http"
	"strconv"

	joboutcomelineform "github.com/erniealice/fayna-golang/domain/operation/job_outcome_line/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	joboutcomelinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"
)

// NewAddAction creates the job outcome line add action (GET = form, POST = create).
func NewAddAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_line", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("job-outcome-line-drawer-form", &joboutcomelineform.Data{
				FormAction:           deps.Routes.AddURL,
				ReportingRoleOptions: joboutcomelineform.DefaultReportingRoleOptions(),
				Labels:               deps.Labels,
				CommonLabels:         nil, // injected by ViewAdapter
			})
		}

		// POST — create job outcome line
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Errors.InvalidFormData)
		}

		r := viewCtx.Request

		record := &joboutcomelinepb.JobOutcomeLine{
			Label:         r.FormValue("label"),
			ReportingRole: reportingRoleFromString(r.FormValue("reporting_role")),
			Active:        true,
		}

		// Optional float fields
		if v := r.FormValue("weight_or_credits"); v != "" {
			f, _ := strconv.ParseFloat(v, 64)
			record.WeightOrCredits = &f
		}
		if v := r.FormValue("output_value"); v != "" {
			f, _ := strconv.ParseFloat(v, 64)
			record.OutputValue = &f
		}
		if v := r.FormValue("output_label"); v != "" {
			record.OutputLabel = &v
		}
		if v := r.FormValue("score_scale_band_id"); v != "" {
			record.ScoreScaleBandId = &v
		}

		_, err := deps.CreateJobOutcomeLine(ctx, &joboutcomelinepb.CreateJobOutcomeLineRequest{
			Data: record,
		})
		if err != nil {
			log.Printf("Failed to create job outcome line: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("job-outcome-lines-table")
	})
}

// NewEditAction creates the job outcome line edit action (GET = pre-filled form, POST = update).
func NewEditAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_line", "update") {
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

			readResp, err := deps.ReadJobOutcomeLine(ctx, &joboutcomelinepb.ReadJobOutcomeLineRequest{
				Data: &joboutcomelinepb.JobOutcomeLine{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job outcome line %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			rec := readData[0]

			return view.OK("job-outcome-line-drawer-form", &joboutcomelineform.Data{
				FormAction:           route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:               true,
				ID:                   id,
				Label:                rec.GetLabel(),
				WeightOrCredits:      rec.GetWeightOrCredits(),
				OutputValue:          rec.GetOutputValue(),
				OutputLabel:          rec.GetOutputLabel(),
				OutputLabelSet:       rec.OutputLabel != nil,
				ScoreScaleBandId:     rec.GetScoreScaleBandId(),
				ScoreScaleBandIdSet:  rec.ScoreScaleBandId != nil,
				ReportingRole:        rec.GetReportingRole().String(),
				ReportingRoleOptions: joboutcomelineform.DefaultReportingRoleOptions(),
				Labels:               deps.Labels,
				CommonLabels:         nil, // injected by ViewAdapter
			})
		}

		// POST — update job outcome line
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

		record := &joboutcomelinepb.JobOutcomeLine{
			Id:            id,
			Label:         r.FormValue("label"),
			ReportingRole: reportingRoleFromString(r.FormValue("reporting_role")),
		}

		// Optional float fields — nil = unset (NULL)
		if v := r.FormValue("weight_or_credits"); v != "" {
			f, _ := strconv.ParseFloat(v, 64)
			record.WeightOrCredits = &f
		}
		if v := r.FormValue("output_value"); v != "" {
			f, _ := strconv.ParseFloat(v, 64)
			record.OutputValue = &f
		}
		if v := r.FormValue("output_label"); v != "" {
			record.OutputLabel = &v
		}
		if v := r.FormValue("score_scale_band_id"); v != "" {
			record.ScoreScaleBandId = &v
		}

		_, err := deps.UpdateJobOutcomeLine(ctx, &joboutcomelinepb.UpdateJobOutcomeLineRequest{
			Data: record,
		})
		if err != nil {
			log.Printf("Failed to update job outcome line %s: %v", id, err)
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

// NewDeleteAction creates the job outcome line delete action (POST only).
func NewDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_line", "delete") {
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

		_, err := deps.DeleteJobOutcomeLine(ctx, &joboutcomelinepb.DeleteJobOutcomeLineRequest{
			Data: &joboutcomelinepb.JobOutcomeLine{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete job outcome line %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("job-outcome-lines-table")
	})
}

// NewBulkDeleteAction creates the job outcome line bulk delete action (POST only).
func NewBulkDeleteAction(deps *ModuleDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_line", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteJobOutcomeLine(ctx, &joboutcomelinepb.DeleteJobOutcomeLineRequest{
				Data: &joboutcomelinepb.JobOutcomeLine{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete job outcome line %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("job-outcome-lines-table")
	})
}

// reportingRoleFromString converts a proto enum name string to the ReportingRole enum.
// Round-trips via ReportingRole_value map (proto-idiomatic enum string round-trip).
func reportingRoleFromString(s string) enums.ReportingRole {
	if v, ok := enums.ReportingRole_value[s]; ok {
		return enums.ReportingRole(v)
	}
	return enums.ReportingRole_REPORTING_ROLE_UNSPECIFIED
}

// strPtrIfNotEmpty returns a pointer to s if non-empty, otherwise nil.
func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
