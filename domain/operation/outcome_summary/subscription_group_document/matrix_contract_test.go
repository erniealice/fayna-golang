package subscription_group_document

// matrix_contract_test.go — the group-matrix data source contract. Like
// outcome_matrix/document/engine_contract_test.go this is the only place the
// package touches the real fycha engine, and only in a test (go.work resolves
// it); production renders go through the injected PDF closure.

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/erniealice/fycha-golang/services/doctemplate"
)

func exampleTemplate(t *testing.T) []byte {
	t.Helper()
	docx, err := os.ReadFile("matrix-template-example.docx")
	if err != nil {
		t.Fatalf("read example template: %v", err)
	}
	return docx
}

func legacyTemplate(t *testing.T) []byte {
	t.Helper()
	docx, err := os.ReadFile("subscription-group-outcome-matrix-single-period-11-v1.docx")
	if err != nil {
		t.Fatalf("read legacy 11-slot template: %v", err)
	}
	return docx
}

func groupMatrix(columns int) Matrix {
	m := Matrix{
		JobCategoryID: "category-1", SheetTitle: "Category", SubscriptionGroupName: "Group A",
		PriceScheduleName: "Schedule 1", PeriodName: "Period 1", ClientNameLabel: "Client",
	}
	for i := 0; i < columns; i++ {
		m.Columns = append(m.Columns, Column{JobTemplateID: fmt.Sprintf("jt-%02d", i), DisplayName: fmt.Sprintf("Subject %02d", i)})
	}
	client := Row{Kind: RowClient, Label: "[1] One, Alpha"}
	for i, column := range m.Columns {
		client.Cells = append(client.Cells, Cell{JobTemplateID: column.JobTemplateID, Value: fmt.Sprintf("%d", 4+i%4)})
	}
	m.Rows = []Row{{Kind: RowBand, Label: "band"}, client, {Kind: RowBlank}}
	return m
}

func render(t *testing.T, template []byte, m Matrix) string {
	t.Helper()
	layout, err := CheckColumnCapacity(template, len(m.Columns))
	if err != nil {
		t.Fatalf("%d columns: CheckColumnCapacity: %v", len(m.Columns), err)
	}
	m.NumberedSlots = layout.NumberedSlots
	data, err := BuildData(GroupMatrixRenderProfile, m)
	if err != nil {
		t.Fatalf("%d columns: BuildData: %v", len(m.Columns), err)
	}
	rendered, err := doctemplate.ProcessTemplate(template, data)
	if err != nil {
		t.Fatalf("%d columns: ProcessTemplate: %v", len(m.Columns), err)
	}
	xml := documentXML(t, rendered)
	if strings.Contains(xml, "{{") {
		t.Fatalf("%d columns: residual template token in output", len(m.Columns))
	}
	return xml
}

func documentXML(t *testing.T, docx []byte) string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatalf("open rendered docx: %v", err)
	}
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			rc, _ := file.Open()
			body, _ := io.ReadAll(rc)
			rc.Close()
			return string(body)
		}
	}
	t.Fatal("rendered docx has no document.xml")
	return ""
}

// A-DATA-1: one column-loop template renders 1, 11 and 13 columns.
func TestSectionMatrix_ColumnLoopTemplateRendersAnyColumnCount(t *testing.T) {
	layout, err := ValidateGroupMatrixTemplate(exampleTemplate(t))
	if err != nil || !layout.ColumnLoop || layout.NumberedSlots != 0 {
		t.Fatalf("example template layout = %+v, %v", layout, err)
	}
	for _, columns := range []int{1, 11, 13} {
		xml := render(t, exampleTemplate(t), groupMatrix(columns))
		if got := strings.Count(xml, "<w:gridCol "); got != columns+1 {
			t.Fatalf("%d columns: grid columns = %d, want %d", columns, got, columns+1)
		}
		for _, want := range []string{"Subject 00", fmt.Sprintf("Subject %02d", columns-1), "[1] One, Alpha", "band", "Group A", "Period 1"} {
			if !strings.Contains(xml, want) {
				t.Fatalf("%d columns: output missing %q", columns, want)
			}
		}
	}
}

// A-DATA-1: cells align by column identity, not by the order they arrive in.
func TestSectionMatrix_ReorderedCellsAlignByIdentity(t *testing.T) {
	m := groupMatrix(3)
	cells := m.Rows[1].Cells
	m.Rows[1].Cells = []Cell{cells[2], cells[0], cells[1]}
	data, err := BuildData(GroupMatrixRenderProfile, m)
	if err != nil {
		t.Fatal(err)
	}
	got := data["rows"].([]map[string]any)[1]["cells"].([]map[string]any)
	for i, cell := range cells {
		if got[i]["cell_scaled_label"] != cell.Value {
			t.Fatalf("cell %d = %v, want %q", i, got[i], cell.Value)
		}
	}
	if data["rows"].([]map[string]any)[1]["job_template3_scaled_label"] != cells[2].Value {
		t.Fatal("numbered view does not follow column identity")
	}
}

// A-DATA-2 / D8: the legacy 11-slot template keeps rendering from the
// numbered view; fewer columns print blank slots (D9).
func TestSectionMatrix_LegacyNumberedTemplateRenders(t *testing.T) {
	layout, err := ValidateGroupMatrixTemplate(legacyTemplate(t))
	if err != nil || layout.ColumnLoop || layout.NumberedSlots != 11 {
		t.Fatalf("legacy layout = %+v, %v", layout, err)
	}
	for _, columns := range []int{11, 10, 1} {
		xml := render(t, legacyTemplate(t), groupMatrix(columns))
		if !strings.Contains(xml, fmt.Sprintf("Subject %02d", columns-1)) || !strings.Contains(xml, "[1] One, Alpha") {
			t.Fatalf("%d columns: legacy output is missing data", columns)
		}
	}
}

// A-CAP-1 / D9: a numbered template refuses a matrix with more columns than
// its slots, with both counts.
func TestSectionMatrix_NumberedTemplateRefusesMoreColumns(t *testing.T) {
	_, err := CheckColumnCapacity(legacyTemplate(t), 13)
	var capacity *ColumnCapacityError
	if !errors.As(err, &capacity) || capacity.TemplateColumns != 11 || capacity.DataColumns != 13 {
		t.Fatalf("err = %v, want ColumnCapacityError{11, 13}", err)
	}
	m := groupMatrix(13)
	m.NumberedSlots = 11
	if _, err := BuildData(GroupMatrixRenderProfile, m); !errors.As(err, &capacity) {
		t.Fatalf("BuildData over capacity err = %v", err)
	}
}

func TestSectionMatrix_BuildDataBoundsAndPayload(t *testing.T) {
	if _, err := BuildData(GroupMatrixRenderProfile, groupMatrix(MaxColumns+1)); err == nil {
		t.Fatal("over-limit columns accepted")
	}
	noCategory := groupMatrix(2)
	noCategory.JobCategoryID = ""
	if _, err := BuildData(GroupMatrixRenderProfile, noCategory); err == nil {
		t.Fatal("missing job category accepted")
	}
	if _, err := BuildData(testClientPhaseProfile, groupMatrix(2)); err == nil {
		t.Fatal("client-phase profile accepted by the group-matrix builder")
	}
	ragged := groupMatrix(2)
	ragged.Rows[1].Cells = ragged.Rows[1].Cells[:1]
	if _, err := BuildData(GroupMatrixRenderProfile, ragged); err == nil {
		t.Fatal("ragged client row accepted")
	}
	m := groupMatrix(4)
	m.NumberedSlots = 6
	data, err := BuildData(GroupMatrixRenderProfile, m)
	if err != nil {
		t.Fatal(err)
	}
	rows := data["rows"].([]map[string]any)
	cells := rows[1]["cells"].([]map[string]any)
	if len(cells) != 4 || cells[1]["cell_scaled_label"] != "5" || cells[1]["cell_fill_hex"] != "FFFF00" {
		t.Fatalf("client cells = %v", cells)
	}
	if rows[0]["row_band_label"] != "band" || rows[0]["row_client_label"] != "" || rows[1]["row_band_label"] != "" || rows[1]["row_client_label"] != "[1] One, Alpha" {
		t.Fatalf("row labels split by kind: %v / %v", rows[0], rows[1])
	}
	if rows[0]["row_label_display"] != "band" || rows[0]["row_bold"] != "true" {
		t.Fatalf("band row must carry the raw grouping value: %v", rows[0])
	}
	if data["job_template6_name_display"] != "" || rows[1]["job_template6_fill_hex"] != "FFFFFF" || rows[1]["job_template5_scaled_label"] != "" {
		t.Fatal("numbered slots beyond the columns must be blank-seeded")
	}
	if _, present := data["job_template7_name_display"]; present {
		t.Fatal("numbered view extends past the template's slots")
	}
	if data["period_name_display"] != "Period 1" || data["job_template_count"] != "4" || len(data["job_templates"].([]map[string]any)) != 4 {
		t.Fatalf("root payload = %v", data)
	}
}

// A-VAL-1: typed, scope-aware allow-list.
func TestSectionMatrix_ValidatorAllowList(t *testing.T) {
	example := documentXML(t, exampleTemplate(t))
	legacy := documentXML(t, legacyTemplate(t))
	accepted := map[string]string{
		"repeated root token":            strings.Replace(example, "{{period_name_display}}", "{{period_name_display}}</w:t></w:r><w:r><w:t>{{period_name_display}}", 1),
		"omitted root token":             strings.Replace(example, "{{period_name_display}}", "Period", 1),
		"omitted sheet title":            strings.Replace(legacy, "{{sheet_title}}", "Title", 1),
		"numbered root with column loop": strings.Replace(example, "{{period_name_display}}", "{{period_name_display}}</w:t></w:r><w:r><w:t>{{job_template1_name_display}}", 1),
	}
	for name, xml := range accepted {
		t.Run("accept/"+name, func(t *testing.T) {
			base := exampleTemplate(t)
			if strings.Contains(name, "sheet title") {
				base = legacyTemplate(t)
			}
			if _, err := ValidateGroupMatrixTemplate(replaceDocumentXML(t, base, xml)); err != nil {
				t.Fatalf("rejected: %v", err)
			}
		})
	}
	rejected := map[string]struct{ xml, want string }{
		"unknown token":                {strings.Replace(example, "{{period_name_display}}", "{{period_name_display}}</w:t></w:r><w:r><w:t>{{student_gpa}}", 1), `unknown template token "student_gpa" at paragraph`},
		"unknown loop":                 {strings.Replace(example, "{{period_name_display}}", "{{#subjects}}", 1), `unknown loop "#subjects"`},
		"cell token in root scope":     {strings.Replace(example, "{{period_name_display}}", "{{cell_scaled_label}}", 1), "outside its cells scope"},
		"row token in root scope":      {strings.Replace(example, "{{period_name_display}}", "{{row_label_display}}", 1), "outside its rows scope"},
		"header loop without close":    {strings.Replace(example, "{{/job_templates}}", "", 1), "must be paired"},
		"row cell loop without close":  {strings.Replace(example, "{{/cells}}", "", 1), "must be paired"},
		"slot beyond the column limit": {strings.Replace(example, "{{period_name_display}}", "{{job_template64_name_display}}", 1), "unknown template token"},
		"row loop without end":         {strings.Replace(legacy, "{{/rows}}", "", 1), "row loop markers"},
		"style token as text":          {strings.Replace(legacy, "{{job_template1_scaled_label}}", "{{job_template1_fill_hex}}", 1), "style token"},
	}
	for name, tc := range rejected {
		t.Run("reject/"+name, func(t *testing.T) {
			base := exampleTemplate(t)
			if strings.Contains(tc.xml, "job_template1_scaled_label") || strings.Contains(name, "row loop without end") || strings.Contains(name, "style token") {
				base = legacyTemplate(t)
			}
			_, err := ValidateGroupMatrixTemplate(replaceDocumentXML(t, base, tc.xml))
			var contract *TemplateContractError
			if err == nil || !errors.As(err, &contract) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want a TemplateContractError containing %q", err, tc.want)
			}
		})
	}
}

// ValidateTemplate (the upload path) dispatches the section profile to the
// allow-list validator.
func TestSectionMatrix_ValidateTemplateUsesVocabulary(t *testing.T) {
	for _, docx := range [][]byte{exampleTemplate(t), legacyTemplate(t)} {
		if err := ValidateTemplate(GroupMatrixRenderProfile, docx); err != nil {
			t.Fatalf("ValidateTemplate: %v", err)
		}
	}
}

func mustDocumentXML(t *testing.T, docx []byte) []byte {
	return []byte(documentXML(t, docx))
}

func replaceDocumentXML(t *testing.T, docx []byte, document string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	writer := zip.NewWriter(&out)
	for _, file := range reader.File {
		rc, _ := file.Open()
		body, _ := io.ReadAll(rc)
		rc.Close()
		if file.Name == "word/document.xml" {
			body = []byte(document)
		}
		w, _ := writer.Create(file.Name)
		w.Write(body)
	}
	writer.Close()
	return out.Bytes()
}
