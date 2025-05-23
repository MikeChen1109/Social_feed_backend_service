package routes

import (
	"feed-service/controllers"
	"feed-service/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterCommentRoutes(c *gin.Engine, commentsController *controllers.CommentsController) {
	commentGroup := c.Group("/comment")
	{
		commentGroup.POST("/create", middleware.RequireAuth, commentsController.CreateComment)
		commentGroup.GET("/paginated", commentsController.PaginatedComments)
	}
}
