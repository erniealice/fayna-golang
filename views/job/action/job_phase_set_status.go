package action

import (
	"context"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
)

// PhaseDeps holds the slim dependency surface for the JobPhase set-status
// handler. Kept separate from the Job-level Deps so the existing wiring stays
// untouched.
//
// 2026-04-29 milestone-billing plan §4 — operator clicks "Mark Complete" on
// the Phases tab; the espyna `UpdateJobPhase` use case fires the
// `OnJobPhaseCompleted` hook which advances any linked BillingEvent rows
// from UNSPECIFIED → READY. fayna does NOT duplicate that logic — it only
// drives the status transition through the existing use case.
type PhaseDeps struct {
	Routes fayna.JobRoutes
	Labels fayna.JobLabels

	ReadJobPhase   func(ctx context.Context, req *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error)
	UpdateJobPhase func(ctx context.Context, req *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error)
}

// NewPhaseSetStatusAction creates the JobPhase status update action (POST only).
//
// Reads `id` and `status` from the query string (form fallback). Resolves the
// status string (PHASE_STATUS_PENDING / PHASE_STATUS_ACTIVE /
// PHASE_STATUS_COMPLETED, plus shorthand pending/active/completed) and calls
// the espyna `UpdateJobPhase` use case which handles the OnJobPhaseCompleted
// hook internally on COMPLETED transitions.
//
// Returns an HTMX redirect back to the phases tab so the row's badge and CTA
// disabled state both refresh in one swap.
func NewPhaseSetStatusAction(deps *PhaseDeps) view.View {
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
			return fayna.HTMXError("Phase ID is required")
		}
		if targetStatus == "" {
			return fayna.HTMXError("Status is required")
		}

		statusEnum := phaseStatusToEnum(targetStatus)
		if statusEnum == jobphasepb.PhaseStatus_PHASE_STATUS_UNSPECIFIED {
			return fayna.HTMXError("Invalid phase status")
		}

		// Read existing phase first — UpdateJobPhase requires Name in the
		// request payload (espyna validation). Fetching also ensures the
		// phase exists before attempting the update.
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
				return fayna.HTMXError("Phase not found")
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

		// Refresh the phases tab partial so the badge + disabled CTA both
		// reflect the new status. When jobID is unknown (read disabled) we
		// fall back to a generic 204 — clients can re-trigger their own
		// refresh on success.
		if jobID == "" || deps.Routes.TabActionURL == "" {
			return view.ViewResult{
				StatusCode: http.StatusNoContent,
				Headers:    map[string]string{"HX-Trigger": `{"jobPhaseStatusChanged":true}`},
			}
		}
		tabActionURL := route.ResolveURL(deps.Routes.TabActionURL, "id", jobID, "tab", "phases")
		return view.ViewResult{
			StatusCode: http.StatusNoContent,
			Headers: map[string]string{
				"HX-Redirect": tabActionURL,
			},
		}
	})
}

// phaseStatusToEnum converts a status string (proto enum name or shorthand)
// to the protobuf PhaseStatus enum.
func phaseStatusToEnum(status string) jobphasepb.PhaseStatus {
	switch status {
	case "PHASE_STATUS_PENDING", "pending":
		return jobphasepb.PhaseStatus_PHASE_STATUS_PENDING
	case "PHASE_STATUS_ACTIVE", "active":
		return jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE
	case "PHASE_STATUS_COMPLETED", "completed":
		return jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED
	default:
		return jobphasepb.PhaseStatus_PHASE_STATUS_UNSPECIFIED
	}
}
