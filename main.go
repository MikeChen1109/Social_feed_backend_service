package main

import (
	"myApp/SocialFeed/controllers"
	"myApp/SocialFeed/initializers"
	"myApp/SocialFeed/middleware"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDatabase()
}

func main() {
	router := gin.Default()

	// Feed routes
	router.POST("/createFeed", middleware.RequireAuth, controllers.CreateFeed)
	router.GET("/feeds", controllers.GetFeeds)
	router.GET("/feeds/:id", controllers.GetFeedByID)
	router.PUT("/feeds/:id", controllers.UpdateFeed)
	router.DELETE("/feeds/:id", controllers.DeleteFeed)

	// User routes
	router.POST("/signup", controllers.Signup)
	router.POST("/login", controllers.Login)

	router.Run()
}
