package tui

import (
	"fmt"
	"le-grimoire/internal/library"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func chapterLabel(c library.Chapter) string {
	if c.Title != "" {
		return c.Title
	}
	return fmt.Sprintf("Chapter %.4g", c.Number)
}

func buildContentMsg(content *library.PageContent, totalPages int, width int) tea.Msg {
	if content.Type == library.ContentImage {
		return contentLoadedMsg{lines: nil, isImagePage: true, totalPages: totalPages, listCursor: -1}
	}
	raw := string(content.Data)
	plain := StripHTML(raw)
	wrapWidth := width - 2
	if wrapWidth < 20 {
		wrapWidth = 20
	}
	lines := WordWrap(plain, wrapWidth)
	lines = trimBlankLines(lines)

	if len(lines) == 0 {
		return contentLoadedMsg{lines: nil, isImagePage: true, totalPages: totalPages, listCursor: -1}
	}
	return contentLoadedMsg{lines: lines, isImagePage: false, totalPages: totalPages, listCursor: -1}
}

func trimBlankLines(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
