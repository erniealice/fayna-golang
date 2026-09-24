package subscription_group_document

import (
	"fmt"
	"strconv"
	"strings"

	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
)

// RowKind distinguishes client rows from presentation-only band/separator rows.
type RowKind uint8

const (
	RowClient RowKind = iota
	RowBand
	RowBlank
)

type Column struct {
	JobTemplateID string
	DisplayName   string
}

type Cell struct {
	JobTemplateID string
	Value         string
}

type Row struct {
	Kind  RowKind
	Label string
	Cells []Cell
}

// Matrix is the already-authorized, normalized semantic model shared by CSV
// and PDF. BuildData validates identity again at the document boundary.
type Matrix struct {
	JobCategoryID         string
	SheetTitle            string
	SubscriptionGroupName string
	PriceScheduleName     string
	JobTemplatePhaseName  string
	// PeriodName is the plain display name of the selected period (the
	// drawer's option label), without any document-name decoration.
	PeriodName      string
	ClientNameLabel string
	Columns         []Column
	Rows            []Row
	// NumberedSlots is the highest numbered job_template{N}_* slot the
	// uploaded template uses (TemplateLayout.NumberedSlots). Numbered fields
	// beyond the data's columns up to this slot are emitted blank so a
	// template with more slots than the group has columns prints them empty.
	NumberedSlots int
}

// BuildData converts a matrix to the group-matrix data source. Every
// template draws from the same payload and uses only what it needs:
//
//   - root display fields and job_template_count;
//   - job_templates[]: one item per column (job_template_name_display,
//     job_template_number) for header column loops;
//   - rows[]: one item per row (client, band heading, or blank spacer) with
//     row_label_display, row_band_label, row_client_label, row_bold and cells[] aligned to job_templates by
//     column identity (cell_scaled_label, cell_fill_hex, cell_text_hex);
//   - the numbered view job_template{N}_name_display and per row
//     job_template{N}_scaled_label/_fill_hex/_text_hex for every column (and
//     blank up to NumberedSlots), so fixed-slot templates keep rendering.
//
// Band headings carry the raw grouping value; a template formats it. Rating
// colours are optional legacy data a template may ignore. It never truncates
// columns or trusts cell order.
func BuildData(profileValue bindingpb.RenderProfile, matrix Matrix) (map[string]any, error) {
	profile, ok := LookupProfile(profileValue)
	if !ok || profile.Enum != GroupMatrixRenderProfile {
		return nil, fmt.Errorf("unsupported render profile")
	}
	if !profile.AcceptsCategoryBinding(matrix.JobCategoryID) {
		return nil, fmt.Errorf("render profile does not accept the requested job-category binding scope")
	}
	if len(matrix.Columns) > MaxColumns {
		return nil, fmt.Errorf("a group matrix supports at most %d job-template columns", MaxColumns)
	}
	if matrix.NumberedSlots < 0 || matrix.NumberedSlots > MaxColumns {
		return nil, fmt.Errorf("numbered slots must be between 0 and %d", MaxColumns)
	}
	if matrix.NumberedSlots > 0 && len(matrix.Columns) > matrix.NumberedSlots {
		return nil, &ColumnCapacityError{TemplateColumns: matrix.NumberedSlots, DataColumns: len(matrix.Columns)}
	}
	slots := max(len(matrix.Columns), matrix.NumberedSlots)

	columnIDs := make(map[string]struct{}, len(matrix.Columns))
	columns := make([]map[string]any, 0, len(matrix.Columns))
	data := map[string]any{
		"sheet_title":                     matrix.SheetTitle,
		"subscription_group_name_display": matrix.SubscriptionGroupName,
		"price_schedule_name_display":     matrix.PriceScheduleName,
		"job_template_phase_name_display": matrix.JobTemplatePhaseName,
		"period_name_display":             matrix.PeriodName,
		"client_name_label":               matrix.ClientNameLabel,
		"job_template_count":              strconv.Itoa(len(matrix.Columns)),
	}
	for i, column := range matrix.Columns {
		id := strings.TrimSpace(column.JobTemplateID)
		if id == "" {
			return nil, fmt.Errorf("job-template column %d has an empty identity", i+1)
		}
		if _, duplicate := columnIDs[id]; duplicate {
			return nil, fmt.Errorf("duplicate job-template column identity")
		}
		columnIDs[id] = struct{}{}
		columns = append(columns, map[string]any{
			"job_template_name_display": column.DisplayName,
			"job_template_number":       strconv.Itoa(i + 1),
		})
	}
	for slot := 1; slot <= slots; slot++ {
		name := ""
		if slot <= len(matrix.Columns) {
			name = matrix.Columns[slot-1].DisplayName
		}
		data[fmt.Sprintf("job_template%d_name_display", slot)] = name
	}
	data["job_templates"] = columns

	rows := make([]map[string]any, 0, len(matrix.Rows))
	for rowIndex, row := range matrix.Rows {
		// row_band_label / row_client_label split the label by row kind so a
		// template can format band headings (e.g. Word capitals) without
		// touching client names.
		item := map[string]any{
			"row_label_display": row.Label,
			"row_bold":          boolText(row.Kind == RowBand),
			"row_band_label":    "",
			"row_client_label":  "",
		}
		switch row.Kind {
		case RowBand:
			item["row_band_label"] = row.Label
		case RowClient:
			item["row_client_label"] = row.Label
		}
		values := make(map[string]string, len(row.Cells))
		if row.Kind == RowClient {
			if len(row.Cells) != len(matrix.Columns) {
				return nil, fmt.Errorf("client row %d does not match the canonical column set", rowIndex+1)
			}
			for _, cell := range row.Cells {
				id := strings.TrimSpace(cell.JobTemplateID)
				if _, known := columnIDs[id]; !known || id == "" {
					return nil, fmt.Errorf("client row %d contains an unknown column identity", rowIndex+1)
				}
				if _, duplicate := values[id]; duplicate {
					return nil, fmt.Errorf("client row %d contains a duplicate column identity", rowIndex+1)
				}
				values[id] = cell.Value
			}
		} else if len(row.Cells) != 0 {
			return nil, fmt.Errorf("presentation row %d contains outcome cells", rowIndex+1)
		}

		cells := make([]map[string]any, 0, len(matrix.Columns))
		for slot := 1; slot <= slots; slot++ {
			value := ""
			if slot <= len(matrix.Columns) {
				value = values[strings.TrimSpace(matrix.Columns[slot-1].JobTemplateID)]
			}
			fill, text := ratingStyle(value)
			prefix := fmt.Sprintf("job_template%d_", slot)
			item[prefix+"scaled_label"] = value
			item[prefix+"fill_hex"] = fill
			item[prefix+"text_hex"] = text
			if slot <= len(matrix.Columns) {
				cells = append(cells, map[string]any{"cell_scaled_label": value, "cell_fill_hex": fill, "cell_text_hex": text})
			}
		}
		item["cells"] = cells
		rows = append(rows, item)
	}
	data["rows"] = rows
	return data, nil
}

// ColumnCapacityError reports a numbered-slot template with fewer job-template
// slots than the group has columns. Printing it would silently drop columns,
// so the export refuses with both counts instead.
type ColumnCapacityError struct {
	TemplateColumns int
	DataColumns     int
}

func (e *ColumnCapacityError) Error() string {
	return fmt.Sprintf("template has %d job-template columns, the matrix has %d", e.TemplateColumns, e.DataColumns)
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func ratingStyle(value string) (fill, text string) {
	switch strings.TrimSpace(value) {
	case "5":
		return "FFFF00", "000000"
	case "6":
		return "008000", "FFFFFF"
	case "7":
		return "0000FF", "FFFFFF"
	default:
		return "FFFFFF", "000000"
	}
}
