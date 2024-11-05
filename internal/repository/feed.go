package repository

import (
	"context"
	"fmt"
	"llrss/internal/models"

	"github.com/google/uuid"
)

type FeedRepository interface {
	GetFeed(ctx context.Context, id string) (*models.Feed, error)
	ListFeeds(ctx context.Context) ([]models.Feed, error)
	SaveFeed(ctx context.Context, feed *models.Feed) error
	DeleteFeed(ctx context.Context, id string) error
	UpdateFeed(ctx context.Context, feed *models.Feed) error
}

// Example memory repository implementation
type memoryFeedRepository struct {
	feeds map[string]*models.Feed
}

func NewMemoryFeedRepository() FeedRepository {
	return &memoryFeedRepository{
		feeds: make(map[string]*models.Feed),
	}
}

func (r *memoryFeedRepository) GetFeed(ctx context.Context, id string) (*models.Feed, error) {
	feed, ok := r.feeds[id]
	if !ok {
		return nil, fmt.Errorf("feed not found: %s", id)
	}
	return feed, nil
}

func (r *memoryFeedRepository) ListFeeds(ctx context.Context) ([]models.Feed, error) {
	feeds := make([]models.Feed, 0, len(r.feeds))
	for _, feed := range r.feeds {
		feeds = append(feeds, *feed)
	}
	return feeds, nil
}

func (r *memoryFeedRepository) SaveFeed(ctx context.Context, feed *models.Feed) error {
	if feed.ID == "" {
		feed.ID = uuid.New().String()
	}
	r.feeds[feed.ID] = feed
	return nil
}

func (r *memoryFeedRepository) DeleteFeed(ctx context.Context, id string) error {
	delete(r.feeds, id)
	return nil
}

func (r *memoryFeedRepository) UpdateFeed(ctx context.Context, feed *models.Feed) error {
	r.feeds[feed.ID] = feed
	return nil
}
