package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"llrss/internal/models"
	"llrss/internal/models/db"
	"llrss/internal/models/rss"
	"llrss/internal/repository"
	"llrss/internal/utils"
	"net/http"
	"time"
)

// TODO: Implement this through a column in Feed table based on provider TTL requested.
const (
	MinRefreshRateMinutes = 0
)

type FeedService interface {
	FetchFeed(ctx context.Context, url string) (*db.Feed, error)

	RefreshFeeds(ctx context.Context) error

	GetFeed(ctx context.Context, id string) (*db.Feed, error)
	GetFeedByURL(ctx context.Context, url string) (*db.Feed, error)
	ListFeeds(ctx context.Context, userId uint64) ([]db.Feed, error)
	SearchFeedItems(ctx context.Context, userId uint64, items models.SearchParams) ([]db.Item, int64, error)

	AddFeed(ctx context.Context, userId uint64, url string) (string, error)
	UpdateFeed(ctx context.Context, feed *db.Feed) error
	MarkFeedItemRead(ctx context.Context, userId uint64, feedItemID string, read bool) error

	DeleteFeed(ctx context.Context, userId uint64, id string) error
	Nuke(ctx context.Context) error
}

type feedService struct {
	feedRepo repository.FeedRepository
	userRepo repository.UserRepository
	client   *http.Client
}

func NewFeedService(repo repository.FeedRepository) FeedService {
	return &feedService{
		feedRepo: repo,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *feedService) FetchFeed(ctx context.Context, url string) (*db.Feed, error) {
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

	var r rss.RSS
	if err := xml.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("parse RSS: %w", err)
	}

	fmt.Println(r)

	var items []db.Item
	for _, item := range r.Channel.Items {
		d, err := utils.ParseRSSDate(item.PubDate)
		if err != nil {
			fmt.Printf("parse date: %v", err)
			continue
		}

		items = append(items, db.Item{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Author:      item.Author,
			Category:    item.Category,
			Comments:    item.Comments,
			Source:      item.Source,
			PubDate:     d,
		})
	}

	// TODO: This should be an RSS.Feed, not db.Feed to keep things well separated, shouldn't be a problem
	feed := &db.Feed{
		URL:         url,
		Title:       r.Channel.Title,
		Description: r.Channel.Description,
		LastFetch:   time.Now(),
		Items:       items,
	}

	return feed, nil
}

func (s *feedService) GetFeed(ctx context.Context, id string) (*db.Feed, error) {
	return s.feedRepo.GetFeed(ctx, id)
}

func (s *feedService) GetFeedByURL(ctx context.Context, url string) (*db.Feed, error) {
	return s.feedRepo.GetFeedByURL(ctx, url)
}

func (s *feedService) ListFeeds(ctx context.Context, userId uint64) ([]db.Feed, error) {
	return s.feedRepo.ListFeeds(ctx, userId)
}

// AddFeed adds a new feed by url and returns its ID.
func (s *feedService) AddFeed(ctx context.Context, userId uint64, url string) (string, error) {
	f, _ := s.feedRepo.GetFeedByURL(ctx, url)
	if f != nil {
		return f.ID, nil
	}

	// TODO: If already existing and last fetched recently we're good
	feed, err := s.FetchFeed(ctx, url)
	if err != nil {
		return "", fmt.Errorf("fetch feed: %w", err)
	}

	id, err := s.feedRepo.SaveFeed(ctx, feed)
	if err != nil {
		return "", fmt.Errorf("save feed: %w", err)
	}

	return id, nil
}

func (s *feedService) DeleteFeed(ctx context.Context, userId uint64, id string) error {
	return s.feedRepo.DeleteFeed(ctx, userId, id)
}

func (s *feedService) UpdateFeed(ctx context.Context, feed *db.Feed) error {
	return s.feedRepo.UpdateFeed(ctx, feed)
}

func (s *feedService) MarkFeedItemRead(ctx context.Context, userId uint64, feedItemID string, read bool) error {
	return s.userRepo.UpdateUserFeedItem(ctx, &models.NewUserItem{
		ItemID: feedItemID,
		UserID: userId,
		IsRead: read,
	})
}

func (s *feedService) SearchFeedItems(ctx context.Context, userId uint64, params models.SearchParams) ([]db.Item, int64, error) {
	return s.feedRepo.SearchFeedItems(ctx, userId, params)
}

// TODO: When refreshing feed I should also update the metadata of the Feed itself, in case they're changed (title, description, etc.)
func (s *feedService) RefreshFeeds(ctx context.Context) error {
	feeds, err := s.feedRepo.ListFeeds(ctx, 0)
	if err != nil {
		return err
	}

	for _, f := range feeds {
		if (f.LastFetch.Add(MinRefreshRateMinutes * time.Minute)).After(time.Now()) {
			// TODO: These must be debug logs
			fmt.Printf("feed %s is not due for refresh\n", f.URL)
			continue
		}

		feed, err := s.FetchFeed(ctx, f.URL)
		if err != nil {
			e := fmt.Errorf("error on fetching feed %s: %w", f.URL, err)
			fmt.Printf("%v\n", e)
			continue
		}

		f.LastFetch = time.Now()
		f.Title = feed.Title
		f.Description = feed.Description
		// Don't update items directly
		f.Items = nil

		err = s.feedRepo.UpdateFeed(ctx, &f)
		if err != nil {
			e := fmt.Errorf("error on updating last_fetch feed %s: %w", f.URL, err)
			fmt.Printf("%v\n", e)
			continue
		}

		err = s.feedRepo.SaveFeedItems(ctx, f.ID, feed.Items)
		if err != nil {
			e := fmt.Errorf("error on saving feed items for feed %s: %w", f.URL, err)
			fmt.Printf("%v\n", e)
			continue
		}
	}

	return nil
}

func (s *feedService) Nuke(ctx context.Context) error {
	return s.feedRepo.Nuke(ctx)
}
