// export.go — sheet-level CSV download for the outcome matrix. Builds the SAME
// grid the HTML view renders — same scope resolution, same ?hide= pruning, same
// row order and group bands (buildGrid end to end) — so the CSV and the page
// can never disagree ("export what you see", plan 20260720 Q3). Registered as a
// raw GET handler wrapped by the ViewAdapter (WrapHandler), so
// view.GetUserPermissions sees the same RBAC context as the HTML view; the
// route lives OUTSIDE /action/* (safe method — no CSRF/action-guard applies,
// mirroring outcome_summary/section/export.go).
package list

import (
	"context"
	"encoding/csv"
	"log"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	outcome_matrix "github.com/erniealice/fayna-golang/domain/operation/outcome_matrix"
	"github.com/erniealice/pyeza-golang/view"

	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

// NewExportHandler creates the outcome-matrix CSV download handler.
func NewExportHandler(deps *PageViewDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		perms := view.GetUserPermissions(ctx)
		// Same Layer-3 gate as the HTML view (NewView) — fail-closed.
		if !perms.Can("task_outcome", "read") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		templateID := strings.TrimSpace(r.PathValue("id"))
		if templateID == "" || deps.GetOutcomeMatrix == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// format: "" / "csv" (default) → CSV; "pdf" → the P5 render (503 stub
		// this wave); anything else → 400 (report-card handler idiom).
		format := strings.TrimSpace(r.URL.Query().Get("format"))
		switch format {
		case "", "csv", "pdf":
		default:
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		// period: "" (all) / "s1" / "s2" (a phase code) / "final" (composite).
		// Validated against the response phases below (unknown → 400).
		period := strings.TrimSpace(r.URL.Query().Get("period"))

		// Scope resolution — byte-identical to NewView (incl. the widened
		// admin default and the server-side workspace:list re-check).
		canSeeAll := perms.Can(scopeEntity, scopeAction)
		scopeParam := r.URL.Query().Get("scope")
		requestedAll := scopeParam == "all"
		if scopeParam == "" {
			requestedAll = canSeeAll
		}
		effectiveAll := requestedAll && canSeeAll

		scope := matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_MINE
		if effectiveAll {
			scope = matrixpb.OutcomeMatrixScope_OUTCOME_MATRIX_SCOPE_ALL
		}
		resp, err := deps.GetOutcomeMatrix(ctx, &matrixpb.GetOutcomeMatrixRequest{
			JobTemplateId: templateID,
			Scope:         scope,
		})
		if err != nil || resp == nil {
			// Same fail-closed response for foreign and missing ids (no leak).
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Validate the period token against the RESPONSE phase codes plus the
		// reserved "final" (and "" = all). Unknown → 400.
		if !periodKnown(period, resp.GetPhases()) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// format=pdf: the composite PDF render is P5. Fail loud with the lyngua
		// "no template configured" body (503) — no embedded fallback in v1.
		// P5 wires the render pipeline (resolve binding → storage bytes → engine).
		if format == "pdf" {
			http.Error(w, deps.Labels.Export.NoTemplateError, http.StatusServiceUnavailable)
			return
		}

		// period=final: the roster composite CSV (student · per-phase final ·
		// year final) — bypasses the grid entirely (a summary read, not a column
		// prune; the year-final has no matrix column).
		if period == "final" {
			// Thread the ALREADY-RESOLVED scope (the same MINE/ALL the grid + the
			// GetOutcomeMatrix call above use) into the roster read — MINE stays MINE
			// so a non-admin never receives the full-workspace year-final roster.
			writeFinalCompositeCSV(ctx, w, deps, resp.GetJobTemplateName(), templateID, scope)
			return
		}

		// period "" / s1 / s2: the grid CSV. s1/s2 prune the OTHER phase(s) by
		// unioning their hide-tokens onto the user ?hide= set (the prune is
		// intentional — added AFTER resolveHidden so its all-hidden fail-safe
		// cannot collapse the period selection back to the full sheet).
		hidden := resolveHidden(r.URL.Query().Get("hide"), resp)
		if toks := periodHideTokens(period, resp.GetPhases()); len(toks) > 0 {
			if hidden == nil {
				hidden = map[string]bool{}
			}
			for _, t := range toks {
				hidden[t] = true
			}
		}
		grid := buildGrid(ctx, deps, perms, resp, effectiveAll, templateID, hidden,
			&view.ViewContext{Request: r})
		if grid.LeafColumnCount() == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		filename := exportFilename(deps.Labels, resp.GetJobTemplateName(), templateID)
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`.csv"`)

		cw := csv.NewWriter(w)

		// One flattened header row: "L1 — L2 — L3" per leaf (the same
		// flattening idiom table-export.js uses for its one grouping level,
		// extended to the grid's two). First column = the roster label.
		header := []string{csvSafe(deps.Labels.Grid.ClientColumn)}
		for _, l1 := range grid.Columns {
			for _, l2 := range l1.Level2 {
				for _, l3 := range l2.Level3 {
					header = append(header, csvSafe(l1.Label+" — "+l2.Label+" — "+l3.Label))
				}
			}
		}
		if err := cw.Write(header); err != nil {
			log.Printf("outcome matrix export: write header: %v", err)
			return
		}

		// Rows in the page's final display order (applyRowOptions has settled
		// it inside buildGrid). A GroupLabel band renders as its own
		// single-cell row, mirroring the on-screen band separators. Cell text
		// is the recorded VALUE only (the secondary descriptor TextValue is a
		// long free-text sentence — deliberately out of the score export).
		record := make([]string, 0, len(header))
		for _, row := range grid.Rows {
			if row.GroupLabel != "" {
				// Band rows are padded to the full header width: CSV has no
				// colspan, and strict readers (encoding/csv included) reject
				// records whose field count differs from the header's.
				band := make([]string, len(header))
				band[0] = csvSafe(row.GroupLabel)
				if err := cw.Write(band); err != nil {
					log.Printf("outcome matrix export: write band: %v", err)
					return
				}
			}
			record = record[:0]
			record = append(record, csvSafe(row.Label))
			for _, l1 := range grid.Columns {
				for _, l2 := range l1.Level2 {
					for _, l3 := range l2.Level3 {
						record = append(record, csvSafe(row.Cells[l3.ColumnKey].Value))
					}
				}
			}
			if err := cw.Write(record); err != nil {
				log.Printf("outcome matrix export: write row: %v", err)
				return
			}
		}
		cw.Flush()
	}
}

// periodKnown reports whether the period token is recognized: "" (all periods),
// "final" (the reserved composite), or one of the response phases' non-empty
// codes. Pure over the response tree — the export handler 400s on anything else.
func periodKnown(period string, phases []*matrixpb.PhaseColumn) bool {
	if period == "" || period == "final" {
		return true
	}
	for _, ph := range phases {
		if c := ph.GetCode(); c != "" && c == period {
			return true
		}
	}
	return false
}

// periodHideTokens returns the job_template_phase_ids to hide for a semester
// period export — every phase whose code is NOT the selected period (so only the
// selected semester's subtree survives). Empty for "" (all) and "final" (no
// column prune). Pure over the response phase tree; the handler unions the result
// onto the user ?hide= set.
func periodHideTokens(period string, phases []*matrixpb.PhaseColumn) []string {
	if period == "" || period == "final" {
		return nil
	}
	var toks []string
	for _, ph := range phases {
		if ph.GetCode() != period {
			toks = append(toks, ph.GetJobTemplatePhaseId())
		}
	}
	return toks
}

// rosterPhaseCol is one composite-CSV phase column (a job_template_phase, in
// sequence order, with its display label).
type rosterPhaseCol struct {
	id    string
	label string
	seq   int32
}

// rosterPhaseColumns derives the canonical, sequence-ordered phase column set
// from a roster response by unioning every row's phase entries (robust to a rare
// student with a different phase subset). Deterministic: seq then id.
func rosterPhaseColumns(rows []*matrixpb.OutcomeSummaryRosterRow) []rosterPhaseCol {
	seen := map[string]rosterPhaseCol{}
	for _, row := range rows {
		for _, pe := range row.GetPhases() {
			id := pe.GetJobTemplatePhaseId()
			if id == "" {
				continue
			}
			if _, ok := seen[id]; !ok {
				seen[id] = rosterPhaseCol{id: id, label: pe.GetLabel(), seq: pe.GetSequenceOrder()}
			}
		}
	}
	cols := make([]rosterPhaseCol, 0, len(seen))
	for _, c := range seen {
		cols = append(cols, c)
	}
	sort.SliceStable(cols, func(i, j int) bool {
		if cols[i].seq != cols[j].seq {
			return cols[i].seq < cols[j].seq
		}
		return cols[i].id < cols[j].id
	})
	return cols
}

// writeFinalCompositeCSV streams the roster composite CSV (Q4 rec A): the
// MMIS-parity gradesheet schema — student · <phase> Final (per phase, sequence
// order) · Final. Every stored value is read VERBATIM from the roster read (D8)
// and every cell passes through csvSafe. Student labels are hydrated the same way
// the grid resolves them (fetchClientNames); a zero-row roster (incl. a foreign
// template id, workspace-scoped to empty, OR a MINE-scoped non-staff caller the
// adapter fails closed) 404s — never an empty CSV. The scope is the SAME resolved
// MINE/ALL the grid CSV path uses, so the composite mirrors the grid's row set.
func writeFinalCompositeCSV(ctx context.Context, w http.ResponseWriter, deps *PageViewDeps, subjectName, templateID string, scope matrixpb.OutcomeMatrixScope) {
	if deps.GetOutcomeSummaryRoster == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	roster, err := deps.GetOutcomeSummaryRoster(ctx, &matrixpb.GetOutcomeSummaryRosterRequest{
		JobTemplateId: templateID,
		Scope:         scope,
	})
	if err != nil || roster == nil || len(roster.GetRows()) == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	cols := rosterPhaseColumns(roster.GetRows())

	// Roster name hydration (same closure + chunking the grid uses).
	ids := make([]string, 0, len(roster.GetRows()))
	for _, row := range roster.GetRows() {
		ids = append(ids, row.GetClientId())
	}
	names := fetchClientNames(ctx, deps, ids)

	filename := exportFilename(deps.Labels, subjectName, templateID) + "-final"
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`.csv"`)

	cw := csv.NewWriter(w)
	finalLabel := deps.Labels.Export.PeriodFinal

	// Header: roster label · "<phase> Final" per phase · Final.
	header := make([]string, 0, len(cols)+2)
	header = append(header, csvSafe(deps.Labels.Grid.ClientColumn))
	for _, c := range cols {
		header = append(header, csvSafe(c.label+" "+finalLabel))
	}
	header = append(header, csvSafe(finalLabel))
	if err := cw.Write(header); err != nil {
		log.Printf("outcome matrix composite export: write header: %v", err)
		return
	}

	record := make([]string, 0, len(header))
	for _, row := range roster.GetRows() {
		byPhase := make(map[string]string, len(row.GetPhases()))
		for _, pe := range row.GetPhases() {
			byPhase[pe.GetJobTemplatePhaseId()] = pe.GetScaledLabel()
		}
		record = record[:0]
		record = append(record, csvSafe(rosterLabel(row, names)))
		for _, c := range cols {
			record = append(record, csvSafe(byPhase[c.id]))
		}
		record = append(record, csvSafe(row.GetYearFinalLabel()))
		if err := cw.Write(record); err != nil {
			log.Printf("outcome matrix composite export: write row: %v", err)
			return
		}
	}
	cw.Flush()
}

// rosterLabel resolves the display name for a composite-CSV student row: the
// hydrated name first (fetchClientNames), else a truncated id. The roster read's
// ClientLabel is the opaque client_id (adapter parity) so it is never a name.
func rosterLabel(row *matrixpb.OutcomeSummaryRosterRow, names map[string]clientName) string {
	if n := names[row.GetClientId()].display; n != "" {
		return n
	}
	return short(row.GetClientId())
}

// exportFilename derives "{page-title}-{subject}" from lyngua'd labels
// ("grade-sheet-arts" on education), falling back to the template id.
func exportFilename(l outcome_matrix.Labels, subjectName, templateID string) string {
	prefix := slug(l.Page.Title)
	if prefix == "" || prefix == "none" {
		prefix = "outcome-matrix"
	}
	suffix := slug(subjectName)
	if suffix == "" || suffix == "none" {
		suffix = slug(templateID)
	}
	return prefix + "-" + suffix
}

// csvSafe neutralizes spreadsheet formula/DDE injection (same OWASP
// neutralization as outcome_summary/section/export.go — a leading formula
// trigger is evaluated by Excel/Sheets; encoding/csv quoting does not prevent
// this). Decodes the first RUNE, not byte: the trigger set includes the
// full-width ＝＋－＠ forms (U+FF1D/0B/0D/20) Excel also honors, plus LF
// alongside TAB/CR as trimmable prefixes (current OWASP guidance).
func csvSafe(s string) string {
	if s == "" {
		return s
	}
	r, _ := utf8.DecodeRuneInString(s)
	switch r {
	case '=', '+', '-', '@', '\t', '\r', '\n', '＝', '＋', '－', '＠':
		return "\t" + s
	}
	return s
}
