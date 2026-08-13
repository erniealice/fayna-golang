package section_document

import (
	"fmt"
	"testing"

	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
)

func elevenColumnMatrix() Matrix {
	result := Matrix{JobCategoryID: "category-1", SheetTitle: "Outcomes", SubscriptionGroupName: "Group", ClientNameLabel: "Client"}
	for i := 1; i <= 11; i++ {
		id := fmt.Sprintf("template-%02d", i)
		result.Columns = append(result.Columns, Column{JobTemplateID: id, DisplayName: fmt.Sprintf("Column %d", i)})
	}
	return result
}

func TestBuildSectionSheetData_PaletteAndBlankSeeding(t *testing.T) {
	matrix := elevenColumnMatrix()
	client := Row{Kind: RowClient, Label: "[1] Synthetic Client"}
	for i := len(matrix.Columns) - 1; i >= 0; i-- { // deliberately permuted input
		value := ""
		if i < 4 {
			value = fmt.Sprint(5 + i)
		}
		client.Cells = append(client.Cells, Cell{JobTemplateID: matrix.Columns[i].JobTemplateID, Value: value})
	}
	matrix.Rows = []Row{{Kind: RowBand, Label: "Band"}, client, {Kind: RowBlank}}
	data, err := BuildData(bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1, matrix)
	if err != nil {
		t.Fatal(err)
	}
	rows := data["rows"].([]map[string]any)
	if rows[0]["row_bold"] != "true" || rows[1]["row_bold"] != "false" || rows[2]["job_template1_scaled_label"] != "" {
		t.Fatalf("row seeding = %+v", rows)
	}
	wants := [][3]string{{"5", "FFFF00", "000000"}, {"6", "008000", "FFFFFF"}, {"7", "0000FF", "FFFFFF"}, {"8", "FFFFFF", "000000"}}
	for i, want := range wants {
		prefix := fmt.Sprintf("job_template%d_", i+1)
		if rows[1][prefix+"scaled_label"] != want[0] || rows[1][prefix+"fill_hex"] != want[1] || rows[1][prefix+"text_hex"] != want[2] {
			t.Fatalf("slot %d = value=%v fill=%v text=%v", i+1, rows[1][prefix+"scaled_label"], rows[1][prefix+"fill_hex"], rows[1][prefix+"text_hex"])
		}
	}
}

func TestBuildSectionSheetData_RequiresExactCellIdentitySet(t *testing.T) {
	profile := bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1
	base := elevenColumnMatrix()
	validCells := make([]Cell, len(base.Columns))
	for i, column := range base.Columns {
		validCells[i] = Cell{JobTemplateID: column.JobTemplateID, Value: "6"}
	}
	tests := map[string]func(*Matrix){
		"ten columns":      func(m *Matrix) { m.Columns = m.Columns[:10] },
		"twelve columns":   func(m *Matrix) { m.Columns = append(m.Columns, Column{JobTemplateID: "template-12"}) },
		"duplicate column": func(m *Matrix) { m.Columns[10].JobTemplateID = m.Columns[0].JobTemplateID },
		"missing category": func(m *Matrix) { m.JobCategoryID = "" },
		"missing cell":     func(m *Matrix) { m.Rows[0].Cells = m.Rows[0].Cells[:10] },
		"duplicate cell":   func(m *Matrix) { m.Rows[0].Cells[10].JobTemplateID = m.Rows[0].Cells[0].JobTemplateID },
		"unknown cell":     func(m *Matrix) { m.Rows[0].Cells[10].JobTemplateID = "unknown" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			matrix := base
			matrix.Columns = append([]Column(nil), base.Columns...)
			matrix.Rows = []Row{{Kind: RowClient, Label: "Synthetic", Cells: append([]Cell(nil), validCells...)}}
			mutate(&matrix)
			if _, err := BuildData(profile, matrix); err == nil {
				t.Fatal("expected exact identity/capacity failure")
			}
		})
	}
	if _, err := BuildData(bindingpb.RenderProfile_RENDER_PROFILE_UNSPECIFIED, base); err == nil {
		t.Fatal("unspecified profile must fail")
	}
}
