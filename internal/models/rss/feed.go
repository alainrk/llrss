package rss

import (
	"encoding/xml"
	"time"
)

type RSS struct {
	XMLName          xml.Name `xml:"rss"`
	Version          string   `xml:"version,attr"`
	ContentNamespace string   `xml:"xmlns:content,attr"`
	Channel          Channel  `xml:"channel"`
}

type Channel struct {
	TextInput      *TextInput
	Image          *Image
	XMLName        xml.Name `xml:"channel"`
	Category       string   `xml:"category,omitempty"`
	Docs           string   `xml:"docs,omitempty"`
	Copyright      string   `xml:"copyright,omitempty"`
	ManagingEditor string   `xml:"managingEditor,omitempty"`
	WebMaster      string   `xml:"webMaster,omitempty"`
	PubDate        string   `xml:"pubDate,omitempty"`
	LastBuildDate  string   `xml:"lastBuildDate,omitempty"`
	Description    string   `xml:"description"`
	Generator      string   `xml:"generator,omitempty"`
	Language       string   `xml:"language,omitempty"`
	Cloud          string   `xml:"cloud,omitempty"`
	Title          string   `xml:"title"`
	Rating         string   `xml:"rating,omitempty"`
	SkipHours      string   `xml:"skipHours,omitempty"`
	SkipDays       string   `xml:"skipDays,omitempty"`
	Link           string   `xml:"link"`
	Items          []Item   `xml:"item"`
	TTL            int      `xml:"ttl,omitempty"`
}

type Image struct {
	XMLName xml.Name `xml:"image"`
	URL     string   `xml:"url"`
	Title   string   `xml:"title"`
	Link    string   `xml:"link"`
	Width   int      `xml:"width,omitempty"`
	Height  int      `xml:"height,omitempty"`
}

type TextInput struct {
	XMLName     xml.Name `xml:"textInput"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Name        string   `xml:"name"`
	Link        string   `xml:"link"`
}

type Content struct {
	XMLName xml.Name `xml:"content:encoded"`
	Content string   `xml:",cdata"`
}

type Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Content     *Content
	Author      string `xml:"author,omitempty"`
	Category    string `xml:"category,omitempty"`
	Comments    string `xml:"comments,omitempty"`
	Enclosure   *Enclosure
	GUID        *GUID
	PubDate     string `xml:"pubDate,omitempty"`
	Source      string `xml:"source,omitempty"`
	FeedID      string
	IsRead      bool
}

type Feed struct {
	ID          string
	URL         string
	Title       string
	Description string `xml:"description"`
	LastFetch   time.Time
	Items       []Item
}

type Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
	Length  string   `xml:"length,attr"`
	Type    string   `xml:"type,attr"`
}

type GUID struct {
	XMLName     xml.Name `xml:"guid"`
	ID          string   `xml:",chardata"`
	IsPermaLink string   `xml:"isPermaLink,attr,omitempty"`
}
