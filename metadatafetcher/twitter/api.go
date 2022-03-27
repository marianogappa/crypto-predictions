package twitter

import (
	"fmt"
	"os"
	"time"

	"github.com/marianogappa/predictions/request"
	"github.com/marianogappa/predictions/types"
)

type Twitter struct {
	apiKey      string
	apiSecret   string
	bearerToken string
	apiURL      string
}

func NewTwitter(apiURL string) Twitter {
	if apiURL == "" {
		apiURL = "https://api.twitter.com/2"
	}
	return Twitter{
		apiKey:      os.Getenv("PREDICTIONS_TWITTER_API_KEY"),
		apiSecret:   os.Getenv("PREDICTIONS_TWITTER_API_SECRET"),
		bearerToken: os.Getenv("PREDICTIONS_TWITTER_BEARER_TOKEN"),
		apiURL:      apiURL,
	}
}

type Tweet struct {
	TweetText      string
	TweetID        string
	TweetCreatedAt time.Time
	UserID         string
	UserName       string
	UserHandle     string
	ProfileImgUrl  string
	UserCreatedAt  time.Time
	FollowersCount int
	Verified       bool

	err error
}

type responseData struct {
	AuthorId  string `json:"author_id"`
	Text      string `json:"text"`
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

type responseIncludesUsersPublicMetrics struct {
	FollowersCount int `json:"followers_count"`
	TweetCount     int `json:"tweet_count"`
}

type responseIncludesUsers struct {
	ID            string                             `json:"id"`
	Name          string                             `json:"name"`
	Username      string                             `json:"username"`
	Verified      bool                               `json:"verified"`
	CreatedAt     types.ISO8601                      `json:"created_at"`
	ProfileImgUrl string                             `json:"profile_image_url"`
	PublicMetrics responseIncludesUsersPublicMetrics `json:"public_metrics"`
}

type responseIncludes struct {
	Users []responseIncludesUsers `json:"users"`
}

type response struct {
	Data     responseData     `json:"data"`
	Includes responseIncludes `json:"includes"`
}

func responseToTweet(r response) (Tweet, error) {
	tweetCreatedAt, err := types.ISO8601(r.Data.CreatedAt).Time()
	if err != nil {
		return Tweet{}, err
	}
	if len(r.Includes.Users) < 1 {
		return Tweet{}, fmt.Errorf("expecting len(r.Includes.Users) to be >= 1, but was %v", len(r.Includes.Users))
	}
	userCreatedAt, err := types.ISO8601(r.Includes.Users[0].CreatedAt).Time()
	if err != nil {
		return Tweet{}, err
	}
	return Tweet{
		TweetText:      r.Data.Text,
		TweetID:        r.Data.ID,
		TweetCreatedAt: tweetCreatedAt,
		UserID:         r.Includes.Users[0].ID,
		UserName:       r.Includes.Users[0].Name,
		UserHandle:     r.Includes.Users[0].Username,
		UserCreatedAt:  userCreatedAt,
		Verified:       r.Includes.Users[0].Verified,
		ProfileImgUrl:  r.Includes.Users[0].ProfileImgUrl,
		FollowersCount: r.Includes.Users[0].PublicMetrics.FollowersCount,
	}, nil
}

func parseError(err error) Tweet {
	return Tweet{err: err}
}

func (t Twitter) GetTweetByID(id string) (Tweet, error) {
	req := request.Request[response, Tweet]{
		BaseUrl: t.apiURL,
		Path:    fmt.Sprintf("tweets/%v?tweet.fields=created_at&user.fields=created_at,name,profile_image_url,public_metrics,verified,username&expansions=author_id,geo.place_id", id),
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %v", t.bearerToken),
			"Cookie":        "guest_id=v1%3A164530605018760344; Path=/; Domain=.twitter.com; Secure; Expires=Wed, 22 Mar 2023 21:27:30 GMT;",
		},
		ParseResponse: responseToTweet,
		ParseError:    parseError,
	}

	tweet := request.MakeRequest(req)
	if tweet.err != nil {
		return tweet, tweet.err
	}

	return tweet, nil
}
