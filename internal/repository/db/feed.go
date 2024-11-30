package sqlite

import (
	"context"
	"errors"
	"fmt"
	"llrss/internal/models"
	"llrss/internal/models/db"
	"llrss/internal/repository"
	"llrss/internal/utils"
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

// ForUser is a Scope for a query with user selected.
func ForUser(userId uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userId)
	}
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
	id := utils.URLToID(url)

	res := r.d.First(&feed, "id = ?", id)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &feed, nil
}

func (r *gormFeedRepository) ListFeeds(_ context.Context, userId uint64) ([]db.Feed, error) {
	var feeds []db.Feed
	if userId == 0 {
		res := r.d.Find(&feeds)
		if res.Error != nil {
			return nil, res.Error
		}
		return feeds, nil
	}

	query := r.d.Joins("JOIN user_feeds ON feeds.id = user_feeds.feed_id").
		Where("user_feeds.user_id = ?", userId)

	err := query.Order("pub_date DESC").Find(&feeds).Error
	return feeds, err
}

func (r *gormFeedRepository) SaveFeed(ctx context.Context, feed *db.Feed) (string, error) {
	feed.ID = utils.URLToID(feed.URL)

	// I want to process and save feed items myself after the feed is saved
	items := feed.Items
	feed.Items = nil

	res := r.d.Clauses(clause.OnConflict{DoNothing: true}).Create(feed)
	if res.Error != nil {
		return "", res.Error
	}

	// TODO: This requires a new function (e.g. subscribeFeed that calls this function before if needed)
	// Assign it to a user, if any
	// if userId != 0 {
	// 	res = r.d.Clauses(clause.OnConflict{DoNothing: true}).Create(&db.UserFeed{
	// 		UserID: userId,
	// 		FeedID: feed.ID,
	// 	})
	// 	if res.Error != nil {
	// 		return "", res.Error
	// 	}
	// }

	err := r.SaveFeedItems(ctx, feed.ID, items)
	if err != nil {
		fmt.Printf("failed to save feed items: %v\n", err)
	}
	return feed.ID, nil
}

func (r *gormFeedRepository) SaveFeedItems(_ context.Context, feedID string, items []db.Item) error {
	for _, item := range items {
		item.ID = utils.URLToID(item.Link)
		item.FeedID = feedID
		item.Title = strings.TrimSpace(item.Title)
		item.Description = utils.CleanDescription(item.Description)

		// NOTE: If multiple feeds try to create the same item (URL as ID), we don't add it but neither fail
		res := r.d.Clauses(clause.OnConflict{DoNothing: true}).Create(&item)
		if res.Error != nil {
			fmt.Printf("failed to save item: %v\n", res.Error)
			continue
		}

		// If the item wasn't actually inserted (already existed), skip creating user items
		if res.RowsAffected == 0 {
			continue
		}

		// Get all users subscribed to this feed
		var userFeeds []db.UserFeed
		err := r.d.Where("feed_id = ?", feedID).Find(&userFeeds).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("failed to get feed subscribers: %v\n", err)
			continue
		}

		// Create UserItem entries for each subscriber
		for _, userFeed := range userFeeds {
			userItem := db.UserItem{
				ItemID: item.ID,
				UserID: userFeed.UserID,
				IsRead: false, // Default to unread for new items
			}

			// TODO: This can be done in bulk or anyway improved performance-wise
			// Use DoNothing clause in case the user item already exists
			if err := r.d.Clauses(clause.OnConflict{DoNothing: true}).Create(&userItem).Error; err != nil {
				fmt.Printf("failed to create user item for user %d: %v\n", userFeed.UserID, err)
				continue
			}
		}
	}

	return nil
}

func (r *gormFeedRepository) DeleteFeed(_ context.Context, userId uint64, id string) error {
	if userId == 0 {
		return fmt.Errorf("userId is required")
	}

	return r.d.Transaction(func(tx *gorm.DB) error {
		err := tx.Scopes(ForUser(userId)).Where("feed_id = ?", id).Delete(&db.UserFeed{}).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		err = tx.Scopes(ForUser(userId)).Where("feed_id = ?", id).Delete(&db.UserItem{}).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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

func (r *gormFeedRepository) SearchFeedItems(_ context.Context, userId uint64, params models.SearchParams) ([]db.Item, int64, error) {
	var items []db.Item
	var total int64
	var err error

	if userId == 0 {
		return nil, 0, fmt.Errorf("userId is required")
	}

	// First of all join with user items.
	query := r.d.Model(&db.Item{}).
		Joins("JOIN user_items ON items.id = user_items.item_id").
		Where("user_items.user_id = ?", userId)

	// Apply unread filter - note that is_read now comes from user_items.
	if params.Unread {
		query = query.Where("user_items.is_read = ?", false)
	}

	// Apply text search if query is provided.
	if params.Query != "" {
		searchPattern := "%" + params.Query + "%"
		query = query.Where(
			"items.title LIKE ? OR items.description LIKE ? OR items.author LIKE ? OR items.category LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Apply date range.
	query = query.Where("items.pub_date BETWEEN ? AND ?", params.FromDate, params.ToDate)

	// Count total before applying pagination.
	err = query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting.
	if params.Sort == "asc" {
		query = query.Order("items.pub_date asc")
	} else {
		query = query.Order("items.pub_date desc")
	}

	// Apply pagination.
	query = query.Offset(params.Offset).Limit(params.Limit)

	// Execute the final query.
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
