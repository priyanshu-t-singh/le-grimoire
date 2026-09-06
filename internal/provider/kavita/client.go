package kavita

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"le-grimoire/internal/library"
	"net/http"
	"time"
)

const defaultRequestTimeout = 8 * time.Second

// doRequest wraps ctx with a request-scoped timeout, so a single slow call
// can't hang forever, but a caller-supplied shorter deadline (e.g. from
// r.Context() upstream) still wins if it's tighter.
func (p *Provider) doRequest(ctx context.Context, method, path string, body any, out any) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		// ctx.Err() tells you WHY it failed: deadline exceeded vs cancelled vs network error
		if ctx.Err() != nil {
			return fmt.Errorf("request to %s timed out or was cancelled: %w", path, ctx.Err())
		}
		return fmt.Errorf("request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return library.ErrUnauthenticated
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("kavita returned status %d for %s", resp.StatusCode, path)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response from %s: %w", path, err)
		}
	}
	return nil
}

func (p *Provider) doRequestWithRetry(ctx context.Context, method, path string, body, out any) error {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		lastErr = p.doRequest(ctx, method, path, body, out)
		if lastErr == nil {
			return nil
		}
		if ctx.Err() != nil {
			return lastErr // caller's context is done, don't retry
		}
		time.Sleep(500 * time.Millisecond)
	}
	return lastErr
}

func (p *Provider) doRequestRawBytes(ctx context.Context, method, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request to %s timed out: %w", path, ctx.Err())
		}
		return nil, fmt.Errorf("request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, library.ErrUnauthenticated
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("kavita returned status %d for %s: %s", resp.StatusCode, path, string(body))
	}

	return io.ReadAll(resp.Body)
}
