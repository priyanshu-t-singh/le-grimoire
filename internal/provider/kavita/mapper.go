package kavita

import (
	"le-grimoire/internal/library"
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
