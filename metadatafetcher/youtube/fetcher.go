package youtube

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/marianogappa/predictions/core"
	mfTypes "github.com/marianogappa/predictions/metadatafetcher/types"
)

// MetadataFetcher is the main struct for fetching metadata from Youtube.
type MetadataFetcher struct {
	apiURL string
}

// NewMetadataFetcher is constructor for fetching metadata from Youtube.
func NewMetadataFetcher(apiURL string) MetadataFetcher {
	return MetadataFetcher{apiURL}
}

// Fetch requests metadata from Youtube API for the specified URL.
func (f MetadataFetcher) Fetch(fetchURL *url.URL) (mfTypes.PostMetadata, error) {
	path := strings.Split(fetchURL.Path, "/")
	if len(path) != 2 || path[0] != "" || path[1] != "watch" {
		return mfTypes.PostMetadata{}, fmt.Errorf("invalid path for Youtube metadata fetching: %v", path)
	}

	m, err := url.ParseQuery(fetchURL.RawQuery)
	if err != nil {
		return mfTypes.PostMetadata{}, fmt.Errorf("invalid queryString for Youtube metadata fetching: %v", err)
	}

	vField, ok := m["v"]
	if !ok || len(vField) != 1 || vField[0] == "" {
		return mfTypes.PostMetadata{}, fmt.Errorf("invalid videoID for Youtube metadata fetching: %v", m["v"])
	}

	youtubeAPI := NewYoutube(f.apiURL)

	videoID := vField[0]
	video, err := youtubeAPI.getVideoByID(videoID)
	if err != nil {
		return mfTypes.PostMetadata{}, err
	}

	channel, err := youtubeAPI.getChannelByID(video.ChannelID)
	if err != nil {
		return mfTypes.PostMetadata{}, err
	}

	chURL, err := url.Parse(channel.URL)
	if err != nil {
		return mfTypes.PostMetadata{}, fmt.Errorf("error parsing channel's URL: %v", err)
	}

	chThumbDefURL, err := url.Parse(channel.ThumbnailDefaultURL)
	if err != nil {
		return mfTypes.PostMetadata{}, fmt.Errorf("error parsing channel's Default Thumbnail URL: %v", err)
	}

	chThumbMediumURL, err := url.Parse(channel.ThumbnailMediumURL)
	if err != nil {
		return mfTypes.PostMetadata{}, fmt.Errorf("error parsing channel's Medium Thumbnail URL: %v", err)
	}

	chThumbHighURL, err := url.Parse(channel.ThumbnailHighURL)
	if err != nil {
		return mfTypes.PostMetadata{}, fmt.Errorf("error parsing channel's High Thumbnail URL: %v", err)
	}

	chPublishedAt, err := channel.PublishedAt.Time()
	if err != nil {
		return mfTypes.PostMetadata{}, fmt.Errorf("error parsing channel's PublishedAt: %v", err)
	}

	return mfTypes.PostMetadata{
		Author: core.Account{
			URL:           chURL,
			AccountType:   "YOUTUBE",
			FollowerCount: channel.SubscriberCount,
			Thumbnails:    []*url.URL{chThumbDefURL, chThumbMediumURL, chThumbHighURL},
			Name:          channel.Title,
			Description:   channel.Description,
			CreatedAt:     &chPublishedAt,
			Handle:        channel.Title,
		},
		ThumbnailImgSmall:  video.ThumbnailDefaultURL,
		ThumbnailImgMedium: video.ThumbnailMediumURL,
		PostTitle:          video.VideoTitle,
		PostText:           video.VideoDescription,
		PostCreatedAt:      video.PublishedAt,
		PostType:           mfTypes.YOUTUBE,
	}, nil
}

// IsCorrectFetcher answers if this fetcher is the correct one for the specified URL.
func (f MetadataFetcher) IsCorrectFetcher(url *url.URL) bool {
	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		host = url.Host
	}
	return host == "youtube.com" || host == "www.youtube.com"
}
