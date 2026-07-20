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
	"encoding/csv"
	"log"
	"net/http"
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

		hidden := resolveHidden(r.URL.Query().Get("hide"), resp)
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
