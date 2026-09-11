package tui

import tea "charm.land/bubbletea/v2"

func (m *Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.view == viewReader {
		return m.handleReaderKey(key)
	}
	return m.handleListKey(key)
}

func (m *Model) handleListKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		p := m.ds.Top()
		if p.State["cursor"] > 0 {
			p.State["cursor"]--
			m.cursor = p.State["cursor"]
		}
	case "down", "j":
		p := m.ds.Top()
		if p.State["cursor"] < len(m.items)-1 {
			p.State["cursor"]++
			m.cursor = p.State["cursor"]
		}
	case "enter", "l", " ":
		return m.selectItem()
	case "h", "backspace":
		return m.goBack()
	}
	return m, nil
}

func (m *Model) handleReaderKey(key string) (tea.Model, tea.Cmd) {
	bodyHeight := m.height - 2
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	totalLines := len(m.content)

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.line > 0 {
			m.line--
		}
	case "down", "j", "enter", "l", " ":
		if m.line+bodyHeight < totalLines {
			m.line++
		} else {
			return m.nextBookPage()
		}
	case "pgup":
		return m.prevChapter()
	case "pgdn":
		return m.nextChapter()
	case "h", "backspace":
		return m.goBack()
	}
	return m, nil
}
