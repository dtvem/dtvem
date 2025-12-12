// Package tui provides styled console output using lipgloss for rich terminal UI.
// This package is only imported by the main dtvem CLI, not the shim executable,
// to keep the shim binary size minimal.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	colorPrimary   = lipgloss.Color("39")  // Cyan
	colorSecondary = lipgloss.Color("213") // Magenta/Pink
	colorSuccess   = lipgloss.Color("42")  // Green
	colorWarning   = lipgloss.Color("214") // Orange/Yellow
	colorError     = lipgloss.Color("196") // Red
	colorMuted     = lipgloss.Color("245") // Gray
)

// Text styles
var (
	// StyleTitle is for main headers and titles
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	// StyleSubtitle is for secondary headers
	StyleSubtitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary)

	// StyleVersion is for version numbers
	StyleVersion = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary)

	// StyleActiveVersion is for currently active versions
	StyleActiveVersion = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSuccess)

	// StyleRuntime is for runtime names (Python, Node.js, etc.)
	StyleRuntime = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	// StyleMuted is for dimmed/secondary text
	StyleMuted = lipgloss.NewStyle().
			Foreground(colorMuted)

	// StyleIndicator is for status indicators (checkmarks, etc.)
	StyleIndicator = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// StyleWarningIndicator is for warning indicators
	StyleWarningIndicator = lipgloss.NewStyle().
				Foreground(colorWarning)
)

// Box styles
var (
	// StyleBox is a generic rounded box
	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	// StyleInfoBox is for informational content
	StyleInfoBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	// StyleSuccessBox is for success messages
	StyleSuccessBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSuccess).
			Padding(0, 1)

	// StyleErrorBox is for error messages
	StyleErrorBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorError).
			Padding(0, 1)
)

// Table styles
var (
	// StyleTableHeader is for table column headers
	StyleTableHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				PaddingRight(2)

	// StyleTableCell is for regular table cells
	StyleTableCell = lipgloss.NewStyle().
			PaddingRight(2)

	// StyleTableRowActive is for highlighted/active rows
	StyleTableRowActive = lipgloss.NewStyle().
				Foreground(colorSuccess)

	// StyleTableBorder wraps the table in a rounded border
	StyleTableBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorMuted).
				Padding(0, 1)
)

// Indicator constants with styles applied
var (
	// CheckMark is a styled checkmark indicator
	CheckMark = StyleIndicator.Render("✓")

	// CrossMark is a styled cross/X indicator
	CrossMark = lipgloss.NewStyle().Foreground(colorError).Render("✗")

	// Bullet is a styled bullet point
	Bullet = StyleMuted.Render("•")

	// Arrow is a styled arrow indicator
	Arrow = lipgloss.NewStyle().Foreground(colorPrimary).Render("→")
)

// RenderTitle renders a styled title
func RenderTitle(text string) string {
	return StyleTitle.Render(text)
}

// RenderRuntime renders a runtime name with styling
func RenderRuntime(name string) string {
	return StyleRuntime.Render(name)
}

// RenderVersion renders a version string with styling
func RenderVersion(version string) string {
	return StyleVersion.Render(version)
}

// RenderActiveVersion renders an active version string with styling
func RenderActiveVersion(version string) string {
	return StyleActiveVersion.Render(version)
}

// RenderMuted renders text in a muted/dim style
func RenderMuted(text string) string {
	return StyleMuted.Render(text)
}

// RenderBox renders content in a rounded box
func RenderBox(content string) string {
	return StyleBox.Render(content)
}

// RenderInfoBox renders content in an info-styled box
func RenderInfoBox(content string) string {
	return StyleInfoBox.Render(content)
}
