package kavita

import "le-grimoire/internal/library"

type KavitaProvider struct {
	library.BookProvider
	baseURL string
	apiKey  string
}

func New(baseURL, apiKey string) *KavitaProvider {
	return &KavitaProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}
