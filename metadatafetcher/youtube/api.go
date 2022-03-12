package youtube

/*
curl \
 'https://youtube.googleapis.com/youtube/v3/videos?part=contentDetails%2Cid%2CliveStreamingDetails%2Clocalizations%2Cplayer%2CrecordingDetails%2Csnippet%2Cstatistics%2Cstatus%2CtopicDetails&id=ozgGPWnVLkY&key=AIzaSyBGIaqdlYg4feSjj5DmmTIMTTRWuXEAcY4' \
  --header 'Accept: application/json' \
  --compressed
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Youtube struct {
	apiKey string
	apiURL string
}

func NewYoutube(apiURL string) Youtube {
	if apiURL == "" {
		apiURL = "https://youtube.googleapis.com/youtube/v3"
	}

	return Youtube{
		apiKey: os.Getenv("PREDICTIONS_YOUTUBE_API_KEY"),
		apiURL: apiURL,
	}
}

type Video struct {
	VideoTitle           string
	VideoDescription     string
	VideoID              string
	PublishedAt          string
	ChannelURL           string
	ChannelTitle         string
	ThumbnailDefaultURL  string
	ThumbnailMediumURL   string
	ThumbnailHighURL     string
	ThumbnailStandardURL string
	ThumbnailMaxresURL   string
}

type responseSnippetThumbnailsThumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
type responseSnippetThumbnails struct {
	Default  responseSnippetThumbnailsThumbnail `json:"default"`
	Medium   responseSnippetThumbnailsThumbnail `json:"medium"`
	High     responseSnippetThumbnailsThumbnail `json:"high"`
	Standard responseSnippetThumbnailsThumbnail `json:"standard"`
	Maxres   responseSnippetThumbnailsThumbnail `json:"maxres"`
}

type responseSnippet struct {
	PublishedAt  string                    `json:"publishedAt"`
	ChannelId    string                    `json:"channelId"`
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	Thumbnails   responseSnippetThumbnails `json:"thumbnails"`
	ChannelTitle string                    `json:"channelTitle"`
}
type responseItems struct {
	Snippet responseSnippet `json:"snippet"`
}

type response struct {
	Items []responseItems `json:"items"`
}

func (r response) toVideo() (Video, error) {
	if len(r.Items) != 1 {
		return Video{}, fmt.Errorf("expected len(Items) == 1, but was %v", len(r.Items))
	}
	return Video{
		ChannelURL:           fmt.Sprintf("https://www.youtube.com/c/%v", r.Items[0].Snippet.ChannelId),
		PublishedAt:          r.Items[0].Snippet.PublishedAt,
		ChannelTitle:         r.Items[0].Snippet.ChannelTitle,
		VideoTitle:           r.Items[0].Snippet.Title,
		VideoDescription:     r.Items[0].Snippet.Description,
		ThumbnailDefaultURL:  r.Items[0].Snippet.Thumbnails.Default.URL,
		ThumbnailMediumURL:   r.Items[0].Snippet.Thumbnails.Medium.URL,
		ThumbnailHighURL:     r.Items[0].Snippet.Thumbnails.High.URL,
		ThumbnailStandardURL: r.Items[0].Snippet.Thumbnails.Standard.URL,
		ThumbnailMaxresURL:   r.Items[0].Snippet.Thumbnails.Maxres.URL,
	}, nil
}

func (t Youtube) GetVideoByID(id string) (Video, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%v/videos?part=snippet&id=%v&key=%v", t.apiURL, id, t.apiKey), nil)

	req.Header.Add("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return Video{}, err
	}
	defer resp.Body.Close()

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("twitter returned broken body response! Was: %v", string(byts))
		return Video{}, err
	}

	res := response{}
	if err := json.Unmarshal(byts, &res); err != nil {
		err2 := fmt.Errorf("twitter returned invalid JSON response! Response was: %v. Error is: %v", string(byts), err)
		return Video{}, err2
	}
	log.Println("tweet", string(byts), "response", res)

	v, err := res.toVideo()
	if err != nil {
		return v, err
	}

	return v, nil
}
