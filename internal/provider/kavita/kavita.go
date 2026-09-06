package kavita

import (
	"context"
	"fmt"
	"le-grimoire/internal/library"
	"net/http"
	"strconv"
	"time"
)

type Provider struct {
	library.BookProvider
	baseURL    string
	apiKey     string
	token      string
	httpClient *http.Client
}

func New(baseURL, apiKey string) *Provider {
	return &Provider{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // hard backstop, slightly above our per-call timeout
		},
	}
}

func (p *Provider) Connect(ctx context.Context) error {
	var resp struct {
		RefreshToken string `json:"refreshToken"`
		Token        string `json:"token"`
	}

	err := p.doRequest(ctx, "POST", "/api/account/login", map[string]string{
		"username": "",
		"password": "",
		"apiKey":   p.apiKey,
	}, &resp)

	if err != nil {
		return fmt.Errorf("kavita connect: %w", err)
	}

	p.token = resp.Token
	return nil
}

func (p *Provider) GetLibraries(ctx context.Context) ([]library.Library, error) {
	var raw []kavitaLibrary

	err := p.doRequest(ctx, "GET", "/api/Library/libraries", nil, &raw)
	if err != nil {
		return nil, fmt.Errorf("get libraries: %w", err)
	}

	return mapLibraryListToLibraries(raw), nil
}

func (p *Provider) GetBooks(ctx context.Context, libraryID string) ([]library.Book, error) {
	libID, err := strconv.Atoi(libraryID)
	if err != nil {
		return nil, fmt.Errorf("invalid library id %q: %w", libraryID, err)
	}

	// TODO: Kavita API v2 supports pagination, but we currently don't support it.
	// For now, we just fetch the first page with a large page size.
	path := fmt.Sprintf("/api/series/v2?libraryId=%d&pageNumber=1&pageSize=100", libID)

	var raw []kavitaSeries
	if err := p.doRequest(ctx, "POST", path, defaultSeriesFilter(), &raw); err != nil {
		return nil, fmt.Errorf("get books for library %s: %w", libraryID, err)
	}

	return mapSeriesListToBooks(raw), nil
}

func (p *Provider) GetChapters(ctx context.Context, bookID string) ([]library.Chapter, error) {
	seriesID, err := strconv.Atoi(bookID)
	if err != nil {
		return nil, fmt.Errorf("invalid book id %q: %w", bookID, err)
	}

	var volumes []kavitaVolume
	path := fmt.Sprintf("/api/Series/volumes?seriesId=%d", seriesID)
	if err := p.doRequest(ctx, "GET", path, nil, &volumes); err != nil {
		return nil, fmt.Errorf("get chapters for book %s: %w", bookID, err)
	}

	chapters := flattenVolumes(bookID, volumes)
	if len(chapters) == 0 {
		return nil, fmt.Errorf("%w: no chapters found for book %s", library.ErrNotFound, bookID)
	}
	return chapters, nil
}
