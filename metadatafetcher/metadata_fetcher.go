package metadatafetcher

import (
	"net/url"

	core "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/metadatafetcher/youtube"

	"github.com/marianogappa/predictions/metadatafetcher/twitter"
)

// SpecificFetcher is the interface for a concrete fetcher, e.g. youtube.Fetcher, twitter.Fetcher.
type SpecificFetcher interface {
	IsCorrectFetcher(url *url.URL) bool
	Fetch(url *url.URL) (core.PostMetadata, error)
}

// MetadataFetcher is the main struct for the social media account metadata fetcher.
type MetadataFetcher struct {
	Fetchers []SpecificFetcher
}

// NewMetadataFetcher constructs a MetadataFetcher.
func NewMetadataFetcher() *MetadataFetcher {
	return &MetadataFetcher{[]SpecificFetcher{twitter.NewMetadataFetcher(""), youtube.NewMetadataFetcher("")}}
}

// Fetch takes a URL string, loops through the available fetchers looking for the correct one, and then requests
// metadata for that URL using the specific fetcher.
func (f MetadataFetcher) Fetch(rawURL string) (core.PostMetadata, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return core.PostMetadata{}, err
	}

	for _, fetcher := range f.Fetchers {
		if !fetcher.IsCorrectFetcher(url) {
			continue
		}
		m, err := fetcher.Fetch(url)
		if err != nil {
			return m, err
		}
		return m, nil
	}
	return core.PostMetadata{}, core.ErrNoMetadataFound
}
