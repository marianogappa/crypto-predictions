package twitter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/marianogappa/predictions/core"
	mfTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/stretchr/testify/require"
)

func TestTwitterHappyCase(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock oEmbed API response
		w.Write([]byte(`{
			"author_name": "Crypto Rover",
			"author_url": "https://twitter.com/rovercrc",
			"html": "<blockquote class=\"twitter-tweet\"><p lang=\"en\" dir=\"ltr\">Where are the bears now? 🐻🔫</p>&mdash; Crypto Rover (@rovercrc) <a href=\"https://twitter.com/rovercrc/status/1507015952621211649?ref_src=twsrc%5Etfw\">March 24, 2022</a></blockquote>",
			"url": "https://twitter.com/rovercrc/status/1507015952621211649"
		}`))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://twitter.com/rovercrc/status/1507015952621211649")
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
		Author: core.Account{
			URL:           mURL("https://twitter.com/rovercrc"),
			AccountType:   "TWITTER",
			Handle:        "rovercrc",
			FollowerCount: 0,            // Not available via oEmbed
			Thumbnails:    []*url.URL{}, // Not available via oEmbed
			Name:          "Crypto Rover",
			Description:   "",
			CreatedAt:     nil,   // Not available via oEmbed
			IsVerified:    false, // Not available via oEmbed
		},
		PostTitle:     "Where are the bears now? 🐻🔫",
		PostText:      "Where are the bears now? 🐻🔫",
		PostCreatedAt: core.ISO8601("2022-03-24T00:00:00Z"), // Date parsed from "March 24, 2022"
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
		// oEmbed response with unparseable date
		w.Write([]byte(`{
			"author_name": "il Capo Of Crypto",
			"author_url": "https://twitter.com/CryptoCapo_",
			"html": "<blockquote class=\"twitter-tweet\"><p lang=\"en\" dir=\"ltr\">sample tweet content</p>&mdash; il Capo Of Crypto (@CryptoCapo_) <a href=\"https://twitter.com/CryptoCapo_/status/1491357566974054400?ref_src=twsrc%5Etfw\">Invalid Date Format</a></blockquote>",
			"url": "https://twitter.com/CryptoCapo_/status/1491357566974054400"
		}`))
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
func TestTwitterNoAuthor(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// oEmbed response without author_url
		w.Write([]byte(`{
			"author_name": "",
			"author_url": "",
			"html": "<blockquote class=\"twitter-tweet\"><p lang=\"en\" dir=\"ltr\">sample tweet content</p></blockquote>",
			"url": "https://twitter.com/CryptoCapo_/status/1491357566974054400"
		}`))
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
		t.Errorf("request should have failed with no author")
		t.FailNow()
	}
}

func TestTwitterInvalidBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
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
		t.Errorf("request should have failed due to invalid response")
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
	// Test with invalid oEmbed base URL - should still work as it defaults to production
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://twitter.com/invalid_tweet_that_does_not_exist_12345/status/9999999999999999999")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	// This might succeed or fail depending on whether the tweet exists
	// So we just test that it doesn't panic
	_, _ = fetcher.Fetch(u)
}
func TestTwitterPathTooLong(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"author_name": "Test", "author_url": "https://twitter.com/test", "html": "<blockquote>test</blockquote>", "url": "https://twitter.com/test/status/123"}`))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://twitter.com/path/is/too/long")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	// Should fail because we can't extract tweet ID from invalid path
	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("should have failed")
	}
}

func TestTwitterPathNoStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"author_name": "Test", "author_url": "https://twitter.com/test", "html": "<blockquote>test</blockquote>", "url": "https://twitter.com/test/status/123"}`))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://twitter.com/CryptoCapo_/not_status/1491357566974054400")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	// Should fail because we can't extract tweet ID (no "status" in path)
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
	// NewTwitter is still used internally but not directly by MetadataFetcher anymore
	// This test can remain for backward compatibility
	y := NewTwitter("")
	if y.apiURL != "https://api.twitter.com/2" {
		t.Errorf("invalid production API URL %v", y.apiURL)
	}
}
