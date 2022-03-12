package twitter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/marianogappa/signal-checker/common"
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
	TweetCreatedAt common.ISO8601
	UserID         string
	UserName       string
	UserHandle     string
}

type responseData struct {
	AuthorId  string `json:"author_id"`
	Text      string `json:"text"`
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

type responseIncludesUsers struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type responseIncludes struct {
	Users []responseIncludesUsers `json:"users"`
}

type response struct {
	Data     responseData     `json:"data"`
	Includes responseIncludes `json:"includes"`
}

func (r response) toTweet() (Tweet, error) {
	_, err := common.ISO8601(r.Data.CreatedAt).Time()
	if err != nil {
		return Tweet{}, err
	}
	if len(r.Includes.Users) < 1 {
		return Tweet{}, fmt.Errorf("expecting len(r.Includes.Users) to be >= 1, but was %v", len(r.Includes.Users))
	}
	return Tweet{
		TweetText:      r.Data.Text,
		TweetID:        r.Data.ID,
		TweetCreatedAt: common.ISO8601(r.Data.CreatedAt),
		UserID:         r.Includes.Users[0].ID,
		UserName:       r.Includes.Users[0].Name,
		UserHandle:     r.Includes.Users[0].Username,
	}, nil
}

func (t Twitter) GetTweetByID(id string) (Tweet, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%v/tweets/%v?tweet.fields=created_at&expansions=author_id", t.apiURL, id), nil)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", t.bearerToken))
	req.Header.Add("Cookie", "guest_id=v1%3A164530605018760344; Path=/; Domain=.twitter.com; Secure; Expires=Wed, 22 Mar 2023 21:27:30 GMT;")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return Tweet{}, err
	}
	defer resp.Body.Close()

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("twitter returned broken body response! Was: %v", string(byts))
		return Tweet{}, err
	}

	res := response{}
	if err := json.Unmarshal(byts, &res); err != nil {
		err2 := fmt.Errorf("twitter returned invalid JSON response! Response was: %v. Error is: %v", string(byts), err)
		return Tweet{}, err2
	}

	tweet, err := res.toTweet()
	if err != nil {
		return tweet, err
	}

	return tweet, nil
}
