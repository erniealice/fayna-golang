package section_document

import (
	"fmt"
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
// and PDF. BuildData validates identity/capacity again at the document boundary.
type Matrix struct {
	JobCategoryID         string
	SheetTitle            string
	SubscriptionGroupName string
	PriceScheduleName     string
	JobTemplatePhaseName  string
	ClientNameLabel       string
	Columns               []Column
	Rows                  []Row
}

// BuildData converts a matrix to the exact blank-seeded positional map expected
// by the registered DOCX. It never truncates, pads columns, or trusts cell order.
func BuildData(profileValue bindingpb.RenderProfile, matrix Matrix) (map[string]any, error) {
	profile, ok := LookupProfile(profileValue)
	if !ok {
		return nil, fmt.Errorf("unsupported render profile")
	}
	if profile.RequiresExactCategory && strings.TrimSpace(matrix.JobCategoryID) == "" {
		return nil, fmt.Errorf("render profile requires an exact job category")
	}
	if len(matrix.Columns) != profile.JobTemplateSlots {
		return nil, fmt.Errorf("render profile requires %d job-template columns", profile.JobTemplateSlots)
	}

	columnIDs := make(map[string]struct{}, len(matrix.Columns))
	data := map[string]any{
		"sheet_title":                     matrix.SheetTitle,
		"subscription_group_name_display": matrix.SubscriptionGroupName,
		"price_schedule_name_display":     matrix.PriceScheduleName,
		"job_template_phase_name_display": matrix.JobTemplatePhaseName,
		"client_name_label":               matrix.ClientNameLabel,
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
		data[fmt.Sprintf("job_template%d_name_display", i+1)] = column.DisplayName
	}

	rows := make([]map[string]any, 0, len(matrix.Rows))
	for rowIndex, row := range matrix.Rows {
		item := map[string]any{
			"row_label_display": row.Label,
			"row_bold":          boolText(row.Kind == RowBand),
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

		for i, column := range matrix.Columns {
			value := values[strings.TrimSpace(column.JobTemplateID)]
			fill, text := ratingStyle(value)
			prefix := fmt.Sprintf("job_template%d_", i+1)
			item[prefix+"scaled_label"] = value
			item[prefix+"fill_hex"] = fill
			item[prefix+"text_hex"] = text
		}
		rows = append(rows, item)
	}
	data["rows"] = rows
	return data, nil
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
