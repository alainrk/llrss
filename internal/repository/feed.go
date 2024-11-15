package repository

import (
	"context"
	"llrss/internal/models"
)

type FeedRepository interface {
	GetFeed(ctx context.Context, id string) (*models.Feed, error)
	GetFeedByURL(ctx context.Context, url string) (*models.Feed, error)
	ListFeeds(ctx context.Context) ([]models.Feed, error)
	SaveFeed(ctx context.Context, feed *models.Feed) (string, error)
	DeleteFeed(ctx context.Context, id string) error
	UpdateFeed(ctx context.Context, feed *models.Feed) error

	GetFeedItem(ctx context.Context, id string) (*models.Item, error)
	UpdateFeedItem(ctx context.Context, s *models.Item) error
	Nuke(ctx context.Context) error
}
