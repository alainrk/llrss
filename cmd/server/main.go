package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RSS represents the root RSS feed structure
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

// Channel represents the main content of the RSS feed
type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Language    string `xml:"language"`
	PubDate     string `xml:"pubDate"`
	LastBuild   string `xml:"lastBuildDate"`
	Items       []Item `xml:"item"`
}

// Item represents an individual RSS feed entry
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// FeedReader handles RSS feed operations
type FeedReader struct {
	client *http.Client
}

// NewFeedReader creates a new FeedReader instance
func NewFeedReader() *FeedReader {
	return &FeedReader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchFeed retrieves and parses an RSS feed from a given URL
func (fr *FeedReader) FetchFeed(url string) (*RSS, error) {
	resp, err := fr.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching feed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("error parsing RSS feed: %v", err)
	}

	return &rss, nil
}

func main() {
	reader := NewFeedReader()

	// Example usage with a sample RSS feed URL
	feedURL := "https://technicalwriting.dev/rss.xml"
	feed, err := reader.FetchFeed(feedURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print feed information
	fmt.Printf("Feed Title: %s\n", feed.Channel.Title)
	fmt.Printf("Feed Description: %s\n", feed.Channel.Description)
	fmt.Printf("\nLatest Items:\n")

	// Print the latest items
	for i, item := range feed.Channel.Items {
		fmt.Printf("\n%d. %s\n", i+1, item.Title)
		fmt.Printf("   Link: %s\n", item.Link)
		fmt.Printf("   Published: %s\n", item.PubDate)
		fmt.Printf("   Description: %s\n", item.Description)
	}
}
