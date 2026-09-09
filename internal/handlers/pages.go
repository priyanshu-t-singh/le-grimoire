package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"le-grimoire/internal/middleware"
	"le-grimoire/internal/render"
	"le-grimoire/internal/state"
)

func (h *Handler) CurrentPageHandler(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := middleware.GetDeviceID(r.Context())
	if !ok {
		h.RespondWithStatusError(w, http.StatusUnauthorized, fmt.Errorf("device ID not found in headers"))
		return
	}

	ds, err := h.Devices.GetDeviceState(deviceID)
	if err != nil || ds == nil {
		ds = state.NewDeviceState(deviceID)
	}

	frame, err := h.renderCurrentState(r.Context(), ds)
	if err != nil {
		frame = render.RenderFallbackErrorBitmap(err.Error())
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(frame)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(frame)
}

func (h *Handler) renderCurrentState(ctx context.Context, ds *state.DeviceState) ([]byte, error) {
	top := ds.Top()
	cursor := top.State["cursor"]

	switch top.Type {
	case state.PageLibrary:
		libraries, err := h.Books.GetLibraries(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetch libraries: %w", err)
		}
		html, err := render.BuildLibraryHTML(libraries, cursor)
		if err != nil {
			return nil, err
		}
		return h.Renderer.RenderListPage(ctx, html)

	case state.PageSeries:
		seriesList, err := h.Books.GetBooks(ctx, top.Params["library_id"])
		if err != nil {
			return nil, fmt.Errorf("fetch series: %w", err)
		}
		html, err := render.BuildSeriesHTML(seriesList, cursor)
		if err != nil {
			return nil, err
		}
		return h.Renderer.RenderListPage(ctx, html)

	case state.PageBookList:
		chapters, err := h.Books.GetChapters(ctx, top.Params["series_id"])
		if err != nil {
			return nil, fmt.Errorf("fetch chapters: %w", err)
		}
		html, err := render.BuildBookListHTML(chapters, cursor)
		if err != nil {
			return nil, err
		}
		return h.Renderer.RenderListPage(ctx, html)

	case state.PageReader:
		return h.renderReaderPage(ctx, top)

	default:
		html := render.BuildPlaceholderHTML(*top)
		return h.Renderer.RenderListPage(ctx, html)
	}
}

func (h *Handler) renderReaderPage(ctx context.Context, p *state.Page) ([]byte, error) {
	chapterID := p.Params["chapter_id"]
	format := p.Params["format"]
	bookPageIndex := p.State["book_page"]
	subPageIndex := p.State["sub_page"]

	h.Log.Debug(fmt.Sprintf(
		"Rendering reader page: chapter=%s, format=%s, book_page=%s, sub_page=%s",
		chapterID, format, bookPageIndex, subPageIndex,
	))

	// Format 0: Manga / Comic
	if format != "epub" {
		imgBytes, err := h.Books.PageContent(ctx, chapterID, subPageIndex)
		if err != nil {
			return nil, fmt.Errorf("fetch manga page %d: %w", subPageIndex, err)
		}
		return render.ProcessMangaImage(imgBytes.Data)
	}

	// Format 1+: Book / EPUB
	var frames [][]byte
	if cachedFrames, exists := h.Cache.GetAllFrames(chapterID, bookPageIndex); exists {
		frames = cachedFrames
	} else {
		rawHTML, err := h.Books.PageContent(ctx, chapterID, bookPageIndex)
		if err != nil {
			return nil, fmt.Errorf("fetch book content (chapter %d, page %d): %w", chapterID, bookPageIndex, err)
		}

		renderedHTML, err := render.BuildReaderHTML(string(rawHTML.Data))
		if err != nil {
			return nil, fmt.Errorf("build reader html: %w", err)
		}

		frames, err = h.Renderer.RenderBookFrames(ctx, renderedHTML, 24)
		if err != nil {
			return nil, fmt.Errorf("render book frames: %w", err)
		}

		h.Cache.Set(chapterID, bookPageIndex, frames)
	}

	if len(frames) == 0 {
		return nil, fmt.Errorf("no rendered frames produced for chapter %d (book_page %d)", chapterID, bookPageIndex)
	}

	// Clamp sub_page within current fragment frames
	if subPageIndex >= len(frames) {
		subPageIndex = len(frames) - 1
		p.State["sub_page"] = subPageIndex
	}
	if subPageIndex < 0 {
		subPageIndex = 0
		p.State["sub_page"] = 0
	}

	return frames[subPageIndex], nil
}
