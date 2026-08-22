// export.go — CSV download for the per-section report-card grid. Serves the
// SAME grid buildSectionTable composes for the HTML view (identical band +
// row order, identical cell text via TableCell.CSVValue), either for the
// whole section or narrowed to one client row (?id=<client id> — the table
// download action's JS appends the row id). Registered as a raw handler
// wrapped by the ViewAdapter (WrapHandler), so view.GetUserPermissions sees
// the same RBAC context as the HTML view — the same Layer-3 gate applies.
package section

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/erniealice/fayna-golang/domain/operation/outcome_summary"
	sectiondoc "github.com/erniealice/fayna-golang/domain/operation/outcome_summary/section_document"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	clientattributepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/entity/client_attribute"
	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

var explicitPhaseCode = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// exportSkipColumns returns the set of column indices the CSV export omits — the
// frozen per-row action column (actionsColumnKey), matched by Key so a header
// rename can never leak raw HTML action anchors into the CSV. Shared with the
// grid builder's actionsColumnKey (T8).
func exportSkipColumns(columns []types.TableColumn) map[int]bool {
	skip := map[int]bool{}
	for i, c := range columns {
		if c.Key == actionsColumnKey {
			skip[i] = true
		}
	}
	return skip
}

// NewExportHandler serves both the frozen legacy section-grid CSV path and the
// explicit category × one-period drawer path. Branching happens before reads:
// no new selectors stays byte-compatible; legacy `jc` mixed with any new
// selector is rejected rather than guessed.
func NewExportHandler(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		perms := view.GetUserPermissions(ctx)
		sectionID := strings.TrimSpace(r.PathValue("id"))
		if sectionID == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		query := r.URL.Query()
		explicit := query.Has("format") || query.Has("job_category_id") || query.Has("period")
		if !explicit {
			if !outcome_summary.CanLegacyDetail(perms) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			serveLegacySectionCSV(w, r, deps, sectionID)
			return
		}
		if deps == nil {
			http.Error(w, "export unavailable", http.StatusServiceUnavailable)
			return
		}
		if !outcome_summary.CanExplicitExport(perms, ctx, deps.ResolvePrincipalKind) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !deps.Options.SectionExportEnabled() {
			http.NotFound(w, r)
			return
		}
		if query.Has("jc") {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		format := strings.TrimSpace(query.Get("format"))
		categoryID := strings.TrimSpace(query.Get("job_category_id"))
		period := strings.TrimSpace(query.Get("period"))
		req, ok := explicitExportRequest(sectionID, categoryID, period)
		if !ok || (format != "csv" && format != "pdf") {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		_, _, bandConfigured, bandErr := deps.Options.ExportRowBandConfig()
		if bandErr != nil {
			http.Error(w, deps.Labels.SectionExport.GroupingConfigError, http.StatusServiceUnavailable)
			return
		}
		if bandConfigured && (!perms.Can("attribute", "list") || !perms.Can("client_attribute", "list")) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if deps.GetSubscriptionGroupOutcomeExport == nil {
			http.Error(w, deps.Labels.SectionExport.DataUnavailableError, http.StatusServiceUnavailable)
			return
		}

		resp, err := deps.GetSubscriptionGroupOutcomeExport(ctx, req)
		if resp != nil && (resp.GetContext() == nil || resp.GetContext().GetSubscriptionGroupId() != sectionID) {
			// Missing, foreign, and unauthorized groups have the same shape.
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			reason, expectedSlots, actualSlots, canonicalOrder := explicitMatrixFailureContext(err)
			if reason != "none" {
				logExplicitExportFailure("selection", reason, expectedSlots, actualSlots, canonicalOrder)
			}
			if resp != nil && !explicitSelectionOffered(resp, categoryID, period) {
				reason := "selection_unoffered"
				reqExpected := len(resp.GetJobTemplateColumns())
				reqActual := reqExpected
				if resp.GetContext() != nil {
					reqActual = len(resp.GetClientRows())
				}
				logExplicitExportFailure("selection", reason, reqExpected, reqActual, columnsInCanonicalOrder(resp.GetJobTemplateColumns()))
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			http.Error(w, deps.Labels.SectionExport.DataUnavailableError, http.StatusServiceUnavailable)
			return
		}
		if !explicitSelectionOffered(resp, categoryID, period) {
			if resp != nil {
				reqExpected := len(resp.GetJobTemplateColumns())
				reqActual := reqExpected
				if resp.GetContext() != nil {
					reqActual = len(resp.GetClientRows())
				}
				logExplicitExportFailure("selection", "selection_unoffered", reqExpected, reqActual, columnsInCanonicalOrder(resp.GetJobTemplateColumns()))
			}
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if len(resp.GetJobTemplateColumns()) == 0 || len(resp.GetClientRows()) == 0 {
			http.Error(w, deps.Labels.Section.NotComputedBanner, http.StatusNotFound)
			return
		}

		matrix, err := normalizeExplicitMatrix(ctx, deps, resp)
		if err != nil {
			reason, expectedSlots, actualSlots, canonicalOrder := explicitMatrixFailureContext(err)
			logExplicitExportFailure("normalize", reason, expectedSlots, actualSlots, canonicalOrder)
			http.Error(w, deps.Labels.SectionExport.DataUnavailableError, explicitMatrixFailureStatus(err))
			return
		}
		if format == "pdf" {
			writeExplicitPDF(ctx, w, deps, matrix, categoryID, period)
			return
		}
		writeExplicitCSV(w, deps, matrix, period)
	}
}

func writeExplicitPDF(ctx context.Context, w http.ResponseWriter, deps *Deps, matrix *explicitMatrix, categoryID, period string) {
	canonicalOrder := columnsInCanonicalOrder(matrix.columns)
	actualSlots := len(matrix.rows)
	category := selectedExplicitCategory(matrix, categoryID)
	if category == nil {
		logExplicitExportFailure("selection", "missing_category", len(matrix.columns), actualSlots, canonicalOrder)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	profile, mapped := deps.Options.SectionExport.ProfileForCategoryCode(category.GetCode())
	if !mapped {
		logExplicitExportFailure("profile", "missing_profile", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.NoProfileError, http.StatusBadRequest)
		return
	}
	registered, supported := sectiondoc.LookupProfile(profile)
	if !supported {
		logExplicitExportFailure("profile", "profile_unavailable", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.IncompatibleTemplateError, http.StatusBadRequest)
		return
	}
	if registered.JobTemplateSlots != len(matrix.columns) || !canonicalOrder {
		logExplicitExportFailure("profile", "slot_count_or_order_mismatch", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.IncompatibleTemplateError, http.StatusBadRequest)
		return
	}
	if deps.ResolveSectionTemplate == nil || deps.GeneratePDF == nil {
		logExplicitExportFailure("profile", "template_dependency_unavailable", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.NoTemplateError, http.StatusServiceUnavailable)
		return
	}

	request := &exportpb.ResolveSubscriptionGroupOutcomeDocumentForRenderRequest{
		SubscriptionGroupId:     matrix.context.GetSubscriptionGroupId(),
		JobCategoryId:           categoryID,
		RenderProfile:           profile,
		ExpectedPlanId:          cloneOptionalString(matrix.context.PlanId),
		ExpectedPriceScheduleId: cloneOptionalString(matrix.context.PriceScheduleId),
	}
	resolved, err := deps.ResolveSectionTemplate(ctx, request)
	if err != nil {
		logExplicitExportFailure("profile", "template_resolve_failed", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.DataUnavailableError, http.StatusServiceUnavailable)
		return
	}
	if resolved == nil || len(resolved.Bytes) == 0 {
		logExplicitExportFailure("profile", "template_not_found", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.NoTemplateError, http.StatusServiceUnavailable)
		return
	}
	if resolved.RenderProfile != profile || resolved.JobCategoryID != categoryID {
		logExplicitExportFailure("profile", "resolver_profile_category_mismatch", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.IncompatibleTemplateError, http.StatusBadRequest)
		return
	}
	if err := sectiondoc.ValidateTemplate(profile, resolved.Bytes); err != nil {
		logExplicitExportFailure("profile", "invalid_template_manifest", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.IncompatibleTemplateError, http.StatusServiceUnavailable)
		return
	}

	data, err := sectiondoc.BuildData(profile, explicitDocumentMatrix(deps, matrix, category, period))
	if err != nil {
		logExplicitExportFailure("profile", "manifest_incompatible", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.IncompatibleTemplateError, http.StatusServiceUnavailable)
		return
	}
	pdf, err := deps.GeneratePDF(resolved.Bytes, data)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case isSectionStyleContractError(err):
			logExplicitExportFailure("render", "style_contract", len(matrix.columns), len(matrix.rows), canonicalOrder)
			status = http.StatusServiceUnavailable
		case isSectionLibreOfficeUnavailable(err):
			logExplicitExportFailure("render", "libreoffice_unavailable", len(matrix.columns), len(matrix.rows), canonicalOrder)
			status = http.StatusServiceUnavailable
		default:
			logExplicitExportFailure("render", "unexpected", len(matrix.columns), len(matrix.rows), canonicalOrder)
		}
		http.Error(w, deps.Labels.SectionExport.DataUnavailableError, status)
		return
	}
	if len(pdf) == 0 || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		logExplicitExportFailure("render", "invalid_pdf", len(matrix.columns), len(matrix.rows), canonicalOrder)
		http.Error(w, deps.Labels.SectionExport.DataUnavailableError, http.StatusInternalServerError)
		return
	}

	filename := slug(deps.Labels.Section.Title) + "-" + slug(matrix.context.GetSubscriptionGroupName()) + "-" + slug(period) + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	if _, err := w.Write(pdf); err != nil {
		log.Printf("section explicit PDF: write response: %v", err)
	}
}

func selectedExplicitCategory(matrix *explicitMatrix, categoryID string) *exportpb.JobCategoryOption {
	if matrix == nil {
		return nil
	}
	for _, category := range matrix.categories {
		if category != nil && category.GetJobCategoryId() == categoryID {
			return category
		}
	}
	return nil
}

func columnsInCanonicalOrder(columns []*exportpb.JobTemplateColumn) bool {
	for index := 1; index < len(columns); index++ {
		left, right := columns[index-1], columns[index]
		leftName, rightName := strings.TrimSpace(left.GetDisplayName()), strings.TrimSpace(right.GetDisplayName())
		if leftName == "" {
			leftName = left.GetJobTemplateId()
		}
		if rightName == "" {
			rightName = right.GetJobTemplateId()
		}
		leftFold, rightFold := strings.ToLower(leftName), strings.ToLower(rightName)
		if leftFold > rightFold || (leftFold == rightFold && left.GetJobTemplateId() > right.GetJobTemplateId()) {
			return false
		}
	}
	return true
}

func explicitColumnName(column *exportpb.JobTemplateColumn) string {
	if column == nil {
		return ""
	}
	if name := strings.TrimSpace(column.GetDisplayName()); name != "" {
		return name
	}
	return column.GetJobTemplateId()
}

func explicitDocumentMatrix(deps *Deps, matrix *explicitMatrix, category *exportpb.JobCategoryOption, period string) sectiondoc.Matrix {
	result := sectiondoc.Matrix{
		JobCategoryID:         category.GetJobCategoryId(),
		SheetTitle:            categoryLabel(deps.Labels, category),
		SubscriptionGroupName: matrix.context.GetSubscriptionGroupName(),
		PriceScheduleName:     matrix.context.GetPriceScheduleName(),
		JobTemplatePhaseName:  explicitPeriodName(deps.Labels, category, period),
		ClientNameLabel:       deps.Labels.Section.ClientColumn,
	}
	for _, column := range matrix.columns {
		result.Columns = append(result.Columns, sectiondoc.Column{JobTemplateID: column.GetJobTemplateId(), DisplayName: explicitColumnName(column)})
	}
	for _, row := range matrix.rows {
		documentRow := sectiondoc.Row{Kind: sectiondoc.RowClient, Label: row.name}
		switch {
		case row.band:
			documentRow.Kind = sectiondoc.RowBand
		case row.blank:
			documentRow.Kind = sectiondoc.RowBlank
		default:
			for index, column := range matrix.columns {
				documentRow.Cells = append(documentRow.Cells, sectiondoc.Cell{JobTemplateID: column.GetJobTemplateId(), Value: row.values[index]})
			}
		}
		result.Rows = append(result.Rows, documentRow)
	}
	return result
}

func explicitPeriodName(labels outcome_summary.Labels, category *exportpb.JobCategoryOption, period string) string {
	if period == "final" {
		return labels.SectionExport.PeriodFinal
	}
	code, _ := strings.CutPrefix(period, "phase:")
	for _, phase := range category.GetJobTemplatePhases() {
		if phase != nil && phase.GetCode() == code {
			if name := strings.TrimSpace(phase.GetName()); name != "" {
				return name
			}
			return code
		}
	}
	return code
}

func cloneOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

type sectionStyleContractError interface{ StyleContractError() bool }
type sectionLibreOfficeUnavailable interface{ LibreOfficeUnavailable() bool }

func isSectionStyleContractError(err error) bool {
	var marker sectionStyleContractError
	return errors.As(err, &marker) && marker.StyleContractError()
}

func isSectionLibreOfficeUnavailable(err error) bool {
	var marker sectionLibreOfficeUnavailable
	return errors.As(err, &marker) && marker.LibreOfficeUnavailable()
}

func serveLegacySectionCSV(w http.ResponseWriter, r *http.Request, deps *Deps, sectionID string) {
	ctx := r.Context()
	group, table, _ := buildSectionTable(ctx, deps, sectionID, r.URL.Query().Get("jc"))
	if group == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if table == nil {
		http.Error(w, deps.Labels.Section.NotComputedBanner, http.StatusNotFound)
		return
	}

	rows := allRows(table)
	prefix := slug(deps.Labels.Section.Title)
	if prefix == "none" {
		prefix = "outcomes"
	}
	filename := prefix + "-" + slug(group.GetName())
	if rowID := strings.TrimSpace(r.URL.Query().Get("id")); rowID != "" {
		for _, row := range rows {
			if row.ID == rowID {
				rows = []types.TableRow{row}
				if len(row.Cells) > 0 {
					filename += "-" + slug(strings.TrimLeft(row.Cells[0].Value, "0123456789 "))
				}
				break
			}
		}
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`.csv"`)
	skipCol := exportSkipColumns(table.Columns)
	cw := csv.NewWriter(w)
	header := make([]string, 0, len(table.Columns))
	for i, c := range table.Columns {
		if !skipCol[i] {
			header = append(header, csvSafe(c.Label))
		}
	}
	if err := cw.Write(header); err != nil {
		log.Printf("report cards export: write header: %v", err)
		return
	}
	record := make([]string, 0, len(table.Columns))
	for _, row := range rows {
		record = record[:0]
		for i, cell := range row.Cells {
			if !skipCol[i] {
				record = append(record, csvSafe(types.CellCSV(cell)))
			}
		}
		if err := cw.Write(record); err != nil {
			log.Printf("report cards export: write row: %v", err)
			return
		}
	}
	cw.Flush()
}

func explicitExportRequest(sectionID, categoryID, period string) (*exportpb.GetSubscriptionGroupOutcomeExportRequest, bool) {
	if sectionID == "" || categoryID == "" || period == "" {
		return nil, false
	}
	req := &exportpb.GetSubscriptionGroupOutcomeExportRequest{
		SubscriptionGroupId: sectionID,
		JobCategoryId:       &categoryID,
	}
	if period == "final" {
		req.OutcomeSelector = &exportpb.GetSubscriptionGroupOutcomeExportRequest_FinalOutcome{FinalOutcome: true}
		return req, true
	}
	code, found := strings.CutPrefix(period, "phase:")
	if !found || !explicitPhaseCode.MatchString(code) {
		return nil, false
	}
	req.OutcomeSelector = &exportpb.GetSubscriptionGroupOutcomeExportRequest_JobTemplatePhaseCode{JobTemplatePhaseCode: code}
	return req, true
}

func explicitSelectionOffered(resp *exportpb.GetSubscriptionGroupOutcomeExportResponse, categoryID, period string) bool {
	if resp == nil {
		return false
	}
	var selected *exportpb.JobCategoryOption
	for _, category := range resp.GetJobCategories() {
		if category != nil && category.GetJobCategoryId() == categoryID {
			selected = category
			break
		}
	}
	if selected == nil {
		return false
	}
	if period == "final" {
		return selected.GetFinalOutcomeAvailable()
	}
	code, found := strings.CutPrefix(period, "phase:")
	if !found {
		return false
	}
	for _, phase := range selected.GetJobTemplatePhases() {
		if phase.GetCode() == code {
			return !phase.GetAmbiguous()
		}
	}
	return false
}

type explicitMatrix struct {
	context    *exportpb.SubscriptionGroupOutcomeExportContext
	categories []*exportpb.JobCategoryOption
	columns    []*exportpb.JobTemplateColumn
	rows       []explicitRow
}

type explicitRow struct {
	clientID, name, firstName, lastName, group string
	values                                     []string
	band, blank                                bool
}

func explicitNormalizeFailureClass(err error) string {
	if err == nil {
		return "none"
	}
	if err == nil {
		return "none"
	}
	message := err.Error()
	for _, candidate := range []string{
		"incomplete explicit matrix",
		"columns not in canonical order",
		"empty job template column",
		"duplicate job template column",
		"invalid client rectangle",
		"missing cell evidence",
		"invalid cell identity set",
		"attribute dependencies unavailable",
		"attribute definition read failed",
		"attribute definition response escaped trusted filter",
		"attribute definition is missing or ambiguous",
		"duplicate client row",
		"client attribute read failed",
		"client attribute response escaped trusted filter",
		"client attribute response escaped scoped roster",
		"duplicate client attribute value",
	} {
		if strings.Contains(message, candidate) {
			return strings.ReplaceAll(candidate, " ", "_")
		}
	}
	return "matrix_contract"
}

type explicitMatrixFailure struct {
	reason         string
	expectedSlots  int
	actualSlots    int
	canonicalOrder bool
}

func (failure explicitMatrixFailure) Error() string {
	if failure.reason == "" {
		return "incomplete explicit matrix"
	}
	return failure.reason
}

func explicitMatrixFailureContext(err error) (string, int, int, bool) {
	if failure, ok := err.(explicitMatrixFailure); ok {
		return explicitNormalizeFailureClass(failure), failure.expectedSlots, failure.actualSlots, failure.canonicalOrder
	}
	if ptr, ok := err.(*explicitMatrixFailure); ok {
		return explicitNormalizeFailureClass(ptr), ptr.expectedSlots, ptr.actualSlots, ptr.canonicalOrder
	}
	return explicitNormalizeFailureClass(err), 0, 0, true
}

func logExplicitExportFailure(stage, reason string, expectedSlots, actualSlots int, canonicalOrder bool) {
	log.Printf("section explicit export unavailable: stage=%s reason=%s expected_slots=%d actual_slots=%d canonical_order=%t", stage, reason, expectedSlots, actualSlots, canonicalOrder)
}

func explicitMatrixFailureStatus(err error) int {
	switch explicitNormalizeFailureClass(err) {
	case "empty_job_template_column", "duplicate_job_template_column", "invalid_client_rectangle", "missing_cell_evidence", "invalid_cell_identity_set", "incomplete_explicit_matrix", "columns_not_in_canonical_order",
		"attribute_definition_response_escaped_trusted_filter", "attribute_definition_is_missing_or_ambiguous", "duplicate_client_row",
		"client_attribute_response_escaped_trusted_filter", "client_attribute_response_escaped_scoped_roster", "duplicate_client_attribute_value":
		return http.StatusBadRequest
	case "attribute_dependencies_unavailable", "attribute_definition_read_failed", "client_attribute_read_failed":
		return http.StatusServiceUnavailable
	case "matrix_contract":
		return http.StatusServiceUnavailable
	default:
		return http.StatusServiceUnavailable
	}
}

func isCanonicalColumnOrder(columns []*exportpb.JobTemplateColumn) bool {
	return columnsInCanonicalOrder(columns)
}

func normalizeExplicitMatrix(ctx context.Context, deps *Deps, resp *exportpb.GetSubscriptionGroupOutcomeExportResponse) (*explicitMatrix, error) {
	expectedSlots := len(resp.GetJobTemplateColumns())
	if resp == nil || !resp.GetSuccess() || resp.GetContext() == nil || expectedSlots == 0 {
		return nil, explicitMatrixFailure{reason: "incomplete explicit matrix", expectedSlots: expectedSlots, actualSlots: 0, canonicalOrder: true}
	}
	if !isCanonicalColumnOrder(resp.GetJobTemplateColumns()) {
		return nil, explicitMatrixFailure{reason: "columns not in canonical order", expectedSlots: expectedSlots, actualSlots: expectedSlots, canonicalOrder: false}
	}
	columnIDs := make(map[string]struct{}, len(resp.GetJobTemplateColumns()))
	for _, column := range resp.GetJobTemplateColumns() {
		id := strings.TrimSpace(column.GetJobTemplateId())
		if id == "" {
			return nil, explicitMatrixFailure{reason: "empty job template column", expectedSlots: expectedSlots, actualSlots: expectedSlots, canonicalOrder: true}
		}
		if _, duplicate := columnIDs[id]; duplicate {
			return nil, explicitMatrixFailure{reason: "duplicate job template column", expectedSlots: expectedSlots, actualSlots: expectedSlots, canonicalOrder: true}
		}
		columnIDs[id] = struct{}{}
	}

	rows := make([]explicitRow, 0, len(resp.GetClientRows()))
	for _, source := range resp.GetClientRows() {
		if source == nil || strings.TrimSpace(source.GetClientId()) == "" || len(source.GetCells()) != len(columnIDs) {
			return nil, explicitMatrixFailure{reason: "invalid client rectangle", expectedSlots: len(columnIDs), actualSlots: len(source.GetCells()), canonicalOrder: true}
		}
		byID := make(map[string]*exportpb.SubscriptionGroupOutcomeCell, len(source.GetCells()))
		for _, cell := range source.GetCells() {
			if cell == nil || cell.GetEnrollmentEvidence() == nil {
				return nil, explicitMatrixFailure{reason: "missing cell evidence", expectedSlots: len(columnIDs), actualSlots: len(source.GetCells()), canonicalOrder: true}
			}
			id := strings.TrimSpace(cell.GetJobTemplateId())
			if _, known := columnIDs[id]; !known || id == "" || byID[id] != nil {
				return nil, explicitMatrixFailure{reason: "invalid cell identity set", expectedSlots: len(columnIDs), actualSlots: len(source.GetCells()), canonicalOrder: true}
			}
			byID[id] = cell
		}
		row := explicitRow{clientID: source.GetClientId(), name: source.GetClientName(), firstName: source.GetClientFirstName(), lastName: source.GetClientLastName()}
		if strings.TrimSpace(row.name) == "" {
			row.name = strings.TrimSpace(strings.Join([]string{row.firstName, row.lastName}, " "))
		}
		row.values = make([]string, 0, len(resp.GetJobTemplateColumns()))
		for _, column := range resp.GetJobTemplateColumns() {
			cell := byID[column.GetJobTemplateId()]
			value := ""
			if cell.ScaledLabel != nil {
				value = strings.TrimSpace(cell.GetScaledLabel())
			}
			if value == "" && cell.ScaledScore != nil {
				value = strconv.FormatFloat(cell.GetScaledScore(), 'f', -1, 64)
			}
			evidence := cell.GetEnrollmentEvidence()
			if outcome_summary.IsNonEnrolledCell(outcome_summary.EnrollmentEvidence{
				HasMarks: evidence.GetHasMarks(), HasPositiveMark: evidence.GetHasPositiveMark(),
			}, value) {
				value = ""
			}
			row.values = append(row.values, value)
		}
		rows = append(rows, row)
	}

	bandValues, err := hydrateExplicitBandValues(ctx, deps, rows)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].group = bandValues[rows[i].clientID]
	}
	rows = orderExplicitRows(rows, deps.Options)
	return &explicitMatrix{context: resp.GetContext(), categories: resp.GetJobCategories(), columns: resp.GetJobTemplateColumns(), rows: rows}, nil
}

func hydrateExplicitBandValues(ctx context.Context, deps *Deps, rows []explicitRow) (map[string]string, error) {
	values := make(map[string]string, len(rows))
	code, module, configured, err := deps.Options.ExportRowBandConfig()
	if err != nil || !configured {
		return values, err
	}
	if deps.ListAttributes == nil || deps.ListClientAttributes == nil {
		return nil, fmt.Errorf("attribute dependencies unavailable")
	}
	definitionResp, err := deps.ListAttributes(ctx, &commonpb.ListAttributesRequest{
		Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
			stringEq("code", code), stringEq("module", module),
			{Field: "active", FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: true}}},
		}},
		Pagination: &commonpb.PaginationRequest{Limit: int32(pageLimit)},
	})
	if err != nil || definitionResp == nil || !definitionResp.GetSuccess() {
		return nil, fmt.Errorf("attribute definition read failed")
	}
	definitionIDs := map[string]struct{}{}
	for _, definition := range definitionResp.GetData() {
		if definition == nil || !definition.GetActive() || definition.GetCode() != code || definition.GetModule() != module || strings.TrimSpace(definition.GetId()) == "" {
			return nil, fmt.Errorf("attribute definition response escaped trusted filter")
		}
		definitionIDs[definition.GetId()] = struct{}{}
	}
	if len(definitionIDs) != 1 {
		return nil, fmt.Errorf("attribute definition is missing or ambiguous")
	}
	attributeID := ""
	for id := range definitionIDs {
		attributeID = id
	}

	allowed := make(map[string]struct{}, len(rows))
	clientIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		if _, duplicate := allowed[row.clientID]; duplicate {
			return nil, fmt.Errorf("duplicate client row")
		}
		allowed[row.clientID] = struct{}{}
		clientIDs = append(clientIDs, row.clientID)
	}
	for start := 0; start < len(clientIDs); start += pageLimit {
		end := start + pageLimit
		if end > len(clientIDs) {
			end = len(clientIDs)
		}
		resp, err := deps.ListClientAttributes(ctx, &clientattributepb.ListClientAttributesRequest{
			Filters: &commonpb.FilterRequest{Filters: []*commonpb.TypedFilter{
				stringEq("attribute_id", attributeID), listIn("client_id", clientIDs[start:end]),
				{Field: "active", FilterType: &commonpb.TypedFilter_BooleanFilter{BooleanFilter: &commonpb.BooleanFilter{Value: true}}},
			}},
			Pagination: &commonpb.PaginationRequest{Limit: int32(pageLimit)},
		})
		if err != nil || resp == nil || !resp.GetSuccess() {
			return nil, fmt.Errorf("client attribute read failed")
		}
		for _, value := range resp.GetData() {
			if value == nil || !value.GetActive() || value.GetAttributeId() != attributeID {
				return nil, fmt.Errorf("client attribute response escaped trusted filter")
			}
			if _, ok := allowed[value.GetClientId()]; !ok {
				return nil, fmt.Errorf("client attribute response escaped scoped roster")
			}
			if _, duplicate := values[value.GetClientId()]; duplicate {
				return nil, fmt.Errorf("duplicate client attribute value")
			}
			values[value.GetClientId()] = strings.TrimSpace(value.GetValue())
		}
	}
	return values, nil
}

func orderExplicitRows(rows []explicitRow, options outcome_summary.Options) []explicitRow {
	groups := make(map[string][]explicitRow)
	groupKeys := make([]string, 0)
	for _, row := range rows {
		if _, exists := groups[row.group]; !exists {
			groupKeys = append(groupKeys, row.group)
		}
		groups[row.group] = append(groups[row.group], row)
	}
	sort.SliceStable(groupKeys, func(i, j int) bool {
		a, b := groupKeys[i], groupKeys[j]
		if a == "" || b == "" {
			return a != ""
		}
		ar, aok := options.Row.GroupValueRank(a)
		br, bok := options.Row.GroupValueRank(b)
		if aok != bok {
			return aok
		}
		if aok && ar != br {
			return ar < br
		}
		return strings.ToLower(a) < strings.ToLower(b)
	})
	for key := range groups {
		sort.SliceStable(groups[key], func(i, j int) bool {
			left, right := groups[key][i], groups[key][j]
			leftValue := explicitSortValue(left, options.Row.SortField)
			rightValue := explicitSortValue(right, options.Row.SortField)
			if leftValue == rightValue {
				return left.clientID < right.clientID
			}
			if options.Row.Direction() == "desc" {
				return leftValue > rightValue
			}
			return leftValue < rightValue
		})
	}

	banded := false
	if _, _, configured, _ := options.ExportRowBandConfig(); configured {
		banded = true
	}
	out := make([]explicitRow, 0, len(rows)+len(groupKeys)*2+1)
	sequence := 0
	for _, key := range groupKeys {
		if banded {
			out = append(out, explicitRow{name: key, band: true})
		}
		for _, row := range groups[key] {
			sequence++
			row.name = fmt.Sprintf("[%d] %s", sequence, row.name)
			out = append(out, row)
		}
		if banded {
			out = append(out, explicitRow{blank: true})
		}
	}
	if !banded {
		out = append(out, explicitRow{blank: true})
	}
	return out
}

func explicitSortValue(row explicitRow, field string) string {
	switch strings.TrimSpace(field) {
	case "first_name":
		return strings.ToLower(strings.TrimSpace(row.firstName))
	case "last_name":
		return strings.ToLower(strings.TrimSpace(row.lastName))
	default:
		return strings.ToLower(strings.TrimSpace(row.name))
	}
}

func writeExplicitCSV(w http.ResponseWriter, deps *Deps, matrix *explicitMatrix, period string) {
	prefix := slug(deps.Labels.Section.Title)
	if prefix == "none" {
		prefix = "outcomes"
	}
	filename := prefix + "-" + slug(matrix.context.GetSubscriptionGroupName()) + "-" + slug(period)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`.csv"`)
	cw := csv.NewWriter(w)
	header := make([]string, 0, len(matrix.columns)+1)
	header = append(header, csvSafe(deps.Labels.Section.ClientColumn))
	for _, column := range matrix.columns {
		name := strings.TrimSpace(column.GetDisplayName())
		if name == "" {
			name = column.GetJobTemplateId()
		}
		header = append(header, csvSafe(name))
	}
	if err := cw.Write(header); err != nil {
		log.Printf("section explicit export: write header: %v", err)
		return
	}
	for _, row := range matrix.rows {
		record := make([]string, len(matrix.columns)+1)
		record[0] = csvSafe(row.name)
		if !row.band && !row.blank {
			for i, value := range row.values {
				record[i+1] = csvSafe(value)
			}
		}
		if err := cw.Write(record); err != nil {
			log.Printf("section explicit export: write row: %v", err)
			return
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("section explicit export: flush: %v", err)
	}
}

// csvSafe neutralizes spreadsheet formula/DDE injection: a cell whose text
// begins with a formula trigger is evaluated by Excel/Sheets on open.
// encoding/csv quoting does NOT prevent this. Prefix such values with a tab so
// the client treats them as literal text (the OWASP-recommended
// neutralization; the tab is invisible in the rendered cell). Decodes the
// first RUNE, not byte: the trigger set includes the full-width ＝＋－＠ forms
// (U+FF1D/0B/0D/20) Excel also honors, plus LF alongside TAB/CR as trimmable
// prefixes (current OWASP guidance; kept in sync with
// outcome_matrix/list/export.go). Empty values pass through untouched.
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
