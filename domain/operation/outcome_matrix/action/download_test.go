package action

// download_test.go — the export-drawer GET view: period-option construction
// (incl. the zero-phase guard-rail) and live ?scope=/?hide= carry-through into
// the drawer's hidden inputs.

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

func TestBuildPeriodOptions(t *testing.T) {
	l := outcome_matrix.DefaultLabels()

	// Two coded phases → All, Semester 1, Semester 2, Final.
	resp := &matrixpb.GetOutcomeMatrixResponse{
		Phases: []*matrixpb.PhaseColumn{
			{JobTemplatePhaseId: "p2", Label: "Semester 2", SequenceOrder: 2, Code: "s2"},
			{JobTemplatePhaseId: "p1", Label: "Semester 1", SequenceOrder: 1, Code: "s1"},
		},
	}
	opts := buildPeriodOptions(l, resp)
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
		opts := buildPeriodOptions(l, resp)
		if len(opts) != 2 || opts[0].Value != "" || opts[1].Value != "final" {
			t.Fatalf("zero/codeless phases: options = %+v, want [all, final]", opts)
		}
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
