package action

import (
	"context"
	"log"
	"net/http"
	"strings"

	outcome_matrix "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
)

// transition.go — the four per-phase approval-bar POST handlers (submit / verify
// / publish / return). Each is a real, query-free HTMX POST form target under
// /action/* (CSRF + action-workspace guards apply). The form carries the
// job_template_phase_id in the body; {id} in the path is the job_template_id.
//
// SECURITY (copya.md gate discipline): the Layer-2 view gate here cites the SAME
// job_phase:<verb> code the espyna use-case ActionGatekeeper.Check + strict
// authorizer verify — no verb mismatch (job_activity's submit-gated-on-update
// defect is NOT repeated). The use case is authoritative (full-set authz, D7
// ownership / admin override, ancestry+workspace, hard-freeze, exact-set
// compare); this handler only forwards the trusted sheet identity. The raw
// server error is NEVER echoed to the client — a single lyngua'd fail-closed
// message is surfaced; the detail is logged server-side.

// TransitionDeps holds the four transition use-case closures + the routes/labels.
// Each closure is optional/nil-safe (a nil closure fails the action closed).
type TransitionDeps struct {
	Routes outcome_matrix.Routes
	Labels outcome_matrix.Labels

	Submit  func(ctx context.Context, req *jobphasepb.SubmitJobPhaseApprovalRequest) (*jobphasepb.SubmitJobPhaseApprovalResponse, error)
	Verify  func(ctx context.Context, req *jobphasepb.VerifyJobPhaseApprovalRequest) (*jobphasepb.VerifyJobPhaseApprovalResponse, error)
	Publish func(ctx context.Context, req *jobphasepb.PublishJobPhaseApprovalRequest) (*jobphasepb.PublishJobPhaseApprovalResponse, error)
	Return  func(ctx context.Context, req *jobphasepb.ReturnJobPhaseApprovalRequest) (*jobphasepb.ReturnJobPhaseApprovalResponse, error)
}

// NewSubmitAction returns the IN_PROGRESS → FOR_REVIEW POST handler.
func NewSubmitAction(deps *TransitionDeps) view.View {
	return newTransitionAction(deps, "submit", func(ctx context.Context, templateID, phaseID, _ string) error {
		if deps.Submit == nil {
			return errNotWired
		}
		_, err := deps.Submit(ctx, &jobphasepb.SubmitJobPhaseApprovalRequest{
			JobTemplateId:      templateID,
			JobTemplatePhaseId: phaseID,
		})
		return err
	})
}

// NewVerifyAction returns the FOR_REVIEW → VERIFIED POST handler.
func NewVerifyAction(deps *TransitionDeps) view.View {
	return newTransitionAction(deps, "verify", func(ctx context.Context, templateID, phaseID, _ string) error {
		if deps.Verify == nil {
			return errNotWired
		}
		_, err := deps.Verify(ctx, &jobphasepb.VerifyJobPhaseApprovalRequest{
			JobTemplateId:      templateID,
			JobTemplatePhaseId: phaseID,
		})
		return err
	})
}

// NewPublishAction returns the VERIFIED → PUBLISHED POST handler.
func NewPublishAction(deps *TransitionDeps) view.View {
	return newTransitionAction(deps, "publish", func(ctx context.Context, templateID, phaseID, _ string) error {
		if deps.Publish == nil {
			return errNotWired
		}
		_, err := deps.Publish(ctx, &jobphasepb.PublishJobPhaseApprovalRequest{
			JobTemplateId:      templateID,
			JobTemplatePhaseId: phaseID,
		})
		return err
	})
}

// NewReturnAction returns the mixed/advanced → IN_PROGRESS normalizer POST
// handler. The reason field is collected here; the server enforces the
// published-return non-blank-reason requirement.
func NewReturnAction(deps *TransitionDeps) view.View {
	return newTransitionAction(deps, "return", func(ctx context.Context, templateID, phaseID, reason string) error {
		if deps.Return == nil {
			return errNotWired
		}
		var reasonArg *string
		if r := strings.TrimSpace(reason); r != "" {
			reasonArg = &r
		}
		_, err := deps.Return(ctx, &jobphasepb.ReturnJobPhaseApprovalRequest{
			JobTemplateId:      templateID,
			JobTemplatePhaseId: phaseID,
			Reason:             reasonArg,
		})
		return err
	})
}

var errNotWired = &transitionError{"approval transition not configured"}

type transitionError struct{ msg string }

func (e *transitionError) Error() string { return e.msg }

// newTransitionAction is the shared handler body: verb gate → parse sheet
// identity → run → HX-Redirect reload on success / HTMXError banner on failure.
func newTransitionAction(deps *TransitionDeps, verb string, run func(ctx context.Context, templateID, phaseID, reason string) error) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		// Layer-2 gate cites the SAME job_phase:<verb> the use case gates on.
		if !perms.Can("job_phase", verb) {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		templateID := strings.TrimSpace(viewCtx.Request.PathValue("id"))
		if templateID == "" {
			return view.HTMXError(deps.Labels.Approval.Errors.ActionFailed)
		}
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError(deps.Labels.Approval.Errors.ActionFailed)
		}
		phaseID := strings.TrimSpace(viewCtx.Request.FormValue("job_template_phase_id"))
		reason := viewCtx.Request.FormValue("reason")
		if phaseID == "" {
			return view.HTMXError(deps.Labels.Approval.Errors.ActionFailed)
		}

		if err := run(ctx, templateID, phaseID, reason); err != nil {
			// Fail closed: log the detail, surface only the generic lyngua'd
			// message (never echo the raw server error — it could enumerate).
			log.Printf("outcome matrix approval %s: template=%s phase=%s: %v", verb, templateID, phaseID, err)
			return view.HTMXError(deps.Labels.Approval.Errors.ActionFailed)
		}

		// Success → client-side redirect back to the matrix page so the bar chips
		// + cell editability re-render from the new sheet state.
		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Redirect": route.ResolveURL(deps.Routes.MatrixURL, "id", templateID),
			},
		}
	})
}
