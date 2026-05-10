package form

import activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"

// BuildRateTypeOptions returns a []Option for the rate type select picker.
// The current value is pre-selected. Accepts both the short label form
// ("STANDARD") and the full proto enum form ("RATE_TYPE_STANDARD") for
// round-trip safety.
func BuildRateTypeOptions(labels RateTypeLabels, current string) []Option {
	type row struct {
		protoVal activitylaborpb.RateType
		strVal   string
		label    string
	}
	rows := []row{
		{activitylaborpb.RateType_RATE_TYPE_STANDARD, "RATE_TYPE_STANDARD", labels.Standard},
		{activitylaborpb.RateType_RATE_TYPE_OVERTIME, "RATE_TYPE_OVERTIME", labels.Overtime},
		{activitylaborpb.RateType_RATE_TYPE_PREMIUM, "RATE_TYPE_PREMIUM", labels.Premium},
	}

	opts := make([]Option, 0, len(rows))
	for _, r := range rows {
		selected := current == r.strVal ||
			current == r.protoVal.String() ||
			current == shortRateType(r.protoVal)
		opts = append(opts, Option{
			Value:    r.strVal,
			Label:    r.label,
			Selected: selected,
		})
	}
	return opts
}

// RateTypeLabels carries the display labels used by BuildRateTypeOptions.
// Populated from fayna.ActivityLaborLabels.Form.RateType* fields by the action layer.
type RateTypeLabels struct {
	Standard string
	Overtime string
	Premium  string
}

// RateTypeFromString converts a form string to the RateType enum.
// Accepts full proto enum names ("RATE_TYPE_STANDARD"), short forms
// ("STANDARD", "standard"), and numeric strings.
func RateTypeFromString(s string) activitylaborpb.RateType {
	switch s {
	case "RATE_TYPE_STANDARD", "STANDARD", "standard":
		return activitylaborpb.RateType_RATE_TYPE_STANDARD
	case "RATE_TYPE_OVERTIME", "OVERTIME", "overtime":
		return activitylaborpb.RateType_RATE_TYPE_OVERTIME
	case "RATE_TYPE_PREMIUM", "PREMIUM", "premium":
		return activitylaborpb.RateType_RATE_TYPE_PREMIUM
	default:
		return activitylaborpb.RateType_RATE_TYPE_UNSPECIFIED
	}
}

// RateTypeToString converts a RateType enum to its canonical form string
// (the value stored in hidden inputs and POSTed by the form).
func RateTypeToString(rt activitylaborpb.RateType) string {
	switch rt {
	case activitylaborpb.RateType_RATE_TYPE_STANDARD:
		return "RATE_TYPE_STANDARD"
	case activitylaborpb.RateType_RATE_TYPE_OVERTIME:
		return "RATE_TYPE_OVERTIME"
	case activitylaborpb.RateType_RATE_TYPE_PREMIUM:
		return "RATE_TYPE_PREMIUM"
	default:
		return ""
	}
}

// shortRateType returns the short name used in legacy / test contexts.
func shortRateType(rt activitylaborpb.RateType) string {
	switch rt {
	case activitylaborpb.RateType_RATE_TYPE_STANDARD:
		return "standard"
	case activitylaborpb.RateType_RATE_TYPE_OVERTIME:
		return "overtime"
	case activitylaborpb.RateType_RATE_TYPE_PREMIUM:
		return "premium"
	default:
		return ""
	}
}
