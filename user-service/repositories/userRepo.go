package repositories

import (
	"context"
	"errors"
	"time"
	apperrors "user-service/common/appErrors"
	"user-service/models"

	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint) (*models.User, error)
}

type UserRepository struct {
	DB *gorm.DB
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	if err := db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
