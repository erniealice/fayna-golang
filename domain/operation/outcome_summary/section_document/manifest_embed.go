package section_document

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/json"
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

const manifestFilename = "subscription-group-outcome-matrix-single-period-11-v1.manifest.json"

const wordprocessingMLNamespace = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

const (
	maxTemplateInputBytes     int64 = 10 << 20
	maxTemplateEntries              = 2000
	maxTemplateEntryBytes     int64 = 64 << 20
	maxTemplateAggregateBytes int64 = 256 << 20
)

//go:embed subscription-group-outcome-matrix-single-period-11-v1.manifest.json
var manifestFiles embed.FS

var completeTemplateToken = regexp.MustCompile(`\{\{[^{}]*\}\}`)

type manifestNode struct {
	Scalars []string                `json:"scalars"`
	Loops   map[string]manifestNode `json:"loops"`
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
	Scalars     []string                `json:"scalars"`
	Loops       map[string]manifestNode `json:"loops"`
	StyleValues struct {
		RowBold []string `json:"row_bold"`
		FillHex []string `json:"fill_hex"`
		TextHex []string `json:"text_hex"`
	} `json:"style_values"`
}

// TemplateContractError marks a trusted template/profile incompatibility. The
// HTTP layer maps it to a fail-loud 503 before invoking the renderer.
type TemplateContractError struct{ reason string }

func (e *TemplateContractError) Error() string {
	return "section document template contract: " + e.reason
}
func (*TemplateContractError) TemplateContractError() bool { return true }

func contractError(format string, args ...any) error {
	return &TemplateContractError{reason: fmt.Sprintf(format, args...)}
}

// ManifestBytes returns a defensive copy of the generic, package-owned
// manifest. It intentionally contains no docs-only evidence/category metadata.
func ManifestBytes(profileValue bindingpb.RenderProfile) ([]byte, error) {
	profile, ok := LookupProfile(profileValue)
	if !ok || profile.Key != SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key {
		return nil, contractError("unsupported render profile")
	}
	raw, err := manifestFiles.ReadFile(manifestFilename)
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
	if manifest.Profile.Key != profile.Key || manifest.Profile.JobTemplateSlots != profile.JobTemplateSlots ||
		manifest.Profile.BindingJobCategoryScope != "exact_required" || manifest.Profile.PeriodCardinality != 1 ||
		manifest.Profile.RowLoop != "rows" {
		return Profile{}, renderManifest{}, contractError("embedded manifest does not match the registered profile")
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
	nextTable     int
	rowText       map[rowPosition]*strings.Builder
	occurrences   []tokenOccurrence
	completeCount int
}

func validateManifestTokens(parts map[string][]byte, manifest renderManifest) error {
	var occurrences []tokenOccurrence
	rowTexts := make(map[rowLocation]string)
	completeCount := 0
	for part, body := range parts {
		scanner := &partScanner{
			part:          part,
			completeCount: len(completeTemplateToken.FindAll(body, -1)),
			rowText:       make(map[rowPosition]*strings.Builder),
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
	s.occurrences = append(s.occurrences, tokenOccurrence{
		key: inner, part: s.part, kind: kind, element: element, attribute: attribute,
		table: row.table, row: row.row, tableDepth: row.tableDepth,
		ancestry: append([]xml.Name(nil), s.stack...),
	})
	return nil
}
