package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"llrss/internal/models"
	"llrss/internal/models/db"
	"llrss/internal/repository"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockService struct {
	feeds map[string]*db.Feed
}

func newMockService() *mockService {
	return &mockService{
		feeds: make(map[string]*db.Feed),
	}
}

func (m *mockService) FetchFeed(ctx context.Context, url string) (*db.Feed, error) {
	feed := &db.Feed{
		URL:   url,
		Title: "Test Feed",
		Items: []db.Item{{Title: "Test Item"}},
	}
	return feed, nil
}

func (m *mockService) GetFeed(ctx context.Context, id string) (*db.Feed, error) {
	feed, ok := m.feeds[id]
	if !ok {
		return nil, repository.ErrFeedNotFound
	}
	return feed, nil
}

func (m *mockService) GetFeedByURL(ctx context.Context, url string) (*db.Feed, error) {
	feed, ok := m.feeds[url]
	if !ok {
		return nil, repository.ErrFeedNotFound
	}
	return feed, nil
}

func (m *mockService) ListFeeds(ctx context.Context) ([]db.Feed, error) {
	feeds := make([]db.Feed, 0, len(m.feeds))
	for _, feed := range m.feeds {
		feeds = append(feeds, *feed)
	}
	return feeds, nil
}

func (m *mockService) AddFeed(ctx context.Context, url string) (string, error) {
	feed := &db.Feed{
		ID:    "test-id",
		URL:   url,
		Title: "Test Feed",
	}
	m.feeds[feed.ID] = feed
	return feed.ID, nil
}

func (m *mockService) DeleteFeed(ctx context.Context, id string) error {
	delete(m.feeds, id)
	return nil
}

func (m *mockService) UpdateFeed(ctx context.Context, feed *db.Feed) error {
	m.feeds[feed.ID] = feed
	return nil
}

func (m *mockService) MarkFeedItemRead(ctx context.Context, feedItemID string, read bool) error {
	// TODO Implement this
	return nil
}

func (m *mockService) SearchFeedItems(ctx context.Context, items models.SearchParams) ([]db.Item, int64, error) {
	// TODO Implement this
	return nil, 0, nil
}

func (m *mockService) RefreshFeeds(ctx context.Context) error {
	// TODO Implement this
	return nil
}

func (m *mockService) Nuke(ctx context.Context) error {
	m.feeds = make(map[string]*db.Feed)
	return nil
}

func setupTestHandler() (*chi.Mux, *mockService) {
	r := chi.NewRouter()
	mockSvc := newMockService()
	handler := NewFeedHandler(mockSvc)
	handler.RegisterRoutes(r)
	return r, mockSvc
}

func TestListFeeds(t *testing.T) {
	r, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/feeds", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var feeds []db.Feed
	if err := json.NewDecoder(w.Body).Decode(&feeds); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func TestAddFeed(t *testing.T) {
	r, _ := setupTestHandler()

	reqBody := map[string]string{"url": "http://example.com/feed.xml"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/feeds", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	ID := strings.TrimSpace(w.Body.String())

	if ID == "" {
		t.Errorf("Expected feed ID, got empty string")
	}
}

// TODO: Test already added feed URL

func TestGetFeed(t *testing.T) {
	r, mockSvc := setupTestHandler()

	ID, _ := mockSvc.AddFeed(context.Background(), "http://example.com/feed.xml")

	req := httptest.NewRequest(http.MethodGet, "/feeds/"+ID, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var responseFeed db.Feed
	if err := json.NewDecoder(w.Body).Decode(&responseFeed); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if responseFeed.ID != ID {
		t.Errorf("Expected feed ID %s, got %s", ID, responseFeed.ID)
	}
}
