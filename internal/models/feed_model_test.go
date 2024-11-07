package models

import (
	"encoding/xml"
	"testing"
)

// mockRSSFeed is a sample RSS feed for testing.
const mockRSSFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Feed</title>
		<link>http://example.com</link>
		<description>Test RSS Feed</description>
		<language>en-us</language>
		<pubDate>Tue, 05 Nov 2024 12:00:00 GMT</pubDate>
		<lastBuildDate>Tue, 05 Nov 2024 12:00:00 GMT</lastBuildDate>
		<item>
			<title>Test Item 1</title>
			<link>http://example.com/item1</link>
			<description>First test item</description>
			<pubDate>Tue, 05 Nov 2024 11:00:00 GMT</pubDate>
			<guid>http://example.com/item1</guid>
		</item>
	</channel>
</rss>`

func TestRSSParsing(t *testing.T) {
	var rss RSS
	err := xml.Unmarshal([]byte(mockRSSFeed), &rss)
	if err != nil {
		t.Fatalf("Failed to unmarshal RSS: %v", err)
	}

	// Test channel metadata
	if rss.Channel.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", rss.Channel.Title)
	}

	if rss.Channel.Link != "http://example.com" {
		t.Errorf("Expected link 'http://example.com', got '%s'", rss.Channel.Link)
	}

	if rss.Channel.Language != "en-us" {
		t.Errorf("Expected language 'en-us', got '%s'", rss.Channel.Language)
	}

	// Test items
	if len(rss.Channel.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(rss.Channel.Items))
	}

	item := rss.Channel.Items[0]
	if item.Title != "Test Item 1" {
		t.Errorf("Expected item title 'Test Item 1', got '%s'", item.Title)
	}
}
