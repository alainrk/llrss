package sqlite

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"llrss/internal/models"
	"llrss/internal/repository"
	"llrss/internal/text"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrFeedNotFound = errors.New("feed not found")

type gormFeedRepository struct {
	db *gorm.DB
}

func NewGormFeedRepository(db *gorm.DB) repository.FeedRepository {
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

func (r *gormFeedRepository) SaveFeed(ctx context.Context, feed *models.Feed) (string, error) {
	if feed.ID == "" {
		feed.ID = uuid.New().String()
	}

	feed.Title = strings.TrimSpace(feed.Title)
	feed.Description = text.CleanDescription(feed.Description)

	items := feed.Items
	feed.Items = nil

	res := r.db.Create(feed)
	if res.Error != nil {
		return "", res.Error
	}

	err := r.SaveFeedItems(ctx, feed.ID, items)
	if err != nil {
		fmt.Printf("failed to save feed items: %v\n", err)
	}
	return feed.ID, nil
}

func (r *gormFeedRepository) SaveFeedItems(_ context.Context, feedID string, items []models.Item) error {
	for _, item := range items {
		item.ID = base64.StdEncoding.EncodeToString([]byte(item.Link))
		item.FeedID = feedID
		item.Title = strings.TrimSpace(item.Title)
		item.Description = text.CleanDescription(item.Description)

		// NOTE: If multiple feeds try to create the same item (URL as ID), we don't add it but neither fail
		res := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&item)
		if res.Error != nil {
			fmt.Printf("failed to save item: %v\n", res.Error)
			continue
		}
	}
	return nil
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

func (r *gormFeedRepository) GetFeedItem(ctx context.Context, id string) (*models.Item, error) {
	i := &models.Item{}
	res := r.db.First(i, id)
	if res.Error != nil {
		fmt.Printf("failed to get feed item: %v\n", res.Error)
		return nil, res.Error
	}
	return i, nil
}

func (r *gormFeedRepository) UpdateFeedItem(_ context.Context, s *models.Item) error {
	return r.db.Save(s).Error
}

func (r *gormFeedRepository) Nuke(_ context.Context) error {
	res := r.db.Unscoped().Where("1 = 1").Delete(&models.Item{})
	if res.Error != nil {
		return res.Error
	}
	res = r.db.Unscoped().Where("1 = 1").Delete(&models.Feed{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
