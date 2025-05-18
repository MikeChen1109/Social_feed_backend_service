package main

import (
	"myApp/SocialFeed/controllers"
	"myApp/SocialFeed/initializers"
	"myApp/SocialFeed/routes"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDatabase()
	controllers.InitAuthController()
}

func main() {
	sqlDB, _ := initializers.DB.DB()
	defer sqlDB.Close()
	
	router := gin.Default()

	routes.RegisterFeedRoutes(router)
	routes.RegisterUserRoutes(router)

	router.Run()
}
