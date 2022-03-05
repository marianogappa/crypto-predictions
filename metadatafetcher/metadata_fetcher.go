package metadatafetcher

import (
	"net/url"

	"github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/metadatafetcher/youtube"

	"github.com/marianogappa/predictions/metadatafetcher/twitter"
)

type SpecificFetcher interface {
	IsCorrectFetcher(url *url.URL) bool
	Fetch(url *url.URL) (types.PostMetadata, error)
}

type MetadataFetcher struct {
	Fetchers []SpecificFetcher
}

func NewMetadataFetcher() *MetadataFetcher {
	return &MetadataFetcher{[]SpecificFetcher{twitter.NewMetadataFetcher(""), youtube.NewMetadataFetcher("")}}
}

func (f MetadataFetcher) Fetch(rawURL string) (types.PostMetadata, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return types.PostMetadata{}, err
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
	return types.PostMetadata{}, types.ErrNoMetadataFound
}
