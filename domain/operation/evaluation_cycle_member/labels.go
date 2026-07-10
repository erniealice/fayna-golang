package evaluation_cycle_member

// labels.go — EvaluationCycleMember label structs + DefaultLabels constructor.
//
// evaluation_cycle_member has NO standalone list page (STR-1): it surfaces only
// via the cycle Members tab + the X-of-Y banner projection. These labels feed
// the members-tab.html partial.

// Labels holds all translatable strings for the cycle-member surface.
type Labels struct {
	Tab     TabLabels    `json:"tab"`
	Columns ColumnLabels `json:"columns"`
	Empty   EmptyLabels  `json:"empty"`
	Badges  BadgeLabels  `json:"badges"`
}

type TabLabels struct {
	Title    string `json:"title"`
	CountFmt string `json:"count_fmt"` // "%d members"
}

type ColumnLabels struct {
	Associate string `json:"associate"`
	Client    string `json:"client"`
	Probation string `json:"probation"`
}

type EmptyLabels struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type BadgeLabels struct {
	Probation string `json:"probation"`
}

// DefaultLabels returns Labels with sensible English defaults.
func DefaultLabels() Labels {
	return Labels{
		Tab: TabLabels{
			Title:    "Members",
			CountFmt: "%d members",
		},
		Columns: ColumnLabels{
			Associate: "Associate",
			Client:    "Client",
			Probation: "Probation",
		},
		Empty: EmptyLabels{
			Title:   "No members enrolled",
			Message: "Open the cycle to enroll members from active seats.",
		},
		Badges: BadgeLabels{
			Probation: "Probation",
		},
	}
}
