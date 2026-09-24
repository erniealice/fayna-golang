package outcome_summary

import (
	"net/http"
	"strings"
)

// IsTableRefresh reports whether r is pyeza's post-action table refresh
// (sheet.js refreshTable → htmx GET of the table card's data-refresh-url,
// swapping #<tableID>-card or #<tableID>). Such a request must receive only
// the "table-card" partial; answering with the full page nests a second page
// shell inside the table card.
func IsTableRefresh(r *http.Request, tableID string) bool {
	if r == nil || r.Header.Get("HX-Request") != "true" {
		return false
	}
	target := strings.TrimPrefix(strings.TrimSpace(r.Header.Get("HX-Target")), "#")
	return target != "" && (target == tableID || target == tableID+"-card")
}
