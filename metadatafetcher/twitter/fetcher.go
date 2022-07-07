package twitter

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/metadatafetcher/types"
)

// MetadataFetcher is the main struct for fetching metadata from Twitter.
type MetadataFetcher struct {
	apiURL string
}

// NewMetadataFetcher is constructor for fetching metadata from Twitter.
func NewMetadataFetcher(apiURL string) MetadataFetcher {
	return MetadataFetcher{apiURL}
}

// Fetch requests metadata from Twitter API for the specified URL.
func (f MetadataFetcher) Fetch(u *url.URL) (types.PostMetadata, error) {
	path := strings.Split(u.Path, "/")
	if len(path) != 4 || path[0] != "" || path[2] != "status" {
		return types.PostMetadata{}, fmt.Errorf("invalid path for Twitter metadata fetching: %v", path)
	}

	tweet, err := NewTwitter(f.apiURL).getTweetByID(path[3])
	if err != nil {
		return types.PostMetadata{}, fmt.Errorf("while getting tweet by ID (%v): %w", path[3], err)
	}

	userURL, err := url.Parse(fmt.Sprintf("https://twitter.com/%v", tweet.UserHandle))
	if err != nil {
		return types.PostMetadata{}, fmt.Errorf("error parsing user's URL: %w", err)
	}

	userProfileImgURL, err := url.Parse(tweet.ProfileImgURL)
	if err != nil {
		return types.PostMetadata{}, fmt.Errorf("error parsing user's profile image URL: %w", err)
	}

	userProfileMediumImgURL, err := url.Parse(strings.Replace(tweet.ProfileImgURL, "_normal.", "_400x400.", -1))
	if err != nil {
		return types.PostMetadata{}, fmt.Errorf("error parsing user's profile medium image URL: %w", err)
	}

	return types.PostMetadata{
		Author: core.Account{
			URL:           userURL,
			AccountType:   "TWITTER",
			FollowerCount: tweet.FollowersCount,
			Thumbnails:    []*url.URL{userProfileImgURL, userProfileMediumImgURL},
			Handle:        tweet.UserHandle,
			Name:          tweet.UserName,
			IsVerified:    tweet.Verified,
			CreatedAt:     &tweet.UserCreatedAt,
		},
		PostTitle:     tweet.TweetText,
		PostText:      tweet.TweetText,
		PostCreatedAt: core.ISO8601(tweet.TweetCreatedAt.Format(time.RFC3339)), // TODO
		PostType:      types.TWITTER,
	}, nil
}

// IsCorrectFetcher answers if this fetcher is the correct one for the specified URL.
func (f MetadataFetcher) IsCorrectFetcher(url *url.URL) bool {
	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		host = url.Host
	}
	return host == "twitter.com" || host == "www.twitter.com"
}
