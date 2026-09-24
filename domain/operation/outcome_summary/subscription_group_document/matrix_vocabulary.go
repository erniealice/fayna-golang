package subscription_group_document

import (
	"encoding/json/v2"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// The group-matrix data source offers a vocabulary, not a layout. A template
// may use any supported token any number of times (or not at all); the
// validator only checks that each token is known and sits in its legal scope:
//
//	root           body paragraphs and static table cells, outside every loop
//	job_templates  inside a header table column-loop cell {{#job_templates}}…{{/job_templates}}
//	rows           in the one template row of the {{#rows}} table row loop
//	cells          inside a column-loop cell {{#cells}}…{{/cells}} of that template row
//
// Numbered tokens job_template{N}_* are the legacy fixed-slot view of the
// same columns; the highest N a template uses is its column capacity (D9).

type matrixScopeVocabulary struct {
	Text          []string          `json:"text"`
	NumberedText  []string          `json:"numbered_text,omitempty"`
	Style         map[string]string `json:"style"`
	NumberedStyle map[string]string `json:"numbered_style,omitempty"`
}

type matrixVocabulary struct {
	Profile struct {
		Key                     string `json:"key"`
		BindingJobCategoryScope string `json:"binding_job_category_scope"`
		PeriodCardinality       int    `json:"period_cardinality"`
		RowLoop                 string `json:"row_loop"`
		ColumnLoop              string `json:"column_loop"`
		RowColumnLoop           string `json:"row_column_loop"`
		MaxColumns              int    `json:"max_columns"`
		NumberedSlotPattern     string `json:"numbered_slot_pattern"`
	} `json:"profile"`
	Scopes map[string]matrixScopeVocabulary `json:"scopes"`
}

const (
	scopeRoot         = "root"
	scopeColumnHeader = "job_templates"
	scopeRow          = "rows"
	scopeCell         = "cells"
)

// matrixToken is one vocabulary entry: its scope and, for style tokens, the
// Word attribute kind it may fill ("bold", "fill", "color"; "" for text).
type matrixToken struct {
	scope, style string
}

var numberedMatrixToken = regexp.MustCompile(`^job_template([1-9][0-9]*)_([a-z_]+)$`)

// TemplateLayout is what a validated group-matrix template uses.
type TemplateLayout struct {
	// NumberedSlots is the highest job_template{N}_* slot the template uses
	// (0 when it uses none).
	NumberedSlots int
	// ColumnLoop reports a table column loop, which repeats for any column count.
	ColumnLoop bool
}

// FitsColumns reports whether the template shows every one of n columns. A
// numbered-slot template without a column loop holds at most NumberedSlots
// columns; a template that uses no column token at all deliberately shows none.
func (l TemplateLayout) FitsColumns(n int) bool {
	if l.ColumnLoop || l.NumberedSlots == 0 {
		return n <= MaxColumns
	}
	return n <= l.NumberedSlots
}

// ValidateGroupMatrixTemplate validates an uploaded or stored group-matrix
// DOCX and reports the layout it uses. Errors are *TemplateContractError.
func ValidateGroupMatrixTemplate(docx []byte) (TemplateLayout, error) {
	return validateGroupMatrixTemplate(docx, productionValidationLimits)
}

// CheckColumnCapacity validates the template and refuses (ColumnCapacityError)
// when it cannot show every one of the matrix's columns.
func CheckColumnCapacity(docx []byte, columns int) (TemplateLayout, error) {
	layout, err := ValidateGroupMatrixTemplate(docx)
	if err != nil {
		return TemplateLayout{}, err
	}
	if !layout.FitsColumns(columns) {
		return layout, &ColumnCapacityError{TemplateColumns: layout.NumberedSlots, DataColumns: columns}
	}
	return layout, nil
}

func validateGroupMatrixTemplate(docx []byte, limits validationLimits) (TemplateLayout, error) {
	vocabulary, lookup, err := loadMatrixVocabulary()
	if err != nil {
		return TemplateLayout{}, err
	}
	parts, err := readTemplateParts(docx, limits)
	if err != nil {
		return TemplateLayout{}, err
	}
	return validateMatrixVocabularyTokens(parts, vocabulary, lookup)
}

func loadMatrixVocabulary() (matrixVocabulary, func(string) (matrixToken, int, bool), error) {
	raw, err := ManifestBytes(GroupMatrixRenderProfile)
	if err != nil {
		return matrixVocabulary{}, nil, err
	}
	var vocabulary matrixVocabulary
	if err := json.Unmarshal(raw, &vocabulary); err != nil {
		return matrixVocabulary{}, nil, contractError("decode embedded manifest: %v", err)
	}
	p := vocabulary.Profile
	if p.Key != SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key || p.BindingJobCategoryScope != "exact_required" ||
		p.PeriodCardinality != 1 || p.RowLoop != scopeRow || p.ColumnLoop != scopeColumnHeader ||
		p.RowColumnLoop != scopeCell || p.MaxColumns != MaxColumns || p.NumberedSlotPattern != "job_template{n}_" {
		return matrixVocabulary{}, nil, contractError("embedded matrix manifest does not match the registered profile")
	}
	plain := make(map[string]matrixToken)
	numbered := make(map[string]matrixToken)
	for _, scope := range []string{scopeRoot, scopeColumnHeader, scopeRow, scopeCell} {
		node, ok := vocabulary.Scopes[scope]
		if !ok {
			return matrixVocabulary{}, nil, contractError("embedded matrix manifest is missing scope %q", scope)
		}
		add := func(into map[string]matrixToken, key string, token matrixToken) error {
			if _, duplicate := into[key]; duplicate || key == "" {
				return contractError("embedded matrix manifest repeats token %q", key)
			}
			if token.style != "" && token.style != "bold" && token.style != "fill" && token.style != "color" {
				return contractError("embedded matrix manifest has unknown style kind %q", token.style)
			}
			into[key] = token
			return nil
		}
		for _, key := range node.Text {
			if err := add(plain, key, matrixToken{scope: scope}); err != nil {
				return matrixVocabulary{}, nil, err
			}
		}
		for key, kind := range node.Style {
			if err := add(plain, key, matrixToken{scope: scope, style: kind}); err != nil {
				return matrixVocabulary{}, nil, err
			}
		}
		for _, suffix := range node.NumberedText {
			if err := add(numbered, suffix, matrixToken{scope: scope}); err != nil {
				return matrixVocabulary{}, nil, err
			}
		}
		for suffix, kind := range node.NumberedStyle {
			if err := add(numbered, suffix, matrixToken{scope: scope, style: kind}); err != nil {
				return matrixVocabulary{}, nil, err
			}
		}
	}
	lookup := func(key string) (matrixToken, int, bool) {
		if token, ok := plain[key]; ok {
			return token, 0, true
		}
		match := numberedMatrixToken.FindStringSubmatch(key)
		if match == nil {
			return matrixToken{}, 0, false
		}
		slot, err := strconv.Atoi(match[1])
		if err != nil || slot < 1 || slot > MaxColumns {
			return matrixToken{}, 0, false
		}
		token, ok := numbered[match[2]]
		return token, slot, ok
	}
	return vocabulary, lookup, nil
}

// occurrenceLocation names where a token sits, for upload error messages.
func occurrenceLocation(o tokenOccurrence) string {
	if o.table > 0 {
		if o.column > 0 {
			return fmt.Sprintf("table %d, row %d, cell %d", o.table, o.row, o.column)
		}
		return fmt.Sprintf("table %d, row %d", o.table, o.row)
	}
	if o.paragraph > 0 {
		return fmt.Sprintf("paragraph %d", o.paragraph)
	}
	return o.part
}

type columnLoopCell struct {
	loop        string
	open, close tokenOccurrence
}

func validateMatrixVocabularyTokens(parts map[string][]byte, vocabulary matrixVocabulary, lookup func(string) (matrixToken, int, bool)) (TemplateLayout, error) {
	var occurrences []tokenOccurrence
	rowTexts := make(map[rowLocation]string)
	completeCount := 0
	for part, body := range parts {
		scanner := &partScanner{
			part:          part,
			completeCount: len(completeTemplateToken.FindAll(body, -1)),
			rowText:       make(map[rowPosition]*strings.Builder),
			paragraphText: make(map[int]*strings.Builder),
		}
		if err := scanner.scan(body); err != nil {
			return TemplateLayout{}, err
		}
		completeCount += scanner.completeCount
		occurrences = append(occurrences, scanner.occurrences...)
		for position, value := range scanner.rowText {
			rowTexts[rowLocation{part: part, table: position.table, row: position.row}] = value.String()
		}
	}
	if completeCount != len(occurrences) {
		return TemplateLayout{}, contractError("template token is split, partial, or outside a supported XML node")
	}
	sort.SliceStable(occurrences, func(i, j int) bool { return occurrences[i].order < occurrences[j].order })

	rowLoop, headerLoop, cellLoop := vocabulary.Profile.RowLoop, vocabulary.Profile.ColumnLoop, vocabulary.Profile.RowColumnLoop
	markers := map[string][]tokenOccurrence{}
	var tokens []tokenOccurrence
	for _, occurrence := range occurrences {
		if occurrence.part != "word/document.xml" {
			return TemplateLayout{}, contractError("template token %q is outside word/document.xml (%s)", occurrence.key, occurrence.part)
		}
		switch occurrence.key {
		case "#" + rowLoop, "/" + rowLoop, "#" + headerLoop, "/" + headerLoop, "#" + cellLoop, "/" + cellLoop:
			if err := validateMarkerPlacement(occurrence); err != nil {
				return TemplateLayout{}, contractError("loop marker %q is misplaced at %s", occurrence.key, occurrenceLocation(occurrence))
			}
			markers[occurrence.key] = append(markers[occurrence.key], occurrence)
			continue
		}
		if strings.HasPrefix(occurrence.key, "#") || strings.HasPrefix(occurrence.key, "/") || strings.HasPrefix(occurrence.key, "^") {
			return TemplateLayout{}, contractError("unknown loop %q at %s", occurrence.key, occurrenceLocation(occurrence))
		}
		if _, _, known := lookup(occurrence.key); !known {
			return TemplateLayout{}, contractError("unknown template token %q at %s", occurrence.key, occurrenceLocation(occurrence))
		}
		tokens = append(tokens, occurrence)
	}

	// Row loop: optional; when present, one start row, one template row and
	// one end row in a depth-1 table, marker rows holding only their marker.
	var hasRowLoop bool
	var loopTable, loopStartRow, templateRow int
	start, end := markers["#"+rowLoop], markers["/"+rowLoop]
	switch {
	case len(start) == 0 && len(end) == 0:
	case len(start) == 1 && len(end) == 1:
		if start[0].table <= 0 || start[0].tableDepth != 1 || start[0].table != end[0].table ||
			end[0].tableDepth != 1 || start[0].row <= 0 || end[0].row != start[0].row+2 {
			return TemplateLayout{}, contractError("row loop must have one start row, one template row, and one end row (%s)", occurrenceLocation(start[0]))
		}
		if strings.TrimSpace(rowTexts[occurrenceRowLocation(start[0])]) != "{{#"+rowLoop+"}}" ||
			strings.TrimSpace(rowTexts[occurrenceRowLocation(end[0])]) != "{{/"+rowLoop+"}}" {
			return TemplateLayout{}, contractError("row loop marker rows must contain only their marker text")
		}
		hasRowLoop = true
		loopTable, loopStartRow, templateRow = start[0].table, start[0].row, start[0].row+1
	default:
		return TemplateLayout{}, contractError("row loop markers {{#%s}} and {{/%s}} must appear exactly once as a pair", rowLoop, rowLoop)
	}
	inLoopRows := func(o tokenOccurrence) bool {
		return hasRowLoop && o.table == loopTable && o.row >= loopStartRow && o.row <= loopStartRow+2
	}
	inTemplateRow := func(o tokenOccurrence) bool {
		return hasRowLoop && o.table == loopTable && o.row == templateRow && o.tableDepth == 1
	}

	// Column-loop cells: each pair opens and closes in one depth-1 table cell;
	// header loops sit outside the row loop, row-cell loops in its template row.
	loopCells := make(map[int]columnLoopCell)
	loopColumnByTable := make(map[int]int)
	loopCellByRow := make(map[rowPosition]int)
	for _, loop := range []string{headerLoop, cellLoop} {
		opens, closes := markers["#"+loop], markers["/"+loop]
		if len(opens) != len(closes) {
			return TemplateLayout{}, contractError("column loop markers {{#%s}} and {{/%s}} must be paired", loop, loop)
		}
		closeByCell := make(map[int]tokenOccurrence, len(closes))
		for _, c := range closes {
			if _, duplicate := closeByCell[c.cell]; duplicate || c.cell == 0 {
				return TemplateLayout{}, contractError("column loop {{/%s}} must close once in a table cell (%s)", loop, occurrenceLocation(c))
			}
			closeByCell[c.cell] = c
		}
		for _, o := range opens {
			c, found := closeByCell[o.cell]
			if o.cell == 0 || !found || c.order < o.order || o.tableDepth != 1 {
				return TemplateLayout{}, contractError("column loop {{#%s}} must open and close in one table cell (%s)", loop, occurrenceLocation(o))
			}
			if _, duplicate := loopCells[o.cell]; duplicate {
				return TemplateLayout{}, contractError("a table cell may hold only one column loop (%s)", occurrenceLocation(o))
			}
			if loop == cellLoop && !inTemplateRow(o) {
				return TemplateLayout{}, contractError("column loop {{#%s}} must sit in the {{#%s}} template row (%s)", loop, rowLoop, occurrenceLocation(o))
			}
			if loop == headerLoop && inLoopRows(o) {
				return TemplateLayout{}, contractError("column loop {{#%s}} must sit outside the {{#%s}} row loop (%s)", loop, rowLoop, occurrenceLocation(o))
			}
			position := rowPosition{table: o.table, row: o.row}
			if _, duplicate := loopCellByRow[position]; duplicate {
				return TemplateLayout{}, contractError("a table row may hold only one column-loop cell (%s)", occurrenceLocation(o))
			}
			loopCellByRow[position] = o.cell
			if column, seen := loopColumnByTable[o.table]; seen && column != o.column {
				return TemplateLayout{}, contractError("every column loop of a table must sit in the same cell position (%s)", occurrenceLocation(o))
			}
			loopColumnByTable[o.table] = o.column
			loopCells[o.cell] = columnLoopCell{loop: loop, open: o, close: c}
			delete(closeByCell, o.cell)
		}
		if len(closeByCell) != 0 {
			return TemplateLayout{}, contractError("column loop markers {{#%s}} and {{/%s}} must be paired", loop, loop)
		}
	}

	layout := TemplateLayout{ColumnLoop: len(loopCells) != 0}
	for _, o := range tokens {
		token, slot, _ := lookup(o.key)
		if slot > layout.NumberedSlots {
			layout.NumberedSlots = slot
		}
		cell, inLoopCell := loopCells[o.cell]
		textInsideMarkers := inLoopCell && o.order > cell.open.order && o.order < cell.close.order
		var legal bool
		switch token.scope {
		case scopeRoot:
			legal = !inLoopCell && !inLoopRows(o) && validateRootTokenPlacement(o) == nil
		case scopeColumnHeader:
			legal = inLoopCell && cell.loop == headerLoop && textInsideMarkers
		case scopeRow:
			legal = inTemplateRow(o) && !inLoopCell
		case scopeCell:
			legal = inLoopCell && cell.loop == cellLoop && (o.kind != "text" || textInsideMarkers)
		}
		if !legal {
			return TemplateLayout{}, contractError("token %q is outside its %s scope at %s", o.key, token.scope, occurrenceLocation(o))
		}
		if err := validateMatrixTokenKind(o, token); err != nil {
			return TemplateLayout{}, err
		}
	}
	return layout, nil
}

// validateMatrixTokenKind checks a token's XML position: text tokens fill a
// run's w:t; style tokens fill exactly the whitelisted Word attribute.
func validateMatrixTokenKind(o tokenOccurrence, token matrixToken) error {
	cellPath := []string{"document", "body", "tbl", "tr", "tc"}
	var ok bool
	switch token.style {
	case "":
		ok = o.kind == "text" && isWordprocessingMLName(o.element, "t")
	case "bold":
		ok = o.kind == "attribute" && isWordprocessingMLName(o.element, "b") && isWordprocessingMLName(o.attribute, "val") &&
			matchesWordprocessingMLPath(o.ancestry, append(cellPath, "p", "r", "rPr", "b")...)
	case "fill":
		ok = o.kind == "attribute" && isWordprocessingMLName(o.element, "shd") && isWordprocessingMLName(o.attribute, "fill") &&
			matchesWordprocessingMLPath(o.ancestry, append(cellPath, "tcPr", "shd")...)
	case "color":
		ok = o.kind == "attribute" && isWordprocessingMLName(o.element, "color") && isWordprocessingMLName(o.attribute, "val") &&
			matchesWordprocessingMLPath(o.ancestry, append(cellPath, "p", "r", "rPr", "color")...)
	}
	if !ok {
		if token.style == "" {
			return contractError("text token %q must be run text at %s", o.key, occurrenceLocation(o))
		}
		return contractError("style token %q is misplaced at %s", o.key, occurrenceLocation(o))
	}
	if token.style == "" && token.scope != scopeRoot &&
		!matchesWordprocessingMLPath(o.ancestry, append(cellPath, "p", "r", "t")...) {
		return contractError("text token %q must be run text in a table cell at %s", o.key, occurrenceLocation(o))
	}
	return nil
}
