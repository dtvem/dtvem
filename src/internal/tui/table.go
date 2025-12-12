package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Table represents a simple table for displaying data
type Table struct {
	title      string
	headers    []string
	rows       []TableRow
	widths     []int
	hideHeader bool
	minWidth   int
}

// TableRow represents a single row in the table
type TableRow struct {
	cells  []string
	active bool
}

// NewTable creates a new table with the given headers
func NewTable(headers ...string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = lipgloss.Width(h)
	}
	return &Table{
		headers: headers,
		widths:  widths,
	}
}

// SetTitle sets a title that spans all columns at the top of the table
func (t *Table) SetTitle(title string) {
	t.title = title
}

// HideHeader hides the column header row
func (t *Table) HideHeader() {
	t.hideHeader = true
}

// SetMinWidth sets a minimum width for the table content
func (t *Table) SetMinWidth(width int) {
	t.minWidth = width
}

// AddRow adds a row to the table
func (t *Table) AddRow(cells ...string) {
	t.addRowInternal(cells, false)
}

// AddActiveRow adds an active/highlighted row to the table
func (t *Table) AddActiveRow(cells ...string) {
	t.addRowInternal(cells, true)
}

func (t *Table) addRowInternal(cells []string, active bool) {
	// Pad or truncate to match header count
	row := make([]string, len(t.headers))
	for i := range row {
		if i < len(cells) {
			row[i] = cells[i]
			// Use lipgloss.Width for accurate width calculation with ANSI codes
			cellWidth := lipgloss.Width(cells[i])
			if cellWidth > t.widths[i] {
				t.widths[i] = cellWidth
			}
		}
	}
	t.rows = append(t.rows, TableRow{cells: row, active: active})
}

// Render returns the rendered table as a string
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}

	// Ensure styles are initialized
	initStyles()

	// Calculate total width for title
	totalWidth := 0
	for _, w := range t.widths {
		totalWidth += w + 2
	}

	// Apply minimum width if set (pad the last column)
	if t.minWidth > 0 && totalWidth < t.minWidth {
		diff := t.minWidth - totalWidth
		t.widths[len(t.widths)-1] += diff
		totalWidth = t.minWidth
	}

	var lines []string

	// Render title if set
	if t.title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Width(totalWidth).
			Align(lipgloss.Center)
		lines = append(lines, titleStyle.Render(t.title))
		lines = append(lines, StyleMuted.Render(strings.Repeat("─", totalWidth)))
	}

	// Render headers (unless hidden)
	if !t.hideHeader {
		var headerLine string
		for i, h := range t.headers {
			style := StyleTableHeader.Width(t.widths[i] + 2)
			headerLine += style.Render(h)
		}
		lines = append(lines, headerLine)

		// Render separator
		var sepLine string
		for i := range t.headers {
			sepLine += StyleMuted.Render(strings.Repeat("─", t.widths[i]+2))
		}
		lines = append(lines, sepLine)
	}

	// Render rows
	for _, row := range t.rows {
		var rowLine string
		for i, cell := range row.cells {
			style := StyleTableCell.Width(t.widths[i] + 2)
			if row.active {
				style = style.Foreground(lipgloss.Color("42")) // Green for active
			}
			rowLine += style.Render(cell)
		}
		lines = append(lines, rowLine)
	}

	// Join lines and wrap in a border
	content := strings.Join(lines, "\n")
	return StyleTableBorder.Render(content)
}

// RowCount returns the number of data rows (excluding header)
func (t *Table) RowCount() int {
	return len(t.rows)
}
