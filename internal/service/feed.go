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
	"llrss/internal/text"
	"net/http"
	"time"
)

const (
	// TODO: Implement this through a column in Feed table based on provider TTL requested
	MinRefreshRateMinutes = 0
)

type FeedService interface {
	FetchFeed(ctx context.Context, url string) (*db.Feed, error)
	GetFeed(ctx context.Context, id string) (*db.Feed, error)
	GetFeedByURL(ctx context.Context, url string) (*db.Feed, error)
	ListFeeds(ctx context.Context) ([]db.Feed, error)
	AddFeed(ctx context.Context, url string) (string, error)
	DeleteFeed(ctx context.Context, id string) error
	UpdateFeed(ctx context.Context, feed *db.Feed) error
	MarkFeedItemRead(ctx context.Context, feedItemID string, read bool) error
	SearchFeedItems(ctx context.Context, items models.SearchParams) ([]db.Item, int64, error)
	RefreshFeeds(ctx context.Context) error
	Nuke(ctx context.Context) error
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

	var items []db.Item
	for _, item := range r.Channel.Items {
		d, err := text.ParseRSSDate(item.PubDate)
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
	return s.repo.GetFeed(ctx, id)
}

func (s *feedService) GetFeedByURL(ctx context.Context, url string) (*db.Feed, error) {
	return s.repo.GetFeedByURL(ctx, url)
}

func (s *feedService) ListFeeds(ctx context.Context) ([]db.Feed, error) {
	return s.repo.ListFeeds(ctx)
}

// AddFeed adds a new feed by url and returns its ID.
func (s *feedService) AddFeed(ctx context.Context, url string) (string, error) {
	f, _ := s.repo.GetFeedByURL(ctx, url)
	if f != nil {
		return f.ID, nil
	}

	feed, err := s.FetchFeed(ctx, url)
	if err != nil {
		return "", fmt.Errorf("fetch feed: %w", err)
	}

	id, err := s.repo.SaveFeed(ctx, feed)
	if err != nil {
		return "", fmt.Errorf("save feed: %w", err)
	}

	return id, nil
}

func (s *feedService) DeleteFeed(ctx context.Context, id string) error {
	return s.repo.DeleteFeed(ctx, id)
}

func (s *feedService) UpdateFeed(ctx context.Context, feed *db.Feed) error {
	return s.repo.UpdateFeed(ctx, feed)
}

func (s *feedService) MarkFeedItemRead(ctx context.Context, feedItemID string, read bool) error {
	i, err := s.repo.GetFeedItem(ctx, feedItemID)
	if err != nil {
		return err
	}
	i.IsRead = read
	return s.repo.UpdateFeedItem(ctx, i)
}

func (s *feedService) SearchFeedItems(ctx context.Context, items models.SearchParams) ([]db.Item, int64, error) {
	return s.repo.SearchFeedItems(ctx, items)
}

func (s *feedService) RefreshFeeds(ctx context.Context) error {
	feeds, err := s.repo.ListFeeds(ctx)
	if err != nil {
		return err
	}

	for _, f := range feeds {
		if (f.LastFetch.Add(MinRefreshRateMinutes * time.Minute)).After(time.Now()) {
			// TODO: These must be debug logs
			fmt.Printf("feed %s is not due for refresh\n", f.URL)
			continue
		}

		f.LastFetch = time.Now()
		err := s.repo.UpdateFeed(ctx, &f)
		if err != nil {
			e := fmt.Errorf("error on updating last_fetch feed %s: %w", f.URL, err)
			fmt.Printf("%v\n", e)
			continue
		}

		feed, err := s.FetchFeed(ctx, f.URL)
		if err != nil {
			e := fmt.Errorf("error on fetching feed %s: %w", f.URL, err)
			fmt.Printf("%v\n", e)
			continue
		}

		err = s.repo.SaveFeedItems(ctx, f.ID, feed.Items)
		if err != nil {
			e := fmt.Errorf("error on saving feed items for feed %s: %w", f.URL, err)
			fmt.Printf("%v\n", e)
			continue
		}
	}

	return nil
}

func (s *feedService) Nuke(ctx context.Context) error {
	return s.repo.Nuke(ctx)
}
