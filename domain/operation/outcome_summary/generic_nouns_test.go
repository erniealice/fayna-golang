package outcome_summary

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestNoVerticalNounsInGenericCode(t *testing.T) {
	root := outcomeSummaryRepoRoot(t)
	target := filepath.Join(root, "packages", "fayna-golang", "domain", "operation", "outcome_summary")
	pattern := regexp.MustCompile(`(?i)\b(section|student)\b`)

	var violations []string
	err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}
		if filepath.Ext(info.Name()) != ".go" && filepath.Ext(info.Name()) != ".html" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		lineNumber := 0
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lineNumber++
			line := scanner.Text()
			if !pattern.MatchString(line) || allowedVerticalNounLine(path, line) {
				continue
			}
			violations = append(violations, filepath.ToSlash(path)+":"+itoa(lineNumber)+": "+strings.TrimSpace(line))
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) > 0 {
		t.Fatalf("generic outcome_summary code contains vertical nouns:\n%s", strings.Join(violations, "\n"))
	}
}

func allowedVerticalNounLine(path, line string) bool {
	slashPath := filepath.ToSlash(path)
	if strings.Contains(slashPath, "/document/") && strings.Contains(line, "Student") {
		return true
	}
	if strings.HasSuffix(slashPath, "/document/formation.go") && strings.Contains(strings.ToUpper(line), "STUDENT FORMATION") {
		return true
	}
	if strings.Contains(line, "sectionStyleMarker") || strings.Contains(line, "sectionLibreMarker") {
		return true
	}
	if strings.Contains(line, "form-section") || strings.Contains(line, "os-narrative-section") || strings.Contains(line, "os-section-heading") {
		return true
	}
	return false
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[index:])
}
