package models

import (
	"llrss/internal/models/db"
	"time"
)

type SearchParams struct {
	FromDate time.Time
	ToDate   time.Time
	Query    string
	Sort     string
	Limit    int
	Offset   int
	Unread   bool
}

type SearchResult struct {
	Items []db.Item
	Len   int64
	Total int64
}

type NewUser struct {
	Name string
	ID   uint64
}

type NewUserItem struct {
	ItemID string
	UserID uint64
	IsRead bool
}

type NewUserFeed struct {
	FeedID string
	UserID uint64
}
