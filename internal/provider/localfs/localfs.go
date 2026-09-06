package localfs

import "le-grimoire/internal/library"

type LocalFSProvider struct {
	library.BookProvider
	rootDir string
}

func New(rootDir string) *LocalFSProvider {
	return &LocalFSProvider{
		rootDir: rootDir,
	}
}
