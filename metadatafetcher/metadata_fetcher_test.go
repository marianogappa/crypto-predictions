package metadatafetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marianogappa/predictions/metadatafetcher/twitter"
	"github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/metadatafetcher/youtube"
)

func TestMetadataFetcherFetchInvalidURL(t *testing.T) {
	_, err := NewMetadataFetcher().Fetch(" http://foo.com")
	if err == nil {
		t.Errorf("should have failed with invalid url")
	}
}

func TestMetadataFetcherFetchNoValidFetchers(t *testing.T) {
	_, err := NewMetadataFetcher().Fetch("https://unsupportedsite.com?v=123456543")
	if err != types.ErrNoMetadataFound {
		t.Errorf("should have failed with no metadata found")
	}
}

func TestMetadataFetcherFetchHappyCase(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"id":"1491357566974054400","created_at":"2022-02-09T10:25:26.000Z","author_id":"988796804446769153","text":"sample tweet content"},"includes":{"users":[{"id":"988796804446769153","name":"il Capo Of Crypto","username":"CryptoCapo_"}]}}`))
	}))
	defer ts.Close()

	mf := NewMetadataFetcher()
	mf.Fetchers = []SpecificFetcher{youtube.NewMetadataFetcher("nope"), twitter.NewMetadataFetcher(ts.URL)}
	_, err := mf.Fetch("https://twitter.com/CryptoCapo_/status/1491357566974054400")
	if err != nil {
		t.Errorf("shouldn't have failed, but had error: %v", err)
	}
}

func TestMetadataFetcherFetchErrorInFetcher(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"id":"1491357566974054400","created_at":"2022 02 09T10:25:26.000Z","author_id":"988796804446769153","text":"sample tweet content"},"includes":{"users":[{"id":"988796804446769153","name":"il Capo Of Crypto","username":"CryptoCapo_"}]}}`))
	}))
	defer ts.Close()

	mf := NewMetadataFetcher()
	mf.Fetchers = []SpecificFetcher{youtube.NewMetadataFetcher("nope"), twitter.NewMetadataFetcher(ts.URL)}
	_, err := mf.Fetch("https://twitter.com/CryptoCapo_/status/1491357566974054400")
	if err == nil {
		t.Errorf("should have failed with invalid date")
	}
}
