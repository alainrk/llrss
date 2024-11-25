package sqlite

import (
	"context"
	"errors"
	"fmt"
	"llrss/internal/models"
	"llrss/internal/models/db"
	"llrss/internal/repository"
	"llrss/internal/text"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrFeedNotFound = errors.New("feed not found")

type gormFeedRepository struct {
	d *gorm.DB
}

func NewGormFeedRepository(d *gorm.DB) repository.FeedRepository {
	return &gormFeedRepository{d: d}
}

func (r *gormFeedRepository) GetFeed(_ context.Context, id string) (*db.Feed, error) {
	var feed db.Feed
	res := r.d.First(&feed, "id = ?", id)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, ErrFeedNotFound
		}
		return nil, res.Error
	}
	return &feed, nil
}

func (r *gormFeedRepository) GetFeedByURL(_ context.Context, url string) (*db.Feed, error) {
	var feed db.Feed
	id := text.URLToID(url)

	res := r.d.First(&feed, "id = ?", id)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &feed, nil
}

func (r *gormFeedRepository) ListFeeds(_ context.Context) ([]db.Feed, error) {
	var feeds []db.Feed
	res := r.d.Find(&feeds)
	if res.Error != nil {
		return nil, res.Error
	}
	return feeds, nil
}

func (r *gormFeedRepository) SaveFeed(ctx context.Context, feed *db.Feed) (string, error) {
	feed.ID = text.URLToID(feed.URL)

	// I want to process and save feed items myself after the feed is saved
	items := feed.Items
	feed.Items = nil

	res := r.d.Create(feed)
	if res.Error != nil {
		return "", res.Error
	}

	err := r.SaveFeedItems(ctx, feed.ID, items)
	if err != nil {
		fmt.Printf("failed to save feed items: %v\n", err)
	}
	return feed.ID, nil
}

func (r *gormFeedRepository) SaveFeedItems(_ context.Context, feedID string, items []db.Item) error {
	for _, item := range items {
		item.ID = text.URLToID(item.Link)
		item.FeedID = feedID
		item.Title = strings.TrimSpace(item.Title)
		item.Description = text.CleanDescription(item.Description)

		// NOTE: If multiple feeds try to create the same item (URL as ID), we don't add it but neither fail
		res := r.d.Clauses(clause.OnConflict{DoNothing: true}).Create(&item)
		if res.Error != nil {
			fmt.Printf("failed to save item: %v\n", res.Error)
			continue
		}
	}
	return nil
}

func (r *gormFeedRepository) DeleteFeed(_ context.Context, id string) error {
	return r.d.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("feed_id = ?", id).Delete(&db.Item{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&db.Feed{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *gormFeedRepository) UpdateFeed(_ context.Context, feed *db.Feed) error {
	res := r.d.Save(feed)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *gormFeedRepository) GetFeedItem(ctx context.Context, id string) (*db.Item, error) {
	i := &db.Item{}
	res := r.d.First(i, "id = ?", id)
	fmt.Println(id, i)
	if res.Error != nil {
		fmt.Printf("failed to get feed item: %v\n", res.Error)
		return nil, res.Error
	}
	return i, nil
}

func (r *gormFeedRepository) UpdateFeedItem(_ context.Context, s *db.Item) error {
	return r.d.Save(s).Error
}

func (r *gormFeedRepository) SearchFeedItems(_ context.Context, params models.SearchParams) ([]db.Item, int64, error) {
	var items []db.Item
	var total int64
	var err error

	// Start building the query
	query := r.d.Model(&db.Item{})

	// Apply text search if query is provided
	if params.Query != "" {
		searchPattern := "%" + params.Query + "%"
		query = query.Where(
			"title LIKE ? OR description LIKE ? OR author LIKE ? OR category LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Apply unread filter
	if params.Unread {
		query = query.Where("is_read = ?", false)
	}

	// Apply date range
	query = query.Where("pub_date BETWEEN ? AND ?", params.FromDate, params.ToDate)

	// Count total before applying pagination
	err = query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if params.Sort == "asc" {
		query = query.Order("pub_date asc")
	} else {
		query = query.Order("pub_date desc")
	}

	// Apply pagination
	query = query.Offset(params.Offset).Limit(params.Limit)

	// Execute the final query
	err = query.Find(&items).Debug().Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *gormFeedRepository) Nuke(_ context.Context) error {
	res := r.d.Unscoped().Where("1 = 1").Delete(&db.Item{})
	if res.Error != nil {
		return res.Error
	}
	res = r.d.Unscoped().Where("1 = 1").Delete(&db.Feed{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
