package twitter

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/marianogappa/predictions/metadatafetcher/types"
	coreTypes "github.com/marianogappa/predictions/types"
)

type TwitterMetadataFetcher struct {
	apiURL string
}

func NewMetadataFetcher(apiURL string) TwitterMetadataFetcher {
	return TwitterMetadataFetcher{apiURL}
}

func (f TwitterMetadataFetcher) Fetch(u *url.URL) (types.PostMetadata, error) {
	path := strings.Split(u.Path, "/")
	if len(path) != 4 || path[0] != "" || path[2] != "status" {
		return types.PostMetadata{}, fmt.Errorf("invalid path for Twitter metadata fetching: %v", path)
	}

	tweet, err := NewTwitter(f.apiURL).GetTweetByID(path[3])
	if err != nil {
		return types.PostMetadata{}, err
	}

	userURL, err := url.Parse(fmt.Sprintf("https://twitter.com/%v", tweet.UserHandle))
	if err != nil {
		return types.PostMetadata{}, fmt.Errorf("error parsing user's URL: %v", err)
	}

	userProfileImgURL, err := url.Parse(tweet.ProfileImgUrl)
	if err != nil {
		return types.PostMetadata{}, fmt.Errorf("error parsing user's profile image URL: %v", err)
	}

	userProfileMediumImgURL, err := url.Parse(strings.Replace(tweet.ProfileImgUrl, "_normal.", "_400x400.", -1))
	if err != nil {
		return types.PostMetadata{}, fmt.Errorf("error parsing user's profile medium image URL: %v", err)
	}

	return types.PostMetadata{
		Author: coreTypes.Account{
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
		PostCreatedAt: coreTypes.ISO8601(tweet.TweetCreatedAt.Format(time.RFC3339)), // TODO
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
