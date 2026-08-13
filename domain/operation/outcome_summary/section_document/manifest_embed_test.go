package section_document

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"testing"

	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
)

var testProfile = bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1

type testZipPart struct {
	name, body string
	method     uint16
}

func makeTestDOCX(t *testing.T, parts ...testZipPart) []byte {
	t.Helper()
	if len(parts) == 0 {
		parts = []testZipPart{
			{name: "[Content_Types].xml", body: `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="xml" ContentType="application/xml"/></Types>`},
			{name: "word/document.xml", body: validManifestDocumentXML()},
		}
	}
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, part := range parts {
		header := &zip.FileHeader{Name: part.name, Method: part.method}
		entry, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(part.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func validManifestDocumentXML() string {
	root := []string{
		"sheet_title", "subscription_group_name_display", "price_schedule_name_display",
		"job_template_phase_name_display", "client_name_label",
	}
	for i := 1; i <= 11; i++ {
		root = append(root, fmt.Sprintf("job_template%d_name_display", i))
	}
	var xml strings.Builder
	xml.WriteString(`<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, key := range root {
		xml.WriteString(`<w:p><w:r><w:t>{{` + key + `}}</w:t></w:r></w:p>`)
	}
	xml.WriteString(`<w:tbl><w:tr><w:tc><w:p><w:r><w:t>{{#rows}}</w:t></w:r></w:p></w:tc></w:tr>`)
	xml.WriteString(`<w:tr><w:tc><w:p><w:r><w:rPr><w:b w:val="{{row_bold}}"/></w:rPr><w:t>{{row_label_display}}</w:t></w:r></w:p></w:tc>`)
	for i := 1; i <= 11; i++ {
		prefix := fmt.Sprintf("job_template%d_", i)
		xml.WriteString(`<w:tc><w:tcPr><w:shd w:fill="{{` + prefix + `fill_hex}}"/></w:tcPr><w:p><w:r><w:rPr><w:color w:val="{{` + prefix + `text_hex}}"/></w:rPr><w:t>{{` + prefix + `scaled_label}}</w:t></w:r></w:p></w:tc>`)
	}
	xml.WriteString(`</w:tr><w:tr><w:tc><w:p><w:r><w:t>{{/rows}}</w:t></w:r></w:p></w:tc></w:tr></w:tbl></w:body></w:document>`)
	return xml.String()
}

func TestRenderManifest_ExcludesEvidenceVocabulary(t *testing.T) {
	raw, err := ManifestBytes(testProfile)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(raw))
	for _, forbidden := range []string{"academic", "evidence_only", "observed_job_category"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("package manifest contains docs-only vocabulary %q", forbidden)
		}
	}
	if !strings.Contains(lower, SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key) {
		t.Fatal("package manifest is missing its canonical profile key")
	}
}

func TestGeneratedAuthoringTemplateMatchesManifest(t *testing.T) {
	docx, err := os.ReadFile("subscription-group-outcome-matrix-single-period-11-v1.docx")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateTemplate(testProfile, docx); err != nil {
		t.Fatalf("generated operator authoring asset: %v", err)
	}
}

func TestRenderManifestValidator_RequiresExactTextAndStyleManifest(t *testing.T) {
	validXML := validManifestDocumentXML()
	tests := map[string]string{
		"missing root":     strings.Replace(validXML, "{{sheet_title}}", "Static title", 1),
		"extra token":      strings.Replace(validXML, "</w:body>", `<w:p><w:r><w:t>{{unexpected}}</w:t></w:r></w:p></w:body>`, 1),
		"misplaced style":  strings.Replace(validXML, `w:fill="{{job_template1_fill_hex}}"`, `w:fill="FFFFFF"/><w:t>{{job_template1_fill_hex}}</w:t><w:shd w:fill="FFFFFF"`, 1),
		"partial token":    strings.Replace(validXML, "{{sheet_title}}", "prefix {{sheet_title}}", 1),
		"missing loop row": strings.Replace(validXML, `<w:tr><w:tc><w:p><w:r><w:t>{{#rows}}</w:t></w:r></w:p></w:tc></w:tr>`, "", 1),
	}
	for name, documentXML := range tests {
		t.Run(name, func(t *testing.T) {
			docx := makeTestDOCX(t,
				testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
				testZipPart{name: "word/document.xml", body: documentXML},
			)
			if err := ValidateTemplate(testProfile, docx); err == nil {
				t.Fatal("expected manifest validation failure")
			}
		})
	}
	if err := ValidateTemplate(testProfile, makeTestDOCX(t)); err != nil {
		t.Fatalf("valid exact manifest rejected: %v", err)
	}
}

func TestRenderManifestValidator_RejectsWrongOOXMLNamespaces(t *testing.T) {
	validXML := validManifestDocumentXML()
	alternatePrefixXML := strings.Replace(validXML, "xmlns:w=", "xmlns:x=", 1)
	alternatePrefixXML = strings.ReplaceAll(alternatePrefixXML, "w:", "x:")
	if alternatePrefixXML == validXML {
		t.Fatal("alternate-prefix fixture did not alter the valid manifest")
	}
	if err := ValidateTemplate(testProfile, makeTestDOCX(t,
		testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
		testZipPart{name: "word/document.xml", body: alternatePrefixXML},
	)); err != nil {
		t.Fatalf("valid alternate WordprocessingML prefix rejected: %v", err)
	}

	wrongNamespace := `urn:ichizen:hostile-template`
	tests := map[string]string{
		"root text": strings.Replace(
			validXML,
			`<w:t>{{sheet_title}}</w:t>`,
			`<x:t xmlns:x="`+wrongNamespace+`">{{sheet_title}}</x:t>`,
			1,
		),
		"loop text": strings.Replace(
			validXML,
			`<w:t>{{#rows}}</w:t>`,
			`<x:t xmlns:x="`+wrongNamespace+`">{{#rows}}</x:t>`,
			1,
		),
		"loop row": strings.Replace(
			validXML,
			`<w:tr><w:tc><w:p><w:r><w:t>{{#rows}}</w:t></w:r></w:p></w:tc></w:tr>`,
			`<x:tr xmlns:x="`+wrongNamespace+`"><w:tc><w:p><w:r><w:t>{{#rows}}</w:t></w:r></w:p></w:tc></x:tr>`,
			1,
		),
		"style element": strings.Replace(
			validXML,
			`<w:shd w:fill="{{job_template1_fill_hex}}"/>`,
			`<x:shd xmlns:x="`+wrongNamespace+`" w:fill="{{job_template1_fill_hex}}"/>`,
			1,
		),
		"style attribute": strings.Replace(
			validXML,
			`<w:shd w:fill="{{job_template1_fill_hex}}"/>`,
			`<w:shd xmlns:x="`+wrongNamespace+`" x:fill="{{job_template1_fill_hex}}"/>`,
			1,
		),
	}

	for name, documentXML := range tests {
		t.Run(name, func(t *testing.T) {
			if documentXML == validXML {
				t.Fatal("hostile namespace fixture did not alter the valid manifest")
			}
			docx := makeTestDOCX(t,
				testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
				testZipPart{name: "word/document.xml", body: documentXML},
			)
			if err := ValidateTemplate(testProfile, docx); err == nil {
				t.Fatal("wrong-namespace manifest token was accepted")
			}
		})
	}
}

func TestRenderManifestValidator_RejectsOffRendererParts(t *testing.T) {
	documentXML := strings.Replace(validManifestDocumentXML(), "{{sheet_title}}", "Static title", 1)
	parts := map[string]string{
		"word/header1.xml":             `<w:hdr xmlns:w="` + wordprocessingMLNamespace + `"><w:p><w:r><w:t>{{sheet_title}}</w:t></w:r></w:p></w:hdr>`,
		"word/footnotes.xml":           `<w:footnotes xmlns:w="` + wordprocessingMLNamespace + `"><w:footnote><w:p><w:r><w:t>{{sheet_title}}</w:t></w:r></w:p></w:footnote></w:footnotes>`,
		"word/_rels/document.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships" xmlns:w="` + wordprocessingMLNamespace + `"><w:t>{{sheet_title}}</w:t></Relationships>`,
	}
	for name, body := range parts {
		t.Run(strings.NewReplacer("/", "_", ".", "_").Replace(name), func(t *testing.T) {
			docx := makeTestDOCX(t,
				testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
				testZipPart{name: "word/document.xml", body: documentXML},
				testZipPart{name: name, body: body},
			)
			if err := ValidateTemplate(testProfile, docx); err == nil {
				t.Fatal("manifest token outside word/document.xml was accepted")
			}
		})
	}
}

func TestRenderManifestValidator_RequiresOneDirectLoopTable(t *testing.T) {
	validXML := validManifestDocumentXML()
	splitTables := strings.Replace(validXML, `</w:tr><w:tr>`, `</w:tr></w:tbl><w:tbl><w:tr>`, 2)
	nestedTable := strings.Replace(validXML, `<w:tbl>`, `<w:tbl><w:tr><w:tc><w:p/></w:tc><w:tc><w:tbl>`, 1)
	nestedTable = strings.Replace(nestedTable, `</w:tbl></w:body>`, `</w:tbl></w:tc></w:tr></w:tbl></w:body>`, 1)
	markerStaticText := strings.Replace(
		validXML,
		`<w:r><w:t>{{#rows}}</w:t></w:r>`,
		`<w:r><w:t>{{#rows}}</w:t></w:r><w:r><w:t>Static</w:t></w:r>`,
		1,
	)

	tests := map[string]string{
		"split across tables": splitTables,
		"nested loop table":   nestedTable,
		"marker static text":  markerStaticText,
	}
	for name, documentXML := range tests {
		t.Run(name, func(t *testing.T) {
			if documentXML == validXML {
				t.Fatal("structural fixture did not alter the valid manifest")
			}
			if err := ValidateTemplate(testProfile, makeTestDOCX(t,
				testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
				testZipPart{name: "word/document.xml", body: documentXML},
			)); err == nil {
				t.Fatal("renderer-incompatible loop structure was accepted")
			}
		})
	}

	whitespaceOnly := strings.Replace(
		validXML,
		`<w:r><w:t>{{#rows}}</w:t></w:r>`,
		`<w:r><w:t>{{#rows}}</w:t></w:r><w:r><w:t>  </w:t></w:r>`,
		1,
	)
	if err := ValidateTemplate(testProfile, makeTestDOCX(t,
		testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
		testZipPart{name: "word/document.xml", body: whitespaceOnly},
	)); err != nil {
		t.Fatalf("marker row whitespace should be ignored: %v", err)
	}
}

func TestRenderManifestValidator_RequiresRendererVisitedAncestry(t *testing.T) {
	validXML := validManifestDocumentXML()
	rowLabelCell := `<w:tc><w:p><w:r><w:rPr><w:b w:val="{{row_bold}}"/></w:rPr><w:t>{{row_label_display}}</w:t></w:r></w:p></w:tc>`
	directUnderRow := strings.Replace(validXML, rowLabelCell, `<w:t>{{row_label_display}}</w:t>`+strings.Replace(rowLabelCell, `{{row_label_display}}`, "Static", 1), 1)
	sheetParagraph := `<w:p><w:r><w:t>{{sheet_title}}</w:t></w:r></w:p>`
	outsideBody := strings.Replace(validXML, sheetParagraph, "", 1)
	outsideBody = strings.Replace(outsideBody, `<w:body>`, sheetParagraph+`<w:body>`, 1)

	for name, documentXML := range map[string]string{
		"row token directly under row": directUnderRow,
		"root token outside body":      outsideBody,
	} {
		t.Run(name, func(t *testing.T) {
			if documentXML == validXML {
				t.Fatal("ancestry fixture did not alter the valid manifest")
			}
			if err := ValidateTemplate(testProfile, makeTestDOCX(t,
				testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
				testZipPart{name: "word/document.xml", body: documentXML},
			)); err == nil {
				t.Fatal("renderer-invisible manifest token ancestry was accepted")
			}
		})
	}
}

func TestRenderManifestValidator_BoundsActualExpansionAndRejectsDuplicateParts(t *testing.T) {
	valid := makeTestDOCX(t)
	if err := validateTemplateWithLimits(testProfile, valid, validationLimits{input: int64(len(valid) - 1), entries: 2000, entry: 1 << 20, aggregate: 2 << 20}); err == nil {
		t.Fatal("compressed input cap was not enforced")
	}
	if err := validateTemplateWithLimits(testProfile, valid, validationLimits{input: 1 << 20, entries: 1, entry: 1 << 20, aggregate: 2 << 20}); err == nil {
		t.Fatal("entry-count cap was not enforced")
	}
	if err := validateTemplateWithLimits(testProfile, valid, validationLimits{input: 1 << 20, entries: 2000, entry: 32, aggregate: 2 << 20}); err == nil {
		t.Fatal("actual per-entry cap was not enforced")
	}
	if err := validateTemplateWithLimits(testProfile, valid, validationLimits{input: 1 << 20, entries: 2000, entry: 1 << 20, aggregate: 64}); err == nil {
		t.Fatal("actual aggregate cap was not enforced")
	}

	duplicate := makeTestDOCX(t,
		testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
		testZipPart{name: "word/document.xml", body: validManifestDocumentXML()},
		testZipPart{name: "WORD/DOCUMENT.XML", body: `<invalid/>`},
	)
	if err := ValidateTemplate(testProfile, duplicate); err == nil {
		t.Fatal("case-folded duplicate document part was not rejected")
	}
	for name, parts := range map[string][]testZipPart{
		"case-shifted document": {
			{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
			{name: "WORD/DOCUMENT.XML", body: validManifestDocumentXML()},
		},
		"case-shifted content types": {
			{name: "[content_types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
			{name: "word/document.xml", body: validManifestDocumentXML()},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateTemplate(testProfile, makeTestDOCX(t, parts...)); err == nil {
				t.Fatal("case-shifted authoritative DOCX part was accepted")
			}
		})
	}
	for _, hostile := range []string{"../word/document.xml", "/word/document.xml", `word\document.xml`, "word//document.xml", "C:/word/document.xml", "word/./document.xml"} {
		t.Run("path_"+strings.NewReplacer("/", "_", "\\", "_").Replace(hostile), func(t *testing.T) {
			docx := makeTestDOCX(t,
				testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`},
				testZipPart{name: hostile, body: validManifestDocumentXML()},
			)
			if err := ValidateTemplate(testProfile, docx); err == nil {
				t.Fatalf("unsafe path %q was not rejected", hostile)
			}
		})
	}
}

func TestRenderManifestValidator_RejectsCRCAndDeclaredSizeDrift(t *testing.T) {
	stored := makeTestDOCX(t,
		testZipPart{name: "[Content_Types].xml", body: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`, method: zip.Store},
		testZipPart{name: "word/document.xml", body: validManifestDocumentXML(), method: zip.Store},
	)
	corrupt := append([]byte(nil), stored...)
	offset := bytes.Index(corrupt, []byte("{{sheet_title}}"))
	if offset < 0 {
		t.Fatal("stored payload token not found")
	}
	corrupt[offset] = '['
	if err := ValidateTemplate(testProfile, corrupt); err == nil {
		t.Fatal("CRC-corrupt entry was not rejected")
	}

	declared := append([]byte(nil), stored...)
	if !alterCentralUncompressedSize(declared, "word/document.xml") {
		t.Fatal("document central-directory entry not found")
	}
	if err := ValidateTemplate(testProfile, declared); err == nil {
		t.Fatal("declared-size drift was not rejected")
	}
}

func alterCentralUncompressedSize(docx []byte, target string) bool {
	for offset := 0; offset+46 <= len(docx); offset++ {
		if binary.LittleEndian.Uint32(docx[offset:]) != 0x02014b50 {
			continue
		}
		nameLength := int(binary.LittleEndian.Uint16(docx[offset+28:]))
		extraLength := int(binary.LittleEndian.Uint16(docx[offset+30:]))
		commentLength := int(binary.LittleEndian.Uint16(docx[offset+32:]))
		end := offset + 46 + nameLength + extraLength + commentLength
		if end > len(docx) {
			return false
		}
		name := string(docx[offset+46 : offset+46+nameLength])
		if name == target {
			value := binary.LittleEndian.Uint32(docx[offset+24:])
			binary.LittleEndian.PutUint32(docx[offset+24:], value-1)
			return true
		}
		offset = end - 1
	}
	return false
}
