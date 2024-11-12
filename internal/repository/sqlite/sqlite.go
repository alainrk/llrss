package sqlite

import (
	"context"
	"errors"
	"llrss/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrFeedNotFound = errors.New("feed not found")

type FeedRepository interface {
	GetFeed(ctx context.Context, id string) (*models.Feed, error)
	GetFeedByURL(ctx context.Context, url string) (*models.Feed, error)
	ListFeeds(ctx context.Context) ([]models.Feed, error)
	SaveFeed(ctx context.Context, feed *models.Feed) (string, error)
	DeleteFeed(ctx context.Context, id string) error
	UpdateFeed(ctx context.Context, feed *models.Feed) error
}

type gormFeedRepository struct {
	db *gorm.DB
}

func NewGormFeedRepository(db *gorm.DB) FeedRepository {
	return &gormFeedRepository{db: db}
}

func (r *gormFeedRepository) GetFeed(_ context.Context, id string) (*models.Feed, error) {
	var feed models.Feed
	if err := r.db.Preload("Items").First(&feed, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeedNotFound
		}
		return nil, err
	}
	return &feed, nil
}

func (r *gormFeedRepository) GetFeedByURL(_ context.Context, url string) (*models.Feed, error) {
	var feed models.Feed
	res := r.db.First(&feed, "url = ?", url)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &feed, nil
}

func (r *gormFeedRepository) ListFeeds(_ context.Context) ([]models.Feed, error) {
	var feeds []models.Feed
	if err := r.db.Preload("Items").Find(&feeds).Error; err != nil {
		return nil, err
	}
	return feeds, nil
}

func (r *gormFeedRepository) SaveFeed(_ context.Context, feed *models.Feed) (string, error) {
	if feed.ID == "" {
		feed.ID = uuid.New().String()
	}
	res := r.db.Create(feed)
	if res.Error != nil {
		return "", res.Error
	}
	return feed.ID, nil
}

func (r *gormFeedRepository) DeleteFeed(_ context.Context, id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Feed{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *gormFeedRepository) UpdateFeed(_ context.Context, feed *models.Feed) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(feed).Error; err != nil {
			return err
		}
		return nil
	})
}
