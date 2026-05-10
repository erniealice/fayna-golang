package detail

import (
	"context"
	"log"

	"github.com/erniealice/hybra-golang/views/auditlog"
	"github.com/erniealice/pyeza-golang/route"
)

// loadAuditHistoryTab populates the audit history fields on PageData for the
// job_template entity. Mirrors the same function in views/job/detail/page.go.
func loadAuditHistoryTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, id string, cursor string) {
	if deps.ListAuditHistory == nil {
		return
	}
	auditResp, err := deps.ListAuditHistory(ctx, &auditlog.ListAuditRequest{
		EntityType:  "job_template",
		EntityID:    id,
		Limit:       20,
		CursorToken: cursor,
	})
	if err != nil {
		log.Printf("Failed to load audit history for job template %s: %v", id, err)
	}
	if auditResp != nil {
		pageData.AuditEntries = auditResp.Entries
		pageData.AuditHasNext = auditResp.HasNext
		pageData.AuditNextCursor = auditResp.NextCursor
	}
	pageData.AuditHistoryURL = route.ResolveURL(deps.Routes.TabActionURL, "id", id, "tab", "") + "audit-history"
}
