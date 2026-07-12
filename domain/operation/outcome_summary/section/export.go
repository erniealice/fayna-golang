// export.go — CSV download for the per-section report-card grid. Serves the
// SAME grid buildSectionTable composes for the HTML view (identical band +
// row order, identical cell text via TableCell.CSVValue), either for the
// whole section or narrowed to one client row (?id=<client id> — the table
// download action's JS appends the row id). Registered as a raw handler
// wrapped by the ViewAdapter (WrapHandler), so view.GetUserPermissions sees
// the same RBAC context as the HTML view — the same Layer-3 gate applies.
package section

import (
	"encoding/csv"
	"log"
	"net/http"
	"strings"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

// NewExportHandler creates the section-grid CSV download handler.
func NewExportHandler(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "list") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		sectionID := strings.TrimSpace(r.PathValue("id"))
		if sectionID == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		group, table := buildSectionTable(ctx, deps, sectionID)
		if group == nil {
			// Workspace EXISTS gate failed — same fail-closed response for
			// foreign and missing ids (no leak).
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if table == nil {
			http.Error(w, deps.Labels.Section.NotComputedBanner, http.StatusNotFound)
			return
		}

		rows := allRows(table)

		// Optional single-row narrowing: the row download action's JS appends
		// ?id=<row id> (a client id). A non-matching id — including the
		// section's own id, which the landing row action appends — keeps the
		// full section.
		prefix := slug(deps.Labels.Section.Title) // lyngua-fied ("report-cards" on education)
		if prefix == "none" {
			prefix = "outcomes"
		}
		filename := prefix + "-" + slug(group.GetName())
		if rowID := strings.TrimSpace(r.URL.Query().Get("id")); rowID != "" {
			for _, row := range rows {
				if row.ID == rowID {
					rows = []types.TableRow{row}
					if len(row.Cells) > 0 {
						// Drop the "{n} " sequence prefix from the filename.
						filename += "-" + slug(strings.TrimLeft(row.Cells[0].Value, "0123456789 "))
					}
					break
				}
			}
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`.csv"`)

		cw := csv.NewWriter(w)
		header := make([]string, 0, len(table.Columns))
		for _, c := range table.Columns {
			header = append(header, csvSafe(c.Label))
		}
		if err := cw.Write(header); err != nil {
			log.Printf("report cards export: write header: %v", err)
			return
		}
		record := make([]string, 0, len(table.Columns))
		for _, row := range rows {
			record = record[:0]
			for _, cell := range row.Cells {
				record = append(record, csvSafe(types.CellCSV(cell)))
			}
			if err := cw.Write(record); err != nil {
				log.Printf("report cards export: write row: %v", err)
				return
			}
		}
		cw.Flush()
	}
}

// csvSafe neutralizes spreadsheet formula/DDE injection: a cell whose text
// begins with = + - @ (or a leading tab/CR that a client may trim onto one of
// those) is evaluated as a formula by Excel/Sheets on open. encoding/csv
// quoting does NOT prevent this. Prefix such values with a tab so the client
// treats them as literal text (the OWASP-recommended neutralization; the tab
// is invisible in the rendered cell). Empty values pass through untouched.
func csvSafe(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "\t" + s
	}
	return s
}
