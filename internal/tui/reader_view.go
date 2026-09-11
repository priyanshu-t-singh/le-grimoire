package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) renderReader() string {
	width := m.width
	height := m.height

	// Header
	chapterTitle := m.chapterTitle
	if chapterTitle == "" {
		chapterTitle = "Reader"
	}

	totalLines := len(m.content)
	bodyHeight := height - 2 // header + footer

	var pageInfo string
	if totalLines > 0 && bodyHeight > 0 {
		totalPageCount := max(1, (totalLines+bodyHeight-1)/bodyHeight)
		currentPageNum := m.line/max(1, bodyHeight) + 1
		pageInfo = fmt.Sprintf("%d/%d", currentPageNum, totalPageCount)
	}

	headerLeft := HeaderStyle.Render(chapterTitle)
	headerRight := HeaderStyle.Render(pageInfo)
	gap := width - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if gap < 0 {
		gap = 0
	}
	header := headerLeft + GapFillStyle.Render(strings.Repeat(" ", gap)) + headerRight

	// Footer
	var footerText string
	switch {
	case m.err != nil:
		footerText = ErrorStyle.Render("ERROR: " + m.err.Error())
	case m.loading:
		footerText = LoadingStyle.Render("  Loading…")
	default:
		footerText = FooterStyle.Render("↑/k Up  ↓/j Down  h/Bsp Back  PgUp Prev-ch  PgDn Next-ch  q Quit")
	}
	footer := lipgloss.PlaceHorizontal(width, lipgloss.Left, footerText)

	// Content body
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	var bodyLines []string
	if m.isImagePage {
		// NOTE: image / comic pages are not renderable in the terminal.
		// TODO: render using ANSI half-block / sixel / braille art.
		placeholder := PlaceholderStyle.Render(
			"[Image page — not renderable in terminal]  ↓/j to advance, ↑/k to go back.",
		)
		bodyLines = append(bodyLines, placeholder)
	} else {
		end := m.line + bodyHeight
		if end > totalLines {
			end = totalLines
		}
		slice := m.content[m.line:end]
		for _, l := range slice {
			// Truncate long lines to terminal width.
			runes := []rune(l)
			if len(runes) > width {
				runes = append(runes[:width-1], '…')
				l = string(runes)
			}
			bodyLines = append(bodyLines, ReaderStyle.Render(l))
		}
	}

	// Pad remaining body height with blank lines.
	for len(bodyLines) < bodyHeight {
		bodyLines = append(bodyLines, "")
	}
	body := strings.Join(bodyLines[:bodyHeight], "\n")

	return strings.Join([]string{header, body, footer}, "\n")
}
