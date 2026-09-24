package document

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/erniealice/espyna-golang/consumer"
	espynaports "github.com/erniealice/espyna-golang/ports"
	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	"github.com/erniealice/pyeza-golang/view"
	"golang.org/x/text/unicode/norm"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

const (
	clientDocumentPeriodYearFinal = "year_final"
)

// handleExplicitClientDocument serves the explicit-export path selected by a
// Period form. The no-period path remains the legacy operator document route.
// Both the group-row action and the client-page header action submit here.
func handleExplicitClientDocument(w http.ResponseWriter, r *http.Request, d *Deps) {
	ctx := r.Context()
	if !outcome_summary.CanExplicitExport(view.GetUserPermissions(ctx), ctx, d.ResolvePrincipalKind) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	clientID := strings.TrimSpace(r.PathValue("client_id"))
	if groupID == "" || clientID == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	period := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("period")))
	if period == "" {
		http.Error(w, "invalid period", http.StatusBadRequest)
		return
	}
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "docx"
	}
	if format != "docx" && format != "pdf" {
		http.Error(w, `invalid format: must be "docx" or "pdf"`, http.StatusBadRequest)
		return
	}
	if d.GetSubscriptionGroupClientReportCard == nil {
		http.Error(w, "client report card projection is not configured", http.StatusServiceUnavailable)
		return
	}
	if format == "docx" && d.GenerateDoc == nil {
		http.Error(w, "report card rendering is not configured", http.StatusServiceUnavailable)
		return
	}
	if format == "pdf" && d.GeneratePDF == nil {
		http.Error(w, "report card PDF rendering is not configured", http.StatusServiceUnavailable)
		return
	}

	response, err := d.GetSubscriptionGroupClientReportCard(ctx, &exportpb.GetSubscriptionGroupClientReportCardRequest{
		SubscriptionGroupId:  groupID,
		ClientId:             clientID,
		ClientAttributeCodes: append([]string(nil), d.Options.Document.ClientAttributeCodes...),
		PlanAttributeCodes:   planAttributeCodes(d.Options.Document),
	})
	if err != nil {
		if errors.Is(err, espynaports.ErrClientReportNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Printf("client report card projection: %v", err)
		http.Error(w, "client report card is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}
	if response == nil || !response.GetSuccess() {
		log.Printf("client report card projection: incomplete response")
		http.Error(w, "client report card is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}
	if response.GetReportCard() == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !clientProjectionMatchesRequest(response, groupID, clientID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	card := response.GetReportCard()
	if period == clientDocumentPeriodYearFinal {
		if !clientProjectionHasYearFinal(card) {
			http.Error(w, "requested report period is not available", http.StatusNotFound)
			return
		}
	} else if !clientProjectionHasPhase(card, period) {
		http.Error(w, "requested report period is not available", http.StatusNotFound)
		return
	}

	gate, gateErr := clientProjectionSheetGate(card, groupID)
	blocked := gate.anyBlocked()
	if gateErr != nil {
		log.Printf("client report render gate: cannot prove document safe: %v", gateErr)
		http.Error(w, "report card cannot be generated right now — please retry", http.StatusServiceUnavailable)
		return
	}
	// A proven unpublished sheet may be downloaded as a structural report; a
	// malformed or unprovable gate above remains a hard 503. Phase documents
	// redact per sheet on the projection (notes kept, scores blank); Year Final
	// blanks recorded values across the whole document after assembly.

	printedBy := firstNonEmpty(consumer.GetUserIDFromContext(ctx), "system")
	now := printedNow(ctx)
	printedAt := now.Format("January 2, 2006 3:04 PM")
	printedByDisplay := printedByName(ctx, d, printedBy)
	var templateBytes []byte
	var data map[string]any
	if period != clientDocumentPeriodYearFinal {
		if d.ResolveTemplateBytes != nil {
			templateBytes, err = d.ResolveTemplateBytes(ctx, card.GetContext().GetPriceScheduleId(), period)
		}
		if err != nil || len(templateBytes) == 0 {
			log.Printf("client phase template resolution: %v", err)
			http.Error(w, "report period template is unavailable", http.StatusServiceUnavailable)
			return
		}
		projection := card
		if !d.DocOptions.ShowUnpublishedValues {
			projection = redactBlockedClientPhaseScores(card, gate)
		}
		data, err = buildClientPhaseReportData(d, projection, period, printedByDisplay, printedAt)
		if err != nil {
			log.Printf("client phase report data: %v", err)
			http.Error(w, "report period data is unavailable", http.StatusServiceUnavailable)
			return
		}
	} else {
		templateBytes = clientYearFinalTemplate(ctx, d, card.GetContext().GetPriceScheduleId())
		data = buildProjectedYearFinalData(d, card, printedByDisplay, printedAt, now)
	}
	if len(templateBytes) == 0 {
		http.Error(w, "report template is unavailable", http.StatusServiceUnavailable)
		return
	}
	if blocked && period == clientDocumentPeriodYearFinal {
		blankDocumentOutcomeValues(data)
	}

	var (
		output      []byte
		contentType string
	)
	if format == "pdf" {
		output, err = d.GeneratePDF(templateBytes, data)
		if err != nil {
			if isLibreOfficeUnavailable(err) {
				log.Printf("client report PDF: LibreOffice unavailable: %v", err)
				http.Error(w, "report card PDF rendering is unavailable — LibreOffice is not installed", http.StatusServiceUnavailable)
				return
			}
			log.Printf("client report PDF: generate: %v", err)
			http.Error(w, "failed to generate report card PDF", http.StatusInternalServerError)
			return
		}
		contentType = pdfContentType
	} else {
		output, err = d.GenerateDoc(templateBytes, data)
		if err != nil {
			log.Printf("client report DOCX: generate: %v", err)
			http.Error(w, "failed to generate report card", http.StatusInternalServerError)
			return
		}
		contentType = docxContentType
	}
	if len(output) == 0 {
		http.Error(w, "failed to generate report card", http.StatusInternalServerError)
		return
	}

	filename := clientDocumentFilename(card.GetClient().GetName(), clientID, now, format)
	w.Header().Set("Content-Disposition", contentDisposition(filename))
	if format == "pdf" {
		w.Header().Set("X-Content-Type-Options", "nosniff")
	}
	w.Header().Set("Content-Type", contentType)
	if _, err := w.Write(output); err != nil {
		log.Printf("client report document: write response: %v", err)
	}
}

func clientProjectionMatchesRequest(response *exportpb.GetSubscriptionGroupClientReportCardResponse, groupID, clientID string) bool {
	if response == nil || !response.GetSuccess() || response.GetReportCard() == nil {
		return false
	}
	card := response.GetReportCard()
	if card.GetContext() == nil || card.GetContext().GetSubscriptionGroupId() != groupID ||
		card.GetClient() == nil || card.GetClient().GetClientId() != clientID ||
		len(card.GetClientSubscriptionIds()) == 0 {
		return false
	}
	return true
}

func clientProjectionHasPhase(card *exportpb.ClientReportCardProjection, phaseCode string) bool {
	if card == nil {
		return false
	}
	for _, phase := range clientProjectionPhaseCatalog(card) {
		if strings.EqualFold(phase.Code, strings.TrimSpace(phaseCode)) {
			return true
		}
	}
	return false
}

func clientProjectionPhaseCatalog(card *exportpb.ClientReportCardProjection) []outcome_summary.ClientReportPhase {
	if card == nil || card.GetContext() == nil {
		return nil
	}
	historical := card.GetContext().GetHistorical()
	activeJobTemplates := make(map[string]string, len(card.GetJobs()))
	for _, job := range card.GetJobs() {
		if job == nil || (!historical && !job.GetActive()) {
			continue
		}
		jobID := strings.TrimSpace(job.GetId())
		templateID := strings.TrimSpace(job.GetJobTemplateId())
		if jobID != "" && templateID != "" {
			activeJobTemplates[jobID] = templateID
		}
	}
	activeJobTemplatePhaseIDs := make(map[string]struct{}, len(card.GetJobPhases()))
	for _, phase := range card.GetJobPhases() {
		if phase == nil || (!historical && !phase.GetActive()) {
			continue
		}
		templateID := activeJobTemplates[strings.TrimSpace(phase.GetJobId())]
		phaseID := strings.TrimSpace(phase.GetTemplatePhaseId())
		if templateID != "" && phaseID != "" {
			activeJobTemplatePhaseIDs[templateID+"\x00"+phaseID] = struct{}{}
		}
	}
	entries := make([]outcome_summary.ClientReportPhaseEntry, 0, len(card.GetJobTemplatePhases()))
	for _, phase := range card.GetJobTemplatePhases() {
		if phase == nil || (!historical && !phase.GetActive()) {
			continue
		}
		key := strings.TrimSpace(phase.GetJobTemplateId()) + "\x00" + strings.TrimSpace(phase.GetId())
		if _, referencedByClient := activeJobTemplatePhaseIDs[key]; !referencedByClient {
			continue
		}
		entries = append(entries, outcome_summary.ClientReportPhaseEntry{
			Code: phase.GetCode(), Name: phase.GetName(), Order: phase.GetPhaseOrder(), Active: phase.GetActive(),
		})
	}
	return outcome_summary.BuildClientReportPhaseCatalog(entries)
}

func clientProjectionHasYearFinal(card *exportpb.ClientReportCardProjection) bool {
	if card == nil {
		return false
	}
	jobs := map[string]bool{}
	for _, job := range card.GetJobs() {
		if job != nil {
			jobs[job.GetId()] = true
		}
	}
	return len(jobs) > 0
}

func clientYearFinalTemplate(ctx context.Context, d *Deps, priceScheduleID string) []byte {
	templateBytes := Template()
	if strings.EqualFold(strings.TrimSpace(d.DocOptions.TemplateVariant), outcome_summary.TemplateVariantBlock) {
		templateBytes = TemplateBlock()
	}
	if d.ResolveTemplateBytes != nil {
		if resolved, err := d.ResolveTemplateBytes(ctx, priceScheduleID, ""); err == nil && len(resolved) > 0 {
			templateBytes = resolved
		}
	}
	return templateBytes
}

// clientDocumentFilename builds "{name-dashed}-{client id}-{unix seconds}.{ext}".
// Accents are folded before slugging so "Añora" becomes "anora", not "a-ora".
func clientDocumentFilename(clientName, clientID string, now time.Time, format string) string {
	name := slug(foldAccents(clientName))
	id := slug(clientID)
	return name + "-" + id + "-" + strconv.FormatInt(now.Unix(), 10) + "." + format
}

// foldAccents strips combining marks after NFD decomposition (ñ → n, é → e).
func foldAccents(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// planAttributeCodes lists the configured plan attribute codes the document
// reads (today only the plan label).
func planAttributeCodes(options outcome_summary.DocumentOptions) []string {
	if code := strings.TrimSpace(options.PlanLabelAttributeCode); code != "" {
		return []string{code}
	}
	return nil
}
