package routes

import (
	"myApp/SocialFeed/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(c *gin.Engine, userController *controllers.UsersController) {
	userGroup := c.Group("/user")
	{
		userGroup.POST("/signup", userController.Signup)
		userGroup.POST("/login", userController.Login)
	}
}
