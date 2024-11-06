package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"llrss/internal/models"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

type FeedRepository interface {
	GetFeed(ctx context.Context, id string) (*models.Feed, error)
	ListFeeds(ctx context.Context) ([]models.Feed, error)
	SaveFeed(ctx context.Context, feed *models.Feed) error
	DeleteFeed(ctx context.Context, id string) error
	UpdateFeed(ctx context.Context, feed *models.Feed) error
}

type jsonFileFeedRepository struct {
	mu       sync.RWMutex
	filePath string
	dir      string
}

type fileData struct {
	Feeds []models.Feed `json:"feeds"`
}

func NewJSONFileFeedRepository(filePath string) (FeedRepository, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	repo := &jsonFileFeedRepository{
		filePath: filePath,
		dir:      dir,
	}

	// Create file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		initialData := fileData{Feeds: []models.Feed{}}
		if err := repo.writeFile(initialData); err != nil {
			return nil, fmt.Errorf("failed to initialize repository: %w", err)
		}
	}

	return repo, nil
}

func (r *jsonFileFeedRepository) readFile() (fileData, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return fileData{}, fmt.Errorf("failed to read file: %w", err)
	}

	var fd fileData
	if err := json.Unmarshal(data, &fd); err != nil {
		return fileData{}, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return fd, nil
}

func (r *jsonFileFeedRepository) writeFile(fd fileData) error {
	// Create a temporary file in the same directory
	tmpFile, err := os.CreateTemp(r.dir, "feeds-*.json.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure temporary file is cleaned up
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	// Write data to temporary file
	data, err := json.MarshalIndent(fd, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write data to temporary file: %w", err)
	}

	// Ensure all data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomically replace the old file with the new one
	if err := os.Rename(tmpPath, r.filePath); err != nil {
		return fmt.Errorf("failed to replace file: %w", err)
	}

	return nil
}

func (r *jsonFileFeedRepository) GetFeed(_ context.Context, id string) (*models.Feed, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fd, err := r.readFile()
	if err != nil {
		return nil, err
	}

	for _, feed := range fd.Feeds {
		if feed.ID == id {
			feedCopy := feed // Create a copy
			return &feedCopy, nil
		}
	}

	return nil, fmt.Errorf("feed not found: %s", id)
}

func (r *jsonFileFeedRepository) ListFeeds(_ context.Context) ([]models.Feed, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fd, err := r.readFile()
	if err != nil {
		return nil, err
	}

	feeds := make([]models.Feed, len(fd.Feeds))
	copy(feeds, fd.Feeds)

	return feeds, nil
}

func (r *jsonFileFeedRepository) SaveFeed(_ context.Context, feed *models.Feed) error {
	if feed == nil {
		return fmt.Errorf("feed cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	fd, err := r.readFile()
	if err != nil {
		return err
	}

	if feed.ID == "" {
		feed.ID = uuid.New().String()
	}

	// Check if feed with this ID already exists
	for i, existingFeed := range fd.Feeds {
		if existingFeed.ID == feed.ID {
			fd.Feeds[i] = *feed // Update existing feed
			return r.writeFile(fd)
		}
	}

	// Add new feed
	feedCopy := *feed
	fd.Feeds = append(fd.Feeds, feedCopy)

	return r.writeFile(fd)
}

func (r *jsonFileFeedRepository) DeleteFeed(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fd, err := r.readFile()
	if err != nil {
		return err
	}

	found := false
	feeds := make([]models.Feed, 0, len(fd.Feeds))
	for _, feed := range fd.Feeds {
		if feed.ID != id {
			feeds = append(feeds, feed)
		} else {
			found = true
		}
	}

	if !found {
		return nil
	}

	fd.Feeds = feeds
	return r.writeFile(fd)
}

func (r *jsonFileFeedRepository) UpdateFeed(_ context.Context, feed *models.Feed) error {
	if feed == nil {
		return fmt.Errorf("feed cannot be nil")
	}
	if feed.ID == "" {
		return fmt.Errorf("feed ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	fd, err := r.readFile()
	if err != nil {
		return err
	}

	found := false
	for i, existingFeed := range fd.Feeds {
		if existingFeed.ID == feed.ID {
			feedCopy := *feed
			fd.Feeds[i] = feedCopy
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("feed not found: %s", feed.ID)
	}

	return r.writeFile(fd)
}
