package routes

import (
	"feed-service/controllers"
	"feed-service/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterCommentRoutes(c *gin.Engine, commentsController *controllers.CommentsController) {
	feedGroup := c.Group("/comment")
	{
		feedGroup.POST("/create", middleware.RequireAuth, commentsController.CreateComment)
		feedGroup.GET("/paginated", commentsController.PaginatedComments)
	}
}
