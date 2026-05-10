// Package action — set_status.go
//
// This file is the LIFTED implementation of NewPhaseSetStatusAction (previously
// at views/job/action/job_phase_set_status.go). It is now the canonical owner of
// this handler. The route /action/job-phase/set-status is registered here via
// JobPhaseModule.RegisterRoutes, keeping the same URL pattern so existing E2E
// tests in phase5-revenue-recognition/ that reference /action/job-phase/set-status
// continue to work without modification.
package action

import (
	"context"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"

	"github.com/erniealice/pyeza-golang/view"
)

// NewSetStatusAction creates the job_phase status update action (POST only).
//
// Reads `id` and `status` from the query string (form fallback). Accepts the
// proto enum name (PHASE_STATUS_PENDING / PHASE_STATUS_ACTIVE /
// PHASE_STATUS_COMPLETED) and lowercase shorthands.
//
// On COMPLETED transitions, espyna's UpdateJobPhase use case fires the
// OnJobPhaseCompleted hook internally (BillingEvent → READY). fayna does not
// duplicate that logic.
//
// Returns HX-Redirect to the phase detail page so the status badge refreshes.
func NewSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_phase", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		targetStatus := viewCtx.Request.URL.Query().Get("status")
		if id == "" || targetStatus == "" {
			_ = viewCtx.Request.ParseForm()
			if id == "" {
				id = viewCtx.Request.FormValue("id")
			}
			if targetStatus == "" {
				targetStatus = viewCtx.Request.FormValue("status")
			}
		}
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}
		if targetStatus == "" {
			return fayna.HTMXError("Status is required")
		}

		statusEnum := phaseStatusToEnum(targetStatus)
		if statusEnum == jobphasepb.PhaseStatus_PHASE_STATUS_UNSPECIFIED {
			return fayna.HTMXError("Invalid phase status")
		}

		// Read existing phase first — UpdateJobPhase requires Name in the
		// request payload (espyna validation). Fetching also verifies existence.
		jobID := ""
		phaseName := ""
		if deps.ReadJobPhase != nil {
			readResp, err := deps.ReadJobPhase(ctx, &jobphasepb.ReadJobPhaseRequest{
				Data: &jobphasepb.JobPhase{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job phase %s: %v", id, err)
				return fayna.HTMXError(err.Error())
			}
			data := readResp.GetData()
			if len(data) == 0 {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			jobID = data[0].GetJobId()
			phaseName = data[0].GetName()
		}

		if deps.UpdateJobPhase == nil {
			return fayna.HTMXError("Phase update not available")
		}

		_, err := deps.UpdateJobPhase(ctx, &jobphasepb.UpdateJobPhaseRequest{
			Data: &jobphasepb.JobPhase{
				Id:     id,
				JobId:  jobID,
				Name:   phaseName,
				Status: statusEnum,
			},
		})
		if err != nil {
			log.Printf("Failed to update phase %s status to %s: %v", id, targetStatus, err)
			return fayna.HTMXError(err.Error())
		}

		// Redirect to the phase detail page so the status badge refreshes.
		// Fall back to a 204 + trigger when the detail URL is unavailable.
		if deps.Routes.DetailURL != "" {
			return view.ViewResult{
				StatusCode: http.StatusNoContent,
				Headers: map[string]string{
					"HX-Redirect": deps.Routes.DetailURL + "?id=" + id,
				},
			}
		}
		_ = jobID // suppress unused warning — jobID used in legacy redirect path above
		return view.ViewResult{
			StatusCode: http.StatusNoContent,
			Headers:    map[string]string{"HX-Trigger": `{"jobPhaseStatusChanged":true}`},
		}
	})
}

// NewBulkSetStatusAction creates the job_phase bulk set-status action (POST only).
// Accepts multiple `id` fields + a single `target_status` field.
func NewBulkSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_phase", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}
		ids := viewCtx.Request.Form["id"]
		targetStatus := viewCtx.Request.FormValue("target_status")

		if len(ids) == 0 {
			return fayna.HTMXError("No IDs provided")
		}
		if targetStatus == "" {
			return fayna.HTMXError("Target status is required")
		}

		statusEnum := phaseStatusToEnum(targetStatus)

		for _, id := range ids {
			// Read first to get Name (required by espyna UpdateJobPhase).
			jobID := ""
			phaseName := ""
			if deps.ReadJobPhase != nil {
				readResp, err := deps.ReadJobPhase(ctx, &jobphasepb.ReadJobPhaseRequest{
					Data: &jobphasepb.JobPhase{Id: id},
				})
				if err != nil {
					log.Printf("Bulk set-status: failed to read phase %s: %v", id, err)
					continue
				}
				if data := readResp.GetData(); len(data) > 0 {
					jobID = data[0].GetJobId()
					phaseName = data[0].GetName()
				}
			}

			if deps.UpdateJobPhase == nil {
				continue
			}
			_, err := deps.UpdateJobPhase(ctx, &jobphasepb.UpdateJobPhaseRequest{
				Data: &jobphasepb.JobPhase{
					Id:     id,
					JobId:  jobID,
					Name:   phaseName,
					Status: statusEnum,
				},
			})
			if err != nil {
				log.Printf("Bulk set-status: failed to update phase %s: %v", id, err)
			}
		}

		return fayna.HTMXSuccess("job-phases-table")
	})
}
