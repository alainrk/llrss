package db

type User struct {
	Name  string `gorm:"not null"`
	Feeds []Feed `gorm:"many2many:user_feeds;"`
	Items []Item `gorm:"many2many:user_items;"`
	ID    uint64 `gorm:"primary_key"`
}

type UserFeed struct {
	FeedID string `gorm:"primaryKey"`
	UserID uint64 `gorm:"primaryKey"`
}

type UserItem struct {
	ItemID string `gorm:"primaryKey"`
	UserID uint64 `gorm:"primaryKey"`
	IsRead bool   `gorm:"default:false"`
}
