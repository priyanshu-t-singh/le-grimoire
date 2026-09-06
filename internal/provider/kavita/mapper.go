package kavita

import (
	"le-grimoire/internal/library"
	"sort"
	"strconv"
)

func mapKavitaLibraryToLibrary(raw kavitaLibrary) library.Library {
	return library.Library{
		ID:   strconv.Itoa(raw.ID),
		Name: raw.Name,
	}
}

func mapLibraryListToLibraries(raw []kavitaLibrary) []library.Library {
	libraries := make([]library.Library, len(raw))
	for i, r := range raw {
		libraries[i] = mapKavitaLibraryToLibrary(r)
	}
	return libraries
}

func mapSeriesToBook(s kavitaSeries) library.Book {
	return library.Book{
		ID:        strconv.Itoa(s.ID),
		LibraryID: strconv.Itoa(s.LibraryID),
		Title:     s.Name,
		Author:    "",                 // Kavita's series list doesn't include author; comes from SeriesDetail
		CoverID:   strconv.Itoa(s.ID), // resolved later via a Cover() call using the same series ID
		Format:    mangaFormat(s.Format).toDomain(),
	}
}

func mapSeriesListToBooks(list []kavitaSeries) []library.Book {
	books := make([]library.Book, 0, len(list))
	for _, s := range list {
		books = append(books, mapSeriesToBook(s))
	}
	return books
}

func mapChapter(seriesID string, c kavitaChapter) library.Chapter {
	num, err := strconv.ParseFloat(c.Number, 64)
	if err != nil {
		num = 0 // fallback for non-numeric chapter labels ("Special", etc.)
	}
	return library.Chapter{
		ID:     strconv.Itoa(c.ID),
		BookID: seriesID,
		Number: num,
		Title:  c.Title,
	}
}

// flattenVolumes walks volumes in order and collects their chapters,
// producing a single reading-order list
func flattenVolumes(seriesID string, volumes []kavitaVolume) []library.Chapter {
	var chapters []library.Chapter
	for _, v := range volumes {
		for _, c := range v.Chapters {
			chapters = append(chapters, mapChapter(seriesID, c))
		}
	}
	// Kavita doesn't guarantee cross-volume chapter ordering in the raw
	// response, so sort explicitly by Number to be safe.
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Number < chapters[j].Number
	})
	return chapters
}
