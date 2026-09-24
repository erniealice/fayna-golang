package subscription_group_document

import (
	"os"
	"strings"
	"testing"
)

// clientPhaseDocXML builds a minimal client-phase template: root scalars, the
// jobs → assessments body loops with their required scalars, plus the extras.
func clientPhaseDocXML(jobExtras, assessmentExtras, trailing []string, withDescriptions, withSections bool) string {
	p := func(text string) string { return `<w:p><w:r><w:t>` + text + `</w:t></w:r></w:p>` }
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, key := range clientPhaseRootScalars {
		b.WriteString(p("{{" + key + "}}"))
	}
	b.WriteString(p("{{#jobs}}"))
	for _, key := range append(append([]string{}, clientPhaseJobScalars...), jobExtras...) {
		b.WriteString(p("{{" + key + "}}"))
	}
	b.WriteString(p("{{#assessments}}"))
	for _, key := range append(append([]string{}, clientPhaseAssessmentScalars...), assessmentExtras...) {
		b.WriteString(p("{{" + key + "}}"))
	}
	if withDescriptions {
		b.WriteString(p("{{#rating_descriptions}}"))
		for _, key := range clientPhaseRatingDescriptionScalars {
			b.WriteString(p("{{" + key + "}}"))
		}
		b.WriteString(p("{{/rating_descriptions}}"))
	}
	b.WriteString(p("{{/assessments}}"))
	b.WriteString(p("{{/jobs}}"))
	if withSections {
		b.WriteString(p("{{#outcome_sections}}") + p("{{section_code}}") + p("{{#rows}}") + p("{{row_name}}") +
			p("{{#cells}}") + p("{{period_name}}") + p("{{task_name}}") + p("{{value}}") + p("{{/cells}}") +
			p("{{total}}") + p("{{/rows}}") + p("{{/outcome_sections}}"))
	}
	for _, text := range trailing {
		b.WriteString(p(text))
	}
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func clientPhaseDocx(t *testing.T, xml string) []byte {
	t.Helper()
	return makeTestDOCX(t,
		testZipPart{name: "[Content_Types].xml", body: `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="xml" ContentType="application/xml"/></Types>`},
		testZipPart{name: "word/document.xml", body: xml},
	)
}

// Owner 2026-09-24: a layout may omit the category label, the rating-band lines
// and the outcome-section appendix, and may repeat identity per job page.
func TestClientPhaseValidator_OptionalScalarsAndLoops(t *testing.T) {
	cases := []struct {
		name    string
		xml     string
		wantErr string
	}{
		{"minimal: no category, no bands, no sections", clientPhaseDocXML(nil, nil, nil, false, false), ""},
		{"all optional tokens present", clientPhaseDocXML(clientPhaseJobOptionalScalars, clientPhaseAssessmentOptionalScalars, nil, true, true), ""},
		{"optional scalar duplicated", clientPhaseDocXML([]string{"job_category_name", "job_category_name"}, nil, nil, false, false), "must occur exactly once"},
		{"root scalar repeated outside loops", clientPhaseDocXML(nil, nil, []string{"{{student_name}}", "{{adviser}}"}, false, false), ""},
		{"root scalar inside a loop", strings.Replace(clientPhaseDocXML(nil, nil, nil, false, false), "{{teacher_name}}", "{{teacher_name}}</w:t></w:r></w:p><w:p><w:r><w:t>{{student_name}}", 1), "inside loop"},
		{"omitted loop's token used outside it", clientPhaseDocXML(nil, []string{"{{description}}"}[:0], []string{"{{description}}"}, false, false), "outside loop"},
		{"required job scalar missing", strings.Replace(clientPhaseDocXML(nil, nil, nil, false, false), "{{teacher_name}}", "teacher", 1), `"teacher_name" is missing`},
		{"required loop omitted", strings.NewReplacer("{{#assessments}}", "", "{{/assessments}}", "", "{{assessment_name}}", "", "{{achievement_level}}", "", "{{comment}}", "").Replace(clientPhaseDocXML(nil, nil, nil, false, false)), "missing"},
		{"optional loop with one marker only", strings.Replace(clientPhaseDocXML(nil, nil, nil, true, false), "{{/rating_descriptions}}", "", 1), "rating_descriptions"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTemplate(testClientPhaseProfile, clientPhaseDocx(t, tc.xml))
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateTemplate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ValidateTemplate() = %v, want error containing %q", err, tc.wantErr)
			}
		})
	}
}

// TestJHSClientPhaseAuthoringTemplateV3MatchesManifest validates the v3 layout
// (owner 2026-09-24): per-page identity, two-column subject heading, criteria
// as a table-row loop, no category label and no rating-band lines.
func TestJHSClientPhaseAuthoringTemplateV3MatchesManifest(t *testing.T) {
	docx, err := os.ReadFile("../../../../../../docs/plan/20260923-individual-report-card-downloads/artifacts/JHS Progress Report - MMIS Template v3.docx")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateTemplate(testClientPhaseProfile, docx); err != nil {
		t.Fatalf("JHS client phase v3 authoring asset: %v", err)
	}
}

// TestJHSClientPhaseAuthoringTemplateV4MatchesManifest validates the v4 layout
// (owner 2026-09-24, output-fidelity addendum R1–R5, R7, R8, R11–R13): cover
// academic year in the cover header, printed-by only in the footer, sample
// identity grid, teacher line on its own row, and the required-but-unprinted
// job tokens kept in hidden runs.
func TestJHSClientPhaseAuthoringTemplateV4MatchesManifest(t *testing.T) {
	docx, err := os.ReadFile("../../../../../../docs/plan/20260923-individual-report-card-downloads/artifacts/JHS Progress Report - MMIS Template v4.docx")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateTemplate(testClientPhaseProfile, docx); err != nil {
		t.Fatalf("JHS client phase v4 authoring asset: %v", err)
	}
}
