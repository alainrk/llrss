package main

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockRSSFeed is a sample RSS feed for testing
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
		<item>
			<title>Test Item 2</title>
			<link>http://example.com/item2</link>
			<description>Second test item</description>
			<pubDate>Tue, 05 Nov 2024 10:00:00 GMT</pubDate>
			<guid>http://example.com/item2</guid>
		</item>
	</channel>
</rss>`

// mockInvalidRSSFeed is an invalid RSS feed for testing error handling
const mockInvalidRSSFeed = `<?xml version="1.0" encoding="UTF-8"?>
<invalid>
	<data>This is not a valid RSS feed</data>
</invalid>`

// setupTestServer creates a test HTTP server with the provided handler
func setupTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// TestNewFeedReader tests the creation of a new FeedReader
func TestNewFeedReader(t *testing.T) {
	reader := NewFeedReader()

	if reader == nil {
		t.Fatal("NewFeedReader returned nil")
	}

	if reader.client == nil {
		t.Fatal("FeedReader client is nil")
	}

	if reader.client.Timeout != 30*time.Second {
		t.Errorf("Expected timeout of 30 seconds, got %v", reader.client.Timeout)
	}
}

// TestFetchFeed tests successful RSS feed fetching
func TestFetchFeed(t *testing.T) {
	server := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSFeed))
	}))
	defer server.Close()

	reader := NewFeedReader()
	feed, err := reader.FetchFeed(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test channel metadata
	if feed.Channel.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", feed.Channel.Title)
	}

	if feed.Channel.Link != "http://example.com" {
		t.Errorf("Expected link 'http://example.com', got '%s'", feed.Channel.Link)
	}

	// Test items
	if len(feed.Channel.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(feed.Channel.Items))
	}

	// Test first item
	item := feed.Channel.Items[0]
	if item.Title != "Test Item 1" {
		t.Errorf("Expected item title 'Test Item 1', got '%s'", item.Title)
	}

	if item.Link != "http://example.com/item1" {
		t.Errorf("Expected item link 'http://example.com/item1', got '%s'", item.Link)
	}
}

// TestFetchFeedInvalidXML tests handling of invalid XML
func TestFetchFeedInvalidXML(t *testing.T) {
	server := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockInvalidRSSFeed))
	}))
	defer server.Close()

	reader := NewFeedReader()
	_, err := reader.FetchFeed(server.URL)

	if err == nil {
		t.Fatal("Expected error for invalid XML, got nil")
	}

	if !strings.Contains(err.Error(), "error parsing RSS feed") {
		t.Errorf("Expected parsing error, got: %v", err)
	}
}

// TestFetchFeedHTTPError tests handling of HTTP errors
func TestFetchFeedHTTPError(t *testing.T) {
	server := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	reader := NewFeedReader()
	_, err := reader.FetchFeed(server.URL)

	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}

	if !strings.Contains(err.Error(), "unexpected status code: 500") {
		t.Errorf("Expected status code error, got: %v", err)
	}
}

// TestFetchFeedInvalidURL tests handling of invalid URLs
func TestFetchFeedInvalidURL(t *testing.T) {
	reader := NewFeedReader()
	_, err := reader.FetchFeed("not-a-valid-url")

	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}
}

// TestFetchFeedTimeout tests handling of timeout scenarios
func TestFetchFeedTimeout(t *testing.T) {
	server := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 2) // Delay response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSFeed))
	}))
	defer server.Close()

	reader := NewFeedReader()
	reader.client.Timeout = time.Millisecond * 100 // Set very short timeout

	_, err := reader.FetchFeed(server.URL)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

// TestRSSStructure tests the RSS struct mappings
func TestRSSStructure(t *testing.T) {
	var rss RSS
	err := xml.Unmarshal([]byte(mockRSSFeed), &rss)
	if err != nil {
		t.Fatalf("Failed to unmarshal RSS: %v", err)
	}

	// Test channel structure
	if rss.Channel.Language != "en-us" {
		t.Errorf("Expected language 'en-us', got '%s'", rss.Channel.Language)
	}

	if rss.Channel.PubDate != "Tue, 05 Nov 2024 12:00:00 GMT" {
		t.Errorf("Expected pubDate 'Tue, 05 Nov 2024 12:00:00 GMT', got '%s'", rss.Channel.PubDate)
	}

	// Test item structure
	if len(rss.Channel.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(rss.Channel.Items))
	}

	item := rss.Channel.Items[1] // Test second item
	expectedTitle := "Test Item 2"
	if item.Title != expectedTitle {
		t.Errorf("Expected item title '%s', got '%s'", expectedTitle, item.Title)
	}
}
