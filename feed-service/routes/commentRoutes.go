package routes

import (
	"feed-service/controllers"
	"feed-service/middleware"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/files"
)

func RegisterCommentRoutes(c *gin.Engine, commentsController *controllers.CommentsController) {
	commentGroup := c.Group("/comment")
	commentGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	{
		commentGroup.POST("/create", middleware.RequireAuth, commentsController.CreateComment)
		commentGroup.GET("/paginated", commentsController.PaginatedComments)
	}
}
