package helpers

import (
	appErrors "myApp/SocialFeed/common/appErrors"
	"myApp/SocialFeed/models"

	"github.com/gin-gonic/gin"
)

func GetCurrentUser(c *gin.Context) (models.User, *appErrors.AppError) {
	user, exists := c.Get("user")
	if !exists {
		return models.User{}, appErrors.ErrUserNotFound
	}

	userModel, ok := user.(models.User)
	if !ok {
		return models.User{}, appErrors.ErrUserInvalidType
	}

	return userModel, nil
}
