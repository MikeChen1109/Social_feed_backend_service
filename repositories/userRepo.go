package repositories

import (
	"errors"
	"myApp/SocialFeed/models"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	FindByUsername(username string) (*models.User, error)
	Create(user *models.User) error
}

type UserRepository struct {
	DB *gorm.DB
}

var ErrUserNotFound = errors.New("user not found")

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(user *models.User) error {
	if err := r.DB.Create(user).Error; err != nil {
		return err
	}
	return nil
}
