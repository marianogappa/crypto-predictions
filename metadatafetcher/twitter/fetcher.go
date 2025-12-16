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

// Fetch requests metadata from Twitter using oEmbed API scraping.
func (f MetadataFetcher) Fetch(u *url.URL) (types.PostMetadata, error) {
	tweetURL := u.String()
	if !strings.HasPrefix(tweetURL, "http://") && !strings.HasPrefix(tweetURL, "https://") {
		tweetURL = "https://" + u.Host + u.Path
	}

	tweet, err := scrapeTweetData(tweetURL, f.apiURL)
	if err != nil {
		return types.PostMetadata{}, fmt.Errorf("while scraping tweet data: %w", err)
	}

	userURL, err := url.Parse(fmt.Sprintf("https://twitter.com/%v", tweet.UserHandle))
	if err != nil {
		return types.PostMetadata{}, fmt.Errorf("error parsing user's URL: %w", err)
	}

	// Build thumbnails array - empty if ProfileImgURL is not available
	thumbnails := []*url.URL{}
	if tweet.ProfileImgURL != "" {
		userProfileImgURL, err := url.Parse(tweet.ProfileImgURL)
		if err == nil {
			thumbnails = append(thumbnails, userProfileImgURL)
			// Also add medium size version if available
			userProfileMediumImgURL, err := url.Parse(strings.Replace(tweet.ProfileImgURL, "_normal.", "_400x400.", -1))
			if err == nil {
				thumbnails = append(thumbnails, userProfileMediumImgURL)
			}
		}
	}

	// Set CreatedAt only if UserCreatedAt is not zero
	var userCreatedAt *time.Time
	if !tweet.UserCreatedAt.IsZero() {
		userCreatedAt = &tweet.UserCreatedAt
	}

	return types.PostMetadata{
		Author: core.Account{
			URL:           userURL,
			AccountType:   "TWITTER",
			FollowerCount: tweet.FollowersCount, // Will be 0 if not available
			Thumbnails:    thumbnails,           // Will be empty slice if not available
			Handle:        tweet.UserHandle,
			Name:          tweet.UserName,
			IsVerified:    tweet.Verified, // Will be false if not available
			CreatedAt:     userCreatedAt,  // Will be nil if not available
		},
		PostTitle:     tweet.TweetText,
		PostText:      tweet.TweetText,
		PostCreatedAt: core.ISO8601(tweet.TweetCreatedAt.Format(time.RFC3339)),
		PostType:      types.TWITTER,
	}, nil
}

// IsCorrectFetcher answers if this fetcher is the correct one for the specified URL.
func (f MetadataFetcher) IsCorrectFetcher(url *url.URL) bool {
	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		host = url.Host
	}
	return host == "twitter.com" || host == "www.twitter.com" || host == "x.com" || host == "www.x.com"
}
