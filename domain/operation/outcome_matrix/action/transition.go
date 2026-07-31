package action

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

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
	return newTransitionAction(deps, "submit", func(ctx context.Context, templateID, phaseID, _, groupID string) error {
		if deps.Submit == nil {
			return errNotWired
		}
		req := &jobphasepb.SubmitJobPhaseApprovalRequest{
			JobTemplateId:      templateID,
			JobTemplatePhaseId: phaseID,
		}
		// Set ONLY when the route supplied a group. Absent ⇒ the whole template,
		// i.e. the template-scoped page behaves exactly as it always has.
		if groupID != "" {
			req.SubscriptionGroupId = &groupID
		}
		_, err := deps.Submit(ctx, req)
		return err
	})
}

// NewVerifyAction returns the FOR_REVIEW → VERIFIED POST handler.
func NewVerifyAction(deps *TransitionDeps) view.View {
	return newTransitionAction(deps, "verify", func(ctx context.Context, templateID, phaseID, _, groupID string) error {
		if deps.Verify == nil {
			return errNotWired
		}
		req := &jobphasepb.VerifyJobPhaseApprovalRequest{
			JobTemplateId:      templateID,
			JobTemplatePhaseId: phaseID,
		}
		// Set ONLY when the route supplied a group. Absent ⇒ the whole template,
		// i.e. the template-scoped page behaves exactly as it always has.
		if groupID != "" {
			req.SubscriptionGroupId = &groupID
		}
		_, err := deps.Verify(ctx, req)
		return err
	})
}

// NewPublishAction returns the VERIFIED → PUBLISHED POST handler.
func NewPublishAction(deps *TransitionDeps) view.View {
	return newTransitionAction(deps, "publish", func(ctx context.Context, templateID, phaseID, _, groupID string) error {
		if deps.Publish == nil {
			return errNotWired
		}
		req := &jobphasepb.PublishJobPhaseApprovalRequest{
			JobTemplateId:      templateID,
			JobTemplatePhaseId: phaseID,
		}
		// Set ONLY when the route supplied a group. Absent ⇒ the whole template,
		// i.e. the template-scoped page behaves exactly as it always has.
		if groupID != "" {
			req.SubscriptionGroupId = &groupID
		}
		_, err := deps.Publish(ctx, req)
		return err
	})
}

// NewReturnAction returns the mixed/advanced → IN_PROGRESS normalizer POST
// handler. The reason field is collected here; the server enforces the
// published-return non-blank-reason requirement.
func NewReturnAction(deps *TransitionDeps) view.View {
	return newTransitionAction(deps, "return", func(ctx context.Context, templateID, phaseID, reason, groupID string) error {
		if deps.Return == nil {
			return errNotWired
		}
		var reasonArg *string
		if r := strings.TrimSpace(reason); r != "" {
			reasonArg = &r
		}
		req := &jobphasepb.ReturnJobPhaseApprovalRequest{
			JobTemplateId:      templateID,
			JobTemplatePhaseId: phaseID,
			Reason:             reasonArg,
		}
		if groupID != "" {
			req.SubscriptionGroupId = &groupID
		}
		_, err := deps.Return(ctx, req)
		return err
	})
}

var errNotWired = &transitionError{"approval transition not configured"}

type transitionError struct{ msg string }

func (e *transitionError) Error() string { return e.msg }

// newTransitionAction is the shared handler body: verb gate → parse sheet
// identity → run → HX-Redirect reload on success / HTMXError banner on failure.
func newTransitionAction(deps *TransitionDeps, verb string, run func(ctx context.Context, templateID, phaseID, reason, groupID string) error) view.View {
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
		// Delivery group from the PATH (the Group* route forms). Empty on the
		// template-scoped routes, so those are byte-identical to before. It comes
		// from the path rather than the form body deliberately: {{actionForm}}
		// signs the exact resolved path, so the signature covers the group and a
		// tampered id fails the workspace-form guard instead of retargeting the
		// transition at another group.
		groupID := strings.TrimSpace(viewCtx.Request.PathValue("group_id"))
		phaseID := strings.TrimSpace(viewCtx.Request.FormValue("job_template_phase_id"))
		reason := viewCtx.Request.FormValue("reason")
		if phaseID == "" {
			return view.HTMXError(deps.Labels.Approval.Errors.ActionFailed)
		}

		if err := run(ctx, templateID, phaseID, reason, groupID); err != nil {
			// Fail closed: log the detail, surface only the generic lyngua'd
			// message (never echo the raw server error — it could enumerate).
			log.Printf("outcome matrix approval %s: template=%s phase=%s group=%s: %v", verb, templateID, phaseID, groupID, err)
			return view.HTMXError(deps.Labels.Approval.Errors.ActionFailed)
		}

		// Success → client-side redirect back to the matrix page so the bar chips
		// + cell editability re-render from the new sheet state.
		//
		// Redirect to the SAME scope the transition was invoked from: a
		// section-scoped (Group*) transition must return to the section sheet, not
		// the template-wide one. Sending a section operator back to MatrixURL drops
		// the group narrowing and dumps them on the all-sections sheet, losing their
		// place after every submit/verify/publish/return.
		//
		// Guarded like the row-link grain decision (job/list/template_summary.go):
		// fall back to template grain when the group route is unconfigured or the
		// path carried no group id — never emit a half-resolved path. Deliberately
		// NOT gated on job.Options.RowLinkScopedByGroup(): that option chooses where
		// the job LIST links to, whereas this is "return where the request came
		// from". The group matrix route is mounted whenever GroupMatrixURL is set
		// (outcome_matrix_module.go), so an operator can reach the section sheet by
		// link or bookmark even in a deployment whose list links at template grain —
		// and bouncing them to another scope would still be wrong.
		redirect := route.ResolveURL(deps.Routes.MatrixURL, "id", templateID)
		if groupID != "" && deps.Routes.GroupMatrixURL != "" {
			redirect = route.ResolveURL(deps.Routes.GroupMatrixURL, "id", templateID, "group_id", groupID)
		}
		// Re-render the sheet IN PLACE rather than reloading the browser.
		// htmx 1.9.10 handles HX-Redirect by setting location.href — a full page
		// reload, which is why the operator watches the OLD table paint again
		// before the new state appears. HX-Location instead issues a client-side
		// AJAX GET and swaps, keeping the app shell, the loading indicator and the
		// scroll position, so the bar comes back already carrying the next verb.
		//
		// target ECHOES the request's HX-Target — the id htmx resolved from the
		// form's own hx-target — so no layout constant ("#main-content") is baked
		// into a domain package; the handler follows whatever the view is
		// configured to swap. A non-htmx caller sends no HX-Target, so that path
		// keeps the plain full-page redirect and still completes the transition.
		//
		// Deliberately NO "select": the matrix GET already answers an HX-Request
		// with the bare content partial (<div class="page-content">…), not a full
		// document, so there is no #main-content element inside the response to
		// select — asking for one selects nothing and blanks the target.
		if t := strings.TrimSpace(viewCtx.Request.Header.Get("HX-Target")); t != "" {
			if loc, err := json.Marshal(map[string]string{
				"path":   redirect,
				"target": "#" + t,
			}); err == nil {
				return view.ViewResult{
					StatusCode: http.StatusOK,
					Headers:    map[string]string{"HX-Location": string(loc)},
				}
			}
		}
		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Redirect": redirect,
			},
		}
	})
}
