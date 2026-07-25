package outcome_matrix

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
)

// lister returns a stub ListJobTemplateSummaries closure whose rows pair the
// given template ids with the requested section.
func lister(templateIDs ...string) SummaryLister {
	return func(_ context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
		out := &summarypb.ListJobTemplateSummariesResponse{Success: true}
		for _, id := range templateIDs {
			out.Summaries = append(out.Summaries, &summarypb.JobTemplateSummary{
				JobTemplateId:         id,
				SubscriptionGroupId:   req.GetSubscriptionGroupId(),
				SubscriptionGroupName: "Grade 10 Tantalum (AY 2026-27)",
			})
		}
		return out, nil
	}
}

func TestResolveGroupScope(t *testing.T) {
	const tmpl = "61b151bb"
	const group = "019f82fb"

	tests := []struct {
		name       string
		groupID    string
		templateID string
		lister     SummaryLister
		wantOK     bool
		wantScoped bool
	}{
		{
			name:       "no group_id — template-scoped route is always allowed",
			templateID: tmpl,
			lister:     lister(tmpl),
			wantOK:     true,
			wantScoped: false,
		},
		{
			name:       "valid pair — section delivers the template",
			groupID:    group,
			templateID: tmpl,
			lister:     lister(tmpl),
			wantOK:     true,
			wantScoped: true,
		},
		{
			// The defect this guard exists for: a section from another academic
			// year shares members with the template, so the raw filter would
			// return a plausible PARTIAL roster instead of failing.
			name:       "foreign section — pairs with a different template only",
			groupID:    group,
			templateID: tmpl,
			lister:     lister("some-other-template"),
			wantOK:     false,
		},
		{
			name:       "no deliveries at all — section is empty or foreign-workspace",
			groupID:    group,
			templateID: tmpl,
			lister:     lister(),
			wantOK:     false,
		},
		{
			name:       "nil lister — cannot verify, so refuse (fail closed)",
			groupID:    group,
			templateID: tmpl,
			lister:     nil,
			wantOK:     false,
		},
		{
			name:       "lister errors — refuse rather than fall through",
			groupID:    group,
			templateID: tmpl,
			lister: func(context.Context, *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
				return nil, errors.New("boom")
			},
			wantOK: false,
		},
		{
			name:       "empty template id with a group_id — refuse",
			groupID:    group,
			templateID: "",
			lister:     lister(tmpl),
			wantOK:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/grade-sheet/x", nil)
			if tc.groupID != "" {
				r.SetPathValue("group_id", tc.groupID)
			}

			got, ok := ResolveGroupScope(context.Background(), r, tc.templateID, tc.lister)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !tc.wantOK {
				// A refused pair must carry no narrowing — a caller that ignored
				// ok must not silently get a usable scope.
				if got.Scoped() {
					t.Fatalf("refused pair still returned a scope: %+v", got)
				}
				return
			}
			if got.Scoped() != tc.wantScoped {
				t.Fatalf("Scoped() = %v, want %v", got.Scoped(), tc.wantScoped)
			}
			if tc.wantScoped {
				if got.GroupID != tc.groupID {
					t.Errorf("GroupID = %q, want %q", got.GroupID, tc.groupID)
				}
				if got.GroupName == "" {
					t.Error("GroupName is empty — the page title and header caption depend on it")
				}
			}
		})
	}
}

// A nil request must not panic — the template-scoped call sites pass whatever
// the view context holds.
func TestResolveGroupScopeNilRequest(t *testing.T) {
	got, ok := ResolveGroupScope(context.Background(), nil, "tmpl", lister("tmpl"))
	if !ok {
		t.Fatal("nil request should be treated as template-scoped, not refused")
	}
	if got.Scoped() {
		t.Fatalf("nil request produced a scope: %+v", got)
	}
}
