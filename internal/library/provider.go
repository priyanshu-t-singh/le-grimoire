package library

import "context"

// TODO: Add pagination support to all methods that return lists of items,
// so we can support large libraries without running into memory issues.
type BookProvider interface {
	Connect(ctx context.Context) error
	GetLibraries(ctx context.Context) ([]Library, error)
	GetBooks(ctx context.Context, libraryID string) ([]Book, error)
	// GetBookDetail(ctx context.Context, bookID string) (*Book, error)

	// Chapters: reading units in order. A single-file book (epub/pdf/cbz)
	// just returns one synthetic chapter.
	GetChapters(ctx context.Context, bookID string) ([]Chapter, error)
	// GetChapterInfo(ctx context.Context, chapterID string) (*ChapterInfo, error)

	// Content returns raw page content — could be an image, HTML, or plain text.
	// The Renderer decides what to do with it based on ContentType.
	PageContent(ctx context.Context, chapterID string, page int) (*PageContent, error)
}
