package rss

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// Create an HTTP client and send a GET request to the feed url
	client := &http.Client{}
	// Create the new request
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	// Set the User-Agent header to "gator" (Common practice)
	request.Header.Set("User-Agent", "gator")

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // Remember to close response body
	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// Unmarshal the response body into an RSSFeed struct
	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}
	// Sanatize the feed data (Mainly title and description)
	feed.Sanitize()

	return &feed, nil
}

func sanitizeHTML(input string) string {
	return html.EscapeString(input)
}

func (item *RSSItem) Sanitize() {
	item.Title = sanitizeHTML(item.Title)
	item.Link = sanitizeHTML(item.Link)
	item.Description = sanitizeHTML(item.Description)
}

func (feed *RSSFeed) Sanitize() {
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Sanitize()
	}
}
