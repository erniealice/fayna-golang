// handler.go — the per-student report-card .docx download. A raw
// http.HandlerFunc (registered through the ViewAdapter, so view.GetUserPermissions
// observes the same RBAC context as the HTML views — the SectionExport
// precedent). It reuses the exact view-3 fetch + IDOR gates (section EXISTS gate
// → membership gate), assembles the split-source data map, and streams the .docx
// produced by the injected GenerateDoc closure (fycha doctemplate.ProcessBytes,
// already wired on both app containers).
package document

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	"github.com/erniealice/espyna-golang/consumer"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/view"

	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	joblinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

const docxContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

// Deps holds the report-card document handler dependencies. Every list closure
// is workspace-bound at the espyna adapter (mirroring view-3). GenerateDoc is
// the injected fycha doctemplate closure (nil → the route fails closed with 503).
type Deps struct {
	Labels       outcome_summary.Labels
	CommonLabels pyeza.CommonLabels

	// DocumentHeaderName is the generic document header (sourced from a lyngua
	// label by the module wiring — no education vocabulary in code).
	DocumentHeaderName string

	// CategoryFilter (a job_category code, e.g. "academic") + ListJobCategories
	// gate the subject set to that category — same-origin deportment jobs are
	// dropped (gate H2). Empty code or nil closure → no filter. Resolved once per
	// request via outcome_summary.ResolveCategoryID.
	CategoryFilter    string
	ListJobCategories func(ctx context.Context, req *jobcategorypb.ListJobCategoriesRequest) (*jobcategorypb.ListJobCategoriesResponse, error)

	// GenerateDoc wraps fycha DocumentService.ProcessBytes (template bytes + data
	// map → processed .docx). Injected by the app container via the block Infra.
	GenerateDoc func(templateData []byte, data map[string]any) ([]byte, error)

	// ResolveTemplateBytes resolves the applicable published report-card template
	// binding for this card's price_schedule and returns the bound template's
	// storage bytes (binding resolver ∘ storage download). Returns (nil, nil) on
	// no-binding / unavailable-object → the handler keeps the embedded Template()
	// fallback. Optional/nil-safe — no download regression when unwired.
	ResolveTemplateBytes func(ctx context.Context, priceScheduleID string) ([]byte, error)

	ListSubscriptionGroups        func(ctx context.Context, req *subscriptiongrouppb.ListSubscriptionGroupsRequest) (*subscriptiongrouppb.ListSubscriptionGroupsResponse, error)
	ListSubscriptionGroupMembers  func(ctx context.Context, req *subscriptiongroupmemberpb.ListSubscriptionGroupMembersRequest) (*subscriptiongroupmemberpb.ListSubscriptionGroupMembersResponse, error)
	ListJobs                      func(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error)
	ListJobTemplates              func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
	ListClients                   func(ctx context.Context, req *clientpb.ListClientsRequest) (*clientpb.ListClientsResponse, error)
	ListJobOutcomeSummarys        func(ctx context.Context, req *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error)
	ListPhaseOutcomeSummarysByJob func(ctx context.Context, req *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error)
	ListJobPhases                 func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)
	// ListJobOutcomeLines is retained for a potential per-subject fallback; the
	// per-criterion transcript now reads task_outcome (see below) since
	// job_outcome_line on education1 is per-subject only.
	ListJobOutcomeLines func(ctx context.Context, req *joblinepb.ListJobOutcomeLinesRequest) (*joblinepb.ListJobOutcomeLinesResponse, error)

	// Per-criterion transcript path (crit_a..crit_d + criteria_total). The
	// authoritative per-criterion marks live on task_outcome, reached through
	// job_phase → job_task → task_outcome and ordered A/B/C/D via
	// template_task_criteria.sequence_order. Scoped to THIS card's jobs (no
	// cross-AY accumulation). All nil-safe: a missing closure leaves the
	// criterion columns blank.
	ListJobTasks              func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)
	ListTaskOutcomes          func(ctx context.Context, req *taskoutcomepb.ListTaskOutcomesRequest) (*taskoutcomepb.ListTaskOutcomesResponse, error)
	ListTemplateTaskCriterias func(ctx context.Context, req *ttcpb.ListTemplateTaskCriteriasRequest) (*ttcpb.ListTemplateTaskCriteriasResponse, error)
}

// NewDownloadHandler returns the per-student report-card .docx download handler.
func NewDownloadHandler(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_outcome_summary", "list") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		sectionID := strings.TrimSpace(r.PathValue("id"))
		clientID := strings.TrimSpace(r.PathValue("client_id"))
		if sectionID == "" || clientID == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if d.GenerateDoc == nil {
			log.Printf("report card doc: GenerateDoc not wired — refusing to serve")
			http.Error(w, "report card rendering is not configured", http.StatusServiceUnavailable)
			return
		}

		rc, ok := collectCard(ctx, d, sectionID, clientID)
		if !ok {
			// Same fail-closed response for foreign/missing section, non-member
			// client, and no-data — no leak of which gate tripped.
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if len(rc.Subjects) == 0 {
			http.Error(w, "no computed grades for this student", http.StatusNotFound)
			return
		}

		if rc.DocumentHeaderName == "" {
			rc.DocumentHeaderName = firstNonEmpty(d.Labels.Landing.Title, "Report Card")
		}
		rc.PrintedBy = firstNonEmpty(consumer.GetUserIDFromContext(ctx), "system")
		rc.PrintedAt = time.Now().Format("2006-01-02 15:04")

		// Template selection: the operator-uploaded, AY-scoped binding is the
		// override; the embedded Template() is the fallback (the invoice_download
		// precedent). A resolver error / no-binding / unavailable storage object
		// keeps the embedded bytes — the proven live download must not regress.
		tpl := Template()
		if d.ResolveTemplateBytes != nil {
			if b, rerr := d.ResolveTemplateBytes(ctx, rc.PriceScheduleID); rerr == nil && len(b) > 0 {
				tpl = b
			}
		}

		docBytes, err := d.GenerateDoc(tpl, buildReportCardData(*rc))
		if err != nil {
			log.Printf("report card doc: generate: %v", err)
			http.Error(w, "failed to generate report card", http.StatusInternalServerError)
			return
		}
		if len(docBytes) == 0 {
			http.Error(w, "failed to generate report card", http.StatusInternalServerError)
			return
		}

		filename := "report-card-" + slug(rc.SectionName) + "-" + slug(rc.ClientName) + ".docx"
		w.Header().Set("Content-Type", docxContentType)
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		if _, err := w.Write(docBytes); err != nil {
			log.Printf("report card doc: write response: %v", err)
		}
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// slug lowercases and hyphenates for a safe download filename.
func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "card"
	}
	return out
}
