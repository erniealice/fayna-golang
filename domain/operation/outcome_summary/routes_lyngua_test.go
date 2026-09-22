package outcome_summary

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func outcomeSummaryRepoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "packages", "lyngua", "translations", "en")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate repository root from test source")
	return ""
}

func readOutcomeSummaryTranslation(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()
	document := readTranslationDocument(t, path)

	var outcomeSummary map[string]json.RawMessage
	if raw, ok := document["outcome_summary"]; ok {
		if err := json.Unmarshal(raw, &outcomeSummary); err != nil {
			t.Fatalf("decode outcome_summary in %s: %v", path, err)
		}
	}
	if outcomeSummary == nil {
		t.Fatalf("translation %s has no outcome_summary object", path)
	}
	return outcomeSummary
}

func readTranslationDocument(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read translation %s: %v", path, err)
	}

	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode translation %s: %v", path, err)
	}
	return document
}

func jsonTaggedFields(t reflect.Type) map[string]struct{} {
	fields := make(map[string]struct{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Tag.Get("json")
		if name == "" || name == "-" {
			continue
		}
		if comma := len(name); comma > 0 {
			for j := 0; j < len(name); j++ {
				if name[j] == ',' {
					name = name[:j]
					break
				}
			}
		}
		fields[name] = struct{}{}
	}
	return fields
}

func TestRoutesLynguaContract(t *testing.T) {
	root := outcomeSummaryRepoRoot(t)
	routePaths, err := filepath.Glob(filepath.Join(root, "packages", "lyngua", "translations", "en", "*", "route.json"))
	if err != nil {
		t.Fatalf("glob route translations: %v", err)
	}
	if len(routePaths) == 0 {
		t.Fatal("no route translations found")
	}
	knownKeys := make(map[string]struct{})
	tierKeys := make(map[string]map[string]json.RawMessage)
	for _, path := range routePaths {
		document := readTranslationDocument(t, path)
		raw, ok := document["outcome_summary"]
		if !ok {
			continue
		}
		var outcomeSummary map[string]json.RawMessage
		if err := json.Unmarshal(raw, &outcomeSummary); err != nil {
			t.Fatalf("decode outcome_summary in %s: %v", path, err)
		}
		tier := filepath.Base(filepath.Dir(path))
		tierKeys[tier] = outcomeSummary
		for key := range outcomeSummary {
			knownKeys[key] = struct{}{}
		}
	}
	for _, tier := range []string{"general", "education"} {
		if _, ok := tierKeys[tier]; !ok {
			t.Fatalf("%s/route.json has no outcome_summary object", tier)
		}
	}

	routeFields := jsonTaggedFields(reflect.TypeOf(Routes{}))
	genericOnly := map[string]struct{}{
		"active_nav":          {},
		"active_sub_nav":      {},
		"list_active_sub_nav": {},
		"job_summary_url":     {},
		"phase_summary_url":   {},
	}

	for tier, keys := range tierKeys {
		for key := range keys {
			if _, ok := routeFields[key]; !ok {
				t.Errorf("%s outcome_summary route key %q has no Routes json tag", tier, key)
			}
		}
	}
	for key := range routeFields {
		if _, ok := knownKeys[key]; ok {
			continue
		}
		if _, ok := genericOnly[key]; ok {
			continue
		}
		t.Errorf("Routes json tag %q has no outcome_summary route key in any tier", key)
	}
}
