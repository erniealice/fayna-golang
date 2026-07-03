package list

import (
	"context"
	"log"
	"strconv"
	"strings"

	outcome_matrix "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	outcomecriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// Scope-toggle permission (same as grade_sheet.go's superAdminScopeCode). A
// teacher-only principal never gets the widened "all clients" view.
const (
	scopeEntity = "workspace"
	scopeAction = "list"
)

// PageViewDeps holds view dependencies for the outcome matrix page.
type PageViewDeps struct {
	Routes       outcome_matrix.Routes
	Labels       outcome_matrix.Labels
	CommonLabels pyeza.CommonLabels

	// GetOutcomeMatrix — the new espyna use case (typed against the generated
	// esqyma request/response). Wired via the module Deps, never raw SQL.
	GetOutcomeMatrix func(ctx context.Context, req *matrixpb.GetOutcomeMatrixRequest) (*matrixpb.GetOutcomeMatrixResponse, error)

	// ResolveStaff maps the acting session user → staff_id ("" == no staff
	// identity, fail-closed). Wired via the module Deps (grade_sheet.go's
	// resolveStaff precedent), never raw SQL from the view.
	ResolveStaff func(ctx context.Context) (string, error)
}

// PageData holds the data for the outcome matrix page. The Grid field is named
// "Grid" so the render pipeline's reflection injector (parallel to TableConfig's
// "Table" branch) populates Grid.Nonce / Grid.WorkspaceID on render.
type PageData struct {
	types.PageData
	Grid   *types.CellGridConfig
	Labels outcome_matrix.Labels

	SubjectName  string
	ScopeActive  string // "mine" | "all"
	ScopeMineURL string
	ScopeAllURL  string
	ShowScopeAll bool
}

// NewView creates the outcome matrix GET view.
func NewView(deps *PageViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("task_outcome", "read") {
			return view.Forbidden("task_outcome:read")
		}

		templateID := viewCtx.Request.PathValue("id")

		// Scope resolution. Explicit ?scope=all|mine is honored as requested. With
		// NO explicit scope, an operator holding workspace:list defaults to ALL: a
		// non-staff MINE view is zero rows by design, so an admin landing on the
		// page should see the roster. The server-side re-check below (effectiveAll)
		// still gates ALL on workspace:list, unchanged.
		canSeeAll := perms.Can(scopeEntity, scopeAction)
		scopeParam := viewCtx.Request.URL.Query().Get("scope")
		requestedAll := scopeParam == "all"
		if scopeParam == "" {
			requestedAll = canSeeAll
		}
		effectiveAll := requestedAll && canSeeAll

		l := deps.Labels

		var resp *matrixpb.GetOutcomeMatrixResponse
		if templateID != "" && deps.GetOutcomeMatrix != nil {
			scope := matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE
			if effectiveAll {
				scope = matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL
			}

			var err error
			resp, err = deps.GetOutcomeMatrix(ctx, &matrixpb.GetOutcomeMatrixRequest{
				JobTemplateId: templateID,
				Scope:         scope,
			})
			if err != nil {
				log.Printf("Failed to load outcome matrix for template %s: %v", templateID, err)
				resp = nil
			}
		}

		grid := buildGrid(deps, perms, resp, effectiveAll, templateID, viewCtx)

		subjectName := ""
		if resp != nil {
			subjectName = resp.GetJobTemplateName()
		}

		scopeActive := "mine"
		if effectiveAll {
			scopeActive = "all"
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:    viewCtx.CacheVersion,
				Title:           l.Page.Title,
				ContentTemplate: "outcome-matrix-content",
				CurrentPath:     viewCtx.CurrentPath,
				ActiveNav:       deps.Routes.ActiveNav,
				ActiveSubNav:    deps.Routes.ActiveSubNav,
				HeaderTitle:     l.Page.Title,
				HeaderIcon:      "icon-grid",
				CommonLabels:    deps.CommonLabels,
			},
			Grid:         grid,
			Labels:       l,
			SubjectName:  subjectName,
			ScopeActive:  scopeActive,
			ScopeMineURL: route.ResolveURL(deps.Routes.MatrixURL, "id", templateID) + "?scope=mine",
			ScopeAllURL:  route.ResolveURL(deps.Routes.MatrixURL, "id", templateID) + "?scope=all",
			ShowScopeAll: canSeeAll,
		}

		return view.OK("outcome-matrix", pageData)
	})
}

// buildGrid converts the proto response into a *types.CellGridConfig. The acting
// staff read-only gating is applied here (view layer, per view-scope.md §4).
func buildGrid(
	deps *PageViewDeps,
	perms *types.UserPermissions,
	resp *matrixpb.GetOutcomeMatrixResponse,
	effectiveAll bool,
	templateID string,
	viewCtx *view.ViewContext,
) *types.CellGridConfig {
	l := deps.Labels

	scopeStr := "mine"
	if effectiveAll {
		scopeStr = "all"
	}

	cfg := &types.CellGridConfig{
		ID:               "outcome-matrix-grid",
		Caption:          l.Page.Title,
		FreezeFirstCol:   true,
		FreezeHeaderRows: 3,
		SaveURL:          route.ResolveURL(deps.Routes.RecordURL, "id", templateID),
		SaveMode:         "batch",
		JobTemplateID:    templateID,
		Scope:            scopeStr,
		SaveDisabled:     !perms.Can("task_outcome", "create"),
		Labels:           l.Grid.CellGridLabels,
		CacheVersion:     viewCtx.CacheVersion,
	}

	if resp == nil {
		return cfg
	}

	// Acting staff for the read-only gate.
	var actingStaff string
	if deps.ResolveStaff != nil {
		actingStaff, _ = deps.ResolveStaff(viewCtx.Request.Context())
	}

	cfg.Columns = buildColumns(resp.GetPhases())
	cfg.Rows = buildRows(resp.GetRows(), actingStaff, l.Grid.ReadOnlyTooltip)
	return cfg
}

// buildColumns maps the proto phase→task→criterion tree into CellGridLevel1/2/3.
func buildColumns(phases []*matrixpb.PhaseColumn) []types.CellGridLevel1 {
	columns := make([]types.CellGridLevel1, 0, len(phases))
	for _, ph := range phases {
		l1 := types.CellGridLevel1{
			Key:   ph.GetJobTemplatePhaseId(),
			Label: ph.GetLabel(),
		}
		for _, tk := range ph.GetTasks() {
			l2 := types.CellGridLevel2{
				Key:   tk.GetJobTemplateTaskId(),
				Label: tk.GetLabel(),
			}
			for _, cr := range tk.GetCriteria() {
				l2.Level3 = append(l2.Level3, types.CellGridLevel3{
					ColumnKey: cr.GetColumnKey(),
					Label:     criterionLabel(cr),
					CellInput: buildCellInput(cr.GetCriteria()),
				})
			}
			l1.Level2 = append(l1.Level2, l2)
		}
		columns = append(columns, l1)
	}
	return columns
}

// buildRows maps the proto rows into CellGridRows with the read-only gate applied.
func buildRows(rows []*matrixpb.OutcomeRow, actingStaff, readOnlyTooltip string) []types.CellGridRow {
	out := make([]types.CellGridRow, 0, len(rows))
	for _, r := range rows {
		clientID := r.GetClientId()
		gr := types.CellGridRow{
			ID:     clientID,
			Label:  rowLabel(r),
			Cells:  make(map[string]types.CellGridCell, len(r.GetCells())),
			TestID: "om-row-" + short(clientID),
		}
		for colKey, cell := range r.GetCells() {
			// Read-only when recorded by a different staff member. Honor the
			// backend editable flag too (job not spawned → not editable).
			readOnly := cell.GetRecordedBy() != "" && cell.GetRecordedBy() != actingStaff
			gr.Cells[colKey] = types.CellGridCell{
				OutcomeID:       cell.GetOutcomeId(),
				JobTaskID:       cell.GetJobTaskId(),
				CriteriaID:      criteriaIDFromColumnKey(colKey),
				Value:           cellValue(cell),
				Editable:        cell.GetEditable(),
				ReadOnly:        readOnly,
				ReadOnlyTooltip: readOnlyTooltip,
				TestID:          cellTestID(clientID, colKey, readOnly),
			}
		}
		out = append(out, gr)
	}
	return out
}

// buildCellInput derives a CellInputDescriptor from the criterion's enforcement
// contract (the embedded outcome_criteria entity).
func buildCellInput(oc *outcomecriteriapb.OutcomeCriteria) types.CellInputDescriptor {
	d := types.CellInputDescriptor{Type: "text"}
	if oc == nil {
		return d
	}
	d.Type = criteriaTypeString(oc.GetCriteriaType())

	if oc.MinScore != nil {
		v := float64(oc.GetMinScore())
		d.Min = &v
	}
	if oc.MaxScore != nil {
		v := float64(oc.GetMaxScore())
		d.Max = &v
	}
	if oc.ScoreIncrement != nil {
		v := oc.GetScoreIncrement()
		d.Step = &v
	}
	d.Decimals = int(oc.GetDecimalPlaces())
	d.Unit = oc.GetUnit()

	d.PassLabel = oc.GetPassLabel()
	d.FailLabel = oc.GetFailLabel()

	for _, opt := range oc.GetAllowedDeterminations() {
		d.Options = append(d.Options, types.SelectOption{Value: opt, Label: opt})
	}

	if oc.MaxTextLength != nil {
		v := int(oc.GetMaxTextLength())
		d.MaxLength = &v
	}
	d.Prompt = oc.GetTextPrompt()
	return d
}

// criteriaTypeString maps the criteria_type enum onto the pyeza descriptor's
// string vocabulary. MULTI_CHECK maps through as "multi_check"; the component
// renders "—" for it (OQ-6: multi_check deferred to V2).
func criteriaTypeString(t enums.CriteriaType) string {
	switch t {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE, enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return "numeric"
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		return "pass_fail"
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return "categorical"
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		return "text"
	case enums.CriteriaType_CRITERIA_TYPE_MULTI_CHECK:
		return "multi_check"
	default:
		return "text"
	}
}

func criterionLabel(cr *matrixpb.CriterionColumn) string {
	if oc := cr.GetCriteria(); oc != nil && oc.GetName() != "" {
		return oc.GetName()
	}
	return cr.GetColumnKey()
}

func rowLabel(r *matrixpb.OutcomeRow) string {
	if lbl := r.GetClientLabel(); lbl != "" {
		return lbl
	}
	return short(r.GetClientId())
}

// cellValue serialises the cell's typed value into a display string.
func cellValue(c *matrixpb.OutcomeCell) string {
	switch {
	case c.NumericValue != nil:
		return strconv.FormatFloat(c.GetNumericValue(), 'f', -1, 64)
	case c.TextValue != nil:
		return c.GetTextValue()
	case c.CategoricalValue != nil:
		return c.GetCategoricalValue()
	case c.PassFailValue != nil:
		if c.GetPassFailValue() {
			return "true"
		}
		return "false"
	}
	return ""
}

// criteriaIDFromColumnKey extracts the outcome_criteria_id from a
// "{job_template_task_id}:{outcome_criteria_id}" column key.
func criteriaIDFromColumnKey(k string) string {
	if i := strings.LastIndex(k, ":"); i >= 0 {
		return k[i+1:]
	}
	return k
}

func cellTestID(clientID, colKey string, readOnly bool) string {
	base := "om-cell-"
	if readOnly {
		base = "om-cell-ro-"
	}
	return base + short(clientID) + "-" + slug(colKey)
}

// short truncates an opaque id for a testid suffix.
func short(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// slug lowercases and keeps [a-z0-9]; space/'-'/':' collapse to '-'. Mirrors
// grade_sheet.go's slug() helper (extended with ':' for the composite column key).
func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == ':':
			b.WriteByte('-')
		}
	}
	return b.String()
}
