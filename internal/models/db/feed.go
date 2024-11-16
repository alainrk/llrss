package db

import (
	"time"
)

type Item struct {
	ID          string `gorm:"primaryKey;type:string"`
	Title       string `gorm:"not null"`
	Link        string `gorm:"not null"`
	Description string
	Author      string
	Category    string
	Comments    string
	PubDate     time.Time `gorm:"index;type:datetime"`
	Source      string
	FeedID      string `gorm:"index"`
	IsRead      bool   `gorm:"default:false"`
}

type Feed struct {
	ID          string    `gorm:"primaryKey"`
	URL         string    `gorm:"uniqueIndex;not null"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"type:text"`
	LastFetch   time.Time `gorm:"index;type:datetime"`
	Items       []Item    `gorm:"foreignKey:FeedID"`
}
