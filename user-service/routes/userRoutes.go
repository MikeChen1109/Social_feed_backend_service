package routes

import (
	"user-service/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(c *gin.Engine, userController *controllers.UsersController) {
	userGroup := c.Group("/user")
	{
		userGroup.POST("/signup", userController.Signup)
		userGroup.POST("/login", userController.Login)
		userGroup.POST("/logout", userController.Logout)
		userGroup.POST("/refresh", userController.Refresh)
	}
}
