package repository

import (
	"context"
	"llrss/internal/models"
	"llrss/internal/models/db"
)

type UserRepository interface {
	GetUser(ctx context.Context, id uint64) (*db.User, error)
	SaveUser(ctx context.Context, user *models.NewUser) (uint64, error)
	SaveUserFeed(ctx context.Context, userFeed *models.NewUserFeed) (uint64, error)
	SaveUserFeedItem(ctx context.Context, userItem *models.NewUserItem) (uint64, error)
	UpdateUserFeedItem(ctx context.Context, userItem *models.NewUserItem) error

	Nuke(ctx context.Context) error
}
