package twitter

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// oEmbedResponse represents the response from Twitter's oEmbed API
type oEmbedResponse struct {
	AuthorName string `json:"author_name"`
	AuthorURL  string `json:"author_url"`
	HTML       string `json:"html"`
	URL        string `json:"url"`
}

// scrapeTweetData fetches tweet data using Twitter's oEmbed API
// oembedBaseURL can be set for testing, defaults to "https://publish.twitter.com"
func scrapeTweetData(tweetURL string, oembedBaseURL string) (tweetData, error) {
	if oembedBaseURL == "" {
		oembedBaseURL = "https://publish.twitter.com"
	}
	oembedURL := fmt.Sprintf("%s/oembed?url=%s", oembedBaseURL, url.QueryEscape(tweetURL))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(oembedURL)
	if err != nil {
		return tweetData{}, fmt.Errorf("error fetching oEmbed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tweetData{}, fmt.Errorf("oEmbed API returned status %d", resp.StatusCode)
	}

	var oembed oEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&oembed); err != nil {
		return tweetData{}, fmt.Errorf("error parsing oEmbed JSON: %w", err)
	}

	// Extract tweet ID from URL
	tweetID := extractTweetIDFromURL(tweetURL)
	if tweetID == "" {
		return tweetData{}, fmt.Errorf("could not extract tweet ID from URL: %s", tweetURL)
	}

	// Extract handle from author URL
	handle := extractHandleFromURL(oembed.AuthorURL)
	if handle == "" {
		return tweetData{}, fmt.Errorf("could not extract handle from author URL: %s", oembed.AuthorURL)
	}

	// Extract tweet text from HTML
	tweetText := extractTweetTextFromHTML(oembed.HTML)

	// Extract date from HTML
	tweetCreatedAt, err := extractDateFromHTML(oembed.HTML)
	if err != nil {
		return tweetData{}, fmt.Errorf("error parsing date: %w", err)
	}

	return tweetData{
		TweetText:      tweetText,
		TweetID:        tweetID,
		TweetCreatedAt: tweetCreatedAt,
		UserName:       oembed.AuthorName,
		UserHandle:     handle,
		// These fields are not available via oEmbed:
		ProfileImgURL:  "",
		UserCreatedAt:  time.Time{},
		FollowersCount: 0,
		Verified:       false,
	}, nil
}

// tweetData holds the scraped tweet data
type tweetData struct {
	TweetText      string
	TweetID        string
	TweetCreatedAt time.Time
	UserName       string
	UserHandle     string
	ProfileImgURL  string
	UserCreatedAt  time.Time
	FollowersCount int
	Verified       bool
}

// extractTweetIDFromURL extracts the tweet ID from a Twitter/X URL
func extractTweetIDFromURL(tweetURL string) string {
	// Pattern: https://x.com/username/status/1234567890
	// Pattern: https://twitter.com/username/status/1234567890
	re := regexp.MustCompile(`/(?:status|statuses)/(\d+)`)
	matches := re.FindStringSubmatch(tweetURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractHandleFromURL extracts the handle from a Twitter/X user URL
func extractHandleFromURL(authorURL string) string {
	// Pattern: https://twitter.com/username or https://x.com/username
	parts := strings.Split(strings.TrimSuffix(authorURL, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// extractTweetTextFromHTML extracts the tweet text from oEmbed HTML
func extractTweetTextFromHTML(htmlContent string) string {
	// Extract text from <p> tags in the HTML
	re := regexp.MustCompile(`<p[^>]*>(.*?)</p>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)
	if len(matches) > 0 {
		text := matches[0][1]
		// Remove HTML tags
		text = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, "")
		// Decode HTML entities
		text = html.UnescapeString(text)
		// Clean up whitespace
		text = strings.TrimSpace(text)
		// Replace multiple spaces with single space
		text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
		return text
	}
	return ""
}

// extractDateFromHTML extracts and parses the date from oEmbed HTML
func extractDateFromHTML(htmlContent string) (time.Time, error) {
	// Comprehensive list of date formats to try
	parseFormats := []string{
		// Full month name formats
		"January 2, 2006",
		"January 02, 2006",
		"2 January 2006",
		"02 January 2006",
		// Abbreviated month formats
		"Jan 2, 2006",
		"Jan 02, 2006",
		"Jan. 2, 2006",
		"Jan. 02, 2006",
		"2 Jan 2006",
		"02 Jan 2006",
		"2 Jan. 2006",
		"02 Jan. 2006",
		// ISO format
		"2006-01-02",
		"2006-1-2",
		// With time (12-hour)
		"January 2, 2006 at 3:04 PM",
		"January 2, 2006 at 03:04 PM",
		"January 02, 2006 at 3:04 PM",
		"January 02, 2006 at 03:04 PM",
		"Jan 2, 2006 at 3:04 PM",
		"Jan 2, 2006 at 03:04 PM",
		"Jan. 2, 2006 at 3:04 PM",
		"Jan. 2, 2006 at 03:04 PM",
		"January 2, 2006 at 3:04 AM",
		"January 2, 2006 at 03:04 AM",
		"January 02, 2006 at 3:04 AM",
		"January 02, 2006 at 03:04 AM",
		"Jan 2, 2006 at 3:04 AM",
		"Jan 2, 2006 at 03:04 AM",
		"Jan. 2, 2006 at 3:04 AM",
		"Jan. 2, 2006 at 03:04 AM",
		// With time (24-hour)
		"January 2, 2006 at 15:04",
		"January 02, 2006 at 15:04",
		"Jan 2, 2006 at 15:04",
		"Jan. 2, 2006 at 15:04",
		// RFC822-like formats
		"2 Jan 06 15:04 MST",
		"02 Jan 06 15:04 MST",
		"2 Jan 2006 15:04 MST",
		"02 Jan 2006 15:04 MST",
		// Other common formats
		"2006/01/02",
		"2006/1/2",
		"01/02/2006",
		"1/2/2006",
		"02-01-2006",
		"2-1-2006",
	}

	// Try to find date patterns in the HTML and parse them
	datePatterns := []struct {
		pattern string
		formats []string
	}{
		{
			// Full month name: "December 5, 2025"
			pattern: `(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{1,2}),\s+(\d{4})`,
			formats: []string{"January 2, 2006", "January 02, 2006"},
		},
		{
			// Abbreviated month: "Dec 5, 2025" or "Dec. 5, 2025"
			pattern: `(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\.?\s+(\d{1,2}),\s+(\d{4})`,
			formats: []string{"Jan 2, 2006", "Jan 02, 2006", "Jan. 2, 2006", "Jan. 02, 2006"},
		},
		{
			// Day first: "5 December 2025"
			pattern: `(\d{1,2})\s+(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{4})`,
			formats: []string{"2 January 2006", "02 January 2006"},
		},
		{
			// ISO format: "2025-12-05"
			pattern: `(\d{4})-(\d{1,2})-(\d{1,2})`,
			formats: []string{"2006-01-02", "2006-1-2"},
		},
		{
			// With time AM/PM: "December 5, 2025 at 10:30 AM"
			pattern: `(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{1,2}),\s+(\d{4})\s+at\s+(\d{1,2}):(\d{2})\s+(AM|PM)`,
			formats: []string{"January 2, 2006 at 3:04 PM", "January 2, 2006 at 03:04 PM", "January 02, 2006 at 3:04 PM", "January 02, 2006 at 03:04 PM"},
		},
		{
			// Abbreviated with time: "Dec 5, 2025 at 10:30 AM"
			pattern: `(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\.?\s+(\d{1,2}),\s+(\d{4})\s+at\s+(\d{1,2}):(\d{2})\s+(AM|PM)`,
			formats: []string{"Jan 2, 2006 at 3:04 PM", "Jan 2, 2006 at 03:04 PM", "Jan. 2, 2006 at 3:04 PM", "Jan. 2, 2006 at 03:04 PM"},
		},
	}

	// Try each pattern
	for _, dp := range datePatterns {
		re := regexp.MustCompile(dp.pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 0 {
			dateStr := matches[0]
			// Try parsing with the specific formats for this pattern
			for _, format := range dp.formats {
				if t, err := time.Parse(format, dateStr); err == nil {
					return t, nil
				}
			}
			// If specific formats don't work, try all formats
			for _, format := range parseFormats {
				if t, err := time.Parse(format, dateStr); err == nil {
					return t, nil
				}
			}
		}
	}

	// Fallback: try to find any "Month Day, Year" pattern and parse it
	fallbackPattern := regexp.MustCompile(`([A-Z][a-z]+)\s+(\d{1,2}),\s+(\d{4})`)
	fallbackMatch := fallbackPattern.FindStringSubmatch(htmlContent)
	if len(fallbackMatch) > 0 {
		dateStr := fallbackMatch[0]
		// Try common formats
		for _, format := range []string{
			"January 2, 2006",
			"January 02, 2006",
			"Jan 2, 2006",
			"Jan 02, 2006",
			"Jan. 2, 2006",
			"Jan. 02, 2006",
		} {
			if t, err := time.Parse(format, dateStr); err == nil {
				return t, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date from HTML: %s", htmlContent[:min(200, len(htmlContent))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
