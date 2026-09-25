package outcome_summary

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

// Report-cell template tokens. {composite} is the stored raw composite
// (summary_score, pre-transmutation); {scaled} is the transmuted output
// (scaled_label, else the formatted scaled_score).
const (
	CellTokenComposite = "{composite}"
	CellTokenScaled    = "{scaled}"
	CellFormatScaled   = CellTokenScaled
)

// ValidateCellFormat accepts a template that uses at least one known token and
// no other {…} placeholder.
func ValidateCellFormat(format string) error {
	rest := strings.NewReplacer(CellTokenComposite, "", CellTokenScaled, "").Replace(format)
	if strings.ContainsAny(rest, "{}") {
		return fmt.Errorf("unknown placeholder in cell format %q (allowed: %s, %s)", format, CellTokenComposite, CellTokenScaled)
	}
	if !strings.Contains(format, CellTokenComposite) && !strings.Contains(format, CellTokenScaled) {
		return fmt.Errorf("cell format %q uses no token (allowed: %s, %s)", format, CellTokenComposite, CellTokenScaled)
	}
	return nil
}

var loggedCellFormats sync.Map

func logInvalidCellFormat(format string, err error) {
	if _, seen := loggedCellFormats.LoadOrStore(format, true); !seen {
		log.Printf("outcome summary: %v — falling back to %s", err, CellFormatScaled)
	}
}

// FormatReportCell renders one report cell through a validated template
// (Options.ReportCellFormat). The non-enrolled placeholder test runs on the
// scaled value exactly as ExportCellValue does, so blanking never depends on
// the template. "{scaled}" is byte-identical to ExportCellValue. For any other
// template, a token with no stored value collapses the cell to the other
// token's value alone (no dangling separator); both missing renders blank.
func FormatReportCell(cell *exportpb.SubscriptionGroupOutcomeCell, blank, format string) string {
	if format == "" || format == CellFormatScaled {
		return ExportCellValue(cell, blank)
	}
	scaled := ExportCellValue(cell, "")
	if scaled == "" {
		// Absent, missing-job, or non-enrolled cell: a composite alone must not
		// resurrect a placeholder the scaled rule blanks.
		if cell == nil || cell.SummaryScore == nil || !enrolledWithoutScaled(cell) {
			return blank
		}
	}
	if v := ApplyCellFormat(format, scaled, cell.SummaryScore); v != "" {
		return v
	}
	return blank
}

// ApplyCellFormat is the pure template step shared by every report-cell
// surface: scaled is the already-resolved transmuted value ("" when none) and
// composite the stored raw composite (nil when none). A token the template
// uses with no value collapses the result to the other value alone; both
// missing returns "". Callers own enrollment blanking and the blank marker.
func ApplyCellFormat(format, scaled string, composite *float64) string {
	if format == "" || format == CellFormatScaled {
		return scaled
	}
	compositeText := ""
	if composite != nil {
		compositeText = strconv.FormatFloat(*composite, 'f', -1, 64)
	}
	return applyCellTokens(format, scaled, compositeText)
}

func applyCellTokens(format, scaled, composite string) string {
	usesComposite := strings.Contains(format, CellTokenComposite)
	usesScaled := strings.Contains(format, CellTokenScaled)
	if (usesComposite && composite == "") || (usesScaled && scaled == "") {
		if composite != "" {
			return composite
		}
		return scaled
	}
	return strings.NewReplacer(CellTokenComposite, composite, CellTokenScaled, scaled).Replace(format)
}

// enrolledWithoutScaled reports a present, enrolled cell that simply has no
// scaled value yet (a real mark exists) — the only case where a composite may
// render without a scaled companion.
func enrolledWithoutScaled(cell *exportpb.SubscriptionGroupOutcomeCell) bool {
	if strings.TrimSpace(cell.GetJobTemplateId()) == "" || cell.GetEnrollmentEvidence() == nil {
		return false
	}
	return cell.GetEnrollmentEvidence().GetHasPositiveMark()
}
