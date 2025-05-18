package routes

import (
	"myApp/SocialFeed/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(c *gin.Engine) {
	userGroup := c.Group("/user")
	{
		userGroup.POST("/signup", controllers.Signup)
		userGroup.POST("/login", controllers.Login)
	}
}
