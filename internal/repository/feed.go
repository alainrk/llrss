package repository

import (
	"context"
	"llrss/internal/models"
	"llrss/internal/models/db"
)

type FeedRepository interface {
	GetFeed(ctx context.Context, id string) (*db.Feed, error)
	GetFeedByURL(ctx context.Context, url string) (*db.Feed, error)
	GetFeedItem(ctx context.Context, id string) (*db.Item, error)
	ListFeeds(ctx context.Context, userId uint64) ([]db.Feed, error)
	SearchFeedItems(ctx context.Context, userId uint64, items models.SearchParams) ([]db.Item, int64, error)

	SaveFeed(ctx context.Context, feed *db.Feed) (string, error)
	SaveFeedItems(ctx context.Context, feedID string, items []db.Item) error

	UpdateFeed(ctx context.Context, feed *db.Feed) error
	UpdateFeedItem(ctx context.Context, s *db.Item) error

	DeleteFeed(ctx context.Context, userId uint64, id string) error
	Nuke(ctx context.Context) error
}
