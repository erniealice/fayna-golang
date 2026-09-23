package outcome_summary

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func translationLeafPaths(raw json.RawMessage, prefix string, out map[string]struct{}) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err == nil && object != nil {
		for key, value := range object {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			if err := translationLeafPaths(value, path, out); err != nil {
				return err
			}
		}
		return nil
	}
	out[prefix] = struct{}{}
	return nil
}

func labelStructLeafPaths(t reflect.Type, prefix string, out map[string]struct{}) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Tag.Get("json")
		if comma := strings.IndexByte(name, ','); comma >= 0 {
			name = name[:comma]
		}
		if name == "" || name == "-" {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		fieldType := field.Type
		if fieldType.Kind() == reflect.Struct {
			labelStructLeafPaths(fieldType, path, out)
			continue
		}
		out[path] = struct{}{}
	}
}

func TestLabelsLynguaContract(t *testing.T) {
	root := outcomeSummaryRepoRoot(t)
	commonPath := root + "/packages/lyngua/translations/en/common/outcome_summary.json"
	generalPath := root + "/packages/lyngua/translations/en/general/outcome_summary.json"
	educationPath := root + "/packages/lyngua/translations/en/education/outcome_summary.json"
	common := readOutcomeSummaryTranslation(t, commonPath)
	general := readOutcomeSummaryTranslation(t, generalPath)
	education := readOutcomeSummaryTranslation(t, educationPath)

	labelPaths := make(map[string]struct{})
	labelStructLeafPaths(reflect.TypeOf(Labels{}), "", labelPaths)
	commonPaths := make(map[string]struct{})
	for key, value := range common {
		if err := translationLeafPaths(value, key, commonPaths); err != nil {
			t.Fatalf("walk common label key %q: %v", key, err)
		}
	}
	generalPaths := make(map[string]struct{})
	for key, value := range general {
		if err := translationLeafPaths(value, key, generalPaths); err != nil {
			t.Fatalf("walk general label key %q: %v", key, err)
		}
	}
	educationPaths := make(map[string]struct{})
	for key, value := range education {
		if err := translationLeafPaths(value, key, educationPaths); err != nil {
			t.Fatalf("walk education label key %q: %v", key, err)
		}
	}

	for path := range commonPaths {
		if _, ok := labelPaths[path]; !ok {
			t.Errorf("common translation label path %q has no Labels json tag", path)
		}
	}
	for path := range generalPaths {
		if _, ok := labelPaths[path]; !ok {
			t.Errorf("general translation label path %q has no Labels json tag", path)
		}
	}
	for path := range educationPaths {
		if _, ok := labelPaths[path]; !ok {
			t.Errorf("education translation label path %q has no Labels json tag", path)
		}
	}
	for path := range labelPaths {
		if _, ok := commonPaths[path]; !ok {
			t.Errorf("Labels json path %q has no common translation value", path)
		}
	}
	if len(commonPaths) == 0 || len(educationPaths) == 0 {
		t.Fatal("common and education outcome_summary label paths must both be loaded")
	}
	if len(commonPaths) != len(labelPaths) {
		t.Fatalf("common label path count = %d, Labels path count = %d", len(commonPaths), len(labelPaths))
	}
}

func TestCategoryLabelTierVocabulary(t *testing.T) {
	root := outcomeSummaryRepoRoot(t)
	for _, tc := range []struct {
		tier string
		want string
	}{
		{tier: "general", want: "Job Category"},
		{tier: "education", want: "Grade Category"},
	} {
		t.Run(tc.tier, func(t *testing.T) {
			labels := readOutcomeSummaryTranslation(t, root+"/packages/lyngua/translations/en/"+tc.tier+"/outcome_summary.json")
			got, ok := translationValue(labels, "subscription_group_export.category_label")
			if !ok || got != tc.want {
				t.Fatalf("%s category_label = %q, present=%t; want %q", tc.tier, got, ok, tc.want)
			}
		})
	}
}

func TestCommonLynguaMatchesCompiledDefaults(t *testing.T) {
	root := outcomeSummaryRepoRoot(t)
	common := readOutcomeSummaryTranslation(t, root+"/packages/lyngua/translations/en/common/outcome_summary.json")
	defaults := DefaultLabels()
	want := map[string]string{
		"columns.job":                                 defaults.Columns.Job,
		"columns.determination":                       defaults.Columns.Determination,
		"columns.score":                               defaults.Columns.Score,
		"columns.scoring_method":                      defaults.Columns.ScoringMethod,
		"columns.total":                               defaults.Columns.Total,
		"columns.pass":                                defaults.Columns.Pass,
		"columns.fail":                                defaults.Columns.Fail,
		"columns.issued_by":                           defaults.Columns.IssuedBy,
		"empty.title":                                 defaults.Empty.Title,
		"empty.message":                               defaults.Empty.Message,
		"landing.download_action":                     defaults.Landing.DownloadAction,
		"subscription_group.download_action":          defaults.SubscriptionGroup.DownloadAction,
		"subscription_group.detail_link":              defaults.SubscriptionGroup.DetailLink,
		"subscription_group.category_tabs_aria_label": defaults.SubscriptionGroup.CategoryTabsAriaLabel,
		"client_card.title":                           defaults.Client.Title,
		"client_card.subtitle":                        defaults.Client.Subtitle,
		"client_card.subject_column":                  defaults.Client.SubjectColumn,
		"client_card.period_1":                        defaults.Client.Period1,
		"client_card.period_2":                        defaults.Client.Period2,
		"client_card.year_column":                     defaults.Client.YearColumn,
		"client_card.progress_column":                 defaults.Client.ProgressColumn,
		"client_card.final_column":                    defaults.Client.FinalColumn,
		"client_card.view_action":                     defaults.Client.ViewAction,
		"client_card.staff_label":                     defaults.Client.StaffLabel,
		"client_card.staff_plural_label":              defaults.Client.StaffPluralLabel,
		"client_card.uncategorized_band":              defaults.Client.UncategorizedBand,
	}
	for path, wantValue := range want {
		got, ok := translationValue(common, path)
		if !ok {
			t.Errorf("common translation is missing %q", path)
			continue
		}
		if got != wantValue {
			t.Errorf("common translation %q = %q, compiled default = %q", path, got, wantValue)
		}
	}
}

func translationValue(root map[string]json.RawMessage, path string) (string, bool) {
	parts := strings.Split(path, ".")
	var current json.RawMessage
	var ok bool
	for index, part := range parts {
		if index == 0 {
			current, ok = root[part]
		} else {
			var object map[string]json.RawMessage
			if json.Unmarshal(current, &object) != nil {
				return "", false
			}
			current, ok = object[part]
		}
		if !ok {
			return "", false
		}
	}
	var value string
	if json.Unmarshal(current, &value) != nil {
		return "", false
	}
	return value, true
}
