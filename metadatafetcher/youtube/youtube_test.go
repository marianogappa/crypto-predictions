package youtube

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/signal-checker/common"
)

func TestYoutubeHappyCase(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"kind": "youtube#videoListResponse", "etag": "qrtHuLMG8lB-t9VFyDBLxyT9UVQ", "items": [{"kind": "youtube#video", "etag": "07sXT3jsXQoEuvQTmAuOw_8Kr08", "id": "ozgGPWnVLkY", "snippet": {"publishedAt": "2022-03-01T16:28:50Z", "channelId": "UCIRYBXDze5krPDzAEOxFGVA", "title": "Dozens of diplomats walk out during Russian foreign minister's UN speech", "description": "Dozens of diplomats walked out of a speech by the Russian foreign minister", "thumbnails": {"default": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/default.jpg", "width": 120, "height": 90 }, "medium": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/mqdefault.jpg", "width": 320, "height": 180 }, "high": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/hqdefault.jpg", "width": 480, "height": 360 }, "standard": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/sddefault.jpg", "width": 640, "height": 480 }, "maxres": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/maxresdefault.jpg", "width": 1280, "height": 720 } }, "channelTitle": "Guardian News", "tags": ["2022"], "categoryId": "25", "liveBroadcastContent": "none", "localized": {"title": "Dozens of diplomats walk out during Russian foreign minister's UN speech", "description": "Dozens of diplomats walked out of a speech by the Russian foreign minister"}, "defaultAudioLanguage": "en-US"} } ], "pageInfo": {"totalResults": 1, "resultsPerPage": 1 } }`))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://www.youtube.com/watch?v=OgDrDVsWEc4")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	pm, err := fetcher.Fetch(u)
	if err != nil {
		t.Errorf("request should have succeeded, but this error happened: %v", err)
		t.FailNow()
	}

	expected := types.PostMetadata{
		Author:        "Guardian News",
		AuthorURL:     "https://www.youtube.com/c/UCIRYBXDze5krPDzAEOxFGVA",
		AuthorImgUrl:  "https://i.ytimg.com/vi/ozgGPWnVLkY/default.jpg",
		PostTitle:     "Dozens of diplomats walk out during Russian foreign minister's UN speech",
		PostText:      "Dozens of diplomats walked out of a speech by the Russian foreign minister",
		PostCreatedAt: common.ISO8601("2022-03-01T16:28:50Z"),
		PostType:      types.YOUTUBE,
	}

	if !reflect.DeepEqual(pm, expected) {
		t.Errorf("expected %v but got %v", expected, pm)
	}
}

func TestYoutubeMultipleResponseItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"kind": "youtube#videoListResponse", "etag": "qrtHuLMG8lB-t9VFyDBLxyT9UVQ", "items": [{}, {"kind": "youtube#video", "etag": "07sXT3jsXQoEuvQTmAuOw_8Kr08", "id": "ozgGPWnVLkY", "snippet": {"publishedAt": "2022-03-01T16:28:50Z", "channelId": "UCIRYBXDze5krPDzAEOxFGVA", "title": "Dozens of diplomats walk out during Russian foreign minister's UN speech", "description": "Dozens of diplomats walked out of a speech by the Russian foreign minister", "thumbnails": {"default": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/default.jpg", "width": 120, "height": 90 }, "medium": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/mqdefault.jpg", "width": 320, "height": 180 }, "high": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/hqdefault.jpg", "width": 480, "height": 360 }, "standard": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/sddefault.jpg", "width": 640, "height": 480 }, "maxres": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/maxresdefault.jpg", "width": 1280, "height": 720 } }, "channelTitle": "Guardian News", "tags": ["2022"], "categoryId": "25", "liveBroadcastContent": "none", "localized": {"title": "Dozens of diplomats walk out during Russian foreign minister's UN speech", "description": "Dozens of diplomats walked out of a speech by the Russian foreign minister"}, "defaultAudioLanguage": "en-US"} } ], "pageInfo": {"totalResults": 1, "resultsPerPage": 1 } }`))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://www.youtube.com/watch?v=OgDrDVsWEc4")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("request should have failed due to multiple items")
		t.FailNow()
	}
}

func TestYoutubeInvalidTime(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"kind": "youtube#videoListResponse", "etag": "qrtHuLMG8lB-t9VFyDBLxyT9UVQ", "items": [{"kind": "youtube#video", "etag": "07sXT3jsXQoEuvQTmAuOw_8Kr08", "id": "ozgGPWnVLkY", "snippet": {"publishedAt": "2022 03 01 16:28:50", "channelId": "UCIRYBXDze5krPDzAEOxFGVA", "title": "Dozens of diplomats walk out during Russian foreign minister's UN speech", "description": "Dozens of diplomats walked out of a speech by the Russian foreign minister", "thumbnails": {"default": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/default.jpg", "width": 120, "height": 90 }, "medium": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/mqdefault.jpg", "width": 320, "height": 180 }, "high": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/hqdefault.jpg", "width": 480, "height": 360 }, "standard": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/sddefault.jpg", "width": 640, "height": 480 }, "maxres": {"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/maxresdefault.jpg", "width": 1280, "height": 720 } }, "channelTitle": "Guardian News", "tags": ["2022"], "categoryId": "25", "liveBroadcastContent": "none", "localized": {"title": "Dozens of diplomats walk out during Russian foreign minister's UN speech", "description": "Dozens of diplomats walked out of a speech by the Russian foreign minister"}, "defaultAudioLanguage": "en-US"} } ], "pageInfo": {"totalResults": 1, "resultsPerPage": 1 } }`))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://www.youtube.com/watch?v=OgDrDVsWEc4")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("request should have failed due to invalid time")
		t.FailNow()
	}
}

func TestYoutubeInvalidBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://www.youtube.com/watch?v=OgDrDVsWEc4")
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

func TestYoutubeInvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	fetcher := NewMetadataFetcher(ts.URL)

	u, err := url.Parse("https://www.youtube.com/watch?v=OgDrDVsWEc4")
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

func TestYoutubeInvalidURL(t *testing.T) {
	fetcher := NewMetadataFetcher("invalid url")

	u, err := url.Parse("https://www.youtube.com/watch?v=OgDrDVsWEc4")
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
func TestYoutubePathTooLong(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://www.youtube.com/too/long/watch?v=OgDrDVsWEc4")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("should have failed")
	}
}

func TestYoutubePathNoWatch(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://www.youtube.com/no_watch?v=OgDrDVsWEc4")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("should have failed because not_status")
	}
}

func TestYoutubePathInvalidQueryString(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://www.youtube.com/watch?v=OgDrDVsWEc4;invalidsemicolon")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("should have failed because of invalid query string")
	}
}

func TestYoutubePathInvalidQueryString2(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://www.youtube.com/watch?video_instead_of_v=OgDrDVsWEc4")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}

	_, err = fetcher.Fetch(u)
	if err == nil {
		t.Errorf("should have failed because of invalid key on query string")
	}
}

func TestYoutubeIsCorrectFetcherTrue(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://www.youtube.com/watch?v=OgDrDVsWEc4")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}
	if !fetcher.IsCorrectFetcher(u) {
		t.Errorf("should have been correct fetcher")
		t.FailNow()
	}
}

func TestYoutubeIsCorrectFetcherFalse(t *testing.T) {
	fetcher := NewMetadataFetcher("")

	u, err := url.Parse("https://www.notyoutube.com/watch?v=OgDrDVsWEc4")
	if err != nil {
		t.Errorf("parsing url shouldn't have failed; test invalid")
		t.FailNow()
	}
	if fetcher.IsCorrectFetcher(u) {
		t.Errorf("should have been incorrect fetcher")
		t.FailNow()
	}
}

func TestNewYoutube(t *testing.T) {
	y := NewYoutube("")
	if y.apiURL != "https://youtube.googleapis.com/youtube/v3" {
		t.Errorf("invalid production API URL %v", y.apiURL)
	}
}
