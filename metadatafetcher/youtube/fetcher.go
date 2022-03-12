package youtube

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	mfTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/types"
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

	videoID := vField[0]
	video, err := NewYoutube(f.apiURL).GetVideoByID(videoID)
	if err != nil {
		return mfTypes.PostMetadata{}, err
	}

	_, err = types.ISO8601(video.PublishedAt).Time()
	if err != nil {
		return mfTypes.PostMetadata{}, fmt.Errorf("could not parse %v into valid time, with error: %v", video.PublishedAt, err)
	}

	return mfTypes.PostMetadata{
		Author:        video.ChannelTitle,
		AuthorURL:     video.ChannelURL,
		AuthorImgUrl:  video.ThumbnailDefaultURL,
		PostTitle:     video.VideoTitle,
		PostText:      video.VideoDescription,
		PostCreatedAt: types.ISO8601(video.PublishedAt),
		PostType:      mfTypes.YOUTUBE,
	}, nil
}

func (f YoutubeMetadataFetcher) IsCorrectFetcher(url *url.URL) bool {
	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		host = url.Host
	}
	return host == "youtube.com" || host == "www.youtube.com"
}
