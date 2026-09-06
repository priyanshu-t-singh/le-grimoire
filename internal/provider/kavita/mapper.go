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
