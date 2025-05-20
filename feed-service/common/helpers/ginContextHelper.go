package helpers

import (
	appErrors "feed-service/common/appErrors"
	"feed-service/models"

	"github.com/gin-gonic/gin"
)

func ParseClaims(c *gin.Context) (*models.Claims, *appErrors.AppError) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, appErrors.ErrUserNotFound
	}

	claimsModel, ok := claims.(models.Claims)
	if !ok {
		return nil, appErrors.ErrUserInvalidType
	}

	return &claimsModel, nil
}
