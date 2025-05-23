package routes

import (
	"feed-service/controllers"
	"feed-service/middleware"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/files"
)

func RegisterFeedRoutes(c *gin.Engine, feedsController *controllers.FeedsController) {
	feedGroup := c.Group("/feed")
	feedGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	{
		feedGroup.POST("/create", middleware.RequireAuth, feedsController.CreateFeed)
		feedGroup.GET("/", feedsController.GetFeeds)
		feedGroup.GET("/paginated", feedsController.PaginatedFeeds)
		feedGroup.GET("/:id", feedsController.GetFeedByID)
		feedGroup.PUT("/:id", middleware.RequireAuth, feedsController.UpdateFeed)
		feedGroup.DELETE("/:id", middleware.RequireAuth, feedsController.DeleteFeed)
	}
}
