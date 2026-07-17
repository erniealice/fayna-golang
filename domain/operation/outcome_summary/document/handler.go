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
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	outcome_summary "github.com/erniealice/fayna-golang/domain/operation/outcome_summary"

	"github.com/erniealice/espyna-golang/consumer"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	clientpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	staffpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/staff"
	workspaceuserpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/workspace_user"
	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobcategorypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_category"
	joblinepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_line"
	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"
	taskoutcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"
	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	subscriptiongrouppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group"
	subscriptiongroupmemberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription_group_member"
)

const docxContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
const pdfContentType = "application/pdf"

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

	// DocOptions carries the app-configured document knobs (group category code,
	// client-reference attribute code). Zero value disables every v2 enrichment —
	// the download renders exactly as before.
	DocOptions outcome_summary.DocumentOptions

	// v2 block-layout enrichment closures (ALL optional/nil-safe — a missing
	// closure blanks the affected field, never a placeholder leak or a panic).
	ListOutcomeCriterias     func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)
	GetStaffListPageData     func(ctx context.Context, req *staffpb.GetStaffListPageDataRequest) (*staffpb.GetStaffListPageDataResponse, error)
	ListPriceSchedules       func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)
	ListClientAttributes     func(ctx context.Context, req *clientattributepb.ListClientAttributesRequest) (*clientattributepb.ListClientAttributesResponse, error)
	ResolveAttributeIDByCode func(ctx context.Context, code string) (string, error)
	ListWorkspaceUsers       func(ctx context.Context, req *workspaceuserpb.ListWorkspaceUsersRequest) (*workspaceuserpb.ListWorkspaceUsersResponse, error)

	// GenerateDoc wraps fycha DocumentService.ProcessBytes (template bytes + data
	// map → processed .docx). Injected by the app container via the block Infra.
	GenerateDoc func(templateData []byte, data map[string]any) ([]byte, error)

	// GeneratePDF wraps fycha DocumentService.ProcessBytesToPDF (template bytes +
	// data map → rendered .docx → .pdf via LibreOffice) — a SECOND injected
	// closure mirroring GenerateDoc. Optional/nil-safe: the route registers on
	// GenerateDoc alone (DOCX baseline); a ?format=pdf request when this is nil is
	// a narrower per-format 503, not a 404. On a host without LibreOffice the
	// closure returns the fycha ErrLibreOfficeUnavailable sentinel (detected here
	// by its stable message — fayna must NOT import fycha) → 503; any other error
	// → 500.
	GeneratePDF func(templateData []byte, data map[string]any) ([]byte, error)

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

		// Format selection: empty/"docx" → the existing DOCX behavior (unchanged
		// default, zero regression — no caller relied on a param before W5); "pdf"
		// → the injected GeneratePDF closure; anything else → 400.
		format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
		if format == "" {
			format = "docx"
		}
		if format != "docx" && format != "pdf" {
			http.Error(w, `invalid format: must be "docx" or "pdf"`, http.StatusBadRequest)
			return
		}

		// Per-format wiring gate. The route only registers when GenerateDoc is
		// wired (DOCX baseline, module gate). DOCX keeps that same nil check; PDF
		// additionally needs GeneratePDF — its absence is a narrower, format-
		// specific 503 (fail-closed), NOT a route-level 404.
		switch format {
		case "docx":
			if d.GenerateDoc == nil {
				log.Printf("report card doc: GenerateDoc not wired — refusing to serve")
				http.Error(w, "report card rendering is not configured", http.StatusServiceUnavailable)
				return
			}
		case "pdf":
			if d.GeneratePDF == nil {
				log.Printf("report card pdf: GeneratePDF not wired — refusing to serve")
				http.Error(w, "report card PDF rendering is not configured", http.StatusServiceUnavailable)
				return
			}
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
		now := time.Now()
		rc.PrintedAt = now.Format("2006-01-02 15:04")
		rc.PrintedAtLong = now.Format("January 2, 2006 03:04 PM")
		rc.PrintedByName = printedByName(ctx, d, rc.PrintedBy)

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

		// The SAME data map feeds both formats — PDF conversion happens AFTER the
		// identical DOCX assembly (fycha renders the DOCX then LibreOffice converts).
		data := buildReportCardData(*rc)

		var (
			outBytes    []byte
			contentType string
			// Content-Disposition is set per-branch below (DOCX keeps its exact
			// pre-W5 header for zero regression; PDF uses the LOCKED filename with an
			// RFC-5987 encoding + nosniff).
			err error
		)
		switch format {
		case "pdf":
			outBytes, err = d.GeneratePDF(tpl, data)
			if err != nil {
				// LibreOffice absent at runtime → 503 (the closure IS wired, the host
				// simply has no soffice). fayna must NOT import fycha, so the fycha
				// ErrLibreOfficeUnavailable sentinel is matched by its stable message
				// rather than errors.Is. Any other error is a genuine 500.
				if isLibreOfficeUnavailable(err) {
					log.Printf("report card pdf: LibreOffice unavailable: %v", err)
					http.Error(w, "report card PDF rendering is unavailable — LibreOffice is not installed", http.StatusServiceUnavailable)
					return
				}
				log.Printf("report card pdf: generate: %v", err)
				http.Error(w, "failed to generate report card PDF", http.StatusInternalServerError)
				return
			}
			if len(outBytes) == 0 {
				http.Error(w, "failed to generate report card PDF", http.StatusInternalServerError)
				return
			}
			contentType = pdfContentType
			// LOCKED filename: "Report Card - {Student} - {AY} - {unixMilli}.pdf"
			// (decisions.md Q-GSE PDF filename lock). AY derives from the section's
			// price_schedule period (rc.SchedulePeriod); unixMilli is the current
			// time in ms. Full conversion completes before any byte is streamed.
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Content-Disposition", contentDisposition(reportCardPDFFilename(rc)))
			if _, err := w.Write(outBytes); err != nil {
				log.Printf("report card pdf: write response: %v", err)
			}
		default: // "docx" — unchanged pre-W5 behavior
			outBytes, err = d.GenerateDoc(tpl, data)
			if err != nil {
				log.Printf("report card doc: generate: %v", err)
				http.Error(w, "failed to generate report card", http.StatusInternalServerError)
				return
			}
			if len(outBytes) == 0 {
				http.Error(w, "failed to generate report card", http.StatusInternalServerError)
				return
			}
			filename := "report-card-" + slug(rc.SectionName) + "-" + slug(rc.ClientName) + ".docx"
			w.Header().Set("Content-Type", docxContentType)
			w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
			if _, err := w.Write(outBytes); err != nil {
				log.Printf("report card doc: write response: %v", err)
			}
		}
	}
}

// libreOfficeUnavailable is the STRUCTURAL contract fycha's
// document.ErrLibreOfficeUnavailable satisfies (its concrete type exposes this
// method). fayna must NOT import fycha (the PDF path is an injected closure —
// architecture boundary), so we cannot reference the sentinel VALUE or its type
// for errors.Is; instead we assert this tiny interface, which needs no import.
type libreOfficeUnavailable interface {
	LibreOfficeUnavailable() bool
}

// isLibreOfficeUnavailable reports whether err is (or wraps) the fycha runtime
// "soffice binary absent" condition that must map to a 503 rather than a 500.
//
// PREFERRED: a structural interface assertion via errors.As — it walks the wrap
// chain to find any error exposing LibreOfficeUnavailable() bool (which the fycha
// sentinel does), so a fycha message reword can no longer silently downgrade 503
// to 500, and an unrelated error that merely mentions LibreOffice is not
// misclassified.
//
// FALLBACK (documented): the stable "LibreOffice is not installed" substring, kept
// as defense-in-depth for an error that crossed a boundary which dropped the typed
// wrapper (e.g. serialized across a process). It is intentionally secondary.
func isLibreOfficeUnavailable(err error) bool {
	if err == nil {
		return false
	}
	var lou libreOfficeUnavailable
	if errors.As(err, &lou) && lou.LibreOfficeUnavailable() {
		return true
	}
	return strings.Contains(err.Error(), "LibreOffice is not installed")
}

// reportCardPDFFilename builds the LOCKED report-card PDF filename
// "Report Card - {Student} - {AY} - {unixMilli}.pdf" (decisions.md). The raw
// (possibly non-ASCII, space/comma-bearing) name is sanitized for transport by
// contentDisposition.
func reportCardPDFFilename(rc *reportCard) string {
	student := firstNonEmpty(rc.ClientName, "Student")
	ay := strings.TrimSpace(rc.SchedulePeriod)
	name := "Report Card - " + student
	if ay != "" {
		name += " - " + ay
	}
	name += " - " + strconv.FormatInt(time.Now().UnixMilli(), 10) + ".pdf"
	return name
}

// contentDisposition builds an attachment Content-Disposition with BOTH an
// ASCII-safe filename="" fallback and an RFC-5987 filename*=UTF-8'' form, so
// names with spaces/commas/non-ASCII (student names) download correctly across
// browsers without header injection.
func contentDisposition(name string) string {
	ascii := asciiFilename(name)
	return `attachment; filename="` + ascii + `"; filename*=UTF-8''` + encodeRFC5987(name)
}

// asciiFilename produces a safe quoted-string fallback: printable-ASCII only,
// with '"' and '\' (the quoted-string escapes) and control chars replaced by '_'.
func asciiFilename(name string) string {
	var b strings.Builder
	for _, r := range name {
		if r >= 0x20 && r < 0x7f && r != '"' && r != '\\' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "report-card.pdf"
	}
	return out
}

// encodeRFC5987 percent-encodes name as a UTF-8 ext-value per RFC 5987 §3.2
// (attr-char stays literal; every other byte → %XX).
func encodeRFC5987(name string) string {
	const attr = "!#$&+-.^_`|~"
	var b strings.Builder
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9',
			strings.IndexByte(attr, c) >= 0:
			b.WriteByte(c)
		default:
			b.WriteByte('%')
			const hex = "0123456789ABCDEF"
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0f])
		}
	}
	return b.String()
}

// printedByName resolves the printing user's display name via the
// workspace_user read (whose adapter hydrates the joined user name). Falls
// back to "" (callers keep the raw principal id). Nil-safe.
func printedByName(ctx context.Context, d *Deps, userID string) string {
	if d.ListWorkspaceUsers == nil || strings.TrimSpace(userID) == "" {
		return ""
	}
	resp, err := d.ListWorkspaceUsers(ctx, &workspaceuserpb.ListWorkspaceUsersRequest{
		Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{stringEq("user_id", userID)}},
	})
	if err != nil {
		log.Printf("report card doc: list workspace users: %v", err)
		return ""
	}
	for _, wu := range resp.GetData() {
		if !wu.GetActive() {
			continue
		}
		if u := wu.GetUser(); u != nil {
			n := strings.TrimSpace(strings.TrimSpace(u.GetFirstName()) + " " + strings.TrimSpace(u.GetLastName()))
			if n != "" {
				return n
			}
		}
	}
	return ""
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
