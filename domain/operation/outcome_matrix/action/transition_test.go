package action

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"unicode/utf8"

	espynaports "github.com/erniealice/espyna-golang/ports"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

func TestNewReturnAction_ReasonNormalization(t *testing.T) {
	tests := []struct {
		name       string
		reason     string
		wantReason *string
	}{
		{
			name:       "trims non-blank reason",
			reason:     "  Needs more evidence  ",
			wantReason: strp("Needs more evidence"),
		},
		{
			name:       "whitespace is blank",
			reason:     " \t\n ",
			wantReason: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got *string
			deps := &TransitionDeps{
				Routes: outcome_matrix.DefaultRoutes(),
				Labels: outcome_matrix.DefaultLabels(),
				Return: func(_ context.Context, req *jobphasepb.ReturnJobPhaseApprovalRequest) (*jobphasepb.ReturnJobPhaseApprovalResponse, error) {
					got = req.Reason
					return &jobphasepb.ReturnJobPhaseApprovalResponse{}, nil
				},
			}

			form := url.Values{
				"job_template_phase_id": []string{"phase-1"},
				"reason":                []string{tc.reason},
			}
			req := httptest.NewRequest(http.MethodPost, "/action/outcome-matrix/template-1/return", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.SetPathValue("id", "template-1")
			ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions([]string{"job_phase:return"}))

			result := NewReturnAction(deps).Handle(ctx, &view.ViewContext{Request: req})
			if result.StatusCode != http.StatusOK {
				t.Fatalf("status code = %d, want %d", result.StatusCode, http.StatusOK)
			}

			if got == nil && tc.wantReason == nil {
				return
			}
			if got == nil || tc.wantReason == nil || *got != *tc.wantReason {
				var gotValue string
				if got != nil {
					gotValue = *got
				}
				var wantValue string
				if tc.wantReason != nil {
					wantValue = *tc.wantReason
				}
				t.Fatalf("reason = %q, want %q", gotValue, wantValue)
			}
		})
	}
}

type transitionVerb struct {
	name string
	deps func(err error) *TransitionDeps
	call func(d *TransitionDeps) view.View
}

func transitionVerbs() []transitionVerb {
	base := func() *TransitionDeps {
		return &TransitionDeps{Routes: outcome_matrix.DefaultRoutes(), Labels: outcome_matrix.DefaultLabels()}
	}
	return []transitionVerb{
		{"submit", func(err error) *TransitionDeps {
			d := base()
			d.Submit = func(context.Context, *jobphasepb.SubmitJobPhaseApprovalRequest) (*jobphasepb.SubmitJobPhaseApprovalResponse, error) {
				return &jobphasepb.SubmitJobPhaseApprovalResponse{}, err
			}
			return d
		}, NewSubmitAction},
		{"verify", func(err error) *TransitionDeps {
			d := base()
			d.Verify = func(context.Context, *jobphasepb.VerifyJobPhaseApprovalRequest) (*jobphasepb.VerifyJobPhaseApprovalResponse, error) {
				return &jobphasepb.VerifyJobPhaseApprovalResponse{}, err
			}
			return d
		}, NewVerifyAction},
		{"publish", func(err error) *TransitionDeps {
			d := base()
			d.Publish = func(context.Context, *jobphasepb.PublishJobPhaseApprovalRequest) (*jobphasepb.PublishJobPhaseApprovalResponse, error) {
				return &jobphasepb.PublishJobPhaseApprovalResponse{}, err
			}
			return d
		}, NewPublishAction},
		{"return", func(err error) *TransitionDeps {
			d := base()
			d.Return = func(context.Context, *jobphasepb.ReturnJobPhaseApprovalRequest) (*jobphasepb.ReturnJobPhaseApprovalResponse, error) {
				return &jobphasepb.ReturnJobPhaseApprovalResponse{}, err
			}
			return d
		}, NewReturnAction},
	}
}

func runTransition(t *testing.T, v transitionVerb, err error, groupID, hxTarget string) (*TransitionDeps, view.ViewResult) {
	t.Helper()
	deps := v.deps(err)
	form := url.Values{"job_template_phase_id": []string{"phase-1"}}
	req := httptest.NewRequest(http.MethodPost, "/action/outcome-matrix/template-1/"+v.name, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if hxTarget != "" {
		req.Header.Set("HX-Target", hxTarget)
	}
	req.SetPathValue("id", "template-1")
	if groupID != "" {
		req.SetPathValue("group_id", groupID)
	}
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions([]string{"job_phase:" + v.name}))
	return deps, v.call(deps).Handle(ctx, &view.ViewContext{Request: req})
}

func TestTransitionAction_ErrorClassification(t *testing.T) {
	st := func(found jobphasepb.PhaseApprovalStatus, mixed bool) error {
		return fmt.Errorf("job_phase approval: wrapped: %w", &espynaports.JobPhaseApprovalStateError{Found: found, Mixed: mixed})
	}
	wrap := func(e error) error { return fmt.Errorf("job_phase x: %w", e) }
	l := outcome_matrix.DefaultLabels().Approval.Errors

	cases := []struct {
		name    string
		err     error
		status  int
		label   string
		refresh bool
	}{
		{"mixed", st(0, true), 409, l.StateChanged, true},
		{"found for review", st(jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW, false), 409, l.AlreadyForReview, true},
		{"found verified", st(jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_VERIFIED, false), 409, l.AlreadyVerified, true},
		{"found published", st(jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_PUBLISHED, false), 409, l.AlreadyPublished, true},
		{"found in progress", st(jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_IN_PROGRESS, false), 409, l.AlreadyInProgress, true},
		{"found unknown", st(0, false), 409, l.StateChanged, true},
		{"frozen", wrap(espynaports.ErrJobPhaseApprovalFrozen), 409, l.Frozen, true},
		{"operator login", wrap(espynaports.ErrJobPhaseApprovalOperatorLogin), 403, l.UseStaffLogin, false},
		{"self verify", wrap(espynaports.ErrJobPhaseApprovalSelfVerify), 403, l.SelfVerify, false},
		{"not allowed", wrap(espynaports.ErrJobPhaseApprovalNotAllowed), 403, l.NotAllowed, false},
		{"unclassified", errors.New("pq: connection refused host=db secret"), 422, l.ActionFailed, false},
	}
	for _, v := range transitionVerbs() {
		for _, tc := range cases {
			for _, group := range []string{"", "grp-1"} {
				t.Run(v.name+"/"+tc.name+"/group="+group, func(t *testing.T) {
					_, res := runTransition(t, v, tc.err, group, "main-content")
					if res.StatusCode != tc.status {
						t.Fatalf("status = %d, want %d", res.StatusCode, tc.status)
					}
					if got := res.Headers["HX-Error-Message"]; got != tc.label {
						t.Fatalf("HX-Error-Message = %q, want %q", got, tc.label)
					}
					if strings.Contains(res.Headers["HX-Error-Message"], "secret") {
						t.Fatal("raw error leaked into the response")
					}
					loc, hasLoc := res.Headers["HX-Location"]
					trig, hasTrig := res.Headers["HX-Trigger"]
					if hasLoc != tc.refresh || hasTrig != tc.refresh {
						t.Fatalf("refresh headers present loc=%v trigger=%v, want %v", hasLoc, hasTrig, tc.refresh)
					}
					if !tc.refresh {
						return
					}
					var l struct{ Path, Target string }
					if err := json.Unmarshal([]byte(loc), &l, json.MatchCaseInsensitiveNames(true)); err != nil {
						t.Fatalf("HX-Location not JSON: %v", err)
					}
					wantPath := "/outcome-matrix/template-1"
					if group != "" {
						wantPath += "/subscription-group/grp-1"
					}
					if l.Path != wantPath || l.Target != "#main-content" {
						t.Fatalf("HX-Location = %+v, want path %q target #main-content", l, wantPath)
					}
					for _, r := range trig {
						if r > 0x7f {
							t.Fatalf("HX-Trigger must be ASCII (Latin-1 header decoding), got %q", trig)
						}
					}
					var tg map[string]map[string]string
					if err := json.Unmarshal([]byte(trig), &tg); err != nil {
						t.Fatalf("HX-Trigger not JSON: %v", err)
					}
					toast := tg["pyeza:toast"]
					if toast["message"] != tc.label || toast["state"] != "warning" || toast["testid"] != "pa-transition-notice" {
						t.Fatalf("toast = %+v", toast)
					}
				})
			}
		}
	}
}

func TestTransitionAction_EmptyLabelFallsBackToActionFailed(t *testing.T) {
	v := transitionVerbs()[1]
	deps := v.deps(fmt.Errorf("x: %w", espynaports.ErrJobPhaseApprovalNotAllowed))
	deps.Labels.Approval.Errors = outcome_matrix.ApprovalErrorLabels{ActionFailed: "generic"}
	form := url.Values{"job_template_phase_id": []string{"phase-1"}}
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "template-1")
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions([]string{"job_phase:verify"}))
	res := v.call(deps).Handle(ctx, &view.ViewContext{Request: req})
	if res.StatusCode != 403 || res.Headers["HX-Error-Message"] != "generic" {
		t.Fatalf("got %d %q, want 403 generic", res.StatusCode, res.Headers["HX-Error-Message"])
	}
}

func TestTransitionAction_SuccessUnchanged(t *testing.T) {
	for _, v := range transitionVerbs() {
		t.Run(v.name, func(t *testing.T) {
			_, res := runTransition(t, v, nil, "", "")
			if res.StatusCode != http.StatusOK || res.Headers["HX-Redirect"] != "/outcome-matrix/template-1" ||
				res.Headers["HX-Location"] != "" || res.Headers["HX-Trigger"] != "" || res.Headers["HX-Error-Message"] != "" {
				t.Fatalf("non-htmx success changed: %d %v", res.StatusCode, res.Headers)
			}
			_, res = runTransition(t, v, nil, "grp-1", "main-content")
			if res.StatusCode != http.StatusOK || res.Headers["HX-Redirect"] != "" ||
				!sameLocation(t, res.Headers["HX-Location"], "/outcome-matrix/template-1/subscription-group/grp-1", "#main-content") ||
				res.Headers["HX-Trigger"] != "" {
				t.Fatalf("htmx group success changed: %d %v", res.StatusCode, res.Headers)
			}
		})
	}
}

// sameLocation compares an HX-Location JSON value by content (map key order is
// not stable).
func sameLocation(t *testing.T, raw, path, target string) bool {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return false
	}
	return len(m) == 2 && m["path"] == path && m["target"] == target
}

// A non-ASCII label (the education labels carry an em dash) must reach the
// client intact. HX-Error-Message is sent as the label's exact UTF-8 bytes (pyeza
// ui-form-result.js re-decodes the Latin-1 view of those bytes as UTF-8); the
// HX-Trigger toast is ASCII-escaped JSON that decodes to the same label.
func TestTransitionAction_NonASCIILabelSurvivesHeaders(t *testing.T) {
	const label = "Already submitted — the sheet was refreshed"
	v := transitionVerbs()[0]
	deps := v.deps(fmt.Errorf("x: %w", &espynaports.JobPhaseApprovalStateError{Found: jobphasepb.PhaseApprovalStatus_PHASE_APPROVAL_STATUS_FOR_REVIEW}))
	deps.Labels.Approval.Errors.AlreadyForReview = label
	form := url.Values{"job_template_phase_id": []string{"phase-1"}}
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Target", "main-content")
	req.SetPathValue("id", "template-1")
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions([]string{"job_phase:submit"}))
	res := v.call(deps).Handle(ctx, &view.ViewContext{Request: req})

	rec := httptest.NewRecorder()
	for k, val := range res.Headers {
		rec.Header().Set(k, val)
	}
	rec.WriteHeader(res.StatusCode)
	if got := rec.Result().Header.Get("HX-Error-Message"); got != label || !utf8.ValidString(got) {
		t.Fatalf("HX-Error-Message = %q, want the exact UTF-8 label %q", got, label)
	}
	trig := rec.Result().Header.Get("HX-Trigger")
	for _, r := range trig {
		if r > 0x7f {
			t.Fatalf("HX-Trigger must be ASCII, got %q", trig)
		}
	}
	var tg map[string]map[string]string
	if err := json.Unmarshal([]byte(trig), &tg); err != nil || tg["pyeza:toast"]["message"] != label {
		t.Fatalf("toast message = %q (%v), want %q", tg["pyeza:toast"]["message"], err, label)
	}
}
