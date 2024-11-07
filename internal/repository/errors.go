package repository

import "errors"

var (
	// ErrFeedNotFound is returned when a feed is not found in the repository.
	ErrFeedNotFound = errors.New("feed not found")

	// ErrEmptyID is returned when an empty ID is provided for operations that require an ID.
	ErrEmptyID = errors.New("empty feed ID")

	// ErrInvalidFeed is returned when trying to save or update an invalid feed.
	ErrInvalidFeed = errors.New("invalid feed")

	// ErrDuplicateFeed is returned when trying to save a feed with a duplicate URL.
	ErrDuplicateFeed = errors.New("duplicate feed URL")
)

// IsNotFound returns true if the error is an ErrFeedNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrFeedNotFound)
}

// IsEmptyID returns true if the error is an ErrEmptyID.
func IsEmptyID(err error) bool {
	return errors.Is(err, ErrEmptyID)
}

// IsInvalidFeed returns true if the error is an ErrInvalidFeed.
func IsInvalidFeed(err error) bool {
	return errors.Is(err, ErrInvalidFeed)
}

// IsDuplicateFeed returns true if the error is an ErrDuplicateFeed.
func IsDuplicateFeed(err error) bool {
	return errors.Is(err, ErrDuplicateFeed)
}
