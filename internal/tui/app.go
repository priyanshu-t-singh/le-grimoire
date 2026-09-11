package tui

import (
	"context"
	"le-grimoire/internal/library"
	"le-grimoire/internal/state"

	tea "charm.land/bubbletea/v2"
)

const (
	pageLibrary  = state.PageLibrary
	pageSeries   = state.PageSeries
	pageBookList = state.PageBookList
	pageReader   = state.PageReader
)

type viewKind int

const (
	viewList viewKind = iota
	viewReader
)

type itemsLoadedMsg struct {
	items []string
	page  *state.Page
}

type contentLoadedMsg struct {
	lines       []string
	isImagePage bool
	totalPages  int
	page        *state.Page
	update      *state.Page
	title       string
	listCursor  int
}

type errMsg struct{ err error }

type Model struct {
	ctx   context.Context
	books library.BookProvider
	ds    *state.DeviceState

	view viewKind

	items  []string
	cursor int

	content      []string
	line         int
	isImagePage  bool
	chapterTitle string
	totalPages   int

	width  int
	height int

	loading bool
	err     error
}

func NewModel(ctx context.Context, books library.BookProvider) *Model {
	return &Model{
		ctx:   ctx,
		books: books,
		ds:    state.NewDeviceState("terminal"),
		view:  viewList,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.fetchCurrentList()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case itemsLoadedMsg:
		m.loading = false
		m.err = nil
		if msg.page != nil {
			m.ds.Push(*msg.page)
			m.cursor = msg.page.State["cursor"]
		}
		m.items = msg.items
		return m, nil

	case contentLoadedMsg:
		m.loading = false
		m.err = nil
		if msg.page != nil {
			m.ds.Push(*msg.page)
			m.view = viewReader
		} else if msg.update != nil {
			p := m.ds.Top()
			for k, v := range msg.update.Params {
				if p.Params == nil {
					p.Params = make(map[string]string)
				}
				p.Params[k] = v
			}
			for k, v := range msg.update.State {
				if p.State == nil {
					p.State = make(map[string]int)
				}
				p.State[k] = v
			}
		}
		if msg.title != "" {
			m.chapterTitle = msg.title
		}
		if msg.listCursor >= 0 && len(m.ds.Stack) >= 2 {
			bl := &m.ds.Stack[len(m.ds.Stack)-2]
			if bl.Type == pageBookList {
				bl.State["cursor"] = msg.listCursor
			}
		}
		m.content = msg.lines
		m.isImagePage = msg.isImagePage
		m.totalPages = msg.totalPages
		m.line = 0
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) View() tea.View {
	var s string
	if m.width == 0 {
		s = "Initialising…"
	} else if m.view == viewReader {
		s = m.renderReader()
	} else {
		s = m.renderList()
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}
