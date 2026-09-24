package subscription_group_document

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/json/v2"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
)

const matrixManifestFilename = "subscription-group-outcome-matrix-single-period-11-v1.manifest.json"
const clientPhaseManifestFilename = "subscription-group-client-phase-outcome-report-v1.manifest.json"

const wordprocessingMLNamespace = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

const (
	maxTemplateInputBytes     int64 = 10 << 20
	maxTemplateEntries              = 2000
	maxTemplateEntryBytes     int64 = 64 << 20
	maxTemplateAggregateBytes int64 = 256 << 20
)

//go:embed subscription-group-outcome-matrix-single-period-11-v1.manifest.json subscription-group-client-phase-outcome-report-v1.manifest.json
var manifestFiles embed.FS

var completeTemplateToken = regexp.MustCompile(`\{\{[^{}]*\}\}`)

type manifestNode struct {
	Scalars []string `json:"scalars"`
	// OptionalScalars may be omitted by a template; when present they obey the
	// same exactly-once and placement rules as Scalars.
	OptionalScalars []string `json:"optional_scalars,omitempty"`
	// Optional marks a loop a template may omit entirely (both markers and
	// every token in its scope). Only meaningful for nested manifest loops.
	Optional bool                    `json:"optional,omitempty"`
	Loops    map[string]manifestNode `json:"loops"`
}

type renderManifest struct {
	Profile struct {
		Key                     string `json:"key"`
		BindingJobCategoryScope string `json:"binding_job_category_scope"`
		JobTemplateSlots        int    `json:"job_template_slots"`
		PeriodCardinality       int    `json:"period_cardinality"`
		RowLoop                 string `json:"row_loop"`
		SubjectOrder            string `json:"subject_order"`
		SubjectOrderOwner       string `json:"subject_order_owner"`
		ColumnMismatch          string `json:"column_mismatch"`
		ProfileMismatch         string `json:"profile_mismatch"`
		MissingOutcomeDisplay   string `json:"missing_outcome_display"`
	} `json:"profile"`
	Scalars []string `json:"scalars"`
	// OptionalScalars are root tokens a template may omit (client phase profile).
	OptionalScalars []string                `json:"optional_scalars,omitempty"`
	Loops           map[string]manifestNode `json:"loops"`
	StyleValues     struct {
		RowBold []string `json:"row_bold"`
		FillHex []string `json:"fill_hex"`
		TextHex []string `json:"text_hex"`
	} `json:"style_values"`
}

// TemplateContractError marks a trusted template/profile incompatibility. The
// HTTP layer maps it to a fail-loud 503 before invoking the renderer.
type TemplateContractError struct{ reason string }

func (e *TemplateContractError) Error() string {
	return "subscription group document template contract: " + e.reason
}
func (*TemplateContractError) TemplateContractError() bool { return true }

func contractError(format string, args ...any) error {
	return &TemplateContractError{reason: fmt.Sprintf(format, args...)}
}

// ManifestBytes returns a defensive copy of the generic, package-owned
// manifest. It intentionally contains no docs-only evidence/category metadata.
func ManifestBytes(profileValue bindingpb.RenderProfile) ([]byte, error) {
	profile, ok := LookupProfile(profileValue)
	if !ok {
		return nil, contractError("unsupported render profile")
	}
	filename := ""
	switch profile.Key {
	case SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key:
		filename = matrixManifestFilename
	case SubscriptionGroupClientPhaseOutcomeReportV1Key:
		filename = clientPhaseManifestFilename
	default:
		return nil, contractError("unsupported render profile")
	}
	raw, err := manifestFiles.ReadFile(filename)
	if err != nil {
		return nil, contractError("read embedded manifest: %v", err)
	}
	return append([]byte(nil), raw...), nil
}

func loadManifest(profileValue bindingpb.RenderProfile) (Profile, renderManifest, error) {
	profile, ok := LookupProfile(profileValue)
	if !ok {
		return Profile{}, renderManifest{}, contractError("unsupported render profile")
	}
	raw, err := ManifestBytes(profileValue)
	if err != nil {
		return Profile{}, renderManifest{}, err
	}
	var manifest renderManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return Profile{}, renderManifest{}, contractError("decode embedded manifest: %v", err)
	}
	if manifest.Profile.Key != profile.Key || manifest.Profile.JobTemplateSlots != profile.JobTemplateSlots {
		return Profile{}, renderManifest{}, contractError("embedded manifest does not match the registered profile")
	}
	switch profile.Key {
	case SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key:
		if manifest.Profile.BindingJobCategoryScope != "exact_required" || manifest.Profile.PeriodCardinality != 1 || manifest.Profile.RowLoop != "rows" {
			return Profile{}, renderManifest{}, contractError("embedded matrix manifest does not match the registered profile")
		}
	case SubscriptionGroupClientPhaseOutcomeReportV1Key:
		if manifest.Profile.BindingJobCategoryScope != "null_required" || manifest.Profile.PeriodCardinality != 0 || manifest.Profile.RowLoop != "jobs" {
			return Profile{}, renderManifest{}, contractError("embedded client phase manifest does not match the registered profile")
		}
		if err := validateClientPhaseManifestShape(manifest); err != nil {
			return Profile{}, renderManifest{}, err
		}
	default:
		return Profile{}, renderManifest{}, contractError("unsupported render profile")
	}
	return profile, manifest, nil
}

type validationLimits struct {
	input, entry, aggregate int64
	entries                 int
}

var productionValidationLimits = validationLimits{
	input: maxTemplateInputBytes, entries: maxTemplateEntries,
	entry: maxTemplateEntryBytes, aggregate: maxTemplateAggregateBytes,
}

// ValidateTemplate validates compressed and actual expansion bounds, archive
// paths/identity, and the exact text/style placeholder manifest.
func ValidateTemplate(profileValue bindingpb.RenderProfile, docx []byte) error {
	return validateTemplateWithLimits(profileValue, docx, productionValidationLimits)
}

func validateTemplateWithLimits(profileValue bindingpb.RenderProfile, docx []byte, limits validationLimits) error {
	_, manifest, err := loadManifest(profileValue)
	if err != nil {
		return err
	}
	if int64(len(docx)) > limits.input {
		return contractError("compressed input exceeds %d bytes", limits.input)
	}
	reader, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		return contractError("open DOCX archive: %v", err)
	}
	if len(reader.File) > limits.entries {
		return contractError("archive contains more than %d entries", limits.entries)
	}

	seen := make(map[string]struct{}, len(reader.File))
	parts := make(map[string][]byte)
	var aggregate int64
	for _, file := range reader.File {
		if file.NonUTF8 {
			return contractError("archive entry name is not UTF-8")
		}
		canonical, err := canonicalArchiveName(file.Name)
		if err != nil {
			return err
		}
		folded := strings.ToLower(canonical)
		if _, duplicate := seen[folded]; duplicate {
			return contractError("duplicate archive part %q", canonical)
		}
		seen[folded] = struct{}{}

		rc, err := file.Open()
		if err != nil {
			return contractError("open archive part %q: %v", canonical, err)
		}
		var sink io.Writer = io.Discard
		var content bytes.Buffer
		if isXMLPart(canonical) {
			sink = &content
		}
		actual, copyErr := io.Copy(sink, io.LimitReader(rc, limits.entry+1))
		closeErr := rc.Close()
		if copyErr != nil {
			return contractError("read archive part %q: %v", canonical, copyErr)
		}
		if closeErr != nil {
			return contractError("close archive part %q: %v", canonical, closeErr)
		}
		if actual > limits.entry {
			return contractError("archive part %q exceeds %d actual bytes", canonical, limits.entry)
		}
		if uint64(actual) != file.UncompressedSize64 {
			return contractError("archive part %q actual size does not match its directory entry", canonical)
		}
		aggregate += actual
		if aggregate > limits.aggregate {
			return contractError("archive actual expansion exceeds %d bytes", limits.aggregate)
		}
		if isXMLPart(canonical) {
			parts[canonical] = append([]byte(nil), content.Bytes()...)
		}
	}
	if _, ok := parts["[Content_Types].xml"]; !ok {
		return contractError("DOCX is missing [Content_Types].xml")
	}
	if _, ok := parts["word/document.xml"]; !ok {
		return contractError("DOCX is missing word/document.xml")
	}
	return validateManifestTokens(parts, manifest)
}

func canonicalArchiveName(name string) (string, error) {
	if !utf8.ValidString(name) || name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, "\\") {
		return "", contractError("unsafe archive path %q", name)
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", contractError("unsafe archive path %q", name)
		}
	}
	directory := strings.HasSuffix(name, "/")
	canonical := strings.TrimSuffix(name, "/")
	if canonical == "" || path.Clean(canonical) != canonical {
		return "", contractError("non-canonical archive path %q", name)
	}
	for _, component := range strings.Split(canonical, "/") {
		if component == "" || component == "." || component == ".." || strings.Contains(component, ":") {
			return "", contractError("unsafe archive path %q", name)
		}
	}
	if directory {
		return canonical + "/", nil
	}
	return canonical, nil
}

func isXMLPart(name string) bool {
	lower := strings.ToLower(strings.TrimSuffix(name, "/"))
	return strings.HasSuffix(lower, ".xml") || strings.HasSuffix(lower, ".rels")
}

type tokenOccurrence struct {
	key, part, kind    string
	element, attribute xml.Name
	table, row         int
	tableDepth         int
	order, paragraph   int
	ancestry           []xml.Name
}

type tableScanContext struct {
	id, depth, nextRow int
}

type rowScanContext struct {
	table, row, tableDepth int
}

type rowPosition struct {
	table, row int
}

type rowLocation struct {
	part       string
	table, row int
}

type partScanner struct {
	part          string
	stack         []xml.Name
	tableStack    []tableScanContext
	rowStack      []rowScanContext
	paragraphs    []int
	nextTable     int
	nextParagraph int
	rowText       map[rowPosition]*strings.Builder
	paragraphText map[int]*strings.Builder
	occurrences   []tokenOccurrence
	completeCount int
}

func validateManifestTokens(parts map[string][]byte, manifest renderManifest) error {
	if manifest.Profile.Key == SubscriptionGroupClientPhaseOutcomeReportV1Key {
		return validateClientPhaseManifestTokens(parts, manifest)
	}
	return validateMatrixManifestTokens(parts, manifest)
}

func validateMatrixManifestTokens(parts map[string][]byte, manifest renderManifest) error {
	var occurrences []tokenOccurrence
	rowTexts := make(map[rowLocation]string)
	completeCount := 0
	for part, body := range parts {
		scanner := &partScanner{
			part:          part,
			completeCount: len(completeTemplateToken.FindAll(body, -1)),
			rowText:       make(map[rowPosition]*strings.Builder),
			paragraphText: make(map[int]*strings.Builder),
		}
		if err := scanner.scan(body); err != nil {
			return err
		}
		completeCount += scanner.completeCount
		occurrences = append(occurrences, scanner.occurrences...)
		for position, value := range scanner.rowText {
			rowTexts[rowLocation{part: part, table: position.table, row: position.row}] = value.String()
		}
	}
	if completeCount != len(occurrences) {
		return contractError("template token is split, partial, or outside a supported XML node")
	}

	rootKeys := make(map[string]struct{}, len(manifest.Scalars))
	for _, key := range manifest.Scalars {
		rootKeys[key] = struct{}{}
	}
	rowNode, ok := manifest.Loops[manifest.Profile.RowLoop]
	if !ok || len(rowNode.Loops) != 0 {
		return contractError("manifest row loop is missing or nested")
	}
	rowKeys := make(map[string]struct{}, len(rowNode.Scalars))
	for _, key := range rowNode.Scalars {
		rowKeys[key] = struct{}{}
	}

	byKey := make(map[string][]tokenOccurrence)
	for _, occurrence := range occurrences {
		key := occurrence.key
		if occurrence.part != "word/document.xml" {
			return contractError("template token %q is outside word/document.xml", key)
		}
		if key == "#"+manifest.Profile.RowLoop || key == "/"+manifest.Profile.RowLoop {
			if err := validateMarkerPlacement(occurrence); err != nil {
				return contractError("row loop token %q is misplaced", key)
			}
			byKey[key] = append(byKey[key], occurrence)
			continue
		}
		if _, root := rootKeys[key]; root {
			if err := validateRootTokenPlacement(occurrence); err != nil {
				return err
			}
		} else if _, row := rowKeys[key]; row {
			if err := validateRowTokenPlacement(occurrence); err != nil {
				return err
			}
		} else {
			return contractError("unexpected template token %q", key)
		}
		byKey[key] = append(byKey[key], occurrence)
	}

	start := byKey["#"+manifest.Profile.RowLoop]
	end := byKey["/"+manifest.Profile.RowLoop]
	if len(start) != 1 || len(end) != 1 || start[0].kind != "text" || end[0].kind != "text" ||
		start[0].part != end[0].part || start[0].table <= 0 || start[0].tableDepth != 1 ||
		start[0].table != end[0].table || end[0].tableDepth != 1 || start[0].row <= 0 ||
		end[0].row != start[0].row+2 {
		return contractError("row loop must have one start row, one template row, and one end row")
	}
	startLocation := occurrenceRowLocation(start[0])
	endLocation := occurrenceRowLocation(end[0])
	if strings.TrimSpace(rowTexts[startLocation]) != "{{#"+manifest.Profile.RowLoop+"}}" ||
		strings.TrimSpace(rowTexts[endLocation]) != "{{/"+manifest.Profile.RowLoop+"}}" {
		return contractError("row loop marker rows must contain only their marker text")
	}
	templatePart, templateTable, templateRow := start[0].part, start[0].table, start[0].row+1
	for key := range rootKeys {
		values := byKey[key]
		if len(values) != 1 || values[0].kind != "text" ||
			(values[0].part == templatePart && values[0].table == templateTable && values[0].row == templateRow) {
			return contractError("root token %q is missing, duplicated, or misplaced", key)
		}
	}
	for key := range rowKeys {
		values := byKey[key]
		if len(values) != 1 || values[0].part != templatePart || values[0].table != templateTable ||
			values[0].row != templateRow || values[0].tableDepth != 1 {
			return contractError("row token %q is missing, duplicated, or outside the template row", key)
		}
	}
	return nil
}

type manifestLoopScope struct {
	name   string
	parent string
	node   manifestNode
}

type manifestScalarScope struct {
	loop     string
	optional bool
}

var clientPhaseRootScalars = []string{
	"school_name", "academic_year", "student_name", "grade_level", "section_name",
	"client_reference", "adviser", "printed_by", "printed_at", "phase_name",
}

var clientPhaseJobScalars = []string{
	"job_name", "staff_name", "phase_grade", "phase_comment",
	"phase_total", "phase_maximum", "progress_to_date_total", "progress_to_date_maximum",
}

// clientPhaseJobOptionalScalars: the category label and the per-job copies of
// the document identity (for a heading repeated on every job's page) are
// available but not every layout prints them.
var clientPhaseJobOptionalScalars = []string{
	"job_category_name",
	"page_student_name", "page_grade_level", "page_section_name",
	"page_academic_year", "page_client_reference", "page_adviser", "page_plan_label",
}

var clientPhaseAssessmentScalars = []string{
	"assessment_name", "achievement_level", "comment",
}

var clientPhaseAssessmentOptionalScalars = []string{"assessment_maximum"}

var clientPhaseRatingDescriptionScalars = []string{
	"rating_label", "minimum", "maximum", "output_value", "description",
}

// Period summaries (optional): per-job rows and per-activity rows of the
// configured summary categories, one value per period (phase order 1..3), plus
// the activity-category job's per-period summary and transmuted label.
var clientPhaseRootOptionalScalars = []string{
	"summary_average_period_1", "summary_average_period_2", "summary_average_period_3",
	"summary_transmuted_period_1", "summary_transmuted_period_2", "summary_transmuted_period_3",
}
var clientPhaseSummaryJobScalars = []string{"summary_job_name", "summary_period_1", "summary_period_2", "summary_period_3"}
var clientPhaseSummaryTaskScalars = []string{"summary_task_name", "summary_task_period_1", "summary_task_period_2", "summary_task_period_3"}

var clientPhaseOutcomeSectionScalars = []string{"section_code"}
var clientPhaseOutcomeRowScalars = []string{"row_name", "total"}
var clientPhaseOutcomeCellScalars = []string{"period_name", "task_name", "value"}

func validateClientPhaseManifestShape(manifest renderManifest) error {
	jobs, jobsOK := manifest.Loops["jobs"]
	sections, sectionsOK := manifest.Loops["outcome_sections"]
	assessments, assessmentsOK := jobs.Loops["assessments"]
	descriptions, descriptionsOK := assessments.Loops["rating_descriptions"]
	rows, rowsOK := sections.Loops["rows"]
	cells, cellsOK := rows.Loops["cells"]
	summaryJobs, summaryJobsOK := manifest.Loops["summary_jobs"]
	summaryTasks, summaryTasksOK := manifest.Loops["summary_tasks"]
	if !summaryJobsOK || !summaryTasksOK || !summaryJobs.Optional || !summaryTasks.Optional ||
		len(summaryJobs.Loops) != 0 || len(summaryTasks.Loops) != 0 ||
		!sameStrings(summaryJobs.Scalars, clientPhaseSummaryJobScalars) || len(summaryJobs.OptionalScalars) != 0 ||
		!sameStrings(summaryTasks.Scalars, clientPhaseSummaryTaskScalars) || len(summaryTasks.OptionalScalars) != 0 ||
		!sameStrings(manifest.OptionalScalars, clientPhaseRootOptionalScalars) {
		return contractError("embedded client phase manifest does not match the registered repeated-table contract")
	}
	if !jobsOK || !sectionsOK || !assessmentsOK || !descriptionsOK || !rowsOK || !cellsOK || len(manifest.Loops) != 4 ||
		len(jobs.Loops) != 1 || len(assessments.Loops) != 1 || len(descriptions.Loops) != 0 ||
		len(sections.Loops) != 1 || len(rows.Loops) != 1 || len(cells.Loops) != 0 ||
		!sameStrings(manifest.Scalars, clientPhaseRootScalars) || len(manifest.Scalars) == 0 ||
		!sameStrings(jobs.Scalars, clientPhaseJobScalars) || !sameStrings(jobs.OptionalScalars, clientPhaseJobOptionalScalars) ||
		!sameStrings(assessments.Scalars, clientPhaseAssessmentScalars) || !sameStrings(assessments.OptionalScalars, clientPhaseAssessmentOptionalScalars) ||
		!sameStrings(descriptions.Scalars, clientPhaseRatingDescriptionScalars) || len(descriptions.OptionalScalars) != 0 ||
		!sameStrings(sections.Scalars, clientPhaseOutcomeSectionScalars) || len(sections.OptionalScalars) != 0 ||
		!sameStrings(rows.Scalars, clientPhaseOutcomeRowScalars) || len(rows.OptionalScalars) != 0 ||
		!sameStrings(cells.Scalars, clientPhaseOutcomeCellScalars) || len(cells.OptionalScalars) != 0 ||
		// Only the rating_descriptions and outcome_sections loops may be omitted.
		jobs.Optional || assessments.Optional || !descriptions.Optional ||
		!sections.Optional || rows.Optional || cells.Optional {
		return contractError("embedded client phase manifest does not match the registered repeated-table contract")
	}
	return nil
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	counts := make(map[string]int, len(got))
	for _, value := range got {
		counts[value]++
	}
	for _, value := range want {
		counts[value]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

func validateClientPhaseManifestTokens(parts map[string][]byte, manifest renderManifest) error {
	loops := make(map[string]manifestLoopScope)
	if err := flattenManifestLoops(manifest.Loops, "", loops); err != nil {
		return err
	}
	scalarScopes := make(map[string]manifestScalarScope)
	for index, list := range [][]string{manifest.Scalars, manifest.OptionalScalars} {
		for _, key := range list {
			if strings.TrimSpace(key) == "" {
				return contractError("manifest contains an empty scalar key")
			}
			if _, duplicate := scalarScopes[key]; duplicate {
				return contractError("manifest scalar %q is declared more than once", key)
			}
			scalarScopes[key] = manifestScalarScope{optional: index == 1}
		}
	}
	for _, scope := range loops {
		for index, list := range [][]string{scope.node.Scalars, scope.node.OptionalScalars} {
			for _, key := range list {
				if strings.TrimSpace(key) == "" {
					return contractError("manifest contains an empty scalar key")
				}
				if _, duplicate := scalarScopes[key]; duplicate {
					return contractError("manifest scalar %q is declared more than once", key)
				}
				scalarScopes[key] = manifestScalarScope{loop: scope.name, optional: index == 1}
			}
		}
	}

	byKey := make(map[string][]tokenOccurrence)
	rowTexts := make(map[rowLocation]string)
	paragraphTexts := make(map[string]map[int]string)
	completeCount := 0
	for part, body := range parts {
		scanner := &partScanner{
			part: part, completeCount: len(completeTemplateToken.FindAll(body, -1)),
			rowText: make(map[rowPosition]*strings.Builder), paragraphText: make(map[int]*strings.Builder),
		}
		if err := scanner.scan(body); err != nil {
			return err
		}
		completeCount += scanner.completeCount
		for _, occurrence := range scanner.occurrences {
			if occurrence.part != "word/document.xml" {
				// Header/footer parts may carry root scalars only (e.g. the
				// printed-by line); every loop token stays in the body.
				scope, known := scalarScopes[occurrence.key]
				if !headerFooterPart.MatchString(occurrence.part) || !known || scope.loop != "" ||
					occurrence.kind != "text" || !isWordprocessingMLName(occurrence.element, "t") {
					return contractError("template token %q is outside word/document.xml", occurrence.key)
				}
			}
			byKey[occurrence.key] = append(byKey[occurrence.key], occurrence)
		}
		for position, value := range scanner.rowText {
			rowTexts[rowLocation{part: part, table: position.table, row: position.row}] = value.String()
		}
		texts := make(map[int]string, len(scanner.paragraphText))
		for id, value := range scanner.paragraphText {
			texts[id] = value.String()
		}
		paragraphTexts[part] = texts
	}
	occurrenceCount := 0
	for _, values := range byKey {
		occurrenceCount += len(values)
	}
	if occurrenceCount != completeCount {
		return contractError("template token is split, partial, or outside a supported XML node")
	}

	for name := range loops {
		for _, marker := range []string{"#" + name, "/" + name} {
			if _, duplicate := scalarScopes[marker]; duplicate {
				return contractError("manifest scalar conflicts with loop marker %q", marker)
			}
		}
	}
	for key, values := range byKey {
		// Document-level (root) scalars may repeat (e.g. identity on a cover and on a
		// summary page); every occurrence is placement-checked below. Loop tokens and
		// markers stay exactly-once.
		if scope, root := scalarScopes[key]; root && scope.loop == "" && len(values) > 1 {
			continue
		}
		if len(values) != 1 {
			return contractError("template token %q must occur exactly once", key)
		}
		if isClientPhaseFixedScalar(key) {
			scalarScopes[key] = manifestScalarScope{}
			continue
		}
		if _, ok := scalarScopes[key]; ok {
			continue
		}
		if _, isLoopMarker := loops[strings.TrimPrefix(strings.TrimPrefix(key, "#"), "/")]; isLoopMarker && (strings.HasPrefix(key, "#") || strings.HasPrefix(key, "/")) {
			continue
		}
		return contractError("unexpected template token %q", key)
	}
	// A loop is omitted when neither marker is present and it (or an omitted
	// ancestor) is optional. Every token of an omitted loop's scope must then
	// be absent too; required tokens of present scopes must be present.
	omitted := make(map[string]bool, len(loops))
	var isOmitted func(name string) bool
	isOmitted = func(name string) bool {
		if value, done := omitted[name]; done {
			return value
		}
		scope := loops[name]
		absent := len(byKey["#"+name]) == 0 && len(byKey["/"+name]) == 0
		result := absent && (scope.node.Optional || (scope.parent != "" && isOmitted(scope.parent)))
		omitted[name] = result
		return result
	}
	for name := range loops {
		isOmitted(name)
	}
	for key, scope := range scalarScopes {
		present := len(byKey[key]) >= 1
		if scope.loop != "" && omitted[scope.loop] {
			if present {
				return contractError("loop token %q is outside loop %q", key, scope.loop)
			}
			continue
		}
		if !present && !scope.optional {
			return contractError("template scalar %q is missing", key)
		}
	}

	loopBounds := make(map[string][2]tokenOccurrence, len(loops))
	loopKinds := make(map[string]string, len(loops))
	for name := range loops {
		if omitted[name] {
			continue
		}
		starts, ends := byKey["#"+name], byKey["/"+name]
		if len(starts) != 1 || len(ends) != 1 {
			return contractError("loop %q must have one start and one end marker", name)
		}
		start, end := starts[0], ends[0]
		if start.part != end.part || start.order >= end.order {
			return contractError("loop %q markers have invalid order or parts", name)
		}
		kind, err := validateClientPhaseLoopMarkers(start, end, rowTexts, paragraphTexts)
		if err != nil {
			return contractError("loop %q: %v", name, err)
		}
		loopBounds[name] = [2]tokenOccurrence{start, end}
		loopKinds[name] = kind
	}
	// The flat map traversal above is unordered, so perform parent containment
	// after collecting every loop's bounds.
	for name, scope := range loops {
		if scope.parent == "" || omitted[name] {
			continue
		}
		child, parent := loopBounds[name], loopBounds[scope.parent]
		if child[0].order <= parent[0].order || child[1].order >= parent[1].order {
			return contractError("nested loop %q is outside parent loop %q", name, scope.parent)
		}
	}

	for key, scalarScope := range scalarScopes {
		if len(byKey[key]) == 0 {
			continue // optional token, or a token of an omitted loop (checked above)
		}
		value := byKey[key][0]
		if scalarScope.loop == "" {
			for _, occurrence := range byKey[key] {
				if occurrence.part != "word/document.xml" {
					continue // header/footer text token, checked when collected
				}
				if err := validateRootTokenPlacement(occurrence); err != nil {
					return err
				}
				for name, bounds := range loopBounds {
					if occurrence.order > bounds[0].order && occurrence.order < bounds[1].order {
						return contractError("root token %q is inside loop %q", key, name)
					}
				}
			}
			continue
		}
		bounds := loopBounds[scalarScope.loop]
		if value.order <= bounds[0].order || value.order >= bounds[1].order {
			return contractError("loop token %q is outside loop %q", key, scalarScope.loop)
		}
		if loopKinds[scalarScope.loop] == "rows" {
			if err := validateRowTokenPlacement(value); err != nil {
				return err
			}
			if value.table != bounds[0].table || value.row != bounds[0].row+1 || value.tableDepth != bounds[0].tableDepth {
				return contractError("row token %q is outside the template row for loop %q", key, scalarScope.loop)
			}
		} else if err := validateRootTokenPlacement(value); err != nil {
			return err
		}
	}
	return nil
}

// Coded outcome paths are selected by each template from the typed projection.
// Their category/template/criterion codes are data, so the shared profile must
// validate the path shape without naming any vertical or academic year.
var headerFooterPart = regexp.MustCompile(`^word/(header|footer)[0-9]*\.xml$`)

func isClientPhaseFixedScalar(key string) bool {
	parts := strings.Split(key, ".")
	validCell := len(parts) == 7 && parts[0] == "outcome_cells" && parts[6] == "numeric_value"
	validTotal := len(parts) == 5 && parts[0] == "outcome_totals" && parts[4] == "numeric_value"
	// Template-independent family: category.criterion.activity / category.criterion.
	validCategoryCell := len(parts) == 5 && parts[0] == "category_cells" && parts[4] == "numeric_value"
	validCategoryTotal := len(parts) == 4 && parts[0] == "category_totals" && parts[3] == "numeric_value"
	if !validCell && !validTotal && !validCategoryCell && !validCategoryTotal {
		return false
	}
	for _, part := range parts[1 : len(parts)-1] {
		if strings.TrimSpace(part) == "" || strings.ContainsAny(part, "{} \t\r\n") {
			return false
		}
	}
	return true
}

func flattenManifestLoops(nodes map[string]manifestNode, parent string, out map[string]manifestLoopScope) error {
	for name, node := range nodes {
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "#/{}") {
			return contractError("manifest contains an invalid loop name")
		}
		if _, duplicate := out[name]; duplicate {
			return contractError("manifest loop %q is declared more than once", name)
		}
		out[name] = manifestLoopScope{name: name, parent: parent, node: node}
		if err := flattenManifestLoops(node.Loops, name, out); err != nil {
			return err
		}
	}
	return nil
}

func validateClientPhaseLoopMarkers(start, end tokenOccurrence, rowTexts map[rowLocation]string, paragraphTexts map[string]map[int]string) (string, error) {
	if start.kind != "text" || end.kind != "text" || !isWordprocessingMLName(start.element, "t") || !isWordprocessingMLName(end.element, "t") {
		return "", contractError("loop markers must be text")
	}
	if start.table == 0 && end.table == 0 {
		if !matchesWordprocessingMLPath(start.ancestry, "document", "body", "p", "r", "t") ||
			!matchesWordprocessingMLPath(end.ancestry, "document", "body", "p", "r", "t") || start.paragraph == 0 || end.paragraph == 0 {
			return "", contractError("body loop markers must be standalone body paragraphs")
		}
		texts := paragraphTexts[start.part]
		if strings.TrimSpace(texts[start.paragraph]) != "{{#"+strings.TrimPrefix(start.key, "#")+"}}" ||
			strings.TrimSpace(texts[end.paragraph]) != "{{/"+strings.TrimPrefix(end.key, "/")+"}}" {
			return "", contractError("body loop marker paragraphs must contain only their marker")
		}
		return "body", nil
	}
	if start.table <= 0 || start.table != end.table || start.tableDepth != 1 || end.tableDepth != 1 ||
		start.row <= 0 || end.row != start.row+2 ||
		!matchesWordprocessingMLPath(start.ancestry, "document", "body", "tbl", "tr", "tc", "p", "r", "t") ||
		!matchesWordprocessingMLPath(end.ancestry, "document", "body", "tbl", "tr", "tc", "p", "r", "t") {
		return "", contractError("table loop must use one start row, one template row, and one end row")
	}
	if strings.TrimSpace(rowTexts[occurrenceRowLocation(start)]) != "{{#"+strings.TrimPrefix(start.key, "#")+"}}" ||
		strings.TrimSpace(rowTexts[occurrenceRowLocation(end)]) != "{{/"+strings.TrimPrefix(end.key, "/")+"}}" {
		return "", contractError("table loop marker rows must contain only their marker text")
	}
	return "rows", nil
}

func occurrenceRowLocation(value tokenOccurrence) rowLocation {
	return rowLocation{part: value.part, table: value.table, row: value.row}
}

func validateMarkerPlacement(value tokenOccurrence) error {
	if value.kind != "text" || !isWordprocessingMLName(value.element, "t") ||
		!matchesWordprocessingMLPath(value.ancestry, "document", "body", "tbl", "tr", "tc", "p", "r", "t") {
		return contractError("row loop token %q is misplaced", value.key)
	}
	return nil
}

func validateRootTokenPlacement(value tokenOccurrence) error {
	if value.kind != "text" || !isWordprocessingMLName(value.element, "t") {
		return contractError("root token %q is misplaced", value.key)
	}
	if matchesWordprocessingMLPath(value.ancestry, "document", "body", "p", "r", "t") ||
		matchesWordprocessingMLPath(value.ancestry, "document", "body", "tbl", "tr", "tc", "p", "r", "t") {
		return nil
	}
	return contractError("root token %q is outside a renderer-visited paragraph", value.key)
}

func validateRowTokenPlacement(value tokenOccurrence) error {
	switch {
	case value.key == "row_bold":
		if value.kind != "attribute" || !isWordprocessingMLName(value.element, "b") || !isWordprocessingMLName(value.attribute, "val") ||
			!matchesWordprocessingMLPath(value.ancestry, "document", "body", "tbl", "tr", "tc", "p", "r", "rPr", "b") {
			return contractError("style token %q is misplaced", value.key)
		}
	case strings.HasSuffix(value.key, "_fill_hex"):
		if value.kind != "attribute" || !isWordprocessingMLName(value.element, "shd") || !isWordprocessingMLName(value.attribute, "fill") ||
			!matchesWordprocessingMLPath(value.ancestry, "document", "body", "tbl", "tr", "tc", "tcPr", "shd") {
			return contractError("style token %q is misplaced", value.key)
		}
	case strings.HasSuffix(value.key, "_text_hex"):
		if value.kind != "attribute" || !isWordprocessingMLName(value.element, "color") || !isWordprocessingMLName(value.attribute, "val") ||
			!matchesWordprocessingMLPath(value.ancestry, "document", "body", "tbl", "tr", "tc", "p", "r", "rPr", "color") {
			return contractError("style token %q is misplaced", value.key)
		}
	default:
		if value.kind != "text" || !isWordprocessingMLName(value.element, "t") ||
			!matchesWordprocessingMLPath(value.ancestry, "document", "body", "tbl", "tr", "tc", "p", "r", "t") {
			return contractError("text token %q is misplaced", value.key)
		}
	}
	return nil
}

func matchesWordprocessingMLPath(ancestry []xml.Name, locals ...string) bool {
	if len(ancestry) != len(locals) {
		return false
	}
	for index, local := range locals {
		if !isWordprocessingMLName(ancestry[index], local) {
			return false
		}
	}
	return true
}

func isWordprocessingMLName(name xml.Name, local string) bool {
	return name.Space == wordprocessingMLNamespace && name.Local == local
}

func (s *partScanner) scan(body []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(body))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return contractError("parse XML part %q: %v", s.part, err)
		}
		switch value := token.(type) {
		case xml.StartElement:
			var parent xml.Name
			if len(s.stack) != 0 {
				parent = s.stack[len(s.stack)-1]
			}
			if isWordprocessingMLName(value.Name, "tbl") {
				s.nextTable++
				s.tableStack = append(s.tableStack, tableScanContext{id: s.nextTable, depth: len(s.tableStack) + 1})
			}
			if isWordprocessingMLName(value.Name, "p") {
				s.nextParagraph++
				s.paragraphs = append(s.paragraphs, s.nextParagraph)
			}
			if isWordprocessingMLName(value.Name, "tr") {
				row := rowScanContext{}
				if isWordprocessingMLName(parent, "tbl") && len(s.tableStack) != 0 {
					table := &s.tableStack[len(s.tableStack)-1]
					table.nextRow++
					row = rowScanContext{table: table.id, row: table.nextRow, tableDepth: table.depth}
					position := rowPosition{table: row.table, row: row.row}
					if s.rowText[position] == nil {
						s.rowText[position] = &strings.Builder{}
					}
				}
				s.rowStack = append(s.rowStack, row)
			}
			s.stack = append(s.stack, value.Name)
			for _, attribute := range value.Attr {
				if err := s.record(attribute.Value, "attribute", value.Name, attribute.Name); err != nil {
					return err
				}
			}
		case xml.CharData:
			var element xml.Name
			if len(s.stack) != 0 {
				element = s.stack[len(s.stack)-1]
			}
			if isWordprocessingMLName(element, "t") {
				if len(s.paragraphs) != 0 {
					paragraph := s.paragraphs[len(s.paragraphs)-1]
					if s.paragraphText[paragraph] == nil {
						s.paragraphText[paragraph] = &strings.Builder{}
					}
					s.paragraphText[paragraph].Write(value)
				}
				for _, row := range s.rowStack {
					if row.table > 0 && row.row > 0 {
						s.rowText[rowPosition{table: row.table, row: row.row}].Write(value)
					}
				}
			}
			if err := s.record(string(value), "text", element, xml.Name{}); err != nil {
				return err
			}
		case xml.EndElement:
			if isWordprocessingMLName(value.Name, "p") {
				if len(s.paragraphs) != 0 {
					s.paragraphs = s.paragraphs[:len(s.paragraphs)-1]
				}
			}
			if isWordprocessingMLName(value.Name, "tr") {
				if len(s.rowStack) != 0 {
					s.rowStack = s.rowStack[:len(s.rowStack)-1]
				}
			}
			if isWordprocessingMLName(value.Name, "tbl") {
				if len(s.tableStack) != 0 {
					s.tableStack = s.tableStack[:len(s.tableStack)-1]
				}
			}
			if len(s.stack) != 0 {
				s.stack = s.stack[:len(s.stack)-1]
			}
		}
	}
}

func (s *partScanner) record(raw, kind string, element, attribute xml.Name) error {
	matches := completeTemplateToken.FindAllString(raw, -1)
	if len(matches) == 0 {
		if strings.Contains(raw, "{{") || strings.Contains(raw, "}}") {
			return contractError("partial template token in %q", s.part)
		}
		return nil
	}
	trimmed := strings.TrimSpace(raw)
	if len(matches) != 1 || matches[0] != trimmed || trimmed != raw {
		return contractError("template token in %q must occupy the whole XML value", s.part)
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "{{"), "}}"))
	if inner == "" {
		return contractError("empty template token in %q", s.part)
	}
	var row rowScanContext
	if len(s.rowStack) != 0 {
		row = s.rowStack[len(s.rowStack)-1]
	}
	paragraph := 0
	if len(s.paragraphs) != 0 {
		paragraph = s.paragraphs[len(s.paragraphs)-1]
	}
	s.occurrences = append(s.occurrences, tokenOccurrence{
		key: inner, part: s.part, kind: kind, element: element, attribute: attribute,
		table: row.table, row: row.row, tableDepth: row.tableDepth,
		order: len(s.occurrences) + 1, paragraph: paragraph,
		ancestry: append([]xml.Name(nil), s.stack...),
	})
	return nil
}
