package asciiart

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// TerminalWidth tries to get the width of the terminal
func TerminalWidth() int {
	// Default width if we can't determine it
	width := 80

	// Try to get terminal width using stty on Unix-like systems
	if runtime.GOOS != "windows" {
		cmd := exec.Command("stty", "size")
		cmd.Stdin = os.Stdin
		if output, err := cmd.Output(); err == nil {
			// Output is like "ROWS COLUMNS"
			parts := strings.Fields(string(output))
			if len(parts) >= 2 {
				// We only care about the columns (width)
				width = 0
				fmt.Sscanf(parts[1], "%d", &width)
			}
		}
	}

	// Ensure minimum width
	if width < 40 {
		width = 80
	}

	return width
}

// CenterText centers text within a given width
func CenterText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder

	for _, line := range lines {
		padding := (width - len(line)) / 2
		if padding > 0 {
			result.WriteString(strings.Repeat(" ", padding))
		}
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// CreateBanner creates a banner with the given text and style
func CreateBanner(text, style string, width int) string {
	if width <= 0 {
		width = TerminalWidth()
	}

	// Make sure the text isn't too long
	if len(text) > width-4 {
		text = text[:width-7] + "..."
	}

	var banner strings.Builder
	switch style {
	case "box":
		border := strings.Repeat("═", len(text)+2)
		banner.WriteString("╔" + border + "╗\n")
		banner.WriteString("║ " + text + " ║\n")
		banner.WriteString("╚" + border + "╝\n")
	case "stars":
		banner.WriteString(strings.Repeat("*", width) + "\n")
		banner.WriteString(CenterText("* "+text+" *", width))
		banner.WriteString(strings.Repeat("*", width) + "\n")
	default: // simple
		banner.WriteString(strings.Repeat("-", width) + "\n")
		banner.WriteString(CenterText(text, width))
		banner.WriteString(strings.Repeat("-", width) + "\n")
	}

	return banner.String()
}
