package youtube

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	mfTypes "github.com/marianogappa/predictions/metadatafetcher/types"
)

type YoutubeMetadataFetcher struct {
	apiURL string
}

func NewMetadataFetcher(apiURL string) YoutubeMetadataFetcher {
	return YoutubeMetadataFetcher{apiURL}
}

func (f YoutubeMetadataFetcher) Fetch(fetchURL *url.URL) (mfTypes.PostMetadata, error) {
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
	video, err := youtubeAPI.GetVideoByID(videoID)
	if err != nil {
		return mfTypes.PostMetadata{}, err
	}

	channel, err := youtubeAPI.GetChannelByID(video.ChannelID)
	if err != nil {
		return mfTypes.PostMetadata{}, err
	}

	return mfTypes.PostMetadata{
		Author: mfTypes.PostAuthor{
			URL:               channel.URL,
			AuthorImgSmall:    channel.ThumbnailMediumURL,
			AuthorImgMedium:   channel.ThumbnailHighURL,
			AuthorName:        channel.Title,
			AuthorDescription: channel.Description,
			FollowerCount:     channel.SubscriberCount,
		},
		ThumbnailImgSmall:  video.ThumbnailDefaultURL,
		ThumbnailImgMedium: video.ThumbnailMediumURL,
		PostTitle:          video.VideoTitle,
		PostText:           video.VideoDescription,
		PostCreatedAt:      video.PublishedAt,
		PostType:           mfTypes.YOUTUBE,
	}, nil
}

func (f YoutubeMetadataFetcher) IsCorrectFetcher(url *url.URL) bool {
	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		host = url.Host
	}
	return host == "youtube.com" || host == "www.youtube.com"
}
