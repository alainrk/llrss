package service

import (
	"context"
	"llrss/internal/models"
	"llrss/internal/models/db"
	"llrss/internal/repository"
)

type UserService interface {
	GetUser(ctx context.Context, id uint64) (*db.User, error)
	SaveUser(ctx context.Context, user *models.NewUser) (uint64, error)

	Nuke(ctx context.Context) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetUser(ctx context.Context, id uint64) (*db.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *userService) SaveUser(ctx context.Context, user *models.NewUser) (uint64, error) {
	return s.repo.SaveUser(ctx, user)
}

func (s *userService) Nuke(ctx context.Context) error {
	return s.repo.Nuke(ctx)
}
