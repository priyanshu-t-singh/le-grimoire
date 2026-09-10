package library

type Book struct {
	ID        string
	LibraryID string
	Title     string
	Author    string
	CoverID   string // opaque ref the provider resolves to bytes, not a URL
	Format    string // "epub", "pdf", "cbz", "comic"...
}

type ContentType int

const (
	ContentImage ContentType = iota // pre-rendered image (comics, PDFs, scans)
	ContentHTML                     // reflowable text (epub-style)
)

type PageContent struct {
	Type ContentType
	Data []byte // image bytes, or HTML/text as bytes
}

type Library struct {
	ID   string
	Name string
}

type Chapter struct {
	ID     string
	BookID string
	Number float64 // 1, 1.5, etc.
	Title  string
}

type ChapterInfo struct {
	ChapterID  string
	TotalPages int
	Format     string // "epub", "pdf", "cbz", "comic"...
}
