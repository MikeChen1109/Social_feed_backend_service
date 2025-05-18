package routes

import (
	"myApp/SocialFeed/controllers"
	"myApp/SocialFeed/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterFeedRoutes(c *gin.Engine) {
	feedGroup := c.Group("/feed")
	{
		feedGroup.POST("/create", middleware.RequireAuth, controllers.CreateFeed)
		feedGroup.GET("/", controllers.GetFeeds)
		feedGroup.GET("/:id", controllers.GetFeedByID)
		feedGroup.PUT("/:id", middleware.RequireAuth, controllers.UpdateFeed)
		feedGroup.DELETE("/:id", middleware.RequireAuth, controllers.DeleteFeed)
	}
}
