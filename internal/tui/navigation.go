package tui

import (
	"fmt"
	"le-grimoire/internal/state"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) selectItem() (tea.Model, tea.Cmd) {
	p := m.ds.Top()
	cursor := p.State["cursor"]
	switch p.Type {
	case pageLibrary:
		return m.selectLibrary(cursor)
	case pageSeries:
		return m.selectSeries(cursor)
	case pageBookList:
		return m.selectChapter(cursor)
	}
	return m, nil
}

func (m *Model) selectLibrary(cursor int) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, func() tea.Msg {
		libs, err := m.books.GetLibraries(m.ctx)
		if err != nil {
			return errMsg{err}
		}
		if cursor >= len(libs) {
			return errMsg{fmt.Errorf("no library at index %d", cursor)}
		}
		selected := libs[cursor]
		books, err := m.books.GetBooks(m.ctx, selected.ID)
		if err != nil {
			return errMsg{err}
		}
		items := make([]string, len(books))
		for i, b := range books {
			items[i] = b.Title
		}
		page := state.Page{
			Type:   pageSeries,
			Params: map[string]string{"library_id": selected.ID},
			State:  map[string]int{"cursor": 0, "scroll": 0},
		}
		return itemsLoadedMsg{items: items, page: &page}
	}
}

func (m *Model) selectSeries(cursor int) (tea.Model, tea.Cmd) {
	m.loading = true
	p := m.ds.Top()
	libID := p.Params["library_id"]
	return m, func() tea.Msg {
		books, err := m.books.GetBooks(m.ctx, libID)
		if err != nil {
			return errMsg{err}
		}
		if cursor >= len(books) {
			return errMsg{fmt.Errorf("no series at index %d", cursor)}
		}
		selected := books[cursor]
		chapters, err := m.books.GetChapters(m.ctx, selected.ID)
		if err != nil {
			return errMsg{err}
		}
		items := make([]string, len(chapters))
		for i, c := range chapters {
			items[i] = chapterLabel(c)
		}
		page := state.Page{
			Type: pageBookList,
			Params: map[string]string{
				"series_id": selected.ID,
				"format":    selected.Format,
			},
			State: map[string]int{"cursor": 0, "scroll": 0},
		}
		return itemsLoadedMsg{items: items, page: &page}
	}
}

func (m *Model) selectChapter(cursor int) (tea.Model, tea.Cmd) {
	m.loading = true
	p := m.ds.Top()
	seriesID := p.Params["series_id"]
	format := p.Params["format"]
	return m, func() tea.Msg {
		chapters, err := m.books.GetChapters(m.ctx, seriesID)
		if err != nil {
			return errMsg{err}
		}
		if cursor >= len(chapters) {
			return errMsg{fmt.Errorf("no chapter at index %d", cursor)}
		}
		selected := chapters[cursor]
		info, err := m.books.GetChapterInfo(m.ctx, selected.ID)
		if err != nil {
			return errMsg{err}
		}
		content, err := m.books.PageContent(m.ctx, selected.ID, 0)
		if err != nil {
			return errMsg{err}
		}
		page := state.Page{
			Type: pageReader,
			Params: map[string]string{
				"series_id":  seriesID,
				"volume_id":  selected.BookID,
				"chapter_id": selected.ID,
				"format":     format,
			},
			State: map[string]int{"book_page": 0, "sub_page": 0},
		}
		msg := buildContentMsg(content, info.TotalPages, m.width)
		if cm, ok := msg.(contentLoadedMsg); ok {
			cm.page = &page
			cm.title = chapterLabel(selected)
			return cm
		}
		return msg
	}
}

func (m *Model) goBack() (tea.Model, tea.Cmd) {
	if m.view == viewReader {
		m.view = viewList
		m.ds.Pop()
		m.cursor = m.ds.Top().State["cursor"]
		m.loading = true
		return m, m.fetchCurrentList()
	}
	if len(m.ds.Stack) <= 1 {
		return m, nil
	}
	m.ds.Pop()
	m.cursor = m.ds.Top().State["cursor"]
	m.loading = true
	return m, m.fetchCurrentList()
}

func (m *Model) nextBookPage() (tea.Model, tea.Cmd) {
	p := m.ds.Top()
	chapterID := p.Params["chapter_id"]
	current := p.State["book_page"]
	next := current + 1
	if next >= m.totalPages {
		return m, nil
	}
	m.loading = true
	p.State["book_page"] = next
	p.State["sub_page"] = 0
	return m, func() tea.Msg {
		content, err := m.books.PageContent(m.ctx, chapterID, next)
		if err != nil {
			return errMsg{err}
		}
		return buildContentMsg(content, m.totalPages, m.width)
	}
}

func (m *Model) prevChapter() (tea.Model, tea.Cmd) { return m.navigateChapter(-1) }
func (m *Model) nextChapter() (tea.Model, tea.Cmd) { return m.navigateChapter(1) }

func (m *Model) navigateChapter(direction int) (tea.Model, tea.Cmd) {
	p := m.ds.Top()
	seriesID := p.Params["series_id"]
	currentChapterID := p.Params["chapter_id"]
	m.loading = true
	return m, func() tea.Msg {
		chapters, err := m.books.GetChapters(m.ctx, seriesID)
		if err != nil {
			return errMsg{err}
		}
		currentIdx := -1
		for i, ch := range chapters {
			if ch.ID == currentChapterID {
				currentIdx = i
				break
			}
		}
		if currentIdx == -1 {
			return errMsg{fmt.Errorf("current chapter not found")}
		}
		targetIdx := currentIdx + direction
		if targetIdx < 0 || targetIdx >= len(chapters) {
			return errMsg{fmt.Errorf("no more chapters in that direction")}
		}
		target := chapters[targetIdx]
		info, err := m.books.GetChapterInfo(m.ctx, target.ID)
		if err != nil {
			return errMsg{err}
		}
		content, err := m.books.PageContent(m.ctx, target.ID, 0)
		if err != nil {
			return errMsg{err}
		}
		updatePage := state.Page{
			Params: map[string]string{
				"chapter_id": target.ID,
				"volume_id":  target.BookID,
			},
			State: map[string]int{"book_page": 0, "sub_page": 0},
		}
		msg := buildContentMsg(content, info.TotalPages, m.width)
		if cm, ok := msg.(contentLoadedMsg); ok {
			cm.update = &updatePage
			cm.title = chapterLabel(target)
			cm.listCursor = targetIdx
			return cm
		}
		return msg
	}
}

func (m *Model) fetchCurrentList() tea.Cmd {
	m.loading = true
	p := m.ds.Top()
	switch p.Type {
	case pageLibrary:
		return func() tea.Msg {
			libs, err := m.books.GetLibraries(m.ctx)
			if err != nil {
				return errMsg{err}
			}
			items := make([]string, len(libs))
			for i, l := range libs {
				items[i] = l.Name
			}
			return itemsLoadedMsg{items: items}
		}
	case pageSeries:
		libID := p.Params["library_id"]
		return func() tea.Msg {
			books, err := m.books.GetBooks(m.ctx, libID)
			if err != nil {
				return errMsg{err}
			}
			items := make([]string, len(books))
			for i, b := range books {
				items[i] = b.Title
			}
			return itemsLoadedMsg{items: items}
		}
	case pageBookList:
		seriesID := p.Params["series_id"]
		return func() tea.Msg {
			chapters, err := m.books.GetChapters(m.ctx, seriesID)
			if err != nil {
				return errMsg{err}
			}
			items := make([]string, len(chapters))
			for i, c := range chapters {
				items[i] = chapterLabel(c)
			}
			return itemsLoadedMsg{items: items}
		}
	}
	return nil
}
