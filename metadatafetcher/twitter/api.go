package twitter

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/drswork/go-twitter/twitter"
	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/request"
)

var (
	errTwitterOauth1CredentialsRequired = errors.New("the following env variables are required in order to post tweets: PREDICTIONS_TWITTER_CONSUMER_KEY, PREDICTIONS_TWITTER_CONSUMER_SECRET, PREDICTIONS_TWITTER_ACCESS_TOKEN, PREDICTIONS_TWITTER_ACCESS_SECRET")
)

// Twitter is the main struct for the Twitter component that interacts with Twitter via the Twitter API.
type Twitter struct {
	apiKey         string
	apiSecret      string
	bearerToken    string
	apiURL         string
	consumerKey    string
	consumerSecret string
	accessToken    string
	accessSecret   string

	api *twitter.Client
}

// NewTwitter is the constructor for the Twitter component that interacts with Twitter via the Twitter API.
func NewTwitter(apiURL string) Twitter {
	if apiURL == "" {
		apiURL = "https://api.twitter.com/2"
	}
	t := Twitter{
		apiKey:         os.Getenv("PREDICTIONS_TWITTER_API_KEY"),
		apiSecret:      os.Getenv("PREDICTIONS_TWITTER_API_SECRET"),
		bearerToken:    os.Getenv("PREDICTIONS_TWITTER_BEARER_TOKEN"),
		consumerKey:    os.Getenv("PREDICTIONS_TWITTER_CONSUMER_KEY"),
		consumerSecret: os.Getenv("PREDICTIONS_TWITTER_CONSUMER_SECRET"),
		accessToken:    os.Getenv("PREDICTIONS_TWITTER_ACCESS_TOKEN"),
		accessSecret:   os.Getenv("PREDICTIONS_TWITTER_ACCESS_SECRET"),

		apiURL: apiURL,
	}

	config := oauth1.NewConfig(t.consumerKey, t.consumerSecret)
	token := oauth1.NewToken(t.accessToken, t.accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	t.api = twitter.NewClient(httpClient)

	return t
}

// Tweet represents a Twitter Tweet.
type Tweet struct {
	TweetText      string
	TweetID        string
	TweetCreatedAt time.Time
	UserID         string
	UserName       string
	UserHandle     string
	ProfileImgURL  string
	UserCreatedAt  time.Time
	FollowersCount int
	Verified       bool

	err error
}

type responseData struct {
	AuthorID  string `json:"author_id"`
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
	CreatedAt     core.ISO8601                       `json:"created_at"`
	ProfileImgURL string                             `json:"profile_image_url"`
	PublicMetrics responseIncludesUsersPublicMetrics `json:"public_metrics"`
}

type responseIncludes struct {
	Users []responseIncludesUsers `json:"users"`
}

type response struct {
	Data     responseData     `json:"data"`
	Includes responseIncludes `json:"includes"`
	Errors   json.RawMessage  `json:"errors"`
}

func responseToTweet(r response) (Tweet, error) {
	if len(r.Errors) > 0 {
		return Tweet{}, fmt.Errorf("there are errors in Twitter's API response: %v", string(r.Errors))
	}
	tweetCreatedAt, err := core.ISO8601(r.Data.CreatedAt).Time()
	if err != nil {
		return Tweet{}, fmt.Errorf("while parsing r.Data.CreatedAt (%v) as ISO8601: %w", r.Data.CreatedAt, err)
	}
	if len(r.Includes.Users) < 1 {
		return Tweet{}, fmt.Errorf("expecting len(r.Includes.Users) to be >= 1, but was %v", len(r.Includes.Users))
	}
	userCreatedAt, err := core.ISO8601(r.Includes.Users[0].CreatedAt).Time()
	if err != nil {
		return Tweet{}, fmt.Errorf("while parsing r.Includes.Users[0].CreatedAt (%v) as ISO8601: %w", r.Includes.Users[0].CreatedAt, err)
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
		ProfileImgURL:  r.Includes.Users[0].ProfileImgURL,
		FollowersCount: r.Includes.Users[0].PublicMetrics.FollowersCount,
	}, nil
}

func parseError(err error) Tweet {
	return Tweet{err: err}
}

func (t Twitter) getTweetByID(id string) (Tweet, error) {
	req := request.Request[response, Tweet]{
		BaseURL: t.apiURL,
		Path:    fmt.Sprintf("tweets/%v?tweet.fields=created_at&user.fields=created_at,name,profile_image_url,public_metrics,verified,username&expansions=author_id,geo.place_id", id),
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %v", t.bearerToken),
			"Cookie":        "guest_id=v1%3A164530605018760344; Path=/; Domain=.twitter.com; Secure; Expires=Wed, 22 Mar 2023 21:27:30 GMT;",
		},
		ParseResponse: responseToTweet,
		ParseError:    parseError,
	}

	tweet := request.MakeRequest(req, false)
	if tweet.err != nil {
		return tweet, fmt.Errorf("while making HTTP request: %w", tweet.err)
	}

	return tweet, nil
}

// Tweet posts a Tweet to Twitter.
//
// Remember that if inReplyToStatusID is set, the text must contain @username, where the username is the
// handle whose tweet this tweet is replying to:
// https://developer.twitter.com/en/docs/twitter-api/v1/tweets/post-and-engage/api-reference/post-statuses-update
func (t Twitter) Tweet(text, imageFilename string, inReplyToStatusID int) (string, error) {
	if t.consumerKey == "" || t.consumerSecret == "" || t.accessToken == "" || t.accessSecret == "" {
		return "", errTwitterOauth1CredentialsRequired
	}

	mediaIDs := []int64{}
	if imageFilename != "" {
		mediaID, err := t.uploadMedia(imageFilename)
		if err != nil {
			return "", err
		}
		mediaIDs = []int64{mediaID}
	}

	tweet, _, err := t.api.Statuses.Update(text, &twitter.StatusUpdateParams{
		MediaIds:          mediaIDs,
		InReplyToStatusID: int64(inReplyToStatusID),
	})
	if err != nil {
		return "", err
	}

	return tweet.IDStr, nil
}

func (t Twitter) uploadMedia(imageFilename string) (int64, error) {
	if t.consumerKey == "" || t.consumerSecret == "" || t.accessToken == "" || t.accessSecret == "" {
		return 0, errTwitterOauth1CredentialsRequired
	}

	fileToBeUploaded := imageFilename
	file, err := os.Open(fileToBeUploaded)
	if err != nil {
		return 0, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, err
	}
	var size int64 = fileInfo.Size()
	bytes := make([]byte, size)

	// read file into bytes
	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)
	if err != nil {
		return 0, err
	}
	res, _, err := t.api.Media.Upload(bytes, "image/png")
	if err != nil {
		return 0, err
	}

	mediaID := res.MediaID
	processingInfo := res.ProcessingInfo
	// TODO: technically, Twitter could potentially make the daemon stuck forever without a timeout here
	for {
		if processingInfo == nil || processingInfo.State == "succeeded" {
			return mediaID, nil
		}
		if processingInfo.State == "failed" {
			return 0, errors.New(processingInfo.Error.Message)
		}
		fmt.Printf("Status of media upload is %v, checking after %v seconds...", processingInfo.State, processingInfo.CheckAfterSecs)
		time.Sleep(time.Duration(processingInfo.CheckAfterSecs) * time.Second)
		res, _, err := t.api.Media.Status(mediaID)
		if err != nil {
			return 0, err
		}
		processingInfo = res.ProcessingInfo
	}
}
