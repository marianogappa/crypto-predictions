package youtube

import (
	"fmt"
	"os"
	"strconv"

	"github.com/marianogappa/predictions/request"
	"github.com/marianogappa/predictions/types"
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
	PublishedAt          types.ISO8601
	ChannelID            string
	ChannelTitle         string
	ThumbnailDefaultURL  string
	ThumbnailMediumURL   string
	ThumbnailHighURL     string
	ThumbnailStandardURL string
	ThumbnailMaxresURL   string

	err error
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
	PublishedAt  types.ISO8601             `json:"publishedAt"`
	ChannelId    string                    `json:"channelId"`
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	Thumbnails   responseSnippetThumbnails `json:"thumbnails"`
	ChannelTitle string                    `json:"channelTitle"`
}
type responseItems struct {
	Snippet responseSnippet `json:"snippet"`
}

type videosResponse struct {
	Items []responseItems `json:"items"`
}

func videosResponseToVideo(r videosResponse) (Video, error) {
	if len(r.Items) != 1 {
		return Video{}, fmt.Errorf("expected len(Items) == 1, but was %v", len(r.Items))
	}

	_, err := r.Items[0].Snippet.PublishedAt.Time()
	if err != nil {
		return Video{}, fmt.Errorf("could not parse %v into valid time, with error: %v", r.Items[0].Snippet.PublishedAt, err)
	}

	return Video{
		ChannelID:            r.Items[0].Snippet.ChannelId,
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

func parseVideoError(err error) Video {
	return Video{err: err}
}

func (t Youtube) GetVideoByID(id string) (Video, error) {
	req := request.Request[videosResponse, Video]{
		BaseUrl:       t.apiURL,
		Path:          fmt.Sprintf("videos?part=snippet&id=%v&key=%v", id, t.apiKey),
		Headers:       map[string]string{"Accept": "application/json"},
		ParseResponse: videosResponseToVideo,
		ParseError:    parseVideoError,
	}

	video := request.MakeRequest(req, false)
	if video.err != nil {
		return video, video.err
	}

	return video, nil
}

type Channel struct {
	ID                  string
	URL                 string
	PublishedAt         types.ISO8601
	Title               string
	Description         string
	ThumbnailDefaultURL string
	ThumbnailMediumURL  string
	ThumbnailHighURL    string
	SubscriberCount     int

	err error
}

func channelsResponseToChannel(r channelsResponse) (Channel, error) {
	if len(r.Items) != 1 {
		return Channel{}, fmt.Errorf("expected 1 item in youtube's API response but got %v", len(r.Items))
	}
	if _, err := r.Items[0].Snippet.PublishedAt.Time(); err != nil {
		return Channel{}, fmt.Errorf("expected valid ISO8601 timestamp from youtube's API response but got %v", r.Items[0].Snippet.PublishedAt)
	}
	subscriberCount, err := strconv.Atoi(r.Items[0].Statistics.SubscriberCount)
	if err != nil {
		return Channel{}, fmt.Errorf("expected a numeric subscriber count but got %v", r.Items[0].Statistics.SubscriberCount)
	}

	return Channel{
		ID:                  r.Items[0].ID,
		URL:                 fmt.Sprintf("https://youtube.com/channel/%v", r.Items[0].ID),
		PublishedAt:         r.Items[0].Snippet.PublishedAt,
		Title:               r.Items[0].Snippet.Title,
		Description:         r.Items[0].Snippet.Description,
		ThumbnailDefaultURL: r.Items[0].Snippet.Thumbnails.Default.URL,
		ThumbnailMediumURL:  r.Items[0].Snippet.Thumbnails.Medium.URL,
		ThumbnailHighURL:    r.Items[0].Snippet.Thumbnails.High.URL,
		SubscriberCount:     subscriberCount,
	}, nil
}

type channelsResponseItem struct {
	ID         string                     `json:"id"`
	Snippet    channelsResponseSnippet    `json:"snippet"`
	Statistics channelsResponseStatistics `json:"statistics"`
}

type channelsResponseStatistics struct {
	SubscriberCount string `json:"subscriberCount"`
}

type channelsResponseSnippet struct {
	Title       string                            `json:"title"`
	Description string                            `json:"description"`
	PublishedAt types.ISO8601                     `json:"publishedAt"`
	Thumbnails  channelsResponseSnippetThumbnails `json:"thumbnails"`
}

type channelsResponseSnippetThumbnails struct {
	Default responseSnippetThumbnailsThumbnail `json:"default"`
	Medium  responseSnippetThumbnailsThumbnail `json:"medium"`
	High    responseSnippetThumbnailsThumbnail `json:"high"`
}

type channelsResponse struct {
	Items []channelsResponseItem `json:"items"`
}

func parseChannelError(err error) Channel {
	return Channel{err: err}
}

func (t Youtube) GetChannelByID(id string) (Channel, error) {
	req := request.Request[channelsResponse, Channel]{
		BaseUrl:       t.apiURL,
		Path:          fmt.Sprintf("channels?part=snippet,statistics&id=%v&key=%v", id, t.apiKey),
		Headers:       map[string]string{"Accept": "application/json"},
		ParseResponse: channelsResponseToChannel,
		ParseError:    parseChannelError,
	}

	channel := request.MakeRequest(req, false)
	if channel.err != nil {
		return channel, channel.err
	}

	return channel, nil
}
