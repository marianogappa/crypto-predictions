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
		w.Write([]byte(`
		{
			"data": {
			  "text": "Where are the bears now? 🐻🔫",
			  "created_at": "2022-03-24T15:26:16.000Z",
			  "id": "1507015952621211649",
			  "author_id": "1353384573435056128"
			},
			"includes": {
			  "users": [
				{
				  "name": "Crypto Rover",
				  "profile_image_url": "https://pbs.twimg.com/profile_images/1492942875373588490/GSC34oOF_normal.jpg",
				  "public_metrics": {
					"followers_count": 93571,
					"following_count": 273,
					"tweet_count": 8591,
					"listed_count": 294
				  },
				  "created_at": "2021-01-24T16:50:08.000Z",
				  "verified": true,
				  "id": "1353384573435056128",
				  "username": "rovercrc"
				}
			  ]
			}
		  }
		`))
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
