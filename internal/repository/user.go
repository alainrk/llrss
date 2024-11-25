package repository

import (
	"context"
	"llrss/internal/models"
	"llrss/internal/models/db"
)

type UserRepository interface {
	GetUser(ctx context.Context, id uint64) (*db.User, error)
	SaveUser(ctx context.Context, user *models.NewUser) (uint64, error)

	Nuke(ctx context.Context) error
}
