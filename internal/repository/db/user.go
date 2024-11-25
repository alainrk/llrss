package sqlite

import (
	"context"
	"fmt"
	"llrss/internal/models"
	"llrss/internal/models/db"
	"llrss/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormUserRepository struct {
	d *gorm.DB
}

func NewGormUserRepository(d *gorm.DB) repository.UserRepository {
	return &gormUserRepository{d: d}
}

func (r *gormUserRepository) GetUser(_ context.Context, id uint64) (*db.User, error) {
	var user db.User
	res := r.d.First(&user, "id = ?", id)
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

func (r *gormUserRepository) SaveUser(_ context.Context, user *models.NewUser) (uint64, error) {
	res := r.d.Clauses(clause.OnConflict{DoNothing: true}).Create(&db.User{
		Name: user.Name,
		ID:   user.ID,
	})
	if res.Error != nil {
		fmt.Printf("failed to save item: %v\n", res.Error)
		return 0, res.Error
	}
	return user.ID, nil
}

func (r *gormUserRepository) Nuke(_ context.Context) error {
	res := r.d.Unscoped().Where("1 = 1").Delete(&db.User{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
