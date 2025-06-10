package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ANSI regex to remove color codes for width calculations
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// StripANSI removes ANSI escape codes from a string
func StripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

// PadRight pads `s` with spaces so its visible (ANSI-stripped) length is exactly `width`.
func PadRight(s string, width int) string {
	visible := StripANSI(s)
	pad := width - len([]rune(visible))
	if pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
}

// Color codes
const (
	Reset = "\033[0m"
	Red   = "\033[31m"
	Green = "\033[32m"
)

// ColorizeBool returns a green "true" or red "false"
func ColorizeBool(b bool) string {
	if b {
		return Green + "TRUE" + Reset
	}
	return Red + "FALSE" + Reset
}

// Table represents a printable table
type Table struct {
	header    []string
	rows      [][]string
	separator string // e.g. "  ", " | "
}

// NewTable creates a new Table with a header and a separator
func NewTable(header []string, separator string) *Table {
	return &Table{
		header:    header,
		rows:      make([][]string, 0),
		separator: separator,
	}
}

// AddRow appends a new row to the table. Convert bools or other types to strings before adding.
func (t *Table) AddRow(cells ...string) {
	t.rows = append(t.rows, cells)
}

// Render computes column widths, prints the header, a separator line, and all rows.
func (t *Table) Render() {
	colCount := len(t.header)
	// compute max width per column
	widths := make([]int, colCount)
	for i, h := range t.header {
		widths[i] = len([]rune(StripANSI(h)))
	}
	for _, row := range t.rows {
		for i, cell := range row {
			visLen := len([]rune(StripANSI(cell)))
			if visLen > widths[i] {
				widths[i] = visLen
			}
		}
	}

	// print header
	for i, h := range t.header {
		fmt.Print(PadRight(h, widths[i]))
		if i < colCount-1 {
			fmt.Print(t.separator)
		}
	}
	fmt.Println()

	// separator line
	for i := range widths {
		fmt.Print(strings.Repeat("-", widths[i]))
		if i < colCount-1 {
			// match separator length with "-+-" if separator is " | "
			dash := strings.Repeat("-", len(t.separator))
			// center dash under separator
			fmt.Print(dash)
		}
	}
	fmt.Println()

	// data rows
	for _, row := range t.rows {
		for i, cell := range row {
			fmt.Print(PadRight(cell, widths[i]))
			if i < colCount-1 {
				fmt.Print(t.separator)
			}
		}
		fmt.Println()
	}
}
