package outcome_summary

import (
	"strconv"
	"strings"

	exportpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/subscription_group_outcome_export"
)

// ExportCellValue applies the shared report-cell display policy: a stored label
// wins, a numeric score is the fallback, and a non-enrolled placeholder is blank.
func ExportCellValue(cell *exportpb.SubscriptionGroupOutcomeCell, blank string) string {
	if cell == nil || strings.TrimSpace(cell.GetJobTemplateId()) == "" || cell.GetEnrollmentEvidence() == nil {
		return blank
	}

	value := strings.TrimSpace(cell.GetScaledLabel())
	if value == "" && cell.ScaledScore != nil {
		value = strconv.FormatFloat(cell.GetScaledScore(), 'f', -1, 64)
	}
	if IsNonEnrolledCell(EnrollmentEvidence{
		HasMarks:        cell.GetEnrollmentEvidence().GetHasMarks(),
		HasPositiveMark: cell.GetEnrollmentEvidence().GetHasPositiveMark(),
	}, value) {
		return blank
	}
	if value == "" {
		return blank
	}
	return value
}
