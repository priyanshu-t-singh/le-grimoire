package state

import (
	"context"
	"fmt"
	"le-grimoire/internal/library"
	"log/slog"
	"strconv"
)

type Machine struct {
	library library.BookProvider
	logger  *slog.Logger
}

func NewMachine(library library.BookProvider, logger *slog.Logger) *Machine {
	return &Machine{
		library: library,
		logger:  logger,
	}
}

func (m *Machine) ApplyButton(ctx context.Context, ds *DeviceState, buttonID, pressType string) (string, error) {
	p := ds.Top()

	switch p.Type {
	case PageLibrary, PageSeries, PageBookList:
		return m.applyListButton(ctx, ds, p, buttonID, pressType)
	case PageReader:
		return m.applyReaderButton(ctx, ds, p, buttonID, pressType)
	default:
		return "unknown page type, no-op", nil
	}
}

func (m *Machine) applyListButton(ctx context.Context, ds *DeviceState, p *Page, buttonID, pressType string) (string, error) {
	if pressType == "long" || pressType == "long_press" {
		switch buttonID {
		case "E":
			return "force refresh requested", nil
		default:
			return buttonID + " long-press: unassigned, no-op", nil
		}
	}

	switch buttonID {
	case "A": // Up
		if p.State["cursor"] > 0 {
			p.State["cursor"]--
			ds.UpdatedAt = ds.UpdatedAt.UTC()
		}
		return fmt.Sprintf("cursor moved up to index %d", p.State["cursor"]), nil

	case "B": // Down
		count, err := m.getItemCount(ctx, p)
		if err != nil {
			return "", err
		}
		if p.State["cursor"] < count-1 {
			p.State["cursor"]++
			ds.UpdatedAt = ds.UpdatedAt.UTC()
		}
		return fmt.Sprintf("cursor moved down to index %d (max: %d)", p.State["cursor"], count-1), nil

	case "C": // Select
		return m.selectFromList(ctx, ds, p)

	case "D": // Back
		if len(ds.Stack) <= 1 {
			return "back at root — no-op", nil
		}
		ds.Pop()
		return "back — popped page", nil

	default:
		return buttonID + " short-press: unassigned, no-op", nil
	}
}

func (m *Machine) selectFromList(ctx context.Context, ds *DeviceState, p *Page) (string, error) {
	switch p.Type {
	case PageLibrary:
		libraries, err := m.library.GetLibraries(ctx)
		if err != nil {
			return "", fmt.Errorf("fetch libraries: %w", err)
		}
		if len(libraries) == 0 || p.State["cursor"] >= len(libraries) {
			return "no library at cursor", nil
		}

		selected := libraries[p.State["cursor"]]
		ds.Push(Page{
			Type:   PageSeries,
			Params: map[string]string{"library_id": selected.ID},
			State:  map[string]int{"cursor": 0, "scroll": 0},
		})
		return fmt.Sprintf("selected library %s (id: %s) -> pushed Series", selected.Name, selected.ID), nil

	case PageSeries:
		libraryID, _ := strconv.Atoi(p.Params["library_id"])
		seriesList, err := m.library.GetBooks(ctx, strconv.Itoa(libraryID))
		if err != nil {
			return "", fmt.Errorf("fetch series: %w", err)
		}
		if len(seriesList) == 0 || p.State["cursor"] >= len(seriesList) {
			return "no series at cursor", nil
		}

		selected := seriesList[p.State["cursor"]]
		ds.Push(Page{
			Type: PageBookList,
			Params: map[string]string{
				"series_id": selected.ID,
				"format":    selected.Format,
			},
			State: map[string]int{"cursor": 0, "scroll": 0},
		})
		return fmt.Sprintf("selected series %s (id: %s) -> pushed BookList", selected.Title, selected.ID), nil

	case PageBookList:
		seriesID, _ := strconv.Atoi(p.Params["series_id"])
		chapters, err := m.library.GetChapters(ctx, strconv.Itoa(seriesID))
		if err != nil {
			return "", fmt.Errorf("fetch chapters: %w", err)
		}
		if len(chapters) == 0 || p.State["cursor"] >= len(chapters) {
			return "no chapter at cursor", nil
		}

		selected := chapters[p.State["cursor"]]
		ds.Push(Page{
			Type: PageReader,
			Params: map[string]string{
				"series_id":  p.Params["series_id"],
				"volume_id":  selected.BookID,
				"chapter_id": selected.ID,
				"format":     p.Params["format"],
			},
			// book_page: which Kavita-level fragment is loaded (0-indexed)
			// sub_page:  which rendered 24-line frame within that fragment
			State: map[string]int{"book_page": 0, "sub_page": 0},
		})
		return fmt.Sprintf("selected chapter %s (id: %s) -> pushed Reader", selected.Title, selected.ID), nil
	}

	return "select: nothing to do", nil
}

func (m *Machine) applyReaderButton(ctx context.Context, ds *DeviceState, p *Page, buttonID, pressType string) (string, error) {
	if pressType == "long" || pressType == "long_press" {
		switch buttonID {
		case "C": // Long C: previous chapter
			return m.navigateChapter(ctx, ds, p, -1)
		case "F": // Long F: next chapter
			return m.navigateChapter(ctx, ds, p, 1)
		case "E": // Long E: force refresh
			return "force refresh requested", nil
		default:
			return buttonID + " long-press: unassigned, no-op", nil
		}
	}

	switch buttonID {
	case "A": // Scroll up within current page fragment
		if p.State["sub_page"] > 0 {
			p.State["sub_page"]--
			ds.UpdatedAt = ds.UpdatedAt.UTC()
			return fmt.Sprintf("scroll up -> sub_page %d", p.State["sub_page"]), nil
		}
		return "already at top of page", nil

	case "B": // Scroll down within current page fragment
		p.State["sub_page"]++
		ds.UpdatedAt = ds.UpdatedAt.UTC()
		return fmt.Sprintf("scroll down -> sub_page %d", p.State["sub_page"]), nil

	case "C": // Previous page index
		return m.navigateBookPage(ctx, ds, p, -1)

	case "F": // Next page index
		return m.navigateBookPage(ctx, ds, p, 1)

	case "D": // Back to BookList
		ds.Pop()
		return "back — popped to BookList", nil

	default:
		return buttonID + " short-press: unassigned, no-op", nil
	}
}

// advances/reverses the Kavita-level book-page fragment within
// the current chapter. Resets sub_page to 0 since scroll position is meaningless
// across a fragment change.
func (m *Machine) navigateBookPage(ctx context.Context, ds *DeviceState, p *Page, direction int) (string, error) {
	chapterID := p.Params["chapter_id"]
	meta, err := m.library.GetChapterInfo(ctx, chapterID)
	if err != nil {
		return "", fmt.Errorf("navigate book page: %w", err)
	}
	totalPages := meta.TotalPages
	if totalPages <= 0 {
		totalPages = 1
	}
	current := p.State["book_page"]
	target := current + direction
	if target < 0 {
		return "already at first page of chapter — no-op", nil
	}
	if target >= totalPages {
		return "already at last page of chapter — no-op", nil
	}
	p.State["book_page"] = target
	p.State["sub_page"] = 0
	ds.UpdatedAt = ds.UpdatedAt.UTC()
	return fmt.Sprintf("page index -> %d/%d", target, totalPages-1), nil
}

// advances or reverses the active chapter inside the Reader view.
func (m *Machine) navigateChapter(ctx context.Context, ds *DeviceState, p *Page, direction int) (string, error) {
	seriesID := p.Params["series_id"]
	currentChapterID := p.Params["chapter_id"]

	chapters, err := m.library.GetChapters(ctx, seriesID)
	if err != nil {
		return "", fmt.Errorf("navigate chapter: %w", err)
	}

	currentIdx := -1
	for i, ch := range chapters {
		if ch.ID == currentChapterID {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 {
		return "current chapter not found in series list", nil
	}

	targetIdx := currentIdx + direction
	if targetIdx < 0 {
		return "already at the first chapter — no-op", nil
	}
	if targetIdx >= len(chapters) {
		return "already at the last chapter — no-op", nil
	}

	targetChapter := chapters[targetIdx]
	p.Params["chapter_id"] = targetChapter.ID
	p.Params["volume_id"] = targetChapter.BookID
	p.State["book_page"] = 0
	p.State["sub_page"] = 0
	ds.UpdatedAt = ds.UpdatedAt.UTC()

	// Keep BookList cursor under Reader synchronized
	if len(ds.Stack) >= 2 {
		bookListIdx := len(ds.Stack) - 2
		if ds.Stack[bookListIdx].Type == PageBookList {
			ds.Stack[bookListIdx].State["cursor"] = targetIdx
		}
	}

	return fmt.Sprintf("switched chapter to %s (id: %s)", targetChapter.Title, targetChapter.ID), nil
}

// calculates dynamic boundary counts for cursor navigation.
func (m *Machine) getItemCount(ctx context.Context, p *Page) (int, error) {
	switch p.Type {
	case PageLibrary:
		libs, err := m.library.GetLibraries(ctx)
		if err != nil {
			return 0, err
		}
		return len(libs), nil

	case PageSeries:
		libID := p.Params["library_id"]
		seriesList, err := m.library.GetBooks(ctx, libID)
		if err != nil {
			return 0, err
		}
		return len(seriesList), nil

	case PageBookList:
		seriesID := p.Params["series_id"]
		chapters, err := m.library.GetChapters(ctx, seriesID)
		if err != nil {
			return 0, err
		}
		return len(chapters), nil

	default:
		return 0, nil
	}
}
