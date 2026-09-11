package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) renderList() string {
	width := m.width
	height := m.height

	// Header
	title := pageTitle(m)
	totalItems := len(m.items)
	pageSize := listPageSize(height)
	totalPages := max(1, (totalItems+pageSize-1)/pageSize)
	currentPage := (m.cursor/pageSize + 1)
	pageIndicator := fmt.Sprintf("%d/%d", currentPage, totalPages)

	headerLeft := HeaderStyle.Render(title)
	headerRight := HeaderStyle.Render(pageIndicator)

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
		footerText = FooterStyle.Render("↑/k Up  ↓/j Down  Enter/l Select  h/Bsp Back  q Quit")
	}
	footer := lipgloss.PlaceHorizontal(width, lipgloss.Left, footerText)

	// List body
	bodyHeight := height - 2 // header + footer
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	startIdx := (m.cursor / pageSize) * pageSize
	endIdx := startIdx + pageSize
	if endIdx > totalItems {
		endIdx = totalItems
	}

	var rows []string
	for i := startIdx; i < endIdx; i++ {
		label := m.items[i]
		maxLabel := width - 4
		if maxLabel < 1 {
			maxLabel = 1
		}
		runes := []rune(label)
		if len(runes) > maxLabel {
			runes = append(runes[:maxLabel-1], '…')
			label = string(runes)
		}

		var row string
		if i == m.cursor {
			row = SelectedStyle.Render(fmt.Sprintf("▶ %s", label))
		} else {
			row = NormalStyle.Render(fmt.Sprintf("  %s", label))
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		rows = append(rows, NormalStyle.Render("  (empty)"))
	}

	for len(rows) < bodyHeight {
		rows = append(rows, "")
	}
	body := strings.Join(rows[:bodyHeight], "\n")

	return strings.Join([]string{header, body, footer}, "\n")
}

func pageTitle(m *Model) string {
	if m.ds == nil || len(m.ds.Stack) == 0 {
		return "Le Grimoire"
	}
	p := m.ds.Top()
	switch p.Type {
	case pageLibrary:
		return "Libraries"
	case pageSeries:
		return "Series"
	case pageBookList:
		return "Chapters"
	default:
		return string(p.Type)
	}
}

func listPageSize(height int) int {
	size := height - 2 // reserve header + footer lines
	if size < 1 {
		size = 1
	}
	return size
}
