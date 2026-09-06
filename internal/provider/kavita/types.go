package kavita

import "strconv"

type kavitaLibrary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type kavitaSeries struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"originalName"`
	LibraryID    int    `json:"libraryId"`
	Format       int    `json:"format"` // Kavita's MangaFormat enum
	Pages        int    `json:"pages"`
}

// mangaFormat mirrors Kavita's MangaFormat enum values.
// Source: Kavita API — 0=Image, 1=Archive, 2=Unknown, 3=Epub, 4=Pdf
type mangaFormat int

const (
	formatImage   mangaFormat = 0
	formatArchive mangaFormat = 1
	formatUnknown mangaFormat = 2
	formatEpub    mangaFormat = 3
	formatPdf     mangaFormat = 4
)

func (f mangaFormat) toDomain() string {
	switch f {
	case formatEpub:
		return "epub"
	case formatPdf:
		return "pdf"
	case formatArchive:
		return "cbz" // comic archive (cbz/cbr/zip)
	case formatImage:
		return "comic"
	default:
		return "unknown"
	}
}

type filterStatement struct {
	Field      int    `json:"field"`
	Value      string `json:"value"`
	Comparison int    `json:"comparison"`
}

type seriesFilterV2Request struct {
	Statements  []filterStatement `json:"statements"`
	Combination int               `json:"combination"`
	LimitTo     int               `json:"limitTo"`
	SortOptions sortOptions       `json:"sortOptions"`
}

type sortOptions struct {
	SortField   int  `json:"sortField"`
	IsAscending bool `json:"isAscending"`
}

func defaultSeriesFilter() seriesFilterV2Request {
	return seriesFilterV2Request{
		Statements:  []filterStatement{},
		Combination: 0,
		SortOptions: sortOptions{SortField: 1, IsAscending: true},
	}
}

func (s seriesFilterV2Request) addLibraryFilter(libraryID int) seriesFilterV2Request {
	s.Statements = append(s.Statements, filterStatement{
		Field:      19, // LibraryId
		Value:      strconv.Itoa(libraryID),
		Comparison: 0, // Equal
	})
	return s
}

type kavitaVolume struct {
	ID       int             `json:"id"`
	Number   int             `json:"number"`
	Chapters []kavitaChapter `json:"chapters"`
}

type kavitaChapter struct {
	ID       int    `json:"id"`
	Number   string `json:"number"` // Kavita stores this as string, e.g. "1", "1.5"
	Title    string `json:"titleName"`
	Pages    int    `json:"pages"`
	VolumeID int    `json:"volumeId"`
}
