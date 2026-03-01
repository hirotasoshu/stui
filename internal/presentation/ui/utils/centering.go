package utils

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func CenterContent(content string, width, height int) string {
	if width <= 0 {
		return content
	}

	contentLines := strings.Split(content, "\n")
	var centered strings.Builder

	for _, line := range contentLines {
		visibleWidth := lipgloss.Width(line)
		padding := (width - visibleWidth) / 2
		if padding > 0 {
			centered.WriteString(strings.Repeat(" ", padding))
		}
		centered.WriteString(line + "\n")
	}

	contentHeight := len(contentLines)
	topPadding := (height - contentHeight) / 2
	if topPadding > 0 {
		return strings.Repeat("\n", topPadding) + centered.String()
	}

	return centered.String()
}
