package action

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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
