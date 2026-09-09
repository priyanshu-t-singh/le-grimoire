package handlers

import (
	"context"
	"log/slog"

	"le-grimoire/internal/library"
	"le-grimoire/internal/state"
)

// BookStore is the subset of library.BookProvider that handlers need.
// library.BookProvider already satisfies this interface — no changes needed there.
type BookStore interface {
	GetLibraries(ctx context.Context) ([]library.Library, error)
	GetBooks(ctx context.Context, libraryID string) ([]library.Book, error)
	GetChapters(ctx context.Context, bookID string) ([]library.Chapter, error)
	PageContent(ctx context.Context, chapterID string, page int) (*library.PageContent, error)
}

// PageRenderer is the subset of render.Renderer that handlers need.
type PageRenderer interface {
	RenderListPage(ctx context.Context, htmlContent string) ([]byte, error)
	RenderBookFrames(ctx context.Context, htmlContent string, lineGap int) ([][]byte, error)
}

// DeviceStore is the subset of device.Repository that handlers need.
type DeviceStore interface {
	GetDeviceState(deviceID string) (*state.DeviceState, error)
	SaveDeviceState(s *state.DeviceState) error
}

// StateMachiner is the subset of state.Machine that handlers need.
type StateMachiner interface {
	ApplyButton(ctx context.Context, ds *state.DeviceState, buttonID, pressType string) (string, error)
}

// FrameStore is the subset of render.FrameCache that handlers need.
type FrameStore interface {
	GetAllFrames(chapterID string, bookPageIndex int) ([][]byte, bool)
	Set(chapterID string, bookPageIndex int, frames [][]byte)
	Invalidate(chapterID string)
}

// Logger is the subset of *slog.Logger that handlers need.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// compile-time proof that *slog.Logger satisfies Logger
var _ Logger = (*slog.Logger)(nil)
