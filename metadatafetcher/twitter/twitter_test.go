package twitter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	mfTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/types"
	"github.com/stretchr/testify/require"
)

func TestTwitterHappyCase(t *testing.T) {
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

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://twitter.com/CryptoCapo_/status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	pm, err := fetcher.Fetch(u)
	if err != nil {
		t.Errorf("request should have succeeded, but this error happened: %v", err)
		t.FailNow()
	}

	expected := mfTypes.PostMetadata{
		Author: types.Account{
			URL:           mURL("https://twitter.com/rovercrc"),
			AccountType:   "TWITTER",
			Handle:        "rovercrc",
			FollowerCount: 93571,
			Thumbnails:    []*url.URL{mURL("https://pbs.twimg.com/profile_images/1492942875373588490/GSC34oOF_normal.jpg"), mURL("https://pbs.twimg.com/profile_images/1492942875373588490/GSC34oOF_400x400.jpg")},
			Name:          "Crypto Rover",
			Description:   "",
			CreatedAt:     ptFromISO("2021-01-24T16:50:08.000Z"),
			IsVerified:    true,
		},
		PostTitle:     "Where are the bears now? 🐻🔫",
		PostText:      "Where are the bears now? 🐻🔫",
		PostCreatedAt: types.ISO8601("2022-03-24T15:26:16Z"),
		PostType:      mfTypes.TWITTER,
	}

	require.Equal(t, expected, pm)
}

func mURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}

func ptFromISO(s string) *time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return &t
}

func TestTwitterInvalidTime(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"id":"1491357566974054400","created_at":"2022 02 09T10:25:26.000Z","author_id":"988796804446769153","text":"sample tweet content"},"includes":{"users":[{"id":"988796804446769153","name":"il Capo Of Crypto","username":"CryptoCapo_"}]}}`))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://twitter.com/CryptoCapo_/status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("request should have failed with invalid time")
		t.FailNow()
	}

}
func TestTwitterNoUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"id":"1491357566974054400","created_at":"2022-02-09T10:25:26.000Z","author_id":"988796804446769153","text":"sample tweet content"},"includes":{"users":[]}}`))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://twitter.com/CryptoCapo_/status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("request should have failed with no users")
		t.FailNow()
	}
}

func TestTwitterInvalidBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://twitter.com/CryptoCapo_/status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("request should have failed due to invalid body")
		t.FailNow()
	}
}

func TestTwitterInvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://twitter.com/CryptoCapo_/status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("request should have failed due to invalid json")
		t.FailNow()
	}
}

func TestTwitterInvalidURL(t *testing.T) {
	fetcher := NewMetadataFetcher("invalid url")

	u, err := url.Parse("https://twitter.com/CryptoCapo_/status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("request should have failed due to invalid URL")
		t.FailNow()
	}
}
func TestTwitterPathTooLong(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://twitter.com/path/is/too/long")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("should have failed")
	}
}

func TestTwitterPathNoStatus(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://twitter.com/CryptoCapo_/not_status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("should have failed because not_status")
	}
}

func TestTwitterIsCorrectFetcherTrue(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://twitter.com/CryptoCapo_/status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}
	if !fetcher.IsCorrectFetcher(u) {
		t.Errorf("should have been correct fetcher")
		t.FailNow()
	}
}

func TestTwitterIsCorrectFetcherFalse(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://nottwitter.com/CryptoCapo_/status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}
	if fetcher.IsCorrectFetcher(u) {
		t.Errorf("should have been incorrect fetcher")
		t.FailNow()
	}
}

func TestNewTwitter(t *testing.T) {
	y := NewTwitter("")
	if y.apiURL != "https://api.twitter.com/2" {
		t.Errorf("invalid production API URL %v", y.apiURL)
	}
}

func TestRefreshCookie(t *testing.T) {
	refreshTime := time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)
	if time.Now().After(refreshTime) {
		t.Errorf("Time to refresh the Twitter Cookie!")
	}
}
