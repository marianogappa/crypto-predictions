package twitter

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/marianogappa/predictions/metadatafetcher/types"
)

type TwitterMetadataFetcher struct {
	apiURL string
}

func NewMetadataFetcher(apiURL string) TwitterMetadataFetcher {
	return TwitterMetadataFetcher{apiURL}
}

func (f TwitterMetadataFetcher) Fetch(url *url.URL) (types.PostMetadata, error) {
	path := strings.Split(url.Path, "/")
	if len(path) != 4 || path[0] != "" || path[2] != "status" {
		return types.PostMetadata{}, fmt.Errorf("invalid path for Twitter metadata fetching: %v", path)
	}

	tweet, err := NewTwitter(f.apiURL).GetTweetByID(path[3])
	if err != nil {
		return types.PostMetadata{}, err
	}

	return types.PostMetadata{
		Author:        tweet.UserHandle,
		AuthorURL:     fmt.Sprintf("https://twitter.com/%v", tweet.UserHandle),
		AuthorImgUrl:  "", // TODO
		PostTitle:     tweet.TweetText,
		PostText:      tweet.TweetText,
		PostCreatedAt: tweet.TweetCreatedAt, // TODO
		PostType:      types.TWITTER,
	}, nil
}

func (f TwitterMetadataFetcher) IsCorrectFetcher(url *url.URL) bool {
	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		host = url.Host
	}
	return host == "twitter.com" || host == "www.twitter.com"
}
