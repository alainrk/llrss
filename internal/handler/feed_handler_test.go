package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"llrss/internal/models"
	"llrss/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockService struct {
	feeds map[string]*models.Feed
}

func newMockService() *mockService {
	return &mockService{
		feeds: make(map[string]*models.Feed),
	}
}

func (m *mockService) FetchFeed(ctx context.Context, url string) (*models.Feed, error) {
	feed := &models.Feed{
		URL:   url,
		Title: "Test Feed",
		Items: []models.Item{{Title: "Test Item"}},
	}
	return feed, nil
}

func (m *mockService) GetFeed(ctx context.Context, id string) (*models.Feed, error) {
	feed, ok := m.feeds[id]
	if !ok {
		return nil, repository.ErrFeedNotFound
	}
	return feed, nil
}

func (m *mockService) ListFeeds(ctx context.Context) ([]models.Feed, error) {
	feeds := make([]models.Feed, 0, len(m.feeds))
	for _, feed := range m.feeds {
		feeds = append(feeds, *feed)
	}
	return feeds, nil
}

func (m *mockService) AddFeed(ctx context.Context, url string) (*models.Feed, error) {
	feed := &models.Feed{
		ID:    "test-id",
		URL:   url,
		Title: "Test Feed",
	}
	m.feeds[feed.ID] = feed
	return feed, nil
}

func (m *mockService) DeleteFeed(ctx context.Context, id string) error {
	delete(m.feeds, id)
	return nil
}

func (m *mockService) UpdateFeed(ctx context.Context, feed *models.Feed) error {
	m.feeds[feed.ID] = feed
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

	req := httptest.NewRequest("GET", "/feeds", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var feeds []models.Feed
	if err := json.NewDecoder(w.Body).Decode(&feeds); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func TestAddFeed(t *testing.T) {
	r, _ := setupTestHandler()

	reqBody := map[string]string{"url": "http://example.com/feed.xml"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/feeds", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var feed models.Feed
	if err := json.NewDecoder(w.Body).Decode(&feed); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if feed.URL != "http://example.com/feed.xml" {
		t.Errorf("Expected URL %s, got %s", "http://example.com/feed.xml", feed.URL)
	}
}

func TestGetFeed(t *testing.T) {
	r, mockSvc := setupTestHandler()

	feed, _ := mockSvc.AddFeed(context.Background(), "http://example.com/feed.xml")

	req := httptest.NewRequest("GET", "/feeds/"+feed.ID, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var responseFeed models.Feed
	if err := json.NewDecoder(w.Body).Decode(&responseFeed); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if responseFeed.ID != feed.ID {
		t.Errorf("Expected feed ID %s, got %s", feed.ID, responseFeed.ID)
	}
}
