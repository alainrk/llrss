package json

import (
	"context"
	"llrss/internal/models"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/exp/rand"
)

func TestJSONFileFeedRepository(t *testing.T) {
	// Create temporary directory for test file
	tmpDir, err := os.MkdirTemp("", "feed-repo-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "feeds.json")
	repo, err := NewJSONFileFeedRepository(filePath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Test empty repository
	t.Run("empty repository", func(t *testing.T) {
		feeds, err := repo.ListFeeds(ctx)
		if err != nil {
			t.Errorf("ListFeeds failed: %v", err)
		}
		if len(feeds) != 0 {
			t.Errorf("Expected empty feeds, got %d feeds", len(feeds))
		}
	})

	// Test saving new feed
	t.Run("save feed", func(t *testing.T) {
		feed1 := &models.Feed{
			URL:         "http://example.com",
			Title:       "Test Feed",
			Description: "Test Description",
			LastFetch:   time.Now().UTC(),
		}
		feed2 := &models.Feed{
			URL:         "http://another.example.com",
			Title:       "Test Another Feed",
			Description: "Test Another Description",
			LastFetch:   time.Now().UTC(),
		}

		err := repo.SaveFeed(ctx, feed1)
		if err != nil {
			t.Errorf("SaveFeed failed: %v", err)
		}
		if feed1.ID == "" {
			t.Error("Expected ID to be generated")
		}

		err = repo.SaveFeed(ctx, feed2)
		if err != nil {
			t.Errorf("SaveFeed2 failed: %v", err)
		}

		savedFeed, err := repo.GetFeed(ctx, feed1.ID)
		if err != nil {
			t.Errorf("GetFeed failed: %v", err)
		}
		if savedFeed.Title != feed1.Title {
			t.Errorf("Expected title %q, got %q", feed1.Title, savedFeed.Title)
		}
	})

	// Test updating feed
	t.Run("update feed", func(t *testing.T) {
		feeds, err := repo.ListFeeds(ctx)
		if err != nil {
			t.Fatalf("ListFeeds failed: %v", err)
		}
		if len(feeds) == 0 {
			t.Fatal("Expected at least one feed")
		}

		feed := feeds[0]
		feed.Title = "Updated Title"

		err = repo.UpdateFeed(ctx, &feed)
		if err != nil {
			t.Errorf("UpdateFeed failed: %v", err)
		}

		updatedFeed, err := repo.GetFeed(ctx, feed.ID)
		if err != nil {
			t.Errorf("GetFeed failed: %v", err)
		}
		if updatedFeed.Title != "Updated Title" {
			t.Errorf("Expected title %q, got %q", "Updated Title", updatedFeed.Title)
		}
	})

	// Test deleting feed
	t.Run("delete feed", func(t *testing.T) {
		feeds, err := repo.ListFeeds(ctx)
		if err != nil {
			t.Fatalf("ListFeeds failed: %v", err)
		}
		if len(feeds) == 0 {
			t.Fatal("Expected at least one feed")
		}

		id := feeds[0].ID

		err = repo.DeleteFeed(ctx, id)
		if err != nil {
			t.Errorf("DeleteFeed failed: %v", err)
		}

		err = repo.DeleteFeed(ctx, id)
		if err != nil {
			t.Error("Not expected error deleting not existing feed, should fail silently")
		}

		f, err := repo.GetFeed(ctx, id)
		if err != nil {
			t.Error("Not expected error deleting not existing feed, should fail silently")
		}
		if f != nil {
			t.Error("Expected empty feed on not existing GetFeed, got:", f)
		}
	})

	// Test concurrent operations
	t.Run("concurrent operations", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				feed := &models.Feed{
					URL:         "http://example.com",
					Title:       "Test Feed",
					Description: "Test Description",
					LastFetch:   time.Now().UTC(),
				}

				err := repo.SaveFeed(ctx, feed)
				if err != nil {
					t.Errorf("Concurrent SaveFeed failed: %v", err)
				}

				feeds, err := repo.ListFeeds(ctx)
				if err != nil {
					t.Errorf("Concurrent ListFeeds failed: %v", err)
				}

				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				n := rand.Intn(len(feeds))

				if len(feeds) > 0 {
					err = repo.DeleteFeed(ctx, feeds[n].ID)
					if err != nil {
						t.Errorf("Concurrent DeleteFeed failed: %v", err)
					}
				}

				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
