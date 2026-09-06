package kavita

import (
	"context"
	"fmt"
	"le-grimoire/internal/library"
	"net/http"
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
