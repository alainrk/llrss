package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"llrss/internal/models"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// MockFeedRepository implements FeedRepository interface for testing.
type MockFeedRepository struct {
	getFeedFunc      func(ctx context.Context, id string) (*models.Feed, error)
	getFeedByURLFunc func(ctx context.Context, url string) (*models.Feed, error)
	listFeedsFunc    func(ctx context.Context) ([]models.Feed, error)
	saveFeedFunc     func(ctx context.Context, feed *models.Feed) (string, error)
	deleteFeedFunc   func(ctx context.Context, id string) error
	updateFeedFunc   func(ctx context.Context, feed *models.Feed) error
	nukeFunc         func(ctx context.Context) error
}

func (m *MockFeedRepository) GetFeed(ctx context.Context, id string) (*models.Feed, error) {
	return m.getFeedFunc(ctx, id)
}

func (m *MockFeedRepository) GetFeedByURL(ctx context.Context, url string) (*models.Feed, error) {
	return m.getFeedByURLFunc(ctx, url)
}

func (m *MockFeedRepository) ListFeeds(ctx context.Context) ([]models.Feed, error) {
	return m.listFeedsFunc(ctx)
}

func (m *MockFeedRepository) SaveFeed(ctx context.Context, feed *models.Feed) (string, error) {
	return m.saveFeedFunc(ctx, feed)
}

func (m *MockFeedRepository) DeleteFeed(ctx context.Context, id string) error {
	return m.deleteFeedFunc(ctx, id)
}

func (m *MockFeedRepository) UpdateFeed(ctx context.Context, feed *models.Feed) error {
	return m.updateFeedFunc(ctx, feed)
}

func (m *MockFeedRepository) Nuke(ctx context.Context) error {
	return m.nukeFunc(ctx)
}

func (m *MockFeedRepository) GetFeedItem(ctx context.Context, id string) (*models.Item, error) {
	// TODO: Implement this
	return nil, nil
}

func (m *MockFeedRepository) UpdateFeedItem(ctx context.Context, s *models.Item) error {
	// TODO: Implement this
	return nil
}

// MockRoundTripper implements http.RoundTripper for testing.
type MockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestFetchFeed(t *testing.T) {
	validXML := `<?xml version="1.0" encoding="UTF-8"?>
		<rss version="2.0">
			<channel>
				<title>Test Feed</title>
				<description>Test Description</description>
				<item>
					<title>Test Item</title>
					<link>http://example.com</link>
				</item>
			</channel>
		</rss>`

	tests := []struct {
		mockError     error
		mockResponse  *http.Response
		expectedFeed  *models.Feed
		name          string
		url           string
		expectedError bool
	}{
		{
			name: "successful fetch",
			url:  "http://example.com/feed",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(validXML)),
			},
			mockError:     nil,
			expectedError: false,
			expectedFeed: &models.Feed{
				URL:         "http://example.com/feed",
				Title:       "Test Feed",
				Description: "Test Description",
			},
		},
		{
			name:          "http client error",
			url:           "http://example.com/feed",
			mockResponse:  nil,
			mockError:     errors.New("connection error"),
			expectedError: true,
			expectedFeed:  nil,
		},
		{
			name: "invalid status code",
			url:  "http://example.com/feed",
			mockResponse: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			mockError:     nil,
			expectedError: true,
			expectedFeed:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTripper := &MockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			service := &feedService{
				client: &http.Client{
					Transport: mockTripper,
				},
			}

			feed, err := service.FetchFeed(context.Background(), tt.url)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if feed != nil {
					t.Error("expected nil feed but got:", feed)
				}
				return
			}

			if err != nil {
				t.Error("unexpected error:", err)
			}

			if feed == nil {
				t.Fatal("expected feed but got nil")
			}

			if feed.Title != tt.expectedFeed.Title {
				t.Errorf("expected title %q, got %q", tt.expectedFeed.Title, feed.Title)
			}

			if feed.Description != tt.expectedFeed.Description {
				t.Errorf("expected description %q, got %q", tt.expectedFeed.Description, feed.Description)
			}

			if feed.URL != tt.expectedFeed.URL {
				t.Errorf("expected URL %q, got %q", tt.expectedFeed.URL, feed.URL)
			}
		})
	}
}

func TestGetFeed(t *testing.T) {
	ctx := context.Background()
	expectedFeed := &models.Feed{
		ID:    "1",
		Title: "Test Feed",
	}

	tests := []struct {
		mockError     error
		mockFeed      *models.Feed
		name          string
		id            string
		expectedError bool
	}{
		{
			name:          "successful get",
			id:            "1",
			mockFeed:      expectedFeed,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "not found",
			id:            "2",
			mockFeed:      nil,
			mockError:     errors.New("not found"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFeedRepository{
				getFeedFunc: func(ctx context.Context, id string) (*models.Feed, error) {
					if id != tt.id {
						t.Errorf("expected id %q, got %q", tt.id, id)
					}
					return tt.mockFeed, tt.mockError
				},
			}

			service := NewFeedService(mockRepo)
			feed, err := service.GetFeed(ctx, tt.id)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if feed != nil {
					t.Error("expected nil feed but got:", feed)
				}
				return
			}

			if err != nil {
				t.Error("unexpected error:", err)
			}

			if !reflect.DeepEqual(feed, tt.mockFeed) {
				t.Errorf("expected feed %+v, got %+v", tt.mockFeed, feed)
			}
		})
	}
}

func TestListFeeds(t *testing.T) {
	ctx := context.Background()
	expectedFeeds := []models.Feed{
		{ID: "1", Title: "Feed 1"},
		{ID: "2", Title: "Feed 2"},
	}

	tests := []struct {
		mockError     error
		name          string
		mockFeeds     []models.Feed
		expectedError bool
	}{
		{
			name:          "successful list",
			mockFeeds:     expectedFeeds,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "empty list",
			mockFeeds:     []models.Feed{},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "error",
			mockFeeds:     nil,
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFeedRepository{
				listFeedsFunc: func(ctx context.Context) ([]models.Feed, error) {
					return tt.mockFeeds, tt.mockError
				},
			}

			service := NewFeedService(mockRepo)
			feeds, err := service.ListFeeds(ctx)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Error("unexpected error:", err)
			}

			if !reflect.DeepEqual(feeds, tt.mockFeeds) {
				t.Errorf("expected feeds %+v, got %+v", tt.mockFeeds, feeds)
			}
		})
	}
}

func TestAddFeed(t *testing.T) {
	ctx := context.Background()
	validXML := `<?xml version="1.0" encoding="UTF-8"?>
		<rss version="2.0">
			<channel>
				<title>Test Feed</title>
				<description>Test Description</description>
			</channel>
		</rss>`

	tests := []struct {
		mockError     error
		mockResponse  *http.Response
		existingFeed  *models.Feed
		name          string
		url           string
		expectedID    string
		expectedError bool
	}{
		{
			name: "successful new feed",
			url:  "http://example.com/feed",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(validXML)),
			},
			mockError:     nil,
			existingFeed:  nil,
			expectedError: false,
			expectedID:    "new-id",
		},
		{
			name: "url already exists",
			url:  "http://example.com/feed",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(validXML)),
			},
			existingFeed: &models.Feed{
				ID:  "existing-id",
				URL: "http://example.com/feed",
			},
			mockError:     nil,
			expectedError: false,
			expectedID:    "existing-id",
		},
		{
			name:          "fetch error",
			url:           "http://example.com/feed",
			mockResponse:  nil,
			mockError:     errors.New("fetch error"),
			existingFeed:  nil,
			expectedError: true,
			expectedID:    "",
		},
		{
			name: "invalid status code",
			url:  "http://example.com/feed",
			mockResponse: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			mockError:     nil,
			existingFeed:  nil,
			expectedError: true,
			expectedID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTripper := &MockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			mockRepo := &MockFeedRepository{
				getFeedByURLFunc: func(ctx context.Context, url string) (*models.Feed, error) {
					if tt.existingFeed != nil && tt.existingFeed.URL == url {
						return tt.existingFeed, nil
					}
					return nil, fmt.Errorf("not found")
				},
				saveFeedFunc: func(ctx context.Context, feed *models.Feed) (string, error) {
					if tt.existingFeed != nil {
						t.Error("save called when feed already exists")
						return "", fmt.Errorf("feed already exists")
					}
					feed.ID = "new-id" // Simulate ID generation
					return "new-id", nil
				},
			}

			service := &feedService{
				repo: mockRepo,
				client: &http.Client{
					Transport: mockTripper,
				},
			}

			id, err := service.AddFeed(ctx, tt.url)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if id != "" {
					t.Error("expected empty ID but got:", id)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if id != tt.expectedID {
				t.Errorf("expected ID %q, got %q", tt.expectedID, id)
			}
		})
	}
}

func TestDeleteFeed(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		mockError     error
		name          string
		id            string
		expectedError bool
	}{
		{
			name:          "successful delete",
			id:            "1",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "delete error",
			id:            "2",
			mockError:     errors.New("delete error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFeedRepository{
				deleteFeedFunc: func(ctx context.Context, id string) error {
					if id != tt.id {
						t.Errorf("expected id %q, got %q", tt.id, id)
					}
					return tt.mockError
				},
			}

			service := NewFeedService(mockRepo)
			err := service.DeleteFeed(ctx, tt.id)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Error("unexpected error:", err)
			}
		})
	}
}

func TestUpdateFeed(t *testing.T) {
	ctx := context.Background()
	feed := &models.Feed{
		ID:    "1",
		Title: "Updated Feed",
	}

	tests := []struct {
		mockError     error
		feed          *models.Feed
		name          string
		expectedError bool
	}{
		{
			name:          "successful update",
			feed:          feed,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "update error",
			feed:          feed,
			mockError:     errors.New("update error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFeedRepository{
				updateFeedFunc: func(ctx context.Context, feed *models.Feed) error {
					if !reflect.DeepEqual(feed, tt.feed) {
						t.Errorf("expected feed %+v, got %+v", tt.feed, feed)
					}
					return tt.mockError
				},
			}

			service := NewFeedService(mockRepo)
			err := service.UpdateFeed(ctx, tt.feed)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Error("unexpected error:", err)
			}
		})
	}
}

func TestNuke(t *testing.T) {
	ctx := context.Background()
	feeds := []models.Feed{
		{ID: "1", Title: "Feed 1"},
		{ID: "2", Title: "Feed 2"},
	}
	// items := []models.Item{
	// 	{Title: "Item 1", FeedID: "1"},
	// 	{Title: "Item 2", FeedID: "1"},
	// 	{Title: "Item 3", FeedID: "2"},
	// }

	mockRepo := &MockFeedRepository{
		listFeedsFunc: func(ctx context.Context) ([]models.Feed, error) {
			return feeds, nil
		},
		nukeFunc: func(ctx context.Context) error {
			feeds = []models.Feed{}
			// items = []models.Item{}
			return nil
		},
	}

	service := NewFeedService(mockRepo)

	err := service.Nuke(ctx)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	feedsRes, err := service.ListFeeds(ctx)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	if len(feedsRes) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feedsRes))
	}

	// TODO: Implement test for single items
	// itemsRes, err := service.ListItems(ctx)
	// if err != nil {
	// 	t.Error("unexpected error:", err)
	// }
}
