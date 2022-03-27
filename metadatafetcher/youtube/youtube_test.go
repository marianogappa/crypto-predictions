package youtube

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

func TestYoutubeHappyCase(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/videos":
			w.Write([]byte(`
			{
				"kind": "youtube#videoListResponse",
				"etag": "qrtHuLMG8lB-t9VFyDBLxyT9UVQ",
				"items":
				[
					{
						"kind": "youtube#video",
						"etag": "07sXT3jsXQoEuvQTmAuOw_8Kr08",
						"id": "ozgGPWnVLkY",
						"snippet":
						{
							"publishedAt": "2022-03-01T16:28:50Z",
							"channelId": "UCIRYBXDze5krPDzAEOxFGVA",
							"title": "Dozens of diplomats walk out during Russian foreign minister's UN speech",
							"description": "Dozens of diplomats walked out of a speech by the Russian foreign minister",
							"thumbnails":
							{
								"default":
								{
									"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/default.jpg",
									"width": 120,
									"height": 90
								},
								"medium":
								{
									"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/mqdefault.jpg",
									"width": 320,
									"height": 180
								},
								"high":
								{
									"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/hqdefault.jpg",
									"width": 480,
									"height": 360
								},
								"standard":
								{
									"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/sddefault.jpg",
									"width": 640,
									"height": 480
								},
								"maxres":
								{
									"url": "https://i.ytimg.com/vi/ozgGPWnVLkY/maxresdefault.jpg",
									"width": 1280,
									"height": 720
								}
							},
							"channelTitle": "Guardian News",
							"tags":
							[
								"2022"
							],
							"categoryId": "25",
							"liveBroadcastContent": "none",
							"localized":
							{
								"title": "Dozens of diplomats walk out during Russian foreign minister's UN speech",
								"description": "Dozens of diplomats walked out of a speech by the Russian foreign minister"
							},
							"defaultAudioLanguage": "en-US"
						}
					}
				],
				"pageInfo":
				{
					"totalResults": 1,
					"resultsPerPage": 1
				}
			}
			`))
		case "/channels":
			w.Write([]byte(`
			{
				"kind": "youtube#channelListResponse",
				"etag": "np4iVL9JNjoxTspKFCROyv48eu0",
				"pageInfo": {
				  "totalResults": 1,
				  "resultsPerPage": 5
				},
				"items": [
				  {
					"kind": "youtube#channel",
					"etag": "_BxQL9NUo7YwurKxZooSwDMbvhs",
					"id": "UCy6kyFxaMqGtpE3pQTflK8A",
					"snippet": {
					  "title": "Real Time with Bill Maher",
					  "description": "He's irrepressible, opinionated, and of course, politically incorrect. Watch new episodes of Real Time with Bill Maher, Fridays at 10PM\n\nSubscribe to the Real Time Channel for the latest on Bill Maher.",
					  "publishedAt": "2006-01-18T21:22:08Z",
					  "thumbnails": {
						"default": {
						  "url": "https://yt3.ggpht.com/dvzkOjrnLZXnpSN62MFUsJGPf7JF00Yzaeg4fsHVe-EvDXIovT9B9UMu4KTrpfVfojdNYQbvkw=s88-c-k-c0x00ffffff-no-rj",
						  "width": 88,
						  "height": 88
						},
						"medium": {
						  "url": "https://yt3.ggpht.com/dvzkOjrnLZXnpSN62MFUsJGPf7JF00Yzaeg4fsHVe-EvDXIovT9B9UMu4KTrpfVfojdNYQbvkw=s240-c-k-c0x00ffffff-no-rj",
						  "width": 240,
						  "height": 240
						},
						"high": {
						  "url": "https://yt3.ggpht.com/dvzkOjrnLZXnpSN62MFUsJGPf7JF00Yzaeg4fsHVe-EvDXIovT9B9UMu4KTrpfVfojdNYQbvkw=s800-c-k-c0x00ffffff-no-rj",
						  "width": 800,
						  "height": 800
						}
					  },
					  "localized": {
						"title": "Real Time with Bill Maher",
						"description": "He's irrepressible, opinionated, and of course, politically incorrect. Watch new episodes of Real Time with Bill Maher, Fridays at 10PM\n\nSubscribe to the Real Time Channel for the latest on Bill Maher."
					  }
					},
					"statistics": {
					  "viewCount": "1241134069",
					  "subscriberCount": "2290000",
					  "hiddenSubscriberCount": false,
					  "videoCount": "1970"
					}
				  }
				]
			  }
			`))
		}
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

	expected := mfTypes.PostMetadata{
		Author: types.Account{
			URL:           mURL("https://youtube.com/channel/UCy6kyFxaMqGtpE3pQTflK8A"),
			AccountType:   "YOUTUBE",
			FollowerCount: 2290000,
			Thumbnails: []*url.URL{
				mURL("https://yt3.ggpht.com/dvzkOjrnLZXnpSN62MFUsJGPf7JF00Yzaeg4fsHVe-EvDXIovT9B9UMu4KTrpfVfojdNYQbvkw=s88-c-k-c0x00ffffff-no-rj"),
				mURL("https://yt3.ggpht.com/dvzkOjrnLZXnpSN62MFUsJGPf7JF00Yzaeg4fsHVe-EvDXIovT9B9UMu4KTrpfVfojdNYQbvkw=s240-c-k-c0x00ffffff-no-rj"),
				mURL("https://yt3.ggpht.com/dvzkOjrnLZXnpSN62MFUsJGPf7JF00Yzaeg4fsHVe-EvDXIovT9B9UMu4KTrpfVfojdNYQbvkw=s800-c-k-c0x00ffffff-no-rj"),
			},
			Name:        "Real Time with Bill Maher",
			Description: "He's irrepressible, opinionated, and of course, politically incorrect. Watch new episodes of Real Time with Bill Maher, Fridays at 10PM\n\nSubscribe to the Real Time Channel for the latest on Bill Maher.",
			CreatedAt:   ptFromISO("2006-01-18T21:22:08Z"),
		},
		ThumbnailImgSmall:  "https://i.ytimg.com/vi/ozgGPWnVLkY/default.jpg",
		ThumbnailImgMedium: "https://i.ytimg.com/vi/ozgGPWnVLkY/mqdefault.jpg",
		PostTitle:          "Dozens of diplomats walk out during Russian foreign minister's UN speech",
		PostText:           "Dozens of diplomats walked out of a speech by the Russian foreign minister",
		PostCreatedAt:      types.ISO8601("2022-03-01T16:28:50Z"),
		PostType:           mfTypes.YOUTUBE,
	}

	require.Equal(t, pm, expected)
}

func mURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}

func ptFromISO(s string) *time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return &t
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
