package rating_description_set_entry

import "github.com/erniealice/espyna-golang/consumer/compose"

// Describe returns the compose Unit for the entry drawer module.
// No Nav — the entry has no standalone page; it surfaces only via the
// rating_description_set detail Descriptors matrix + this drawer
// (evaluation_template_item pattern).
func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "operation.rating_description_set_entry",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "rating_description_set_entry"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "rating_description_set_entry.json", Key: "rating_description_set_entry"},
		LabelName: "RatingDescriptionSetEntryLabels",
		Templates: TemplatesFS,
	}
}
