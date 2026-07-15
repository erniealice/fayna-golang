package list

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
)

// TestListContentTemplateDispatch (T9) pins the boosted-nav dispatch invariant:
// the ViewAdapter derives the content-partial name as {result.Template}-content,
// so PageData.ContentTemplate MUST equal result.Template + "-content". A drift
// here reproduces the "vanishing tabs" bug (the full-page and boosted-nav render
// paths desync). templates_golden pins the template NAMES; nothing pinned the
// dispatch relationship until now.
func TestListContentTemplateDispatch(t *testing.T) {
	v := NewView(&ListViewDeps{
		// Non-nil so renderFlat (the zero-Options path) does not nil-deref; an
		// empty response renders the flat empty table.
		ListJobOutcomeSummarys: func(context.Context, *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error) {
			return &jobsumpb.ListJobOutcomeSummarysResponse{}, nil
		},
	})

	req := httptest.NewRequest("GET", "/outcomes/summaries", nil)
	ctx := view.WithUserPermissions(req.Context(), types.NewUserPermissions([]string{"job_outcome_summary:list"}))
	res := v.Handle(ctx, &view.ViewContext{Request: req, CurrentPath: "/outcomes/summaries", CacheVersion: "test"})

	if res.Template != "outcome-summary-list" {
		t.Fatalf("Template = %q, want %q", res.Template, "outcome-summary-list")
	}
	pd, ok := res.Data.(*PageData)
	if !ok {
		t.Fatalf("Data type = %T, want *PageData", res.Data)
	}
	if want := res.Template + "-content"; pd.ContentTemplate != want {
		t.Fatalf("ContentTemplate = %q, want %q (boosted-nav dispatch invariant)", pd.ContentTemplate, want)
	}
}
