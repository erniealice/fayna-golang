package action

// download_test.go — the export-drawer GET view: period-option construction
// (incl. the zero-phase guard-rail), the ?period= pre-selection the column-header
// triggers rely on, the section-scoped form action, and live ?scope=/?hide=
// carry-through into the drawer's hidden inputs.

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	summarypb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/job_template_summary"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// selectedPeriod returns the Value of the single pre-selected option, or a
// sentinel when zero / more than one is marked (either would be a bug: the
// browser silently resolves both to something plausible).
func selectedPeriod(t *testing.T, opts []types.SelectOption) string {
	t.Helper()
	sel := []string{}
	for _, o := range opts {
		if o.Selected {
			sel = append(sel, o.Value)
		}
	}
	if len(sel) != 1 {
		return "<none/multiple>"
	}
	return sel[0]
}

func TestBuildPeriodOptions(t *testing.T) {
	l := outcome_matrix.DefaultLabels()

	// Two coded phases → All, Semester 1, Semester 2, Final.
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			{JobTemplatePhaseId: "p2", Label: "Semester 2", SequenceOrder: 2, Code: "s2"},
			{JobTemplatePhaseId: "p1", Label: "Semester 1", SequenceOrder: 1, Code: "s1"},
		},
	}
	opts := buildPeriodOptions(l, resp, "")
	gotVals := make([]string, len(opts))
	for i, o := range opts {
		gotVals[i] = o.Value
	}
	want := []string{"", "s1", "s2", "final"} // All + seq-ordered phases + Final
	if len(gotVals) != len(want) {
		t.Fatalf("period options = %v, want %v", gotVals, want)
	}
	for i := range want {
		if gotVals[i] != want[i] {
			t.Fatalf("period option[%d] = %q, want %q (full: %v)", i, gotVals[i], want[i], gotVals)
		}
	}
	// Labels: phase options carry the DB phase label; reserved options are lyngua.
	if opts[0].Label != l.Export.PeriodAll || opts[3].Label != l.Export.PeriodFinal {
		t.Errorf("reserved option labels wrong: %q / %q", opts[0].Label, opts[3].Label)
	}
	if opts[1].Label != "Semester 1" {
		t.Errorf("phase option label = %q, want DB label Semester 1", opts[1].Label)
	}
}

func TestBuildPeriodOptions_ZeroPhase(t *testing.T) {
	l := outcome_matrix.DefaultLabels()
	// Guard-rail: a zero-phase / codeless template yields only All + Final.
	for _, resp := range []*matrixpb.GetOutcomeMatrixResponse{
		nil,
		{},
		{Phases: []*matrixpb.PhaseColumn{{JobTemplatePhaseId: "p1", Label: "Nameless", SequenceOrder: 1}}}, // no code
	} {
		opts := buildPeriodOptions(l, resp, "")
		if len(opts) != 2 || opts[0].Value != "" || opts[1].Value != "final" {
			t.Fatalf("zero/codeless phases: options = %+v, want [all, final]", opts)
		}
	}
}

// TestBuildPeriodOptions_Preselect pins the contract the per-column header
// triggers depend on: the ?period= token they carry is a phase CODE (or the
// reserved "final"), and exactly that option opens selected. An unrecognized
// token must select NOTHING so the browser falls back to "All periods" — the
// same fail-safe direction as export.go's periodKnown guard.
func TestBuildPeriodOptions_Preselect(t *testing.T) {
	l := outcome_matrix.DefaultLabels()
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			{JobTemplatePhaseId: "p1", Label: "Semester 1", SequenceOrder: 1, Code: "s1"},
			{JobTemplatePhaseId: "p2", Label: "Semester 2", SequenceOrder: 2, Code: "s2"},
			{JobTemplatePhaseId: "p3", Label: "Nameless", SequenceOrder: 3}, // no code ⇒ no option
		},
	}
	tests := []struct {
		name, preset, want string
	}{
		{"toolbar trigger — no token selects All", "", ""},
		{"phase column", "s1", "s1"},
		{"other phase column", "s2", "s2"},
		{"final column", "final", "final"},
		{"phase ID is NOT the token — ?hide= is the id-keyed axis", "p1", "<none/multiple>"},
		{"codeless phase offers nothing to select", "p3", "<none/multiple>"},
		{"stale/unknown code falls back to the browser default", "s9", "<none/multiple>"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := selectedPeriod(t, buildPeriodOptions(l, resp, tc.preset))
			if got != tc.want {
				t.Errorf("preset %q selected %q, want %q", tc.preset, got, tc.want)
			}
		})
	}
}

func TestNewDownloadDrawer_CarriesViewState(t *testing.T) {
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			{JobTemplatePhaseId: "p1", Label: "Semester 1", SequenceOrder: 1, Code: "s1"},
			{JobTemplatePhaseId: "p2", Label: "Semester 2", SequenceOrder: 2, Code: "s2"},
		},
	}
	deps := &DrawerDeps{
		Routes: outcome_matrix.DefaultRoutes(),
		Labels: outcome_matrix.DefaultLabels(),
		GetOutcomeMatrix: func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error) {
			return resp, nil
		},
	}
	v := NewDownloadDrawer(deps)

	req := httptest.NewRequest("GET", "/action/outcome-matrix/tmpl-1/download?scope=all&hide=p2", nil)
	req.SetPathValue("id", "tmpl-1")
	ctx := view.WithUserPermissions(context.Background(),
		types.NewUserPermissions([]string{"task_outcome:read", "workspace:list"}))

	result := v.Handle(ctx, &view.ViewContext{Request: req})
	if result.Template != "outcome-matrix-download-drawer-form" {
		t.Fatalf("template = %q", result.Template)
	}
	data, ok := result.Data.(*DrawerData)
	if !ok {
		t.Fatalf("data type = %T, want *DrawerData", result.Data)
	}
	// Live view state carried into the hidden inputs.
	if data.Scope != "all" {
		t.Errorf("Scope = %q, want all (workspace:list granted)", data.Scope)
	}
	if data.Hide != "p2" {
		t.Errorf("Hide = %q, want p2", data.Hide)
	}
	// The native GET form posts to the resolved ExportURL (bare — the form
	// fields supply the query).
	if data.ExportAction != "/outcome-matrix/tmpl-1/export" {
		t.Errorf("ExportAction = %q", data.ExportAction)
	}
	// Period options reflect the response phases.
	if len(data.PeriodOptions) != 4 {
		t.Errorf("period options = %d, want 4 (all+s1+s2+final)", len(data.PeriodOptions))
	}
	if len(data.FormatOptions) != 2 || data.FormatOptions[0].Value != "csv" {
		t.Errorf("format options wrong: %+v", data.FormatOptions)
	}
}

// TestNewDownloadDrawer_SectionScoped pins the two things the column-header
// triggers changed about this view on a section page:
//   - the ?period= token opens PRE-SELECTED (end to end, not just in the builder);
//   - the native GET form submits to the SECTION-scoped export.
//
// The second was a live defect: the group route already reached this view (both
// paths are registered to it), but ExportAction was resolved from the
// template-scoped ExportURL, so a download started from a 29-student section
// sheet would have produced every student under the template.
func TestNewDownloadDrawer_SectionScoped(t *testing.T) {
	const tmpl, group = "tmpl-1", "grp-9"
	deps := &DrawerDeps{
		Routes: outcome_matrix.DefaultRoutes(),
		Labels: outcome_matrix.DefaultLabels(),
		GetOutcomeMatrix: func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error) {
			if req.GetSubscriptionGroupId() != group {
				t.Errorf("matrix read section = %q, want %q", req.GetSubscriptionGroupId(), group)
			}
			return &matrixpb.GetOutcomeMatrixResponse{
				Phases: []*matrixpb.PhaseColumn{
					{JobTemplatePhaseId: "p1", Label: "Semester 1", SequenceOrder: 1, Code: "s1"},
					{JobTemplatePhaseId: "p2", Label: "Semester 2", SequenceOrder: 2, Code: "s2"},
				},
			}, nil
		},
		// The (template, section) pair guard — fail-closed, so the stub must pair.
		ListJobTemplateSummaries: func(_ context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
			return &summarypb.ListJobTemplateSummariesResponse{
				Success: true,
				Summaries: []*summarypb.JobTemplateSummary{{
					JobTemplateId:       tmpl,
					SubscriptionGroupId: req.GetSubscriptionGroupId(),
				}},
			}, nil
		},
	}

	req := httptest.NewRequest("GET",
		"/action/outcome-matrix/"+tmpl+"/subscription-group/"+group+"/download?scope=all&period=s2", nil)
	req.SetPathValue("id", tmpl)
	req.SetPathValue("group_id", group)
	ctx := view.WithUserPermissions(context.Background(),
		types.NewUserPermissions([]string{"task_outcome:read", "workspace:list"}))

	result := NewDownloadDrawer(deps).Handle(ctx, &view.ViewContext{Request: req})
	data, ok := result.Data.(*DrawerData)
	if !ok {
		t.Fatalf("data type = %T (result %+v)", result.Data, result)
	}
	wantAction := "/outcome-matrix/" + tmpl + "/subscription-group/" + group + "/export"
	if data.ExportAction != wantAction {
		t.Errorf("ExportAction = %q, want %q (section download must not widen to the template)",
			data.ExportAction, wantAction)
	}
	if got := selectedPeriod(t, data.PeriodOptions); got != "s2" {
		t.Errorf("pre-selected period = %q, want s2", got)
	}
}

// TestNewDownloadDrawer_ForeignSection pins the pair guard still fails CLOSED on
// this route — a section that does not deliver the template 404s rather than
// rendering a drawer whose export would slice a foreign cohort.
func TestNewDownloadDrawer_ForeignSection(t *testing.T) {
	deps := &DrawerDeps{
		Routes: outcome_matrix.DefaultRoutes(),
		Labels: outcome_matrix.DefaultLabels(),
		GetOutcomeMatrix: func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error) {
			t.Fatal("matrix must not be read for an unpaired (template, section)")
			return nil, nil
		},
		ListJobTemplateSummaries: func(_ context.Context, req *summarypb.ListJobTemplateSummariesRequest) (*summarypb.ListJobTemplateSummariesResponse, error) {
			return &summarypb.ListJobTemplateSummariesResponse{
				Success:   true,
				Summaries: []*summarypb.JobTemplateSummary{{JobTemplateId: "other-tmpl"}},
			}, nil
		},
	}
	req := httptest.NewRequest("GET", "/action/outcome-matrix/tmpl-1/subscription-group/grp-9/download", nil)
	req.SetPathValue("id", "tmpl-1")
	req.SetPathValue("group_id", "grp-9")
	ctx := view.WithUserPermissions(context.Background(),
		types.NewUserPermissions([]string{"task_outcome:read"}))

	result := NewDownloadDrawer(deps).Handle(ctx, &view.ViewContext{Request: req})
	if result.StatusCode != 404 {
		t.Fatalf("foreign section: status = %d, want 404 (%+v)", result.StatusCode, result)
	}
}

func TestNewDownloadDrawer_Forbidden(t *testing.T) {
	v := NewDownloadDrawer(&DrawerDeps{Routes: outcome_matrix.DefaultRoutes(), Labels: outcome_matrix.DefaultLabels()})
	req := httptest.NewRequest("GET", "/action/outcome-matrix/tmpl-1/download", nil)
	req.SetPathValue("id", "tmpl-1")
	// No task_outcome:read → fail-closed.
	ctx := view.WithUserPermissions(context.Background(), types.NewEmptyUserPermissions())
	result := v.Handle(ctx, &view.ViewContext{Request: req})
	if result.StatusCode != 403 && result.Error == nil {
		t.Fatalf("expected forbidden, got %+v", result)
	}
}
