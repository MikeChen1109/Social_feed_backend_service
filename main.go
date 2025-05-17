package main

import (
	"myApp/SocialFeed/controllers"
	"myApp/SocialFeed/initializers"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDatabase()
}

func main() {
	router := gin.Default()

	router.POST("/createFeed", controllers.CreateFeed)
	router.GET("/feeds", controllers.GetFeeds)
	router.GET("/feeds/:id", controllers.GetFeedByID)
	router.PUT("/feeds/:id", controllers.UpdateFeed)
	router.DELETE("/feeds/:id", controllers.DeleteFeed)

	router.Run()
}
