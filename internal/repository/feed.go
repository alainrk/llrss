package repository

import (
	"context"
	"llrss/internal/models"
	"llrss/internal/models/db"
)

type FeedRepository interface {
	GetFeed(ctx context.Context, id string) (*db.Feed, error)
	GetFeedByURL(ctx context.Context, url string) (*db.Feed, error)
	ListFeeds(ctx context.Context) ([]db.Feed, error)
	SaveFeed(ctx context.Context, feed *db.Feed) (string, error)
	DeleteFeed(ctx context.Context, id string) error
	UpdateFeed(ctx context.Context, feed *db.Feed) error

	GetFeedItem(ctx context.Context, id string) (*db.Item, error)
	UpdateFeedItem(ctx context.Context, s *db.Item) error

	SearchFeedItems(ctx context.Context, items models.SearchParams) ([]db.Item, int64, error)

	Nuke(ctx context.Context) error
}
