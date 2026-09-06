package library

import "context"

type BookProvider interface {
	Connect(ctx context.Context) error
	Libraries(ctx context.Context) ([]Library, error)
	Books(ctx context.Context, libraryID string) ([]Book, error)
	BookDetail(ctx context.Context, bookID string) (*Book, error)

	// Chapters: reading units in order. A single-file book (epub/pdf/cbz)
	// just returns one synthetic chapter.
	Chapters(ctx context.Context, bookID string) ([]Chapter, error)
	ChapterInfo(ctx context.Context, chapterID string) (*ChapterInfo, error)

	// Content returns raw page content — could be an image, HTML, or plain text.
	// The Renderer decides what to do with it based on ContentType.
	PageContent(ctx context.Context, chapterID string, page int) (*PageContent, error)
}
