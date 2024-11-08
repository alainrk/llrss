package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"llrss/internal/models"
	"llrss/internal/repository"
	"net/http"
	"time"
)

type FeedService interface {
	FetchFeed(ctx context.Context, url string) (*models.Feed, error)
	GetFeed(ctx context.Context, id string) (*models.Feed, error)
	GetFeedByURL(ctx context.Context, url string) (*models.Feed, error)
	ListFeeds(ctx context.Context) ([]models.Feed, error)
	AddFeed(ctx context.Context, url string) (string, error)
	DeleteFeed(ctx context.Context, id string) error
	UpdateFeed(ctx context.Context, feed *models.Feed) error
}

type feedService struct {
	repo   repository.FeedRepository
	client *http.Client
}

func NewFeedService(repo repository.FeedRepository) FeedService {
	return &feedService{
		repo: repo,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *feedService) FetchFeed(ctx context.Context, url string) (*models.Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var rss models.RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("parse RSS: %w", err)
	}

	feed := &models.Feed{
		URL:         url,
		Title:       rss.Channel.Title,
		Description: rss.Channel.Description,
		LastFetch:   time.Now(),
		Items:       rss.Channel.Items,
	}

	return feed, nil
}

func (s *feedService) GetFeed(ctx context.Context, id string) (*models.Feed, error) {
	return s.repo.GetFeed(ctx, id)
}

func (s *feedService) GetFeedByURL(ctx context.Context, url string) (*models.Feed, error) {
	return s.repo.GetFeedByURL(ctx, url)
}

func (s *feedService) ListFeeds(ctx context.Context) ([]models.Feed, error) {
	return s.repo.ListFeeds(ctx)
}

// AddFeed adds a new feed by url and returns its ID.
func (s *feedService) AddFeed(ctx context.Context, url string) (string, error) {
	f, err := s.repo.GetFeedByURL(ctx, url)
	if err != nil {
		return "", fmt.Errorf("get feed by URL: %w", err)
	}
	if f != nil {
		return f.ID, nil
	}

	feed, err := s.FetchFeed(ctx, url)
	if err != nil {
		return "", fmt.Errorf("fetch feed: %w", err)
	}

	if err := s.repo.SaveFeed(ctx, feed); err != nil {
		return "", fmt.Errorf("save feed: %w", err)
	}

	return feed.ID, nil
}

func (s *feedService) DeleteFeed(ctx context.Context, id string) error {
	return s.repo.DeleteFeed(ctx, id)
}

func (s *feedService) UpdateFeed(ctx context.Context, feed *models.Feed) error {
	return s.repo.UpdateFeed(ctx, feed)
}
