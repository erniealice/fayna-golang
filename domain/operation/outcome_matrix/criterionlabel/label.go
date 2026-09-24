package criterionlabel

import (
	"strings"
	"unicode"
)

// Label derives a criterion's display label from its position in an activity.
func Label(name string, index int, prefixPattern string, breakAfter string) string {
	label := name
	if prefixPattern != "" && index >= 0 && index < 26 {
		letter := string(rune('A' + index))
		label = strings.Replace(prefixPattern, "{letter}", letter, 1) + name
	}
	if breakAfter == "" {
		return label
	}
	for offset, r := range label {
		if strings.ContainsRune(breakAfter, r) {
			after := offset + len(string(r))
			return label[:after] + "\n" + strings.TrimLeftFunc(label[after:], unicode.IsSpace)
		}
	}
	return label
}
