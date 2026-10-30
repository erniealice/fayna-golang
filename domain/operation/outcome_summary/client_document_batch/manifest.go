package client_document_batch

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/render"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"
	"golang.org/x/text/unicode/norm"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

// clientDocumentRouteKey is the per-request (workspace-prefixed) route-map key
// of the per-client document route the manifest points at.
const clientDocumentRouteKey = "outcome_summary.client_document"

// Manifest is the read-only batch preflight. It carries no document bytes.
// Expected always equals len(Clients); the browser re-checks that invariant.
type Manifest struct {
	SubscriptionGroupID   string           `json:"subscription_group_id"`
	SubscriptionGroupName string           `json:"subscription_group_name"`
	Period                string           `json:"period"`
	PeriodLabel           string           `json:"period_label"`
	Format                string           `json:"format"`
	ZipFilename           string           `json:"zip_filename"`
	Expected              int              `json:"expected"`
	Clients               []ManifestClient `json:"clients"`
}

// ManifestClient is one member and the existing per-client document URL that
// renders their document for the requested period and format.
type ManifestClient struct {
	ClientID    string `json:"client_id"`
	Name        string `json:"name"`
	DocumentURL string `json:"document_url"`
}

// manifestError is the JSON error body; Code is a stable machine value and
// Message the lyngua-backed user text.
type manifestError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
}

// NewManifestHandler serves GET {manifest}?period=&format=. Mounted as a raw
// HandleFunc so the view adapter's WrapHandler injects permissions, identity
// and the workspace route map exactly as for the per-client document route.
func NewManifestHandler(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if deps == nil || !outcome_summary.CanExplicitExport(view.GetUserPermissions(ctx), ctx, deps.ResolvePrincipalKind) {
			writeManifestError(w, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		l := deps.Labels.ClientDocumentsDownload
		if !Configured(deps.Options, deps.Routes) || deps.GetSubscriptionGroupOutcomeExport == nil {
			writeManifestError(w, http.StatusServiceUnavailable, "not_configured", l.ErrorFallback)
			return
		}
		groupID := strings.TrimSpace(r.PathValue("id"))
		if groupID == "" {
			writeManifestError(w, http.StatusNotFound, "not_found", l.ErrorFallback)
			return
		}
		period := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("period")))
		if period == "" {
			writeManifestError(w, http.StatusBadRequest, "invalid_period", l.PeriodUnavailableError)
			return
		}
		format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
		if format == "" {
			format = FormatPDF
		}
		switch {
		case format != FormatPDF && format != FormatDOCX:
			writeManifestError(w, http.StatusBadRequest, "invalid_format", l.FormatUnavailableError)
			return
		case format == FormatPDF && !deps.PDFAvailable, format == FormatDOCX && !deps.DOCXAvailable:
			writeManifestError(w, http.StatusUnprocessableEntity, "format_unavailable", l.FormatUnavailableError)
			return
		}

		options, err := deps.GetSubscriptionGroupOutcomeExport(ctx, &exportpb.GetSubscriptionGroupOutcomeExportRequest{
			SubscriptionGroupId: groupID,
		})
		if err != nil || !contextMatches(options, groupID) {
			if err != nil {
				log.Printf("client document batch manifest: options: %v", err)
			}
			writeManifestError(w, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		selected, ok := findPeriod(groupPeriods(deps.Labels, options.GetJobCategories()), period)
		if !ok {
			writeManifestError(w, http.StatusNotFound, "period_unavailable", l.PeriodUnavailableError)
			return
		}
		exportContext := options.GetContext()

		// Template preflight: an explicit phase document requires a published
		// binding (the per-client route 503s every file otherwise). Year Final
		// keeps the embedded fallback, so it needs no preflight.
		if period != PeriodYearFinal {
			var templateBytes []byte
			if deps.ResolveTemplateBytes != nil {
				templateBytes, err = deps.ResolveTemplateBytes(ctx, exportContext.GetPriceScheduleId(), period)
			}
			if deps.ResolveTemplateBytes == nil || err != nil || len(templateBytes) == 0 {
				if err != nil {
					log.Printf("client document batch manifest: template preflight: %v", err)
				}
				writeManifestError(w, http.StatusUnprocessableEntity, "template_unavailable", l.TemplateUnavailableError)
				return
			}
		}

		query := memberQuery(groupID, period, options.GetJobCategories())
		if query == nil {
			writeManifestError(w, http.StatusNotFound, "period_unavailable", l.PeriodUnavailableError)
			return
		}
		members, err := deps.GetSubscriptionGroupOutcomeExport(ctx, query)
		if err != nil {
			log.Printf("client document batch manifest: members: %v", err)
			writeManifestError(w, http.StatusServiceUnavailable, "members_unavailable", l.ErrorFallback)
			return
		}
		if !contextMatches(members, groupID) {
			writeManifestError(w, http.StatusForbidden, "forbidden", "forbidden")
			return
		}

		documentPattern := deps.Routes.ClientDocumentURL
		if prefixed := strings.TrimSpace(render.RouteMapFromContext(ctx)[clientDocumentRouteKey]); prefixed != "" {
			documentPattern = prefixed
		}
		documentQuery := url.Values{"period": {period}, "format": {format}}.Encode()

		seen := make(map[string]struct{}, len(members.GetClientRows()))
		clients := make([]ManifestClient, 0, len(members.GetClientRows()))
		for _, row := range members.GetClientRows() {
			clientID := strings.TrimSpace(row.GetClientId())
			if clientID == "" {
				continue
			}
			if _, duplicate := seen[clientID]; duplicate {
				continue
			}
			seen[clientID] = struct{}{}
			clients = append(clients, ManifestClient{
				ClientID:    clientID,
				Name:        clientDisplayName(row),
				DocumentURL: route.ResolveURL(documentPattern, "id", url.PathEscape(groupID), "client_id", url.PathEscape(clientID)) + "?" + documentQuery,
			})
		}
		if len(clients) == 0 {
			writeManifestError(w, http.StatusNotFound, "empty", l.EmptyError)
			return
		}

		manifest := Manifest{
			SubscriptionGroupID:   groupID,
			SubscriptionGroupName: exportContext.GetSubscriptionGroupName(),
			Period:                period,
			PeriodLabel:           selected.Label,
			Format:                format,
			ZipFilename:           zipFilename(exportContext.GetSubscriptionGroupName(), selected.Label, period, format),
			Expected:              len(clients),
			Clients:               clients,
		}
		writeManifestJSON(w, http.StatusOK, manifest)
	}
}

func writeManifestError(w http.ResponseWriter, status int, code, message string) {
	if strings.TrimSpace(message) == "" {
		message = code
	}
	writeManifestJSON(w, status, manifestError{Code: code, Message: message})
}

func writeManifestJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("client document batch manifest: write response: %v", err)
	}
}

// zipFilename builds "{group}-{period}-{format}.zip" with the per-client
// document filename's accent folding and slug rules (document.clientDocumentFilename).
func zipFilename(groupName, periodLabel, periodCode, format string) string {
	group := slug(foldAccents(groupName))
	if group == "" {
		group = "documents"
	}
	period := slug(foldAccents(periodLabel))
	if period == "" {
		period = slug(periodCode)
	}
	parts := []string{group}
	if period != "" {
		parts = append(parts, period)
	}
	parts = append(parts, format)
	return strings.Join(parts, "-") + ".zip"
}

func foldAccents(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

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
	return strings.Trim(b.String(), "-")
}
