package routes

import (
	"user-service/controllers"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/files"
)

func RegisterUserRoutes(c *gin.Engine, userController *controllers.UsersController) {
	userGroup := c.Group("/user")
	userGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	{
		userGroup.POST("/signup", userController.Signup)
		userGroup.POST("/login", userController.Login)
		userGroup.POST("/logout", userController.Logout)
		userGroup.POST("/refresh", userController.Refresh)
	}
}
