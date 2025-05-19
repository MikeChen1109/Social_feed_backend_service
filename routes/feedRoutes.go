package routes

import (
	"myApp/SocialFeed/controllers"
	"myApp/SocialFeed/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterFeedRoutes(c *gin.Engine, authMiddleware *middleware.AuthMiddleware, feedsController *controllers.FeedsController) {
	feedGroup := c.Group("/feed")
	{
		feedGroup.POST("/create", authMiddleware.RequireAuth, feedsController.CreateFeed)
		feedGroup.GET("/", feedsController.GetFeeds)
		feedGroup.GET("/paginated", feedsController.PaginatedFeeds)
		feedGroup.GET("/:id", feedsController.GetFeedByID)
		feedGroup.PUT("/:id", authMiddleware.RequireAuth, feedsController.UpdateFeed)
		feedGroup.DELETE("/:id", authMiddleware.RequireAuth, feedsController.DeleteFeed)
	}
}
