package twitter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Twitter struct {
	apiKey      string
	apiSecret   string
	bearerToken string
}

func NewTwitter() Twitter {
	return Twitter{
		apiKey:      "QdmLenpIKLH28DYY1jc1VPLQc",
		apiSecret:   "Ag8rw5sZfQUgXOCYIYydWDDoBGZDXogDL64Md9yuw7uwJQ9oVH",
		bearerToken: "AAAAAAAAAAAAAAAAAAAAABavZQEAAAAAF5aVr9QJGBBpmQ0SaSvzMvalLoc%3D3Atd8zTPdeuq1VVK7nRaKg08EDwh03Aao9VlPTJrPP5CgrKAoG",
	}
}

type Tweet struct {
	TweetText      string
	TweetID        string
	TweetCreatedAt string
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

func (r response) toTweet() Tweet {
	return Tweet{
		TweetText:      r.Data.Text,
		TweetID:        r.Data.ID,
		TweetCreatedAt: r.Data.CreatedAt,
		UserID:         r.Includes.Users[0].ID,
		UserName:       r.Includes.Users[0].Name,
		UserHandle:     r.Includes.Users[0].Username,
	}
}

func (t Twitter) GetTweetByID(id string) (Tweet, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/2/tweets/%v?tweet.fields=created_at&expansions=author_id", id), nil)

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
	log.Println("tweet", string(byts), "response", res)

	return res.toTweet(), nil
}
