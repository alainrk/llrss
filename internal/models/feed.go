package models

import (
	"encoding/xml"
	"time"
)

// RSS represents an RSS feed
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

// Channel represents the RSS channel
type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Language    string `xml:"language"`
	PubDate     string `xml:"pubDate"`
	LastBuild   string `xml:"lastBuildDate"`
	Items       []Item `xml:"item"`
}

// Item represents an RSS item
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// Feed represents a feed in our system
type Feed struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	LastFetch   time.Time `json:"last_fetch"`
	Items       []Item    `json:"items"`
}
